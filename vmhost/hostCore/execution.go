package hostCore

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-go/executor"
	"github.com/multiversx/mx-chain-vm-go/math"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
	"github.com/multiversx/mx-chain-vm-go/vmhost/contexts"
	"github.com/multiversx/mx-chain-vm-go/vmhost/vmhooks"
)

func (host *vmHost) doRunSmartContractCreate(input *vmcommon.ContractCreateInput) *vmcommon.VMOutput {
	host.InitState()
	defer func() {
		errs := host.GetRuntimeErrors()
		if errs != nil {
			log.Trace("doRunSmartContractCreate full error list", "error", errs)
		}
	}()

	_, blockchain, metering, output, runtime, _, storage := host.GetContexts()

	var vmOutput *vmcommon.VMOutput
	defer func() {
		if vmOutput == nil || vmOutput.ReturnCode == vmcommon.ExecutionFailed {
			runtime.CleanInstance()
		}
	}()

	address, err := blockchain.NewAddress(input.CallerAddr)
	if err != nil {
		vmOutput = output.CreateVMOutputInCaseOfError(err)
		return vmOutput
	}

	contractCallInput := &vmcommon.ContractCallInput{
		VMInput:       input.VMInput,
		RecipientAddr: address,
		Function:      vmhost.InitFunctionName,
	}
	runtime.SetVMInput(contractCallInput)
	runtime.SetCodeAddress(address)
	metering.InitStateFromContractCallInput(&input.VMInput)

	output.AddTxValueToAccount(address, input.CallValue)
	storage.SetAddress(runtime.GetContextAddress())

	codeDeployInput := vmhost.CodeDeployInput{
		ContractCode:         input.ContractCode,
		ContractCodeMetadata: input.ContractCodeMetadata,
		ContractAddress:      address,
		CodeDeployerAddress:  input.CallerAddr,
	}

	vmOutput, err = host.performCodeDeploymentAtContractCreate(codeDeployInput)
	if err != nil {
		log.Trace("doRunSmartContractCreate", "error", err)
		vmOutput = output.CreateVMOutputInCaseOfError(err)
		return vmOutput
	}

	log.Trace("doRunSmartContractCreate",
		"retCode", vmOutput.ReturnCode,
		"message", vmOutput.ReturnMessage,
		"data", vmOutput.ReturnData)

	return vmOutput
}

func (host *vmHost) performCodeDeployment(input vmhost.CodeDeployInput, initFunction func() error) (*vmcommon.VMOutput, error) {
	log.Trace("performCodeDeployment", "address", input.ContractAddress, "len(code)", len(input.ContractCode), "metadata", input.ContractCodeMetadata)

	_, _, metering, output, runtime, _, _ := host.GetContexts()

	err := metering.DeductInitialGasForDirectDeployment(input)
	if err != nil {
		output.SetReturnCode(vmcommon.OutOfGas)
		return nil, err
	}

	runtime.MustVerifyNextContractCode()

	err = runtime.StartWasmerInstance(input.ContractCode, metering.GetGasForExecution(), true)
	if err != nil {
		log.Trace("performCodeDeployment/StartWasmerInstance", "err", err)
		return nil, vmhost.ErrContractInvalid
	}

	err = initFunction()
	if err != nil {
		return nil, err
	}

	output.DeployCode(input)
	output.RemoveNonUpdatedStorage()

	vmOutput := output.GetVMOutput()
	return vmOutput, nil
}

func (host *vmHost) performCodeDeploymentAtContractCreate(input vmhost.CodeDeployInput) (*vmcommon.VMOutput, error) {
	return host.performCodeDeployment(input, host.callInitFunction)
}

func (host *vmHost) performCodeDeploymentAtContractUpgrade(input vmhost.CodeDeployInput) (*vmcommon.VMOutput, error) {
	return host.performCodeDeployment(input, host.callUpgradeFunction)
}

// doRunSmartContractUpgrade upgrades a contract directly
func (host *vmHost) doRunSmartContractUpgrade(input *vmcommon.ContractCallInput) *vmcommon.VMOutput {
	host.InitState()
	defer func() {
		errs := host.GetRuntimeErrors()
		if errs != nil {
			log.Trace("doRunSmartContractUpgrade full error list", "error", errs)
		}
	}()

	_, _, metering, output, runtime, _, storage := host.GetContexts()

	var vmOutput *vmcommon.VMOutput
	defer func() {
		if vmOutput == nil || vmOutput.ReturnCode == vmcommon.ExecutionFailed {
			runtime.CleanInstance()
		}
	}()

	err := host.checkUpgradePermission(input)
	if err != nil {
		log.Trace("doRunSmartContractUpgrade", "error", vmhost.ErrUpgradeNotAllowed)
		vmOutput = output.CreateVMOutputInCaseOfError(err)
		return vmOutput
	}

	runtime.InitStateFromContractCallInput(input)
	metering.InitStateFromContractCallInput(&input.VMInput)
	output.AddTxValueToAccount(input.RecipientAddr, input.CallValue)
	storage.SetAddress(runtime.GetContextAddress())

	code, codeMetadata, err := runtime.ExtractCodeUpgradeFromArgs()
	if err != nil {
		vmOutput = output.CreateVMOutputInCaseOfError(vmhost.ErrInvalidUpgradeArguments)
		return vmOutput
	}

	codeDeployInput := vmhost.CodeDeployInput{
		ContractCode:         code,
		ContractCodeMetadata: codeMetadata,
		ContractAddress:      input.RecipientAddr,
		CodeDeployerAddress:  input.CallerAddr,
	}

	vmOutput, err = host.performCodeDeploymentAtContractUpgrade(codeDeployInput)
	if err != nil {
		log.Trace("doRunSmartContractUpgrade", "error", err)
		vmOutput = output.CreateVMOutputInCaseOfError(err)
		return vmOutput
	}

	return vmOutput
}

func (host *vmHost) checkGasForGetCode(input *vmcommon.ContractCallInput, metering vmhost.MeteringContext) error {
	getCodeBaseCost := metering.GasSchedule().BaseOperationCost.GetCode
	if input.GasProvided < getCodeBaseCost {
		return vmhost.ErrNotEnoughGas
	}

	return nil
}

