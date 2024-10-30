package contexts

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	"github.com/multiversx/mx-chain-vm-go/executor"
	"github.com/multiversx/mx-chain-vm-go/math"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
)

var _ vmhost.OutputContext = (*outputContext)(nil)

var logOutput = logger.GetOrCreate("vm/output")

type outputContext struct {
	host             vmhost.VMHost
	outputState      *vmcommon.VMOutput
	stateStack       []*vmcommon.VMOutput
	codeUpdates      map[string]struct{}
	crtTransferIndex uint32
	callArgsParser   vmcommon.CallArgsParser
}

// NewOutputContext creates a new outputContext
func NewOutputContext(host vmhost.VMHost) (*outputContext, error) {
	if check.IfNil(host) {
		return nil, vmhost.ErrNilVMHost
	}

	context := &outputContext{
		host:             host,
		stateStack:       make([]*vmcommon.VMOutput, 0),
		crtTransferIndex: 1,
		callArgsParser:   parsers.NewCallArgsParser(),
	}

	context.InitState()

	return context, nil
}

// InitState initializes the output state and the code updates.
func (context *outputContext) InitState() {
	context.outputState = newVMOutput()
	context.codeUpdates = make(map[string]struct{})
	context.crtTransferIndex = 1
}

func newVMOutput() *vmcommon.VMOutput {
	return &vmcommon.VMOutput{
		ReturnData:      make([][]byte, 0),
		ReturnCode:      vmcommon.Ok,
		ReturnMessage:   "",
		GasRemaining:    0,
		GasRefund:       big.NewInt(0),
		OutputAccounts:  make(map[string]*vmcommon.OutputAccount),
		DeletedAccounts: make([][]byte, 0),
		TouchedAccounts: make([][]byte, 0),
		Logs:            make([]*vmcommon.LogEntry, 0),
	}
}

// NewVMOutputAccount creates a new output account and sets the given address
func NewVMOutputAccount(address []byte) *vmcommon.OutputAccount {
	return &vmcommon.OutputAccount{
		Address:                 address,
		Nonce:                   0,
		BalanceDelta:            big.NewInt(0),
		Balance:                 nil,
		StorageUpdates:          make(map[string]*vmcommon.StorageUpdate),
		BytesAddedToStorage:     0,
		BytesDeletedFromStorage: 0,
	}
}

// PushState appends the current vmOutput to the state stack
func (context *outputContext) PushState() {
	newState := newVMOutput()
	mergeVMOutputs(newState, context.outputState)
	context.stateStack = append(context.stateStack, newState)
}

// PopSetActiveState removes the latest entry from the state stack and sets it as the current vm output
func (context *outputContext) PopSetActiveState() {
	stateStackLen := len(context.stateStack)
	if stateStackLen == 0 {
		return
	}

	prevState := context.stateStack[stateStackLen-1]
	context.stateStack = context.stateStack[:stateStackLen-1]
	context.outputState = prevState
}

// PopMergeActiveState merges the current state into the head of the stateStack,
// then pop the head of the stateStack into the current state.
// Doing this allows the VM to execute a SmartContract into a context on top
// of an existing context (a previous SC) without allowing access to it, but
// later merging the output of the two SCs in chronological order.
func (context *outputContext) PopMergeActiveState() {
	stateStackLen := len(context.stateStack)
	if stateStackLen == 0 {
		return
	}

	prevState := context.stateStack[stateStackLen-1]
	context.stateStack = context.stateStack[:stateStackLen-1]

	mergeVMOutputs(prevState, context.outputState)
	context.outputState = newVMOutput()
	mergeVMOutputs(context.outputState, prevState)
}

// PopDiscard removes the latest entry from the state stack, but maintaining
// all GasUsed values.
func (context *outputContext) PopDiscard() {
	stateStackLen := len(context.stateStack)
	if stateStackLen == 0 {
		return
	}

	context.stateStack = context.stateStack[:stateStackLen-1]
}

// ClearStateStack reinitializes the state stack.
func (context *outputContext) ClearStateStack() {
	context.stateStack = make([]*vmcommon.VMOutput, 0)
}

