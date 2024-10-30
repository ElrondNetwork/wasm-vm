package contexts

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-go/math"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
)

type lastTransferInfo struct {
	callValue         *big.Int
	lastESDTTransfers []*vmcommon.ESDTTransfer
}

func (context *asyncContext) executeAsyncLocalCalls() error {
	localCalls := make([]*vmhost.AsyncCall, 0)
	hasAnyRemoteCallbacks := false

	for _, group := range context.asyncCallGroups {
		for _, call := range group.AsyncCalls {
			if call.IsLocal() {
				localCalls = append(localCalls, call)
			}

			if call.IsRemote() && call.HasCallback() {
				hasAnyRemoteCallbacks = true
			}
		}
	}

	if hasAnyRemoteCallbacks {
		currentCall := context.GetAsyncCallByCallID(context.GetCallID()).GetAsyncCall()
		if currentCall != nil && currentCall.IsAsyncV3 {
			currentCall.MarkSkippedCallback(currentCall.GetGasLocked())
		}
	}

	for _, call := range localCalls {
		err := context.executeAsyncLocalCall(call)
		if err != nil {
			return err
		}
	}

	return nil
}

func (context *asyncContext) executeAsyncLocalCall(asyncCall *vmhost.AsyncCall) error {
	destinationCallInput, err := context.createContractCallInput(asyncCall)
	if err != nil {
		logAsync.Trace("executeAsyncLocalCall failed", "error", err)
		return err
	}

	logAsync.Trace("executeAsyncLocalCall",
		"caller", destinationCallInput.CallerAddr,
		"dest", destinationCallInput.RecipientAddr,
		"func", destinationCallInput.Function,
		"args", destinationCallInput.Arguments,
		"gasProvided", destinationCallInput.GasProvided,
		"gasLocked", destinationCallInput.GasLocked)

	// Briefly restore the AsyncCall GasLimit, after it was consumed in its
	// entirety by addAsyncCall(); this is required, because ExecuteOnDestContext()
	// must also consume the GasLimit in its entirety, before starting execution,
	// but will restore any GasRemaining to the current instance.
	metering := context.host.Metering()
	metering.RestoreGas(asyncCall.GetGasLimit())

	vmOutput, isComplete, err := context.host.ExecuteOnDestContext(destinationCallInput)
	if vmOutput == nil {
		return vmhost.ErrNilDestinationCallVMOutput
	}

	if destinationCallInput.Function == vmhost.UpgradeFunctionName {
		context.host.CompleteLogEntriesWithCallType(vmOutput, vmhost.UpgradeFromSourceString)
	} else {
		context.host.CompleteLogEntriesWithCallType(vmOutput, vmhost.AsyncCallString)
	}

	logAsync.Trace("executeAsyncLocalCall",
		"retCode", vmOutput.ReturnCode,
		"message", vmOutput.ReturnMessage,
		"data", vmOutput.ReturnData,
		"gasRemaining", vmOutput.GasRemaining,
		"error", err)

	asyncCall.UpdateStatus(vmOutput.ReturnCode)

	if isComplete {
		return context.executeAsyncCallbackAndComplete(asyncCall, vmOutput, destinationCallInput, err)
	}

	return nil
}

func (context *asyncContext) executeAsyncCallbackAndComplete(
	asyncCall *vmhost.AsyncCall,
	destinationVMOutput *vmcommon.VMOutput,
	destinationCallInput *vmcommon.ContractCallInput,
	destinationErr error,
) error {
	callbackGasRemaining := uint64(0)
	if asyncCall.HasCallback() {
		// Restore gas locked while still on the caller instance; otherwise, the
		// locked gas will appear to have been used twice by the caller instance.
		isCallbackComplete, callbackVMOutput := context.ExecuteLocalCallbackAndFinishOutput(asyncCall, destinationVMOutput, destinationCallInput, 0, destinationErr)
		if callbackVMOutput == nil {
			return vmhost.ErrAsyncNoOutputFromCallback
		}

		context.host.CompleteLogEntriesWithCallType(callbackVMOutput, vmhost.AsyncCallbackString)

		if isCallbackComplete {
			callbackGasRemaining = callbackVMOutput.GasRemaining
			callbackVMOutput.GasRemaining = 0
		}
	}

	return context.completeChild(asyncCall.CallID, callbackGasRemaining)
}