// doRunSmartContractDelete deletes a contract directly
func (host *vmHost) doRunSmartContractDelete(input *vmcommon.ContractCallInput) *vmcommon.VMOutput {
	output := host.Output()
	err := host.checkUpgradePermission(input)
	if err != nil {
		log.Trace("doRunSmartContractDelete", "error", vmhost.ErrUpgradeNotAllowed)
		return output.CreateVMOutputInCaseOfError(err)
	}

	vmOutput := output.GetVMOutput()
	vmOutput.DeletedAccounts = append(vmOutput.DeletedAccounts, input.RecipientAddr)
	return vmOutput
}

func (host *vmHost) doRunSmartContractCall(input *vmcommon.ContractCallInput) *vmcommon.VMOutput {
	host.InitState()
	defer func() {
		errs := host.GetRuntimeErrors()
		if errs != nil {
			log.Trace(fmt.Sprintf("doRunSmartContractCall full error list for %s", input.Function), "error", errs)
		}
	}()

	_, _, metering, output, runtime, async, storage := host.GetContexts()

	var vmOutput *vmcommon.VMOutput
	defer func() {
		if vmOutput == nil || vmOutput.ReturnCode == vmcommon.ExecutionFailed {
			host.Runtime().CleanInstance()
		}
	}()

	runtime.InitStateFromContractCallInput(input)

	err := async.InitStateFromInput(input)
	if err != nil {
		log.Trace("doRunSmartContractCall init async", "error", vmhost.ErrAsyncInit)
		vmOutput = output.CreateVMOutputInCaseOfError(err)
		return vmOutput
	}
	metering.InitStateFromContractCallInput(&input.VMInput)
	output.AddTxValueToAccount(input.RecipientAddr, input.CallValue)
	storage.SetAddress(runtime.GetContextAddress())

	err = host.checkGasForGetCode(input, metering)
	if err != nil {
		log.Trace("doRunSmartContractCall check gas for GetSCCode", "error", vmhost.ErrNotEnoughGas)
		vmOutput = output.CreateVMOutputInCaseOfError(vmhost.ErrNotEnoughGas)
		return vmOutput
	}

	contract, err := runtime.GetSCCode()
	if err != nil {
		log.Trace("doRunSmartContractCall get code", "error", vmhost.ErrContractNotFound)
		vmOutput = output.CreateVMOutputInCaseOfError(vmhost.ErrContractNotFound)
		return vmOutput
	}

	err = metering.DeductInitialGasForExecution(contract)
	if err != nil {
		log.Trace("doRunSmartContractCall initial gas", "error", vmhost.ErrNotEnoughGas)
		vmOutput = output.CreateVMOutputInCaseOfError(vmhost.ErrNotEnoughGas)
		return vmOutput
	}

	err = runtime.StartWasmerInstance(contract, metering.GetGasForExecution(), false)
	if err != nil {
		vmOutput = output.CreateVMOutputInCaseOfError(vmhost.ErrContractInvalid)
		return vmOutput
	}

	err = host.callSCMethod()
	if err != nil {
		log.Trace("doRunSmartContractCall", "error", err)
		vmOutput = output.CreateVMOutputInCaseOfError(err)
		return vmOutput
	}

	output.RemoveNonUpdatedStorage()
	vmOutput = output.GetVMOutput()
	host.CompleteLogEntriesWithCallType(vmOutput, vmhost.DirectCallString)

	log.Trace("doRunSmartContractCall finished",
		"retCode", vmOutput.ReturnCode,
		"message", vmOutput.ReturnMessage,
		"data", vmOutput.ReturnData)

	return vmOutput
}

func copyTxHashesFromContext(runtime vmhost.RuntimeContext, input *vmcommon.ContractCallInput) {
	if input.CallType != vm.DirectCall {
		return
	}
	currentVMInput := runtime.GetVMInput()
	if len(currentVMInput.OriginalTxHash) > 0 {
		input.OriginalTxHash = currentVMInput.OriginalTxHash
	}
	if len(currentVMInput.CurrentTxHash) > 0 {
		input.CurrentTxHash = currentVMInput.CurrentTxHash
	}
	if len(currentVMInput.PrevTxHash) > 0 {
		input.PrevTxHash = currentVMInput.PrevTxHash
	}

}

// ExecuteOnDestContext pushes each context to the corresponding stack
// and initializes new contexts for executing the contract call with the given input
func (host *vmHost) ExecuteOnDestContext(input *vmcommon.ContractCallInput) (vmOutput *vmcommon.VMOutput, isChildComplete bool, err error) {
	log.Trace("ExecuteOnDestContext", "caller", input.CallerAddr, "dest", input.RecipientAddr, "function", input.Function, "gas", input.GasProvided)

	scExecutionInput := input

	blockchain := host.Blockchain()

	blockchain.PushState()

	if host.IsBuiltinFunctionName(input.Function) {
		scExecutionInput, vmOutput, err = host.handleBuiltinFunctionCall(input)
		if err != nil {
			blockchain.PopSetActiveState()
			host.Runtime().AddError(err, input.Function)
			vmOutput = host.Output().CreateVMOutputInCaseOfError(err)
			isChildComplete = true
			return
		}
		host.completeLogEntriesAfterBuiltinCall(input, vmOutput)
	}

	isChildComplete = true
	if scExecutionInput != nil {
		vmOutput, isChildComplete, err = host.executeOnDestContextNoBuiltinFunction(scExecutionInput)
		host.addNewBackTransfersFromVMOutput(vmOutput, scExecutionInput.CallerAddr, scExecutionInput.RecipientAddr)
	}

	if err != nil {
		blockchain.PopSetActiveState()
	} else {
		blockchain.PopDiscard()
	}

	return
}

func (host *vmHost) isESDTTransferWithoutExecution(transferData []byte, parent, child []byte) (*vmcommon.ParsedESDTTransfers, bool) {
	function, args, err := host.callArgsParser.ParseData(string(transferData))
	if err != nil {
		return nil, false
	}

	esdtTransfers, err := host.esdtTransferParser.ParseESDTTransfers(child, parent, function, args)
	if err != nil {
		return nil, false
	}
	if esdtTransfers.CallFunction != "" {
		return nil, false
	}

	return esdtTransfers, true
}