// CensorVMOutput will cause the next executed SC to appear isolated, as if
// nothing was executed before. Required for ExecuteOnDestContext().
// StorageUpdates are not deleted from context.outputState.OutputAccounts,
// preserving the storage cache.
func (context *outputContext) CensorVMOutput() {
	context.outputState.ReturnData = make([][]byte, 0)
	context.outputState.ReturnCode = vmcommon.Ok
	context.outputState.ReturnMessage = ""
	context.outputState.GasRemaining = 0
	context.outputState.GasRefund = big.NewInt(0)
	context.outputState.Logs = make([]*vmcommon.LogEntry, 0)

	for _, account := range context.outputState.OutputAccounts {
		newTransfers := make([]vmcommon.OutputTransfer, 0)
		for _, existingTransfer := range account.OutputTransfers {
			if isNonAsyncCallTransfer(existingTransfer) {
				newTransfers = append(newTransfers, existingTransfer)
			}
		}
		account.OutputTransfers = newTransfers
	}

	logOutput.Trace("state content censored")
}

// GetOutputAccount returns the output account present at the given address,
// and a bool that is true if the account is new. If no output account is present at that address,
// a new account will be created and added to the output accounts.
func (context *outputContext) GetOutputAccount(address []byte) (*vmcommon.OutputAccount, bool) {
	accountIsNew := false
	account, ok := context.outputState.OutputAccounts[string(address)]
	if !ok {
		account = NewVMOutputAccount(address)
		context.outputState.OutputAccounts[string(address)] = account
		accountIsNew = true
	}

	return account, accountIsNew
}

// GetOutputAccounts returns all the OutputAccounts in the current outputState.
func (context *outputContext) GetOutputAccounts() map[string]*vmcommon.OutputAccount {
	return context.outputState.OutputAccounts
}

// DeleteOutputAccount removes the given address from the output accounts and code updates
func (context *outputContext) DeleteOutputAccount(address []byte) {
	delete(context.outputState.OutputAccounts, string(address))
	delete(context.codeUpdates, string(address))
}

// GetRefund returns the value of the gas refund for the current output state.
func (context *outputContext) GetRefund() uint64 {
	return uint64(context.outputState.GasRefund.Int64())
}

// SetRefund sets the given value as gas refund for the current output state.
func (context *outputContext) SetRefund(refund uint64) {
	context.outputState.GasRefund = big.NewInt(int64(refund))
}

// ReturnData returns the data of the current output state.
func (context *outputContext) ReturnData() [][]byte {
	return context.outputState.ReturnData
}

// ReturnCode returns the code of the current output state
func (context *outputContext) ReturnCode() vmcommon.ReturnCode {
	return context.outputState.ReturnCode
}

// SetReturnCode sets the given return code as the return code for the current output state.
func (context *outputContext) SetReturnCode(returnCode vmcommon.ReturnCode) {
	context.outputState.ReturnCode = returnCode
}

// ReturnMessage returns a string that represents the return message for the current output state.
func (context *outputContext) ReturnMessage() string {
	return context.outputState.ReturnMessage
}

// SetReturnMessage sets the given string as a return message for the current output state.
func (context *outputContext) SetReturnMessage(returnMessage string) {
	context.outputState.ReturnMessage = returnMessage
}

// ClearReturnData reinitializes the return data for the current output state.
func (context *outputContext) ClearReturnData() {
	context.outputState.ReturnData = make([][]byte, 0)
}

// RemoveReturnData removes the return data item located at the specified index
func (context *outputContext) RemoveReturnData(index uint32) {
	returnData := context.outputState.ReturnData
	if index >= uint32(len(returnData)) {
		return
	}
	context.outputState.ReturnData = append(returnData[:index], returnData[index+1:]...)
}

// Finish appends the given data to the return data of the current output state.
func (context *outputContext) Finish(data []byte) {
	context.outputState.ReturnData = append(context.outputState.ReturnData, data)
	logOutput.Trace("finish", "data", data)
}