// ExecuteLocalCallbackAndFinishOutput executes the callback and finishes the output
func (context *asyncContext) ExecuteLocalCallbackAndFinishOutput(
	asyncCall *vmhost.AsyncCall,
	vmOutput *vmcommon.VMOutput,
	_ *vmcommon.ContractCallInput,
	gasAccumulated uint64,
	err error) (bool, *vmcommon.VMOutput) {
	callbackVMOutput, isComplete, _ := context.executeLocalCallback(asyncCall, vmOutput, gasAccumulated, err)
	context.finishAsyncLocalCallbackExecution()
	return isComplete, callbackVMOutput
}

func (context *asyncContext) executeLocalCallback(
	asyncCall *vmhost.AsyncCall,
	destinationVMOutput *vmcommon.VMOutput,
	gasAccumulated uint64,
	destinationErr error,
) (*vmcommon.VMOutput, bool, error) {
	callbackInput, err := context.createCallbackInput(asyncCall, destinationVMOutput, gasAccumulated, destinationErr)
	if err != nil {
		logAsync.Trace("executeLocalCallback", "error", err)
		return nil, true, err
	}

	logAsync.Trace("executeLocalCallback",
		"caller", callbackInput.CallerAddr,
		"dest", callbackInput.RecipientAddr,
		"func", callbackInput.Function,
		"args", callbackInput.Arguments,
		"gasProvided", callbackInput.GasProvided,
		"gasLocked", callbackInput.GasLocked)

	context.host.Metering().RestoreGas(asyncCall.GasLocked)
	callbackVMOutput, isComplete, callbackErr := context.host.ExecuteOnDestContext(callbackInput)
	if callbackVMOutput != nil {
		logAsync.Trace("async call: sync callback call",
			"retCode", callbackVMOutput.ReturnCode,
			"message", callbackVMOutput.ReturnMessage,
			"data", callbackVMOutput.ReturnData,
			"gasRemaining", callbackVMOutput.GasRemaining,
			"error", callbackErr)
	}

	return callbackVMOutput, isComplete, callbackErr
}

// executeSyncHalfOfBuiltinFunction will synchronously call the requested
// built-in function. This is required for all cross-shard calls to built-in
// functions, because they will handle cross-shard calls themselves, by
// generating entries in vmOutput.OutputAccounts, and they need to be executed
// synchronously to do that. As a consequence, it is not necessary to call
// sendAsyncCallCrossShard(). The vmOutput produced by the built-in function,
// containing the cross-shard call, has ALREADY been merged into the main
// output by the inner call to host.ExecuteOnDestContext(). Moreover, the
// status of the AsyncCall is not updated here - it will be updated by
// PostprocessCrossShardCallback(), when the cross-shard call returns.
func (context *asyncContext) executeSyncHalfOfBuiltinFunction(asyncCall *vmhost.AsyncCall) error {
	destinationCallInput, err := context.createContractCallInput(asyncCall)
	if err != nil {
		return err
	}

	// Briefly restore the AsyncCall GasLimit, after it was consumed in its
	// entirety by addAsyncCall(); this is required, because ExecuteOnDestContext()
	// must also consume the GasLimit in its entirety, before starting execution,
	// but will restore any GasRemaining to the current instance.
	metering := context.host.Metering()
	metering.RestoreGas(asyncCall.GetGasLimit())

	vmOutput, _, err := context.host.ExecuteOnDestContext(destinationCallInput)
	if err != nil {
		return err
	}

	// If the in-shard half of the built-in function call has failed, go no
	// further and execute the error callback of this AsyncCall.
	if vmOutput.ReturnCode != vmcommon.Ok {
		asyncCall.Reject()
		if asyncCall.HasCallback() {
			_, _, _ = context.executeLocalCallback(asyncCall, vmOutput, 0, err)
			context.finishAsyncLocalCallbackExecution()
		}
	}

	// The gas that remains after executing the in-shard half of the built-in
	// function is provided to the cross-shard half.
	asyncCall.GasLimit = vmOutput.GasRemaining

	return nil
}