func (host *vmHost) addNewBackTransfersFromVMOutput(vmOutput *vmcommon.VMOutput, parent, child []byte) {
	if vmOutput == nil || vmOutput.ReturnCode != vmcommon.Ok {
		return
	}
	callerOutAcc, ok := vmOutput.OutputAccounts[string(parent)]
	if !ok {
		return
	}

	for _, transfer := range callerOutAcc.OutputTransfers {
		if !bytes.Equal(transfer.SenderAddress, child) {
			continue
		}
		if transfer.CallType == vm.AsynchronousCallBack {
			continue
		}

		if transfer.Value.Cmp(vmhost.Zero) > 0 {
			if len(transfer.Data) == 0 {
				host.managedTypesContext.AddValueOnlyBackTransfer(transfer.Value)
			}
			continue
		}

		esdtTransfers, isWithoutExec := host.isESDTTransferWithoutExecution(transfer.Data, parent, child)
		if !isWithoutExec {
			continue
		}

		host.managedTypesContext.AddBackTransfers(esdtTransfers.ESDTTransfers)
	}
}

func (host *vmHost) completeLogEntriesAfterBuiltinCall(input *vmcommon.ContractCallInput, vmOutput *vmcommon.VMOutput) {
	switch input.CallType {
	case vm.AsynchronousCall:
		host.CompleteLogEntriesWithCallType(vmOutput, vmhost.AsyncCallString)
	case vm.AsynchronousCallBack:
		host.CompleteLogEntriesWithCallType(vmOutput, vmhost.AsyncCallbackString)
	default:
		host.CompleteLogEntriesWithCallType(vmOutput, vmhost.ExecuteOnDestContextString)
	}
}

func (host *vmHost) handleFunctionCallOnOtherVM(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	output := host.Output()

	vmOutput, err := host.callFunctionOnOtherVM(input)
	if err != nil {
		log.Trace("ExecuteOnDestContext function on other VM", "error", err)
		return nil, err
	}

	output.AddToActiveState(vmOutput)

	return vmOutput, nil
}

func (host *vmHost) handleBuiltinFunctionCall(input *vmcommon.ContractCallInput) (*vmcommon.ContractCallInput, *vmcommon.VMOutput, error) {
	output := host.Output()

	postBuiltinInput, builtinOutput, err := host.callBuiltinFunction(input)
	if err != nil {
		log.Trace("ExecuteOnDestContext builtin function", "error", err)
		return nil, nil, err
	}

	if postBuiltinInput != nil {
		postBuiltinInput.AsyncArguments = input.AsyncArguments
	}

	err = contexts.AddAsyncArgumentsToOutputTransfers(
		input.AsyncArguments,
		vm.AsynchronousCall,
		builtinOutput)
	if err != nil {
		log.Trace("ExecuteOnDestContext builtin function", "error", err)
		return nil, nil, err
	}

	output.AddToActiveState(builtinOutput)

	return postBuiltinInput, builtinOutput, nil
}

func (host *vmHost) executeOnDestContextNoBuiltinFunction(input *vmcommon.ContractCallInput) (vmOutput *vmcommon.VMOutput, isChildComplete bool, err error) {
	if host.IsOutOfVMFunctionExecution(input) {
		vmOutput, err = host.handleFunctionCallOnOtherVM(input)
		if err != nil {
			host.Runtime().AddError(err, input.Function)
			vmOutput = host.Output().CreateVMOutputInCaseOfError(err)
		}

		return vmOutput, true, err
	}

	managedTypes, _, metering, output, runtime, async, storage := host.GetContexts()
	managedTypes.PushState()
	managedTypes.InitState()
	managedTypes.PopBackTransferIfAsyncCallBack(input)

	output.PushState()
	output.CensorVMOutput()

	copyTxHashesFromContext(runtime, input)
	runtime.PushState()
	runtime.InitStateFromContractCallInput(input)

	async.PushState()
	err = async.InitStateFromInput(input)
	if err != nil {
		runtime.AddError(err, input.Function)
		return nil, true, err
	}

	metering.PushState()
	metering.InitStateFromContractCallInput(&input.VMInput)

	storage.PushState()
	storage.SetAddress(runtime.GetContextAddress())

	defer func() {
		vmOutput = host.finishExecuteOnDestContext(err)
		if err == nil && vmOutput.ReturnCode != vmcommon.Ok {
			err = vmhost.ErrExecutionFailed
		}
		runtime.AddError(err, input.Function)
	}()

	// Perform a value transfer to the called SC. If the execution fails, this
	// transfer will not persist.
	isZeroBaseTransfer := input.CallValue.Cmp(vmhost.Zero) == 0
	if len(input.ESDTTransfers) == 0 && (input.CallType != vm.AsynchronousCallBack || isZeroBaseTransfer) {
		err = output.TransferValueOnly(input.RecipientAddr, input.CallerAddr, input.CallValue, false)
		if err != nil {
			log.Trace("ExecuteOnDestContext transfer", "error", err)
			return vmOutput, true, err
		}
	}
	isAsyncCallBackWithBaseTransfer := !isZeroBaseTransfer && input.CallType == vm.AsynchronousCallBack
	if len(input.ESDTTransfers) == 0 {
		tmpCallValue := big.NewInt(0).Set(input.CallValue)
		if isAsyncCallBackWithBaseTransfer {
			tmpCallValue = big.NewInt(0)
		}

		output.WriteLogWithIdentifier(
			input.CallerAddr,
			[][]byte{tmpCallValue.Bytes(), input.RecipientAddr},
			vmcommon.FormatLogDataForCall("", input.Function, input.Arguments),
			[]byte(vmhost.TransferValueOnlyString),
		)
	}

	err = host.execute(input)
	if err != nil {
		log.Trace("ExecuteOnDestContext execution", "error", err)
		return vmOutput, true, err
	}

	err = async.Execute()

	return vmOutput, async.IsComplete(), err
}