// PrependFinish appends the given data to the return data of the current output state.
func (context *outputContext) PrependFinish(data []byte) {
	context.outputState.ReturnData = append([][]byte{data}, context.outputState.ReturnData...)
}

// DeleteFirstReturnData deletes the first return data, to be used after prepend
func (context *outputContext) DeleteFirstReturnData() {
	if len(context.outputState.ReturnData) > 0 {
		context.outputState.ReturnData = context.outputState.ReturnData[1:]
	}
}

// WriteLogWithIdentifier creates a new LogEntry and appends it to the logs of the current output state.
func (context *outputContext) WriteLogWithIdentifier(address []byte, topics [][]byte, data [][]byte, identifier []byte) {
	if context.host.Runtime().ReadOnly() {
		logOutput.Trace("log entry", "error", "cannot write logs in readonly mode")
		return
	}

	newLogEntry := &vmcommon.LogEntry{
		Address:    address,
		Data:       data,
		Identifier: identifier,
	}
	logOutput.Trace("log entry", "address", address, "data", data)

	if len(topics) == 0 {
		context.outputState.Logs = append(context.outputState.Logs, newLogEntry)
		return
	}

	newLogEntry.Topics = topics

	context.outputState.Logs = append(context.outputState.Logs, newLogEntry)
	logOutput.Trace("log entry", "endpoint", newLogEntry.Identifier, "topics", newLogEntry.Topics)
}

// WriteLog creates a new LogEntry and appends it to the logs of the current output state.
func (context *outputContext) WriteLog(address []byte, topics [][]byte, data [][]byte) {
	context.WriteLogWithIdentifier(address, topics, data, []byte(context.host.Runtime().FunctionName()))
}

// TransferValueOnly will transfer the big.int value and checks if it is possible
func (context *outputContext) TransferValueOnly(destination []byte, sender []byte, value *big.Int, checkPayable bool) error {
	logOutput.Trace("transfer value", "sender", sender, "dest", destination, "value", value)

	if value.Cmp(vmhost.Zero) < 0 {
		logOutput.Trace("transfer value", "error", vmhost.ErrTransferNegativeValue)
		return vmhost.ErrTransferNegativeValue
	}

	if !context.hasSufficientBalance(sender, value) {
		logOutput.Trace("transfer value", "error", vmhost.ErrTransferInsufficientFunds)
		return vmhost.ErrTransferInsufficientFunds
	}

	payable, err := context.host.Blockchain().IsPayable(sender, destination)
	if err != nil {
		logOutput.Trace("transfer value", "error", err)
		return err
	}

	isAsyncCall := context.host.Runtime().GetVMInput().CallType == vm.AsynchronousCall
	hasValue := value.Cmp(vmhost.Zero) > 0
	if checkPayable && !payable && hasValue && !isAsyncCall {
		logOutput.Trace("transfer value", "error", vmhost.ErrAccountNotPayable)
		return vmhost.ErrAccountNotPayable
	}

	senderAcc, _ := context.GetOutputAccount(sender)
	destAcc, _ := context.GetOutputAccount(destination)

	senderAcc.BalanceDelta = big.NewInt(0).Sub(senderAcc.BalanceDelta, value)
	destAcc.BalanceDelta = big.NewInt(0).Add(destAcc.BalanceDelta, value)

	if value.Cmp(vmhost.Zero) > 0 {
		if context.host.Runtime().ReadOnly() {
			return vmhost.ErrInvalidCallOnReadOnlyMode
		}
	}

	return nil
}

func (context *outputContext) isBackTransferWithoutExecution(sender, destination []byte, input []byte) bool {
	if len(input) != 0 {
		return false
	}
	if !core.IsSmartContractAddress(destination) {
		return false
	}

	vmInput := context.host.Runtime().GetVMInput()

	currentExecutionCallerAddress := vmInput.CallerAddr
	currentExecutionDestinationAddress := vmInput.RecipientAddr

	if vmInput.CallType == vm.AsynchronousCallBack {
		currentExecutionCallerAddress = context.host.Async().GetParentAddress()
	}

	if !bytes.Equal(currentExecutionCallerAddress, destination) ||
		!bytes.Equal(currentExecutionDestinationAddress, sender) {
		return false
	}

	return true
}