func (context *asyncContext) finishAsyncLocalCallbackExecution() {
	runtime := context.host.Runtime()
	runtime.GetVMInput().GasProvided = 0
}

func (context *asyncContext) createContractCallInput(asyncCall *vmhost.AsyncCall) (*vmcommon.ContractCallInput, error) {
	host := context.host
	runtime := host.Runtime()
	sender := runtime.GetContextAddress()
	originalCaller := runtime.GetOriginalCallerAddress()

	function, arguments, err := context.callArgsParser.ParseData(string(asyncCall.GetData()))
	if err != nil {
		return nil, err
	}

	gasLimit := asyncCall.GetGasLimit()
	gasToUse := host.Metering().GasSchedule().BaseOpsAPICost.AsyncCallStep
	if gasLimit <= gasToUse {
		return nil, vmhost.ErrNotEnoughGas
	}

	contractCallInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			OriginalCallerAddr: originalCaller,
			CallerAddr:         sender,
			Arguments:          arguments,
			CallValue:          big.NewInt(0).SetBytes(asyncCall.GetValue()),
			CallType:           vm.AsynchronousCall,
			GasPrice:           runtime.GetVMInput().GasPrice,
			GasProvided:        gasLimit,
			GasLocked:          asyncCall.GetGasLocked(),
			CurrentTxHash:      runtime.GetCurrentTxHash(),
			OriginalTxHash:     runtime.GetOriginalTxHash(),
			PrevTxHash:         runtime.GetPrevTxHash(),
		},
		RecipientAddr: asyncCall.GetDestination(),
		Function:      function,
	}
	context.SetAsyncArgumentsForCall(contractCallInput)
	asyncCall.CallID = contractCallInput.AsyncArguments.CallID

	return contractCallInput, nil
}

func (context *asyncContext) createCallbackInput(
	asyncCall *vmhost.AsyncCall,
	vmOutput *vmcommon.VMOutput,
	gasAccumulated uint64,
	destinationErr error,
) (*vmcommon.ContractCallInput, error) {
	runtime := context.host.Runtime()

	actualCallbackInitiator, err := context.determineDestinationForAsyncCall(asyncCall.GetDestination(), asyncCall.GetData())
	if err != nil {
		return nil, err
	}

	arguments := context.getArgumentsForCallback(vmOutput, destinationErr)
	returnWithError := false
	if destinationErr != nil || vmOutput.ReturnCode != vmcommon.Ok {
		returnWithError = true
	}

	callbackFunction := asyncCall.GetCallbackName()
	dataLength := computeDataLengthFromArguments(callbackFunction, arguments)
	gasLimit, err := context.computeGasLimitForCallback(asyncCall, vmOutput, dataLength)
	if err != nil {
		return nil, err
	}

	originalCaller := runtime.GetOriginalCallerAddress()
	caller := context.address
	lastTransferData := context.extractLastTransferWithoutData(caller, vmOutput)

	// Return to the sender SC, calling its specified callback method.
	contractCallInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			OriginalCallerAddr:   originalCaller,
			CallerAddr:           actualCallbackInitiator,
			Arguments:            arguments,
			CallValue:            lastTransferData.callValue,
			CallType:             vm.AsynchronousCallBack,
			GasPrice:             runtime.GetVMInput().GasPrice,
			GasProvided:          gasLimit,
			GasLocked:            0,
			CurrentTxHash:        runtime.GetCurrentTxHash(),
			OriginalTxHash:       runtime.GetOriginalTxHash(),
			PrevTxHash:           runtime.GetPrevTxHash(),
			ReturnCallAfterError: returnWithError,
			ESDTTransfers:        lastTransferData.lastESDTTransfers,
		},
		RecipientAddr: caller,
		Function:      callbackFunction,
	}
	context.SetAsyncArgumentsForCallback(contractCallInput, asyncCall, gasAccumulated)

	return contractCallInput, nil
}