func (host *vmHost) finishExecuteOnDestContext(executeErr error) *vmcommon.VMOutput {
	managedTypes, _, metering, output, runtime, async, storage := host.GetContexts()

	var vmOutput *vmcommon.VMOutput
	if executeErr != nil {
		// Execution failed: restore contexts as if the execution didn't happen,
		// but first create a vmOutput to capture the error.
		vmOutput = output.CreateVMOutputInCaseOfError(executeErr)
	} else {
		// Retrieve the VMOutput before popping the Runtime state and the previous
		// instance, to ensure accurate GasRemaining
		vmOutput = output.GetVMOutput()
	}

	async.SetResults(vmOutput)
	if !async.IsComplete() || async.HasLegacyGroup() {
		saveErr := async.Save()
		if saveErr != nil {
			vmOutput = output.CreateVMOutputInCaseOfError(saveErr)
		}
	}

	gasSpentByChildContract := metering.GasSpentByContract()

	// Restore the previous context states
	managedTypes.PopSetActiveState()
	storage.PopSetActiveState()

	if vmOutput.ReturnCode == vmcommon.Ok {
		metering.PopMergeActiveState()
		output.PopMergeActiveState()
	} else {
		metering.PopSetActiveState()
		output.PopSetActiveState()
	}

	log.Trace("ExecuteOnDestContext finished", "sc", string(runtime.GetContextAddress()), "function", runtime.FunctionName())
	log.Trace("ExecuteOnDestContext finished", "gas spent", gasSpentByChildContract, "gas remaining", vmOutput.GasRemaining)

	isAsyncCall := runtime.GetVMInput().CallType == vm.AsynchronousCall
	isAsyncComplete := async.IsComplete()

	// Return to the caller context completely
	runtime.PopSetActiveState()

	async.PopSetActiveState()

	// Restore remaining gas to the caller Wasmer instance
	if !isAsyncCall || isAsyncComplete {
		metering.RestoreGas(vmOutput.GasRemaining)
	}

	return vmOutput
}

// ExecuteOnSameContext executes the contract call with the given input
// on the same runtime context. Some other contexts are backed up.
func (host *vmHost) ExecuteOnSameContext(input *vmcommon.ContractCallInput) error {
	log.Trace("ExecuteOnSameContext", "function", input.Function)

	if host.IsBuiltinFunctionName(input.Function) {
		return vmhost.ErrBuiltinCallOnSameContextDisallowed
	}

	managedTypes, blockchain, metering, output, runtime, _, _ := host.GetContexts()

	// Back up the states of the contexts (except Storage and Async, which aren't affected
	// by ExecuteOnSameContext())
	managedTypes.PushState()
	managedTypes.InitState()
	output.PushState()

	librarySCAddress := make([]byte, len(input.RecipientAddr))
	copy(librarySCAddress, input.RecipientAddr)

	input.RecipientAddr = input.CallerAddr
	copyTxHashesFromContext(runtime, input)
	runtime.PushState()
	runtime.InitStateFromContractCallInput(input)
	runtime.SetCodeAddress(librarySCAddress)

	metering.PushState()
	metering.InitStateFromContractCallInput(&input.VMInput)

	blockchain.PushState()

	var err error

	defer host.finishExecuteOnSameContext(err)

	// Perform a value transfer to the called SC. If the execution fails, this
	// transfer will not persist.
	err = output.TransferValueOnly(input.RecipientAddr, input.CallerAddr, input.CallValue, false)
	if err != nil {
		runtime.AddError(err, input.Function)
		return err
	}
	output.WriteLogWithIdentifier(
		input.CallerAddr,
		[][]byte{input.CallValue.Bytes(), input.RecipientAddr},
		vmcommon.FormatLogDataForCall(vmhost.ExecuteOnSameContextString, input.Function, input.Arguments),
		[]byte(vmhost.TransferValueOnlyString),
	)

	err = host.execute(input)
	runtime.AddError(err, input.Function)
	return err
}

func (host *vmHost) finishExecuteOnSameContext(executeErr error) {
	managedTypes, blockchain, metering, output, runtime, _, _ := host.GetContexts()

	if output.ReturnCode() != vmcommon.Ok || executeErr != nil {
		// Execution failed: restore contexts as if the execution didn't happen.
		managedTypes.PopSetActiveState()
		metering.PopSetActiveState()
		output.PopSetActiveState()
		blockchain.PopSetActiveState()
		runtime.PopSetActiveState()
		return
	}

	// Execution successful; retrieve the VMOutput before popping the Runtime
	// state and the previous instance, to ensure accurate GasRemaining and
	// GasUsed for all accounts.
	vmOutput := output.GetVMOutput()

	metering.PopMergeActiveState()
	output.PopDiscard()
	blockchain.PopDiscard()
	managedTypes.PopSetActiveState()
	runtime.PopSetActiveState()
	// Restore remaining gas to the caller (parent) Wasmer instance
	metering.RestoreGas(vmOutput.GasRemaining)
}

func (host *vmHost) isInitFunctionBeingCalled() bool {
	functionName := host.Runtime().FunctionName()
	return functionName == vmhost.InitFunctionName
}

// IsOutOfVMFunctionExecution returns true if the call should be executed on ahother VM
func (host *vmHost) IsOutOfVMFunctionExecution(input *vmcommon.ContractCallInput) bool {
	isSmartContract := host.Blockchain().IsSmartContract(input.RecipientAddr)
	if isSmartContract {
		vmType, err := vmcommon.ParseVMTypeFromContractAddress(input.RecipientAddr)
		if err != nil {
			return false
		}
		if !bytes.Equal(host.Runtime().GetVMType(), vmType) {
			return true
		}
	}
	return false
}

// IsBuiltinFunctionName returns true if the given function name is the same as any protocol builtin function
func (host *vmHost) IsBuiltinFunctionName(functionName string) bool {
	function, err := host.builtInFuncContainer.Get(functionName)
	if err != nil {
		return false
	}

	return function.IsActive()
}

// IsBuiltinFunctionCall returns true if the given data contains a call to a protocol builtin function
func (host *vmHost) IsBuiltinFunctionCall(data []byte) bool {
	functionName, _, _ := host.callArgsParser.ParseData(string(data))
	return host.IsBuiltinFunctionName(functionName)
}