// Transfer handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
func (context *outputContext) Transfer(
	destination []byte,
	sender []byte,
	gasLimit uint64,
	gasLocked uint64,
	value *big.Int,
	asyncData []byte,
	input []byte,
	callType vm.CallType,
) error {
	checkPayableIfNotCallback := gasLimit > 0 && callType != vm.AsynchronousCallBack
	isBackTransfer := context.isBackTransferWithoutExecution(sender, destination, input)
	checkPayable := checkPayableIfNotCallback || !isBackTransfer
	err := context.TransferValueOnly(destination, sender, value, checkPayable)
	if err != nil {
		return err
	}

	if (callType == vm.AsynchronousCall || callType == vm.AsynchronousCallBack) && len(asyncData) == 0 {
		return vmcommon.ErrAsyncParams
	}

	destAcc, _ := context.GetOutputAccount(destination)
	outputTransfer := vmcommon.OutputTransfer{
		Index:         context.NextOutputTransferIndex(),
		Value:         big.NewInt(0).Set(value),
		GasLimit:      gasLimit,
		GasLocked:     gasLocked,
		AsyncData:     asyncData,
		Data:          input,
		CallType:      callType,
		SenderAddress: sender,
	}
	AppendOutputTransfers(destAcc, destAcc.OutputTransfers, outputTransfer)

	logOutput.Trace("transfer value added")

	function, args, errNotCritical := context.callArgsParser.ParseData(string(input))

	isSimpleTransfer := errNotCritical != nil || !core.IsSmartContractAddress(destination) || gasLimit == 0
	if !isBackTransfer && isSimpleTransfer {
		context.WriteLogWithIdentifier(
			sender,
			[][]byte{value.Bytes(), destination},
			[][]byte{[]byte(vmhost.DirectCallString), input},
			[]byte(vmhost.TransferValueOnlyString),
		)
		return nil
	}

	executionType := callType
	if executionType == vm.DirectCall {
		executionType = vm.ESDTTransferAndExecute
	}

	context.WriteLogWithIdentifier(
		sender,
		[][]byte{value.Bytes(), destination},
		vmcommon.FormatLogDataForCall(getExecutionTypeString(executionType, isBackTransfer), function, args),
		[]byte(vmhost.TransferValueOnlyString),
	)

	return nil
}

func getExecutionTypeString(callType vm.CallType, isBackTransfer bool) string {
	if isBackTransfer {
		return vmhost.BackTransferString
	}

	switch callType {
	case vm.ESDTTransferAndExecute:
		return vmhost.TransferAndExecuteString
	case vm.AsynchronousCall:
		return vmhost.AsyncCallString
	case vm.AsynchronousCallBack:
		return vmhost.AsyncCallbackString
	}

	return vmhost.DirectCallString
}