func (context *asyncContext) extractLastTransferWithoutData(caller []byte, vmOutput *vmcommon.VMOutput) lastTransferInfo {
	callValue := big.NewInt(0)
	emptyLastTransferInfo := lastTransferInfo{
		callValue:         big.NewInt(0),
		lastESDTTransfers: nil,
	}

	callBackReceiver := context.host.Runtime().GetContextAddress()
	outAcc, ok := vmOutput.OutputAccounts[string(callBackReceiver)]
	if !ok || len(outAcc.OutputTransfers) == 0 || len(vmOutput.ReturnData) > 0 {
		return emptyLastTransferInfo
	}

	lastOutTransfer := outAcc.OutputTransfers[len(outAcc.OutputTransfers)-1]
	if len(lastOutTransfer.Data) == 0 || len(vmOutput.ReturnData) == 0 {
		callValue.Set(lastOutTransfer.Value)
	}

	var lastESDTTransfers []*vmcommon.ESDTTransfer
	functionName, args, err := context.callArgsParser.ParseData(string(lastOutTransfer.Data))
	if err != nil {
		return lastTransferInfo{
			callValue:         callValue,
			lastESDTTransfers: lastESDTTransfers,
		}
	}

	builtInFunction := context.host.IsBuiltinFunctionName(functionName)
	if !builtInFunction {
		return lastTransferInfo{
			callValue:         callValue,
			lastESDTTransfers: lastESDTTransfers,
		}
	}

	parsedESDTTransfers, err := context.esdtTransferParser.ParseESDTTransfers(lastOutTransfer.SenderAddress, caller, functionName, args)
	if err == nil && parsedESDTTransfers.CallFunction == "" {
		lastESDTTransfers = parsedESDTTransfers.ESDTTransfers
	}

	return lastTransferInfo{
		callValue:         callValue,
		lastESDTTransfers: lastESDTTransfers,
	}
}

// ReturnCodeToBytes returns the provided returnCode as byte slice
func ReturnCodeToBytes(returnCode vmcommon.ReturnCode) []byte {
	if returnCode == vmcommon.Ok {
		return []byte{0}
	}
	return big.NewInt(int64(returnCode)).Bytes()
}

func (context *asyncContext) computeGasLimitForCallback(asyncCall *vmhost.AsyncCall, vmOutput *vmcommon.VMOutput, dataLength int) (uint64, error) {
	metering := context.host.Metering()
	gasLimit := math.AddUint64(vmOutput.GasRemaining, asyncCall.GetGasLocked())

	gasToUse := metering.GasSchedule().BaseOpsAPICost.AsyncCallStep
	copyPerByte := metering.GasSchedule().BaseOperationCost.DataCopyPerByte
	gas := math.MulUint64(copyPerByte, uint64(dataLength))
	gasToUse = math.AddUint64(gasToUse, gas)
	if gasLimit <= gasToUse {
		return 0, vmhost.ErrNotEnoughGas
	}
	gasLimit -= gasToUse

	return gasLimit, nil
}

func (context *asyncContext) getArgumentsForCallback(vmOutput *vmcommon.VMOutput, err error) [][]byte {
	// always provide return code as the first argument to callback function
	arguments := [][]byte{
		ReturnCodeToBytes(vmOutput.ReturnCode),
	}
	if err == nil && vmOutput.ReturnCode == vmcommon.Ok {
		// when execution went Ok, callBack arguments are:
		// [0, result1, result2, ....]
		arguments = append(arguments, vmOutput.ReturnData...)
	} else {
		// when execution returned error, callBack arguments are:
		// [error code, error message]
		arguments = append(arguments, []byte(vmOutput.ReturnMessage))
	}

	return arguments
}