// CreateNewContract creates a new contract indirectly (from another Smart Contract)
func (host *vmHost) CreateNewContract(input *vmcommon.ContractCreateInput, createContractCallType int) (newContractAddress []byte, err error) {
	newContractAddress = nil
	err = nil

	defer func() {
		if err != nil {
			newContractAddress = nil
		}
	}()

	_, blockchain, metering, output, runtime, _, _ := host.GetContexts()

	codeDeployInput := vmhost.CodeDeployInput{
		ContractCode:         input.ContractCode,
		ContractCodeMetadata: input.ContractCodeMetadata,
		ContractAddress:      nil,
		CodeDeployerAddress:  input.CallerAddr,
	}
	err = metering.DeductInitialGasForIndirectDeployment(codeDeployInput)
	if err != nil {
		return
	}

	if runtime.ReadOnly() {
		err = vmhost.ErrInvalidCallOnReadOnlyMode
		return
	}

	newContractAddress, err = blockchain.NewAddress(input.CallerAddr)
	if err != nil {
		return
	}

	if blockchain.AccountExists(newContractAddress) {
		err = vmhost.ErrDeploymentOverExistingAccount
		return
	}

	codeDeployInput.ContractAddress = newContractAddress
	output.DeployCode(codeDeployInput)

	defer func() {
		if err != nil {
			output.DeleteOutputAccount(newContractAddress)
		}
	}()

	runtime.MustVerifyNextContractCode()

	initCallInput := &vmcommon.ContractCallInput{
		RecipientAddr:     newContractAddress,
		Function:          vmhost.InitFunctionName,
		AllowInitFunction: true,
		VMInput:           input.VMInput,
	}

	var isChildComplete bool
	host.Async().SetAsyncArgumentsForCall(initCallInput)
	initVmOutput, isChildComplete, err := host.ExecuteOnDestContext(initCallInput)
	if err != nil {
		return
	}

	if createContractCallType == vmhooks.DeployContract {
		host.CompleteLogEntriesWithCallType(initVmOutput, vmhost.DeployFromSourceString)
	} else {
		host.CompleteLogEntriesWithCallType(initVmOutput, vmhost.DeploySmartContractString)
	}

	err = host.Async().CompleteChildConditional(isChildComplete, nil, 0)
	if err != nil {
		return
	}

	blockchain.IncreaseNonce(input.CallerAddr)

	return
}

func (host *vmHost) checkUpgradePermission(vmInput *vmcommon.ContractCallInput) error {
	contract, err := host.Blockchain().GetUserAccount(vmInput.RecipientAddr)
	if err != nil {
		return err
	}
	if check.IfNilReflect(contract) {
		return vmhost.ErrNilContract
	}

	codeMetadata := vmcommon.CodeMetadataFromBytes(contract.GetCodeMetadata())
	isUpgradeable := codeMetadata.Upgradeable
	callerAddress := vmInput.CallerAddr
	ownerAddress := contract.GetOwnerAddress()
	isCallerOwner := bytes.Equal(callerAddress, ownerAddress)

	if isUpgradeable && isCallerOwner {
		return nil
	}

	return vmhost.ErrUpgradeNotAllowed
}

// executeUpgrade upgrades a contract indirectly (from another contract). This
// function follows the convention of executeSmartContractCall().
func (host *vmHost) executeUpgrade(input *vmcommon.ContractCallInput) error {
	_, _, metering, output, runtime, _, _ := host.GetContexts()

	err := host.checkUpgradePermission(input)
	if err != nil {
		return err
	}

	code, codeMetadata, err := runtime.ExtractCodeUpgradeFromArgs()
	if err != nil {
		return vmhost.ErrInvalidUpgradeArguments
	}

	codeDeployInput := vmhost.CodeDeployInput{
		ContractCode:         code,
		ContractCodeMetadata: codeMetadata,
		ContractAddress:      input.RecipientAddr,
		CodeDeployerAddress:  input.CallerAddr,
	}

	err = metering.DeductInitialGasForDirectDeployment(codeDeployInput)
	if err != nil {
		output.SetReturnCode(vmcommon.OutOfGas)
		return err
	}

	runtime.MustVerifyNextContractCode()

	err = runtime.StartWasmerInstance(codeDeployInput.ContractCode, metering.GetGasForExecution(), true)
	if err != nil {
		log.Trace("performCodeDeployment/StartWasmerInstance", "err", err)
		return vmhost.ErrContractInvalid
	}

	err = host.callUpgradeFunction()
	if err != nil {
		return err
	}

	output.DeployCode(codeDeployInput)
	if output.ReturnCode() != vmcommon.Ok {
		return vmhost.ErrReturnCodeNotOk
	}

	return nil
}

func (host *vmHost) executeDelete(input *vmcommon.ContractCallInput) error {
	host.doRunSmartContractDelete(input)
	return nil
}

// execute executes an indirect call to a smart contract, assuming there is an
// already-running Wasmer instance of another contract that has requested the
// indirect call. This method creates a new Wasmer instance and pushes the
// previous one onto the Runtime instance stack, but it will not pop the
// previous instance back - that remains the responsibility of the calling
// code. Also, this method does not restore the gas remaining after the
// indirect call, it does not push the states of any Host Context onto their
// respective stacks, nor does it pop any state stack. Handling the state
// stacks and the remaining gas are responsibilities of the calling code, which
// must push and pop as required, before and after calling this method, and
// handle the remaining gas. These principles also apply to indirect contract
// upgrading (via host.executeUpgrade(), which also does not pop the previous
// instance from the Runtime instance stack, nor does it restore the remaining
// gas).
func (host *vmHost) execute(input *vmcommon.ContractCallInput) error {
	_, _, metering, output, runtime, _, _ := host.GetContexts()

	if host.isInitFunctionBeingCalled() && !input.AllowInitFunction {
		return vmhost.ErrInitFuncCalledInRun
	}

	// Use all gas initially, on the Wasmer instance of the caller. In case of
	// successful execution, the unused gas will be restored.
	metering.UseGasForContractInit(input.GasProvided)

	isUpgrade := input.Function == vmhost.UpgradeFunctionName
	if isUpgrade {
		return host.executeUpgrade(input)
	}

	isDelete := input.Function == vmhost.DeleteFunctionName
	if isDelete {
		return host.executeDelete(input)
	}

	contract, err := runtime.GetSCCode()
	if err != nil {
		return err
	}

	err = metering.DeductInitialGasForExecution(contract)
	if err != nil {
		return err
	}

	// Replace the current Wasmer instance of the Runtime with a new one; this
	// assumes that the instance was preserved on the Runtime instance stack
	// before calling executeSmartContractCall().
	err = runtime.StartWasmerInstance(contract, metering.GetGasForExecution(), false)
	if err != nil {
		return err
	}

	err = host.callSCMethodIndirect()
	if err != nil {
		return err
	}

	if output.ReturnCode() != vmcommon.Ok {
		return vmhost.ErrReturnCodeNotOk
	}

	return nil
}