// TransferESDT makes the esdt/nft transfer and exports the data if it is cross shard
func (context *outputContext) TransferESDT(
	transfersArgs *vmhost.ESDTTransfersArgs,
	callInput *vmcommon.ContractCallInput,
) (uint64, error) {
	if len(transfersArgs.Transfers) == 0 {
		return 0, vmhost.ErrTransferValueOnESDTCall
	}

	isSmartContract := context.host.Blockchain().IsSmartContract(transfersArgs.Destination)
	sameShard := context.host.AreInSameShard(transfersArgs.Sender, transfersArgs.Destination)
	callType := vm.DirectCall
	isExecution := isSmartContract && callInput != nil
	isBackTransfer := !isExecution && context.isBackTransferWithoutExecution(transfersArgs.Sender, transfersArgs.Destination, nil)

	if callInput != nil {
		callType = callInput.CallType
		transfersArgs.Function = callInput.Function
		transfersArgs.Arguments = callInput.Arguments
	}
	executionType := callType
	if callType == vm.DirectCall && (isExecution || isBackTransfer) {
		executionType = vm.ESDTTransferAndExecute
	}

	vmOutput, gasConsumedByTransfer, err := context.host.ExecuteESDTTransfer(transfersArgs, executionType)
	if err != nil {
		return 0, err
	}

	gasRemaining := uint64(0)

	if callInput != nil && isSmartContract {
		if gasConsumedByTransfer > callInput.GasProvided {
			logOutput.Trace("ESDT post-transfer execution", "error", vmhost.ErrNotEnoughGas)
			return 0, vmhost.ErrNotEnoughGas
		}
		gasRemaining = callInput.GasProvided - gasConsumedByTransfer
	}

	if isExecution {
		if gasRemaining > context.host.Metering().GasLeft() {
			logOutput.Trace("ESDT post-transfer execution", "error", vmhost.ErrNotEnoughGas)
			return 0, vmhost.ErrNotEnoughGas
		}

		if !sameShard {
			err = context.host.Metering().UseGasBounded(gasRemaining)
			if err != nil {
				logOutput.Trace("ESDT post-transfer execution", "error", vmhost.ErrNotEnoughGas)
				return 0, vmhost.ErrNotEnoughGas
			}
		}
	}

	destAcc, _ := context.GetOutputAccount(transfersArgs.Destination)
	outputAcc, ok := vmOutput.OutputAccounts[string(transfersArgs.Destination)]

	if ok && len(outputAcc.OutputTransfers) == 1 {
		esdtOutTransfer := outputAcc.OutputTransfers[0]
		esdtOutTransfer.GasLimit = gasRemaining
		esdtOutTransfer.CallType = callType
		esdtOutTransfer.SenderAddress = transfersArgs.SenderForExec
		if sameShard {
			esdtOutTransfer.GasLimit = 0
		}

		AppendOutputTransfers(destAcc, destAcc.OutputTransfers, esdtOutTransfer)
	}

	context.host.CompleteLogEntriesWithCallType(vmOutput, getExecutionTypeString(executionType, isBackTransfer))
	context.outputState.Logs = append(context.outputState.Logs, vmOutput.Logs...)

	return gasRemaining, nil
}

func AppendOutputTransfers(account *vmcommon.OutputAccount, existingTransfers []vmcommon.OutputTransfer, transfers ...vmcommon.OutputTransfer) {
	account.OutputTransfers = append(existingTransfers, transfers...)
	for _, transfer := range transfers {
		account.BytesConsumedByTxAsNetworking =
			math.AddUint64(account.BytesConsumedByTxAsNetworking, uint64(len(transfer.Data)))
	}
}

func (context *outputContext) hasSufficientBalance(address []byte, value *big.Int) bool {
	senderBalance := context.host.Blockchain().GetBalanceBigInt(address)
	return senderBalance.Cmp(value) >= 0
}

// AddTxValueToAccount adds the given value to the BalanceDelta of the account that is mapped to the given address
func (context *outputContext) AddTxValueToAccount(address []byte, value *big.Int) {
	destAcc, _ := context.GetOutputAccount(address)
	destAcc.BalanceDelta = big.NewInt(0).Add(destAcc.BalanceDelta, value)
}

// RemoveNonUpdatedStorage removes non updated storage from output state
func (context *outputContext) RemoveNonUpdatedStorage() {
	for _, outAcc := range context.outputState.OutputAccounts {
		for _, storageUpdate := range outAcc.StorageUpdates {
			if !storageUpdate.Written {
				delete(outAcc.StorageUpdates, string(storageUpdate.Offset))
			}
		}
	}
}

// GetVMOutput updates the current VMOutput and returns it
func (context *outputContext) GetVMOutput() *vmcommon.VMOutput {
	context.removeNonUpdatedCode()

	metering := context.host.Metering()
	context.outputState.GasRemaining = metering.GasLeft()

	err := metering.UpdateGasStateOnSuccess(context.outputState)
	if err != nil {
		return context.CreateVMOutputInCaseOfError(err)
	}

	return context.outputState
}