func (host *vmHost) callSCMethodIndirect() error {
	log.Trace("callSCMethodIndirect")
	functionName, err := host.Runtime().FunctionNameChecked()
	if err != nil {
		if errors.Is(err, vmhost.ErrNilCallbackFunction) {
			return nil
		}
		return err
	}

	err = host.Runtime().CallSCFunction(functionName)
	if err != nil {
		err = host.handleBreakpointIfAny(err)
	}

	return err
}

// ExecuteESDTTransfer calls the process built in function with the given transfer for ESDT/ESDTNFT if nonce > 0
// there are no NFTs with nonce == 0, it will call multi transfer if multiple tokens are sent
func (host *vmHost) ExecuteESDTTransfer(transfersArgs *vmhost.ESDTTransfersArgs, callType vm.CallType) (*vmcommon.VMOutput, uint64, error) {
	if len(transfersArgs.Transfers) == 0 {
		return nil, 0, vmhost.ErrFailedTransfer
	}

	if host.Runtime().ReadOnly() {
		return nil, 0, vmhost.ErrInvalidCallOnReadOnlyMode
	}

	_, _, metering, _, runtime, _, _ := host.GetContexts()

	esdtTransferInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			OriginalCallerAddr:   transfersArgs.OriginalCaller,
			CallerAddr:           transfersArgs.Sender,
			Arguments:            make([][]byte, 0),
			CallValue:            big.NewInt(0),
			CallType:             callType,
			GasPrice:             runtime.GetVMInput().GasPrice,
			GasProvided:          metering.GasLeft(),
			GasLocked:            0,
			ReturnCallAfterError: transfersArgs.ReturnAfterError,
		},
		RecipientAddr:     transfersArgs.Destination,
		Function:          core.BuiltInFunctionESDTTransfer,
		AllowInitFunction: false,
	}

	transfers := transfersArgs.Transfers
	if len(transfers) == 1 {
		if transfers[0].ESDTTokenNonce > 0 {
			esdtTransferInput.Function = core.BuiltInFunctionESDTNFTTransfer
			esdtTransferInput.RecipientAddr = esdtTransferInput.CallerAddr
			nonceAsBytes := big.NewInt(0).SetUint64(transfers[0].ESDTTokenNonce).Bytes()
			esdtTransferInput.Arguments = append(esdtTransferInput.Arguments, transfers[0].ESDTTokenName, nonceAsBytes, transfers[0].ESDTValue.Bytes(), transfersArgs.Destination)
		} else {
			esdtTransferInput.Arguments = append(esdtTransferInput.Arguments, transfers[0].ESDTTokenName, transfers[0].ESDTValue.Bytes())
		}
	} else {
		esdtTransferInput.Function = core.BuiltInFunctionMultiESDTNFTTransfer
		esdtTransferInput.RecipientAddr = esdtTransferInput.CallerAddr
		esdtTransferInput.Arguments = append(esdtTransferInput.Arguments, transfersArgs.Destination, big.NewInt(int64(len(transfers))).Bytes())
		for _, transfer := range transfers {
			nonceAsBytes := big.NewInt(0).SetUint64(transfer.ESDTTokenNonce).Bytes()
			esdtTransferInput.Arguments = append(esdtTransferInput.Arguments, transfer.ESDTTokenName, nonceAsBytes, transfer.ESDTValue.Bytes())
		}
	}

	if len(transfersArgs.Function) > 0 {
		esdtTransferInput.Arguments = append(esdtTransferInput.Arguments, []byte(transfersArgs.Function))
	}
	if len(transfersArgs.Arguments) > 0 {
		esdtTransferInput.Arguments = append(esdtTransferInput.Arguments, transfersArgs.Arguments...)
	}

	vmOutput, err := host.Blockchain().ProcessBuiltInFunction(esdtTransferInput)
	log.Trace("ESDT transfer", "sender", transfersArgs.Sender, "dest", transfersArgs.Destination)
	for _, transfer := range transfers {
		log.Trace("ESDT transfer", "token", transfer.ESDTTokenName, "nonce", transfer.ESDTTokenNonce, "value", transfer.ESDTValue)
	}
	if err != nil {
		log.Trace("ESDT transfer", "error", err)
		return vmOutput, esdtTransferInput.GasProvided, err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		log.Trace("ESDT transfer", "error", err, "retcode", vmOutput.ReturnCode, "message", vmOutput.ReturnMessage)
		return vmOutput, esdtTransferInput.GasProvided, vmhost.ErrExecutionFailed
	}

	err = vmOutput.ReindexTransfers(host.Output())
	if err != nil {
		return nil, 0, err
	}

	host.addESDTTransferToVMOutputSCIntraShardCall(esdtTransferInput, vmOutput)

	gasConsumed := math.SubUint64(esdtTransferInput.GasProvided, vmOutput.GasRemaining)
	for _, outAcc := range vmOutput.OutputAccounts {
		for _, transfer := range outAcc.OutputTransfers {
			gasConsumed = math.SubUint64(gasConsumed, transfer.GasLimit)
		}
	}
	if callType != vm.AsynchronousCallBack {
		if metering.GasLeft() < gasConsumed {
			log.Trace("ESDT transfer", "error", vmhost.ErrNotEnoughGas)
			return vmOutput, esdtTransferInput.GasProvided, vmhost.ErrNotEnoughGas
		}
		err = metering.UseGasBounded(gasConsumed)
		if err != nil {
			log.Trace("ESDT transfer", "error", vmhost.ErrNotEnoughGas)
			return vmOutput, esdtTransferInput.GasProvided, vmhost.ErrNotEnoughGas
		}
	}

	return vmOutput, gasConsumed, nil
}