// DeployCode sets the given code to a an account, and creates a new codeUpdates entry at the accounts address.
func (context *outputContext) DeployCode(input vmhost.CodeDeployInput) {
	newSCAccount, _ := context.GetOutputAccount(input.ContractAddress)
	newSCAccount.Code = input.ContractCode
	newSCAccount.CodeMetadata = input.ContractCodeMetadata
	newSCAccount.CodeDeployerAddress = input.CodeDeployerAddress

	var empty struct{}
	context.codeUpdates[string(input.ContractAddress)] = empty
}

// createVMOutputInCaseOfErrorOfAsyncCallback appends the deletion of the async context to the output
func (context *outputContext) createVMOutputInCaseOfErrorOfAsyncCallback(returnCode vmcommon.ReturnCode, returnMessage string) *vmcommon.VMOutput {
	async := context.host.Async()
	metering := context.host.Metering()

	callId := async.GetCallbackAsyncInitiatorCallID()

	context.outputState = &vmcommon.VMOutput{
		GasRemaining:   0,
		GasRefund:      big.NewInt(0),
		ReturnCode:     returnCode,
		ReturnMessage:  returnMessage,
		OutputAccounts: make(map[string]*vmcommon.OutputAccount),
	}

	err := async.DeleteFromCallID(callId)
	if err != nil {
		logOutput.Trace("failed to delete Async Context", "callId", callId, "err", err)
	}

	metering.UpdateGasStateOnFailure(context.outputState)

	return context.outputState
}

// CreateVMOutputInCaseOfError creates a new vmOutput with the given error set as return message.
func (context *outputContext) CreateVMOutputInCaseOfError(err error) *vmcommon.VMOutput {
	runtime := context.host.Runtime()
	metering := context.host.Metering()

	callType := runtime.GetVMInput().CallType

	runtime.AddError(err, runtime.FunctionName())

	returnCode := context.resolveReturnCodeFromError(err)
	returnMessage := context.resolveReturnMessageFromError(err)

	if context.host.EnableEpochsHandler().IsFlagEnabled(vmhost.AsyncV3Flag) && callType == vm.AsynchronousCallBack {
		return context.createVMOutputInCaseOfErrorOfAsyncCallback(returnCode, returnMessage)
	}

	vmOutput := &vmcommon.VMOutput{
		GasRemaining:  0,
		GasRefund:     big.NewInt(0),
		ReturnCode:    returnCode,
		ReturnMessage: returnMessage,
	}

	metering.UpdateGasStateOnFailure(vmOutput)

	return vmOutput
}

func (context *outputContext) removeNonUpdatedCode() {
	for address, account := range context.outputState.OutputAccounts {
		_, ok := context.codeUpdates[address]
		if !ok {
			account.Code = nil
			account.CodeMetadata = nil
			account.CodeDeployerAddress = nil
		}
	}
}

func (context *outputContext) resolveReturnMessageFromError(err error) string {
	if errors.Is(err, vmhost.ErrSignalError) {
		return context.ReturnMessage()
	}
	if errors.Is(err, vmhost.ErrMemoryLimit) {
		// ErrMemoryLimit will still produce the 'execution failed' message.
		return vmhost.ErrExecutionFailed.Error()
	}
	if len(context.outputState.ReturnMessage) > 0 {
		// Another return message was already set.
		return context.outputState.ReturnMessage
	}

	return err.Error()
}