func (host *vmHost) callFunctionOnOtherVM(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	metering := host.Metering()

	vmOutput, err := host.Blockchain().ExecuteSmartContractCallOnOtherVM(input)
	if err != nil {
		_ = metering.UseGasBounded(input.GasProvided)
		return nil, err
	}

	err = vmOutput.ReindexTransfers(host.Output())
	if err != nil {
		return nil, err
	}

	metering.TrackGasUsedByOutOfVMFunction(input, vmOutput, nil)

	host.addESDTTransferToVMOutputSCIntraShardCall(input, vmOutput)

	return vmOutput, nil
}

func (host *vmHost) callBuiltinFunction(input *vmcommon.ContractCallInput) (*vmcommon.ContractCallInput, *vmcommon.VMOutput, error) {
	metering := host.Metering()

	if host.Runtime().ReadOnly() {
		return nil, nil, vmhost.ErrInvalidCallOnReadOnlyMode
	}

	vmOutput, err := host.Blockchain().ProcessBuiltInFunction(input)
	if err != nil {
		_ = metering.UseGasBounded(input.GasProvided)
		return nil, nil, err
	}

	newVMInput, err := host.isSCExecutionAfterBuiltInFunc(input, vmOutput)
	if err != nil {
		_ = metering.UseGasBounded(input.GasProvided)
		return nil, nil, err
	}

	if newVMInput != nil {
		for _, outAcc := range vmOutput.OutputAccounts {
			outAcc.OutputTransfers = make([]vmcommon.OutputTransfer, 0)
		}
	}

	// reindex only for the case of no execution after builtin call
	err = vmOutput.ReindexTransfers(host.Output())
	if err != nil {
		return nil, nil, err
	}

	metering.TrackGasUsedByOutOfVMFunction(input, vmOutput, newVMInput)

	host.addESDTTransferToVMOutputSCIntraShardCall(input, vmOutput)

	return newVMInput, vmOutput, nil
}

// add output transfer of esdt transfer when sc calling another sc intra shard to log the transfer information
func (host *vmHost) addESDTTransferToVMOutputSCIntraShardCall(
	input *vmcommon.ContractCallInput,
	output *vmcommon.VMOutput,
) {
	if output.ReturnCode != vmcommon.Ok {
		return
	}

	parsedTransfer, err := host.esdtTransferParser.ParseESDTTransfers(input.CallerAddr, input.RecipientAddr, input.Function, input.Arguments)
	if err != nil {
		return
	}

	if !host.AreInSameShard(input.CallerAddr, parsedTransfer.RcvAddr) {
		return
	}

	host.addOutputTransferToVMOutput(input.Function, input.Arguments, input.CallerAddr, parsedTransfer.RcvAddr, input.CallType, output)
}

func (host *vmHost) addOutputTransferToVMOutput(
	function string,
	arguments [][]byte,
	sender []byte,
	recipient []byte,
	callType vm.CallType,
	vmOutput *vmcommon.VMOutput,
) {
	esdtTransferTxData := function
	for _, arg := range arguments {
		esdtTransferTxData += "@" + hex.EncodeToString(arg)
	}
	outTransfer := vmcommon.OutputTransfer{
		Index:         host.Output().NextOutputTransferIndex(),
		Value:         big.NewInt(0),
		Data:          []byte(esdtTransferTxData),
		CallType:      callType,
		SenderAddress: sender,
	}

	if len(vmOutput.OutputAccounts) == 0 {
		vmOutput.OutputAccounts = make(map[string]*vmcommon.OutputAccount)
	}
	outAcc, ok := vmOutput.OutputAccounts[string(recipient)]
	if !ok {
		outAcc = &vmcommon.OutputAccount{
			Address:         recipient,
			OutputTransfers: make([]vmcommon.OutputTransfer, 0),
		}
	}
	contexts.AppendOutputTransfers(outAcc, outAcc.OutputTransfers, outTransfer)
	vmOutput.OutputAccounts[string(recipient)] = outAcc
}

func (host *vmHost) checkFinalGasAfterExit() error {
	totalUsedPoints := host.Runtime().GetPointsUsed()
	if totalUsedPoints > host.Metering().GetGasForExecution() {
		log.Trace("checkFinalGasAfterExit", "failed")
		return vmhost.ErrNotEnoughGas
	}

	log.Trace("checkFinalGasAfterExit", "ok")
	return nil
}

func (host *vmHost) checkValidFunctionName(name string) error {
	if name == "" {
		return executor.ErrInvalidFunction
	}

	return nil
}

func (host *vmHost) callInitFunction() error {
	return host.callSCFunction(vmhost.InitFunctionName)
}

func (host *vmHost) callUpgradeFunction() error {
	return host.callSCFunction(vmhost.ContractsUpgradeFunctionName)
}

func (host *vmHost) callSCFunction(functionName string) error {
	err := host.checkValidFunctionName(functionName)
	if err != nil {
		log.Trace("call SC method failed", "error", err, "src", "checkValidFunctionName")
		return err
	}

	runtime := host.Runtime()
	if !runtime.HasFunction(functionName) {
		return executor.ErrFuncNotFound
	}

	err = runtime.CallSCFunction(functionName)
	if err != nil {
		err = host.handleBreakpointIfAny(err)
	}

	if err == nil {
		err = host.checkFinalGasAfterExit()
	}

	return err
}

func (host *vmHost) callSCMethod() error {
	log.Trace("callSCMethod")

	runtime := host.Runtime()
	callType := runtime.GetVMInput().CallType

	var err error
	switch callType {
	case vm.DirectCall:
		err = host.callSCMethodDirectCall()
	case vm.AsynchronousCall:
		err = host.callSCMethodAsynchronousCall()
	case vm.AsynchronousCallBack:
		err = host.callSCMethodAsynchronousCallBack()
	default:
		err = vmhost.ErrUnknownCallType
	}

	if err != nil {
		log.Trace("call SC method failed", "error", err, "callType", callType)
	}

	return err
}

func (host *vmHost) callSCMethodDirectCall() error {
	_, err := host.callFunctionAndExecuteAsync()
	return err
}

func (host *vmHost) callSCMethodAsynchronousCall() error {
	_, err := host.callFunctionAndExecuteAsync()
	return err
}

func (host *vmHost) callSCMethodAsynchronousCallBack() error {
	runtime := host.Runtime()
	async := host.Async()

	callerCallID := async.GetCallerCallID()

	asyncCall, isLegacy, err := async.UpdateCurrentAsyncCallStatus(
		runtime.GetContextAddress(),
		callerCallID,
		&runtime.GetVMInput().VMInput)
	if err != nil {
		return err
	}

	if asyncCall == nil {
		return nil
	}

	if asyncCall.HasPendingCallback {
		return nil
	}

	async.SetCallbackParentCall(asyncCall)

	if asyncCall.HasCallback() {
		callbackName := asyncCall.GetCallbackName()
		runtime.SetCustomCallFunction(callbackName)
		isCallComplete, callbackErr := host.callFunctionAndExecuteAsync()

		if callbackErr != nil {
			metering := host.Metering()
			_ = metering.UseGasBounded(metering.GasLeft())
		}

		if !isCallComplete {
			return callbackErr
		}
	}

	if isLegacy {
		return async.DeleteFromCallID(async.GetCallbackAsyncInitiatorCallID())
	}

	err = async.LoadParentContext()
	if err != nil {
		return err
	}

	return async.NotifyChildIsComplete(callerCallID, host.Metering().GasLeft())
}

func (host *vmHost) callFunctionAndExecuteAsync() (bool, error) {
	runtime := host.Runtime()
	async := host.Async()

	err := host.checkValidFunctionName(runtime.FunctionName())
	if err != nil {
		log.Trace("call SC method failed", "error", err, "src", "checkValidFunctionName")
		return false, err
	}

	err = host.verifyAllowedFunctionCall()
	if err != nil {
		log.Trace("call SC method failed", "error", err, "src", "verifyAllowedFunctionCall")
		return false, err
	}

	functionName, err := runtime.FunctionNameChecked()
	if err != nil {
		log.Trace("call SC method failed", "error", err, "src", "FunctionNameChecked")
		return false, err
	}

	err = runtime.CallSCFunction(functionName)
	if err != nil {
		err = host.handleBreakpointIfAny(err)
		log.Trace("breakpoint detected and handled", "err", err)
	}
	if err == nil {
		err = host.checkFinalGasAfterExit()
	}
	if err != nil {
		log.Trace("call SC method failed", "error", err, "src", "sc function")
		return true, err
	}

	err = async.Execute()
	if err != nil {
		log.Trace("call SC method failed", "error", err, "src", "async execution")
		return false, err
	}

	if !async.IsComplete() || async.HasLegacyGroup() {
		async.SetResults(host.Output().GetVMOutput())
		err = async.Save()
		return false, err
	}

	return true, nil
}

func (host *vmHost) verifyAllowedFunctionCall() error {
	runtime := host.Runtime()
	functionName := runtime.FunctionName()

	isInit := functionName == vmhost.InitFunctionName
	if isInit {
		return vmhost.ErrInitFuncCalledInRun
	}
	isUpgrade := functionName == vmhost.ContractsUpgradeFunctionName
	if isUpgrade {
		return vmhost.ErrInitFuncCalledInRun
	}

	isCallBack := functionName == vmhost.CallbackFunctionName
	isInAsyncCallBack := runtime.GetVMInput().CallType == vm.AsynchronousCallBack
	if isCallBack && !isInAsyncCallBack {
		return vmhost.ErrCallBackFuncCalledInRun
	}

	return nil
}

func (host *vmHost) isSCExecutionAfterBuiltInFunc(
	vmInput *vmcommon.ContractCallInput,
	vmOutput *vmcommon.VMOutput,
) (*vmcommon.ContractCallInput, error) {
	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, nil
	}
	if vmInput.ReturnCallAfterError && vmInput.CallType != vm.AsynchronousCallBack {
		return nil, nil
	}

	parsedTransfer, err := host.esdtTransferParser.ParseESDTTransfers(vmInput.CallerAddr, vmInput.RecipientAddr, vmInput.Function, vmInput.Arguments)
	if err != nil {
		return nil, nil
	}

	if !host.AreInSameShard(vmInput.CallerAddr, parsedTransfer.RcvAddr) {
		return nil, nil
	}
	if !host.Blockchain().IsSmartContract(parsedTransfer.RcvAddr) {
		return nil, nil
	}

	outAcc, ok := vmOutput.OutputAccounts[string(parsedTransfer.RcvAddr)]
	if !ok {
		return nil, nil
	}
	if len(outAcc.OutputTransfers) != 1 {
		return nil, nil
	}

	callType := vmInput.CallType
	scCallOutTransfer := outAcc.OutputTransfers[0]

	function, arguments, err := host.callArgsParser.ParseData(string(scCallOutTransfer.Data))
	if err != nil {
		return nil, err
	}

	// TODO CurrentTxHash might equal PrevTxHash, because VM does not generate
	// SCRs between async steps; analyze and fix this (e.g. create dummy SCRs
	// just for hashing, or hash the VMInput itself)
	newVMInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			OriginalCallerAddr: vmInput.OriginalCallerAddr,
			CallerAddr:         vmInput.CallerAddr,
			Arguments:          arguments,
			CallValue:          big.NewInt(0),
			CallType:           callType,
			GasPrice:           vmInput.GasPrice,
			GasProvided:        scCallOutTransfer.GasLimit,
			GasLocked:          scCallOutTransfer.GasLocked,
			OriginalTxHash:     vmInput.OriginalTxHash,
			CurrentTxHash:      vmInput.CurrentTxHash,
			PrevTxHash:         vmInput.PrevTxHash,
			RelayerAddr:        vmInput.RelayerAddr,
		},
		RecipientAddr:     parsedTransfer.RcvAddr,
		Function:          function,
		AllowInitFunction: false,
	}

	newVMInput.ESDTTransfers = parsedTransfer.ESDTTransfers

	return newVMInput, nil
}