func (context *outputContext) resolveReturnCodeFromError(err error) vmcommon.ReturnCode {
	if err == nil {
		return vmcommon.Ok
	}

	if errors.Is(err, vmhost.ErrSignalError) {
		return vmcommon.UserError
	}
	if errors.Is(err, executor.ErrFuncNotFound) {
		return vmcommon.FunctionNotFound
	}
	if errors.Is(err, executor.ErrFunctionNonvoidSignature) {
		return vmcommon.FunctionWrongSignature
	}
	if errors.Is(err, executor.ErrInvalidFunction) {
		return vmcommon.UserError
	}
	if errors.Is(err, vmhost.ErrInitFuncCalledInRun) {
		return vmcommon.UserError
	}
	if errors.Is(err, vmhost.ErrCallBackFuncCalledInRun) {
		return vmcommon.UserError
	}
	if errors.Is(err, vmhost.ErrNotEnoughGas) {
		return vmcommon.OutOfGas
	}
	if errors.Is(err, vmhost.ErrContractNotFound) {
		return vmcommon.ContractNotFound
	}
	if errors.Is(err, vmhost.ErrContractInvalid) {
		return vmcommon.ContractInvalid
	}
	if errors.Is(err, vmhost.ErrUpgradeFailed) {
		return vmcommon.UpgradeFailed
	}
	if errors.Is(err, vmhost.ErrTransferInsufficientFunds) {
		return vmcommon.OutOfFunds
	}

	return vmcommon.ExecutionFailed
}

// AddToActiveState merges the given vmOutput with the outputState.
func (context *outputContext) AddToActiveState(rightOutput *vmcommon.VMOutput) {
	if rightOutput.GasRefund != nil {
		rightOutput.GasRefund.Add(rightOutput.GasRefund, context.outputState.GasRefund)
	}

	for _, rightAccount := range rightOutput.OutputAccounts {
		leftAccount, ok := context.outputState.OutputAccounts[string(rightAccount.Address)]
		if !ok {
			continue
		}

		if rightAccount.BalanceDelta != nil {
			rightAccount.BalanceDelta.Add(rightAccount.BalanceDelta, leftAccount.BalanceDelta)
		}
	}

	mergeVMOutputsConditionally(context.outputState, rightOutput, true)
}

// NextOutputTransferIndex returns next available output transfer index
func (context *outputContext) NextOutputTransferIndex() uint32 {
	index := context.crtTransferIndex
	context.crtTransferIndex++
	return index
}

// GetCrtTransferIndex returns the current output transfer index
func (context *outputContext) GetCrtTransferIndex() uint32 {
	return context.crtTransferIndex
}

// SetCrtTransferIndex sets the current output transfer index
func (context *outputContext) SetCrtTransferIndex(index uint32) {
	context.crtTransferIndex = index
}

func mergeVMOutputs(leftOutput *vmcommon.VMOutput, rightOutput *vmcommon.VMOutput) {
	mergeVMOutputsConditionally(leftOutput, rightOutput, false)
}

func mergeVMOutputsConditionally(leftOutput *vmcommon.VMOutput, rightOutput *vmcommon.VMOutput, mergeAllTransfers bool) {
	if leftOutput.OutputAccounts == nil {
		leftOutput.OutputAccounts = make(map[string]*vmcommon.OutputAccount)
	}

	for _, rightAccount := range rightOutput.OutputAccounts {
		leftAccount, ok := leftOutput.OutputAccounts[string(rightAccount.Address)]
		if !ok {
			leftAccount = &vmcommon.OutputAccount{}
			leftOutput.OutputAccounts[string(rightAccount.Address)] = leftAccount
		}
		mergeOutputAccounts(leftAccount, rightAccount, mergeAllTransfers)
	}

	leftOutput.Logs = append(leftOutput.Logs, rightOutput.Logs...)
	leftOutput.ReturnData = append(leftOutput.ReturnData, rightOutput.ReturnData...)
	leftOutput.GasRemaining = rightOutput.GasRemaining
	leftOutput.GasRefund = rightOutput.GasRefund
	if leftOutput.GasRefund == nil {
		leftOutput.GasRefund = big.NewInt(0)
	}

	leftOutput.ReturnCode = rightOutput.ReturnCode
	leftOutput.ReturnMessage = rightOutput.ReturnMessage

	leftOutput.DeletedAccounts = append(leftOutput.DeletedAccounts, rightOutput.DeletedAccounts...)
}

func mergeOutputAccounts(
	leftAccount *vmcommon.OutputAccount,
	rightAccount *vmcommon.OutputAccount,
	mergeAllTransfers bool,
) {
	if len(rightAccount.Address) != 0 {
		leftAccount.Address = rightAccount.Address
	}

	mergeStorageUpdates(leftAccount, rightAccount)

	if rightAccount.Balance != nil {
		leftAccount.Balance = rightAccount.Balance
	}
	if leftAccount.BalanceDelta == nil {
		leftAccount.BalanceDelta = big.NewInt(0)
	}
	if rightAccount.BalanceDelta != nil {
		leftAccount.BalanceDelta = rightAccount.BalanceDelta
	}
	if len(rightAccount.Code) > 0 {
		leftAccount.Code = rightAccount.Code
	}
	if len(rightAccount.CodeMetadata) > 0 {
		leftAccount.CodeMetadata = rightAccount.CodeMetadata
	}
	if rightAccount.Nonce > leftAccount.Nonce {
		leftAccount.Nonce = rightAccount.Nonce
	}

	mergeTransfers(leftAccount, rightAccount, mergeAllTransfers)

	leftAccount.GasUsed = rightAccount.GasUsed

	if rightAccount.CodeDeployerAddress != nil {
		leftAccount.CodeDeployerAddress = rightAccount.CodeDeployerAddress
	}

	if rightAccount.BytesAddedToStorage > leftAccount.BytesAddedToStorage {
		leftAccount.BytesAddedToStorage = rightAccount.BytesAddedToStorage
	}
	if rightAccount.BytesDeletedFromStorage > leftAccount.BytesDeletedFromStorage {
		leftAccount.BytesDeletedFromStorage = rightAccount.BytesDeletedFromStorage
	}
}

func mergeTransfers(leftAccount *vmcommon.OutputAccount, rightAccount *vmcommon.OutputAccount, mergeAllTransfers bool) {
	leftAsyncCallTransfers, leftOtherTransfers := splitTransfers(leftAccount)
	rightAsyncCallTransfers, rightOtherTransfers := splitTransfers(rightAccount)

	leftAsyncCallTransfers = append(leftAsyncCallTransfers, rightAsyncCallTransfers...)

	lenLeftOtherTransfers := len(leftOtherTransfers)
	lenRightOtherTransfers := len(rightOtherTransfers)
	if mergeAllTransfers {
		leftOtherTransfers = append(leftOtherTransfers, rightOtherTransfers...)
	} else if lenRightOtherTransfers > lenLeftOtherTransfers {
		leftOtherTransfers = append(leftOtherTransfers, rightOtherTransfers[lenLeftOtherTransfers:]...)
	}

	leftAccount.BytesConsumedByTxAsNetworking = 0
	AppendOutputTransfers(leftAccount, leftAsyncCallTransfers, leftOtherTransfers...)
}

func splitTransfers(account *vmcommon.OutputAccount) ([]vmcommon.OutputTransfer, []vmcommon.OutputTransfer) {
	if account.OutputTransfers == nil {
		return nil, nil
	}
	asyncCallTransfers := make([]vmcommon.OutputTransfer, 0)
	otherTransfers := make([]vmcommon.OutputTransfer, 0)
	for _, transfer := range account.OutputTransfers {
		if isNonAsyncCallTransfer(transfer) {
			otherTransfers = append(otherTransfers, transfer)
		} else {
			asyncCallTransfers = append(asyncCallTransfers, transfer)
		}
	}
	return asyncCallTransfers, otherTransfers
}

func isNonAsyncCallTransfer(transfer vmcommon.OutputTransfer) bool {
	return transfer.CallType != vm.AsynchronousCall && transfer.CallType != vm.AsynchronousCallBack
}

func mergeStorageUpdates(
	leftAccount *vmcommon.OutputAccount,
	rightAccount *vmcommon.OutputAccount,
) {
	if leftAccount.StorageUpdates == nil {
		leftAccount.StorageUpdates = make(map[string]*vmcommon.StorageUpdate)
	}

	for key, update := range rightAccount.StorageUpdates {
		leftAccount.StorageUpdates[key] = update
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (context *outputContext) IsInterfaceNil() bool {
	return context == nil
}
