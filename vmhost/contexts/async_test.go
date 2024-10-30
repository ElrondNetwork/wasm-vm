package contexts

import (
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-vm-go/wasmer2"

	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-scenario-go/worldmock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	"github.com/multiversx/mx-chain-vm-go/config"
	"github.com/multiversx/mx-chain-vm-go/crypto/factory"
	"github.com/multiversx/mx-chain-vm-go/executor"
	contextmock "github.com/multiversx/mx-chain-vm-go/mock/context"
	"github.com/multiversx/mx-chain-vm-go/testcommon/testexecutor"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
	"github.com/multiversx/mx-chain-vm-go/vmhost/vmhooks"
	"github.com/stretchr/testify/require"
)

var mockWasmerInstance *contextmock.InstanceMock
var OriginalCaller = []byte("address_original_caller")
var Alice = []byte("address_alice")
var Bob = []byte("address_bob")

const GasForAsyncStep = config.GasValueForTests

var marshalizer = &marshal.GogoProtoMarshalizer{}

func makeAsyncContext(t testing.TB, host vmhost.VMHost, address []byte) *asyncContext {
	callParser := parsers.NewCallArgsParser()
	esdtParser, _ := parsers.NewESDTTransferParser(worldmock.WorldMarshalizer)
	async, err := NewAsyncContext(
		host,
		callParser,
		esdtParser,
		marshalizer,
	)
	require.Nil(t, err)
	require.NotNil(t, async)

	async.address = address

	return async
}

func initializeVMAndWasmerAsyncContext(tb testing.TB) (*contextmock.VMHostMock, *worldmock.MockWorld) {
	return initializeVMAndWasmerAsyncContextWithBuiltIn(tb, false)
}

func initializeVMAndWasmerAsyncContextWithBuiltIn(tb testing.TB, isBuiltinFunc bool) (*contextmock.VMHostMock, *worldmock.MockWorld) {
	testVMType := []byte("type")

	gasSchedule := config.MakeGasMapForTests()
	gasCostConfig, err := config.CreateGasConfig(gasSchedule)
	require.Nil(tb, err)
	wasmerExecutor, _ := wasmer2.CreateExecutor()
	wasmerExecutor.SetOpcodeCosts(gasCostConfig.WASMOpcodeCost)

	host := &contextmock.VMHostMock{
		EnableEpochsHandlerField: &worldmock.EnableEpochsHandlerStub{},
	}

	mockMetering := &contextmock.MeteringContextMock{GasLeftMock: 10000}
	mockMetering.SetGasSchedule(gasSchedule)
	host.MeteringContext = mockMetering

	world := worldmock.NewMockWorld()
	host.BlockchainContext, err = NewBlockchainContext(host, world)
	require.Nil(tb, err)

	mockWasmerInstance = contextmock.NewInstanceMock(nil)
	execFactory := testexecutor.NewDefaultTestExecutorFactory(tb)
	exec, err := execFactory.CreateExecutor(executor.ExecutorFactoryArgs{
		VMHooks:     vmhooks.NewVMHooksImpl(host),
		OpcodeCosts: gasCostConfig.WASMOpcodeCost,
	})
	require.Nil(tb, err)
	runtimeCtx, err := NewRuntimeContext(
		host,
		testVMType,
		builtInFunctions.NewBuiltInFunctionContainer(),
		exec,
		defaultHasher,
	)
	require.Nil(tb, err)

	runtimeCtx.iTracker.instance = mockWasmerInstance
	host.RuntimeContext = runtimeCtx

	storageCtx, err := NewStorageContext(host, world, reservedTestPrefix)
	require.Nil(tb, err)
	host.StorageContext = storageCtx

	host.OutputContext, _ = NewOutputContext(host)
	host.CryptoHook, _ = factory.NewVMCrypto()
	host.StorageContext, _ = NewStorageContext(host, world, reservedTestPrefix)
	host.EnableEpochsHandlerField = worldmock.EnableEpochsHandlerStubNoFlags()
	host.IsBuiltinFunc = isBuiltinFunc

	return host, world
}

func initializeVMAndWasmerAsyncContextWithAliceAndBob(tb testing.TB) (
	*contextmock.VMHostMock,
	*worldmock.MockWorld,
	*vmcommon.ContractCallInput,
) {
	return initializeVMAndWasmerAsyncContextWithAliceAndBobWithBuiltIn(tb, false)
}

func initializeVMAndWasmerAsyncContextWithAliceAndBobWithBuiltIn(tb testing.TB, isBuiltIn bool) (
	*contextmock.VMHostMock,
	*worldmock.MockWorld,
	*vmcommon.ContractCallInput,
) {
	host, world := initializeVMAndWasmerAsyncContextWithBuiltIn(tb, isBuiltIn)
	world.AcctMap.PutAccount(&worldmock.Account{
		Address: Alice,
		Balance: big.NewInt(88),
	})
	world.AcctMap.PutAccount(&worldmock.Account{
		Address: Bob,
		Balance: big.NewInt(12),
	})

	originalVMInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			OriginalCallerAddr: OriginalCaller,
			CallerAddr:         OriginalCaller,
			Arguments:          nil,
			CallType:           vm.DirectCall,
			GasPrice:           1,
			CurrentTxHash:      []byte("txhash"),
			PrevTxHash:         []byte("txhash"),
			OriginalTxHash:     []byte("txhash"),
		},
		RecipientAddr: Alice,
		Function:      "alice_function",
	}

	return host, world, originalVMInput
}

func TestNewAsyncContext(t *testing.T) {
	host, _ := initializeVMAndWasmerAsyncContext(t)
	async := makeAsyncContext(t, host, nil)

	require.NotNil(t, async.host)
	require.Nil(t, async.stateStack)
	require.Nil(t, async.callerAddr)
	require.Nil(t, async.returnData)
	require.NotNil(t, async.asyncCallGroups)
	require.Empty(t, async.asyncCallGroups)
}

func TestAsyncContext_InitState(t *testing.T) {
	host, _ := initializeVMAndWasmerAsyncContext(t)
	async := makeAsyncContext(t, host, nil)

	async.callerAddr = []byte("some address")
	async.returnData = []byte("some return data")
	async.asyncCallGroups = nil

	async.InitState()

	require.NotNil(t, async.callerAddr)
	require.Empty(t, async.callerAddr)
	require.NotNil(t, async.returnData)
	require.Empty(t, async.returnData)
	require.NotNil(t, async.asyncCallGroups)
	require.Empty(t, async.asyncCallGroups)
}

func TestAsyncContext_InitStateFromContractCallInput(t *testing.T) {
	contract := []byte("contract")
	host, _ := initializeVMAndWasmerAsyncContext(t)
	async := makeAsyncContext(t, host, nil)

	async.callerAddr = []byte("some address")
	async.returnData = []byte("some return data")
	async.asyncCallGroups = nil

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: []byte("real caller addr"),
			GasPrice:   42,
		},
		RecipientAddr: contract,
	}

	host.Runtime().InitStateFromContractCallInput(input)
	err := async.InitStateFromInput(input)

	require.Nil(t, err)
	require.Equal(t, input.CallerAddr, async.callerAddr)
	require.NotNil(t, async.returnData)
	require.Empty(t, async.returnData)
	require.NotNil(t, async.asyncCallGroups)
	require.Empty(t, async.asyncCallGroups)
	require.Equal(t, contract, async.address)
}

func TestAsyncContext_GettersAndSetters(t *testing.T) {
	host, _ := initializeVMAndWasmerAsyncContext(t)
	async := makeAsyncContext(t, host, nil)

	async.callerAddr = []byte("some address")
	async.returnData = []byte("some return data")

	require.Equal(t, []byte("some address"), async.GetCallerAddress())
	require.Equal(t, []byte("some return data"), async.GetReturnData())

	async.SetReturnData([]byte("rockets"))
	require.Equal(t, []byte("rockets"), async.GetReturnData())
}

func TestAsyncContext_RegisterAsyncCall_NewGroup_DeleteGroup(t *testing.T) {
	host, _ := initializeVMAndWasmerAsyncContext(t)

	async := makeAsyncContext(t, host, nil)

	require.False(t, async.HasPendingCallGroups())

	group, exists := async.GetCallGroup("testGroup")
	require.Nil(t, group)
	require.False(t, exists)

	err := async.RegisterAsyncCall("testGroup", &vmhost.AsyncCall{
		Destination: []byte("somewhere"),
		Data:        []byte("something"),
	})
	require.Nil(t, err)
	require.True(t, async.HasPendingCallGroups())
	require.Equal(t, 1, len(async.asyncCallGroups))

	group, exists = async.GetCallGroup("testGroup")
	require.NotNil(t, group)
	require.True(t, exists)
	require.True(t, group.HasPendingCalls())
	require.False(t, group.HasCallback())

	async.deleteCallGroupByID("testGroup")
	group, exists = async.GetCallGroup("testGroup")
	require.Nil(t, group)
	require.False(t, exists)
}

func TestAsyncContext_RegisterAsyncCall_ExistingGroup(t *testing.T) {
	host, _ := initializeVMAndWasmerAsyncContext(t)

	async := makeAsyncContext(t, host, nil)
	require.Equal(t, 0, len(async.asyncCallGroups))

	err := async.AddCallGroup(vmhost.NewAsyncCallGroup("testGroup"))
	require.Nil(t, err)
	require.Equal(t, 1, len(async.asyncCallGroups))
	require.True(t, async.HasPendingCallGroups())

	group, exists := async.GetCallGroup("testGroup")
	require.NotNil(t, group)
	require.True(t, exists)

	err = async.RegisterAsyncCall("testGroup", &vmhost.AsyncCall{
		Destination: []byte("somewhere"),
		Data:        []byte("something"),
	})
	require.Nil(t, err)
	require.Equal(t, 1, len(async.asyncCallGroups))
	require.True(t, async.HasPendingCallGroups())
}

func TestAsyncContext_RegisterAsyncCall_ValidationAndFields(t *testing.T) {
	// TODO execution mode
	// TODO non-nil destination
	// TODO locked gas
}

func TestAsyncContext_DetermineExecutionMode(t *testing.T) {
	leftAddress := []byte("left")
	leftAccount := &worldmock.Account{
		Address: leftAddress,
		Code:    []byte("left code"),
		ShardID: 0,
	}

	rightAddress := []byte("right")
	rightAccount := &worldmock.Account{
		Address: rightAddress,
		Code:    []byte("right code"),
		ShardID: 0,
	}

	host, world := initializeVMAndWasmerAsyncContext(t)
	world.AcctMap.PutAccount(leftAccount)
	world.AcctMap.PutAccount(rightAccount)
	runtime := host.Runtime()

	async := makeAsyncContext(t, host, nil)

	initRuntime(runtime, leftAddress)
	asyncCall := &vmhost.AsyncCall{Destination: rightAddress, Data: []byte("func")}
	execMode, err := async.determineExecutionMode(asyncCall)
	require.Nil(t, err)
	require.Equal(t, vmhost.SyncExecution, execMode)

	host.IsBuiltinFunc = true
	initRuntime(runtime, leftAddress)
	asyncCall = &vmhost.AsyncCall{Destination: rightAddress, Data: []byte("func"), IsBuiltinFunctionCall: true}
	execMode, err = async.determineExecutionMode(asyncCall)
	require.Nil(t, err)
	require.Equal(t, vmhost.AsyncBuiltinFuncIntraShard, execMode)

	asyncCall = &vmhost.AsyncCall{Destination: rightAddress, Data: []byte("func")}
	host.IsBuiltinFunc = false
	rightAccount.Code = []byte{}
	rightAccount.ShardID = 1

	// Erase the code of the rightAccount from the Output context cache
	outputAccount, _ := host.Output().GetOutputAccount(rightAddress)
	outputAccount.Code = []byte{}

	initRuntime(runtime, leftAddress)
	execMode, err = async.determineExecutionMode(asyncCall)
	require.Nil(t, err)
	require.Equal(t, vmhost.AsyncUnknown, execMode)

	asyncCall = &vmhost.AsyncCall{Destination: rightAddress, Data: []byte("func"), IsBuiltinFunctionCall: true}
	host.IsBuiltinFunc = true
	rightAccount.Code = []byte{}
	rightAccount.ShardID = 1

	// Erase the code of the rightAccount from the Output context cache
	outputAccount, _ = host.Output().GetOutputAccount(rightAddress)
	outputAccount.Code = []byte{}

	initRuntime(runtime, leftAddress)
	execMode, err = async.determineExecutionMode(asyncCall)
	require.Nil(t, err)
	require.Equal(t, vmhost.AsyncBuiltinFuncCrossShard, execMode)
}

func initRuntime(runtime vmhost.RuntimeContext, address []byte) {
	runtime.InitStateFromContractCallInput(&vmcommon.ContractCallInput{
		RecipientAddr: address,
	})
}

func TestAsyncContext_IsValidCallbackName(t *testing.T) {
	host, _ := initializeVMAndWasmerAsyncContext(t)
	async := makeAsyncContext(t, host, nil)

	mockWasmerInstance.AddMockMethod("a", nil)
	mockWasmerInstance.AddMockMethod("my_contract_method_22", nil)
	mockWasmerInstance.AddMockMethod("not_builtin", nil)
	mockWasmerInstance.AddMockMethod("callBack", nil)
	mockWasmerInstance.AddMockMethod("callback", nil)
	mockWasmerInstance.AddMockMethod("function_do", nil)

	require.True(t, async.isValidCallbackName("a"))
	require.True(t, async.isValidCallbackName("my_contract_method_22"))

	require.True(t, async.isValidCallbackName("not_builtin"))
	host.IsBuiltinFunc = true
	require.False(t, async.isValidCallbackName("builtin"))
	host.IsBuiltinFunc = false

	require.True(t, async.isValidCallbackName("callBack"))
	require.True(t, async.isValidCallbackName("callback"))
	require.True(t, async.isValidCallbackName("function_do"))

	require.False(t, async.isValidCallbackName("function-do"))
	require.False(t, async.isValidCallbackName("3_my_contract_method_22"))
	require.False(t, async.isValidCallbackName("init"))
	require.False(t, async.isValidCallbackName("München"))
	require.False(t, async.isValidCallbackName("Göteborg"))
	require.False(t, async.isValidCallbackName("東京"))
	require.False(t, async.isValidCallbackName("function.org"))
	require.False(t, async.isValidCallbackName("Ainulindalë"))
}

func TestAsyncContext_UpdateCurrentCallStatus(t *testing.T) {
	contract := []byte("contract")

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: []byte("caller"),
			Arguments:  [][]byte{{0}},
			CallType:   vm.DirectCall,
		},
		RecipientAddr: contract,
	}

	host, _ := initializeVMAndWasmerAsyncContext(t)
	async := makeAsyncContext(t, host, contract)

	storedAsync := &asyncContext{}
	storedAsync.host = host
	err := storedAsync.Save()
	require.Nil(t, err)

	// CallType == DirectCall, async.UpdateCurrentCallStatus() does nothing
	host.Runtime().InitStateFromContractCallInput(vmInput)
	asyncCall, isLegacy, err := async.UpdateCurrentAsyncCallStatus(contract, []byte{}, &vmInput.VMInput)
	require.Nil(t, asyncCall)
	require.False(t, isLegacy)
	require.Nil(t, err)

	// CallType == AsynchronousCall, async.UpdateCurrentCallStatus() does nothing
	vmInput.CallType = vm.AsynchronousCall
	host.Runtime().InitStateFromContractCallInput(vmInput)
	asyncCall, isLegacy, err = async.UpdateCurrentAsyncCallStatus(contract, []byte{}, &vmInput.VMInput)
	require.Nil(t, asyncCall)
	require.False(t, isLegacy)
	require.Nil(t, err)

	// CallType == AsynchronousCallback, but no AsyncCalls registered in the
	// AsyncContext, so async.UpdateCurrentCallStatus() returns an error
	vmInput.CallType = vm.AsynchronousCallBack
	vmInput.Arguments = nil
	host.Runtime().InitStateFromContractCallInput(vmInput)
	asyncCall, isLegacy, err = async.UpdateCurrentAsyncCallStatus(contract, []byte{}, &vmInput.VMInput)
	require.Nil(t, asyncCall)
	require.False(t, isLegacy)
	require.True(t, errors.Is(err, vmhost.ErrCannotInterpretCallbackArgs))

	// CallType == AsynchronousCallback, but no AsyncCalls registered in the
	// AsyncContext, so async.UpdateCurrentCallStatus() returns an error
	vmInput.CallType = vm.AsynchronousCallBack
	vmInput.Arguments = [][]byte{{0}}
	host.Runtime().InitStateFromContractCallInput(vmInput)
	_, _, err = async.UpdateCurrentAsyncCallStatus(contract, []byte{}, &vmInput.VMInput)
	require.NotNil(t, err)

	// CallType == AsynchronousCallback, and there is an AsyncCall registered,
	// but it's not the expected one.
	err = async.RegisterAsyncCall("testGroup", &vmhost.AsyncCall{
		CallID:      []byte("callID_1"),
		Destination: []byte("some_address"),
		Data:        []byte("function"),
	})
	require.Nil(t, err)
	err = async.Save()
	require.Nil(t, err)

	vmInput.CallType = vm.AsynchronousCallBack
	vmInput.Arguments = [][]byte{{0}}
	host.Runtime().InitStateFromContractCallInput(vmInput)
	_, _, err = async.UpdateCurrentAsyncCallStatus(contract, []byte("callID_2"), &vmInput.VMInput)
	require.NotNil(t, err)

	// CallType == AsynchronousCallback, but this time there is a corresponding AsyncCall
	// registered, causing async.UpdateCurrentCallStatus() to find and update the AsyncCall
	err = async.RegisterAsyncCall("testGroup", &vmhost.AsyncCall{
		Destination: vmInput.CallerAddr,
		Data:        []byte("function"),
	})
	require.Nil(t, err)

	asyncCtx := &asyncContext{
		asyncCallGroups: []*vmhost.AsyncCallGroup{
			{
				Identifier: "",
				AsyncCalls: []*vmhost.AsyncCall{
					{
						Destination: []byte("caller"),
					},
				},
			},
		},
	}
	asyncCtx.host = host
	err = asyncCtx.Save()
	require.Nil(t, err)

	vmInput.CallType = vm.AsynchronousCallBack
	vmInput.Arguments = [][]byte{{0}}
	host.Runtime().InitStateFromContractCallInput(vmInput)
	asyncCall, isLegacy, err = async.UpdateCurrentAsyncCallStatus(contract, []byte{}, &vmInput.VMInput)
	require.Nil(t, err)
	require.False(t, isLegacy)
	require.NotNil(t, asyncCall)
	require.Equal(t, vmhost.AsyncCallResolved, asyncCall.Status)

	// CallType == AsynchronousCallback, there is a corresponding AsyncCall
	// registered, causing async.UpdateCurrentCallStatus() to find and update the
	// AsyncCall, but with AsyncCallRejected
	vmInput.CallType = vm.AsynchronousCallBack
	vmInput.Arguments = [][]byte{{1}}
	host.Runtime().InitStateFromContractCallInput(vmInput)
	asyncCall, isLegacy, err = async.UpdateCurrentAsyncCallStatus(contract, []byte{}, &vmInput.VMInput)
	require.Nil(t, err)
	require.False(t, isLegacy)
	require.NotNil(t, asyncCall)
	require.Equal(t, vmhost.AsyncCallRejected, asyncCall.Status)
}

func TestAsyncContext_OutputInCaseOfErrorInCallback(t *testing.T) {
	user := []byte("user")
	contractA := []byte("contractA")
	contractB := []byte("contractB")

	host, _ := initializeVMAndWasmerAsyncContext(t)
	host.EnableEpochsHandlerField = &worldmock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == vmhost.AsyncV3Flag
		},
	}

	async := makeAsyncContext(t, host, contractA)
	host.Storage().SetAddress(contractA)
	host.AsyncContext = async

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: user,
			Arguments:  [][]byte{{0}},
			CallType:   vm.DirectCall,
		},
		RecipientAddr: contractA,
	}
	host.Runtime().InitStateFromContractCallInput(vmInput)

	err := async.RegisterAsyncCall("", &vmhost.AsyncCall{
		Destination: contractB,
		Data:        []byte("function"),
	})
	require.Nil(t, err)

	err = async.Save()
	require.Nil(t, err)

	asyncCallId := async.GetCallID()
	asyncStoragePrefix := host.Storage().GetVmProtectedPrefix(vmhost.AsyncDataPrefix)
	asyncCallKey := vmhost.CustomStorageKey(string(asyncStoragePrefix), asyncCallId)

	data, _, _, _ := host.Storage().GetStorageUnmetered(asyncCallKey)
	require.NotEqual(t, len(data), 0)

	vmInput = &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: contractB,
			Arguments:  [][]byte{{0}},
			CallType:   vm.AsynchronousCallBack,
		},
		RecipientAddr: contractA,
	}
	host.Runtime().InitStateFromContractCallInput(vmInput)

	async.callbackAsyncInitiatorCallID = asyncCallId
	async.callType = vmInput.CallType
	err = async.LoadParentContext()
	require.Nil(t, err)

	vmOutput := host.Output().CreateVMOutputInCaseOfError(vmhost.ErrNotEnoughGas)
	outputAccount := vmOutput.OutputAccounts[string(contractA)]

	require.NotNil(t, outputAccount)

	storageUpdates := outputAccount.StorageUpdates
	require.Equal(t, len(storageUpdates), 1)

	asyncContextDeletionUpdate := storageUpdates[string(asyncCallKey)]
	require.NotNil(t, asyncContextDeletionUpdate)
	require.Equal(t, len(asyncContextDeletionUpdate.Data), 0)

	data, _, _, _ = host.Storage().GetStorageUnmetered(asyncCallKey)
	require.Equal(t, len(data), 0)
}

func TestAsyncContext_SendAsyncCallCrossShard(t *testing.T) {
	host, world := initializeVMAndWasmerAsyncContext(t)
	world.AcctMap.PutAccount(&worldmock.Account{
		Address: []byte("smartcontract"),
		Balance: big.NewInt(88),
	})

	initRuntime(host.Runtime(), []byte("smartcontract"))
	async := makeAsyncContext(t, host, nil)

	asyncCall := &vmhost.AsyncCall{
		Destination: []byte("destination"),
		GasLimit:    42,
		GasLocked:   98,
		ValueBytes:  big.NewInt(88).Bytes(),
		Data:        []byte("some_data"),
	}

	host.Runtime().GetVMInput().GasProvided = 200

	err := async.sendAsyncCallCrossShard(asyncCall)
	require.Nil(t, err)

	mockMetering := host.Metering().(*contextmock.MeteringContextMock)
	mockMetering.GasProvidedMock = 200
	mockMetering.GasLeftMock = 60

	vmOutput := host.Output().GetVMOutput()
	require.NotNil(t, vmOutput)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)

	smartcontract, ok := vmOutput.OutputAccounts["smartcontract"]
	require.True(t, ok)
	require.Equal(t, big.NewInt(-88), smartcontract.BalanceDelta)
	require.Empty(t, smartcontract.OutputTransfers)

	destination, ok := vmOutput.OutputAccounts["destination"]
	require.True(t, ok)
	require.Equal(t, big.NewInt(88), destination.BalanceDelta)
	require.Len(t, destination.OutputTransfers, 1)

	asyncTransfer := destination.OutputTransfers[0]
	require.Equal(t, big.NewInt(88), asyncTransfer.Value)
	require.Equal(t, uint64(42), asyncTransfer.GasLimit)
	require.Equal(t, uint64(98), asyncTransfer.GasLocked)

	callParser := parsers.NewCallArgsParser()
	function, _, _ := callParser.ParseData(string(asyncTransfer.Data))
	require.Equal(t, "some_data", function)
	require.Equal(t, vm.AsynchronousCall, asyncTransfer.CallType)
}

func TestAsyncContext_ExecuteSyncCall_EarlyOutOfGas(t *testing.T) {
	// Scenario 1
	// Assert error propagation in async.executeSyncCall() from
	// async.createContractCallInput()
	host, _, originalVMInput := initializeVMAndWasmerAsyncContextWithAliceAndBob(t)
	host.Runtime().InitStateFromContractCallInput(originalVMInput)
	async := makeAsyncContext(t, host, nil)

	asyncCall := defaultAsyncCallAliceToBob()
	asyncCall.Data = []byte("function")
	asyncCall.GasLimit = 1
	err := async.executeAsyncLocalCall(asyncCall)
	require.True(t, errors.Is(err, vmhost.ErrNotEnoughGas))
}

func TestAsyncContext_ExecuteSyncCall_Successful(t *testing.T) {
	// Scenario 3
	// Successful execution at destination, and successful callback execution;
	// the AsyncCall contains sufficient gas this time.
	host, _, originalVMInput := initializeVMAndWasmerAsyncContextWithAliceAndBob(t)
	host.Runtime().InitStateFromContractCallInput(originalVMInput)

	mockWasmerInstance.AddMockMethod("successCallback", nil)
	mockWasmerInstance.AddMockMethod("errorCallback", nil)

	async := makeAsyncContext(t, host, Alice)

	asyncCall := defaultAsyncCallAliceToBob()
	asyncCall.GasLimit = 200
	asyncCall.GasLocked = 90
	gasConsumedByDestination := uint64(23)
	gasConsumedByCallback := uint64(22)

	// The expected input passed to host.ExecuteOnDestContext() to call Bob as destination
	destInput := defaultCallInputAliceToBob(originalVMInput)
	destInput.GasProvided = asyncCall.GasLimit
	destInput.GasLocked = asyncCall.GasLocked

	// Prepare the output of Bob (the destination call)
	destOutput := defaultDestOutputOk()
	destOutput.GasRemaining = destInput.GasProvided - gasConsumedByDestination

	// Prepare the input to Alice's callback
	callbackInput := defaultCallbackInputBobToAlice(originalVMInput)
	callbackInput.GasProvided = destOutput.GasRemaining + asyncCall.GasLocked
	callbackInput.GasProvided -= defaultOutputDataLengthAsArgs(asyncCall, destOutput)
	callbackInput.GasProvided -= GasForAsyncStep
	callbackInput.GasLocked = 0

	// Prepare the output of Alice's callback
	callbackOutput := defaultCallbackOutputOk()
	callbackOutput.GasRemaining = callbackInput.GasProvided - gasConsumedByCallback

	// Enqueue the prepared VMOutputs
	host.EnqueueVMOutput(destOutput)
	host.EnqueueVMOutput(callbackOutput)

	host.Metering().RestoreGas(10000)

	err := async.RegisterAsyncCall("test", asyncCall)
	require.Nil(t, err)
	err = async.executeAsyncLocalCall(asyncCall)
	require.Nil(t, err)
	require.Equal(t, vmhost.AsyncCallResolved, asyncCall.Status)

	// There were two calls to host.ExecuteOnDestContext()
	require.Len(t, host.StoredInputs, 2)

	host.StoredInputs[0].AsyncArguments = nil
	require.Equal(t, destInput, host.StoredInputs[0])
	host.StoredInputs[1].AsyncArguments = nil
	require.Equal(t, callbackInput, host.StoredInputs[1])

	// Verify the final output of the execution; GasRemaining is set to 0 because
	// the test uses a mocked host.ExecuteOnDestContext(), which does not know to
	// manipulate the state stack of the OutputContext, therefore VMOutputs are
	// not merged between executions.
	expectedOutput := vmhost.MakeEmptyVMOutput()
	expectedOutput.ReturnCode = vmcommon.Ok
	expectedOutput.GasRemaining = host.Metering().GasLeft()

	// The expectedOutput must also contain an OutputAccount corresponding to
	// Alice, because of a call to host.Output().GetOutputAccount() in
	// host.Output().GetVMOutput(), which creates and caches an empty account for
	// her.
	vmhost.AddNewOutputAccount(expectedOutput, Alice, 0, nil)

	host.Output().GetOutputAccount(Alice)
	actualOutput := host.Output().GetVMOutput()
	// Bob entry is retuned empty by the MockWorld.GetStorageData()
	delete(actualOutput.OutputAccounts, string(Bob))
	require.Equal(t, expectedOutput, actualOutput)
}

func TestAsyncContext_CreateContractCallInput(t *testing.T) {
	host, _, originalVMInput := initializeVMAndWasmerAsyncContextWithAliceAndBob(t)
	host.Runtime().InitStateFromContractCallInput(originalVMInput)
	async := makeAsyncContext(t, host, nil)
	asyncCall := &vmhost.AsyncCall{
		Destination: Bob,
		ValueBytes:  big.NewInt(88).Bytes(),
	}

	asyncCall.Data = []byte{}
	input, err := async.createContractCallInput(asyncCall)
	require.Nil(t, input)
	require.Error(t, err)

	asyncCall.Data = []byte("function")
	asyncCall.GasLimit = 1
	input, err = async.createContractCallInput(asyncCall)
	require.Nil(t, input)
	require.True(t, errors.Is(err, vmhost.ErrNotEnoughGas))

	asyncCall.Data = []byte("function@0A0B0C@03")
	asyncCall.GasLimit = 2
	input, err = async.createContractCallInput(asyncCall)
	require.Nil(t, err)
	require.NotNil(t, input)

	expectedInput := defaultCallInputAliceToBob(originalVMInput)
	expectedInput.GasProvided = 2
	input.AsyncArguments = nil
	require.Equal(t, expectedInput, input)
}

func TestAsyncContext_CreateCallbackInput_DestinationCallSuccessful(t *testing.T) {
	host, _, originalVMInput := initializeVMAndWasmerAsyncContextWithAliceAndBob(t)
	host.Runtime().InitStateFromContractCallInput(originalVMInput)
	async := makeAsyncContext(t, host, Alice)

	asyncCall := defaultAsyncCallAliceToBob()
	asyncCall.Status = vmhost.AsyncCallResolved
	asyncCall.GasLocked = 300

	vmOutput := defaultDestOutputOk()
	vmOutput.GasRemaining = 12

	destinationErr := error(nil)
	callbackInput, err := async.createCallbackInput(asyncCall, vmOutput, 0, destinationErr)
	require.Nil(t, err)

	expectedGasProvided := asyncCall.GasLocked + vmOutput.GasRemaining
	expectedGasProvided -= defaultOutputDataLengthAsArgs(asyncCall, vmOutput)
	expectedGasProvided -= host.Metering().GasSchedule().BaseOpsAPICost.AsyncCallStep

	expectedInput := defaultCallbackInputBobToAlice(originalVMInput)
	expectedInput.GasProvided = expectedGasProvided
	callbackInput.AsyncArguments = nil
	require.Equal(t, expectedInput, callbackInput)
}

func TestAsyncContext_CreateCallbackInput_LastTransfer_ValueAndNoESDTTransfer_WithReturnData(t *testing.T) {

	lastTransfer := &vmcommon.OutputTransfer{
		Value: big.NewInt(2),
	}
	vmOutput := createDefaultVMOutput(lastTransfer)

	originalVMInput, callbackInput, expectedGasProvided := createCallbackInputWithVMOutput(t,
		vmOutput,
		lastTransfer)

	expectedInput := defaultCallbackInputBobToAlice(originalVMInput)
	expectedInput.GasProvided = expectedGasProvided
	expectedInput.CallValue = big.NewInt(0)
	expectedInput.ESDTTransfers = nil
	callbackInput.AsyncArguments = nil
	require.Equal(t, expectedInput, callbackInput)
}

func TestAsyncContext_CreateCallbackInput_LastTransfer_ValueAndNoESDTTransfer_NoReturnData(t *testing.T) {

	lastTransfer := &vmcommon.OutputTransfer{
		Value: big.NewInt(2),
	}
	vmOutput := createDefaultVMOutput(lastTransfer)
	vmOutput.ReturnData = nil

	originalVMInput, callbackInput, expectedGasProvided := createCallbackInputWithVMOutput(t,
		vmOutput,
		lastTransfer)

	expectedInput := defaultCallbackInputBobToAlice(originalVMInput)
	expectedInput.Arguments = [][]byte{{0}}
	expectedInput.GasProvided = expectedGasProvided
	expectedInput.CallValue = big.NewInt(2)
	expectedInput.ESDTTransfers = nil
	callbackInput.AsyncArguments = nil
	require.Equal(t, expectedInput, callbackInput)
}

func TestAsyncContext_CreateCallbackInput_LastTransferValueAndESDTTransfer_WithReturnData(t *testing.T) {
	originalVMInput, callbackInput, expectedGasProvided := createCallbackInput(t,
		&vmcommon.OutputTransfer{
			Value: big.NewInt(2),
			Data:  []byte("ESDTTransfer@6d696975746f6b656e@05"),
		})

	expectedInput := defaultCallbackInputBobToAlice(originalVMInput)
	expectedInput.GasProvided = expectedGasProvided
	expectedInput.CallValue = big.NewInt(0)
	expectedInput.ESDTTransfers = nil
	callbackInput.AsyncArguments = nil
	require.Equal(t, expectedInput, callbackInput)
}

func TestAsyncContext_CreateCallbackInput_LastTransferValueAndESDTTransfer_NoReturnData(t *testing.T) {

	lastTransfer := &vmcommon.OutputTransfer{
		Value: big.NewInt(2),
		Data:  []byte("ESDTTransfer@6d696975746f6b656e@05"),
	}
	vmOutput := createDefaultVMOutput(lastTransfer)
	vmOutput.ReturnData = nil

	originalVMInput, callbackInput, expectedGasProvided := createCallbackInputWithVMOutput(t,
		vmOutput,
		lastTransfer)

	expectedInput := defaultCallbackInputBobToAlice(originalVMInput)
	expectedInput.Arguments = [][]byte{{0}}
	expectedInput.GasProvided = expectedGasProvided
	expectedInput.CallValue = big.NewInt(2)
	expectedInput.ESDTTransfers = []*vmcommon.ESDTTransfer{
		{
			ESDTTokenName: []byte("miiutoken"),
			ESDTValue:     big.NewInt(5),
		},
	}
	callbackInput.AsyncArguments = nil
	require.Equal(t, expectedInput, callbackInput)
}

func TestAsyncContext_CreateCallbackInput_LastTransferValueAndESDTTransferWithCall(t *testing.T) {
	originalVMInput, callbackInput, expectedGasProvided := createCallbackInput(t,
		&vmcommon.OutputTransfer{
			Value: big.NewInt(2),
			Data:  []byte("ESDTTransfer@6d696975746f6b656e@05@6d696975"),
		})

	expectedInput := defaultCallbackInputBobToAlice(originalVMInput)
	expectedInput.GasProvided = expectedGasProvided
	expectedInput.CallValue = big.NewInt(0)
	expectedInput.ESDTTransfers = nil
	callbackInput.AsyncArguments = nil
	require.Equal(t, expectedInput, callbackInput)
}

func createDefaultVMOutput(lastTransfer *vmcommon.OutputTransfer) *vmcommon.VMOutput {
	vmOutput := defaultDestOutputOk()
	vmOutput.OutputAccounts = make(map[string]*vmcommon.OutputAccount)
	vmOutput.OutputAccounts[string(Alice)] = &vmcommon.OutputAccount{
		OutputTransfers: []vmcommon.OutputTransfer{
			{
				Value: big.NewInt(1),
			},
			*lastTransfer,
		},
	}
	vmOutput.GasRemaining = 12
	return vmOutput
}

func createCallbackInput(t *testing.T, lastTransfer *vmcommon.OutputTransfer) (*vmcommon.ContractCallInput, *vmcommon.ContractCallInput, uint64) {
	vmOutput := createDefaultVMOutput(lastTransfer)
	return createCallbackInputWithVMOutput(t, vmOutput, lastTransfer)
}

func createCallbackInputWithVMOutput(t *testing.T, vmOutput *vmcommon.VMOutput, _ *vmcommon.OutputTransfer) (*vmcommon.ContractCallInput, *vmcommon.ContractCallInput, uint64) {
	host, _, originalVMInput := initializeVMAndWasmerAsyncContextWithAliceAndBobWithBuiltIn(t, true)
	host.Runtime().InitStateFromContractCallInput(originalVMInput)
	async := makeAsyncContext(t, host, Alice)

	asyncCall := defaultAsyncCallAliceToBob()
	asyncCall.Status = vmhost.AsyncCallResolved
	asyncCall.GasLocked = 300

	destinationErr := error(nil)
	callbackInput, err := async.createCallbackInput(asyncCall, vmOutput, 0, destinationErr)
	require.Nil(t, err)

	expectedGasProvided := asyncCall.GasLocked + vmOutput.GasRemaining
	expectedGasProvided -= defaultOutputDataLengthAsArgs(asyncCall, vmOutput)
	expectedGasProvided -= host.Metering().GasSchedule().BaseOpsAPICost.AsyncCallStep
	return originalVMInput, callbackInput, expectedGasProvided
}

func TestAsyncContext_CreateCallbackInput_DestinationCallFailed(t *testing.T) {
	host, _, originalVMInput := initializeVMAndWasmerAsyncContextWithAliceAndBob(t)
	host.Runtime().InitStateFromContractCallInput(originalVMInput)
	async := makeAsyncContext(t, host, Alice)

	asyncCall := defaultAsyncCallAliceToBob()
	asyncCall.Status = vmhost.AsyncCallRejected
	asyncCall.GasLocked = 200

	vmOutput := defaultDestOutputUserError()
	destinationErr := vmhost.ErrSignalError
	callbackInput, err := async.createCallbackInput(asyncCall, vmOutput, 0, destinationErr)
	require.Nil(t, err)

	expectedGasProvided := asyncCall.GasLocked + vmOutput.GasRemaining
	expectedGasProvided -= defaultOutputDataLengthAsArgs(asyncCall, vmOutput)
	expectedGasProvided -= host.Metering().GasSchedule().BaseOpsAPICost.AsyncCallStep

	expectedInput := vmhost.MakeContractCallInput(Bob, Alice, "errorCallback", 0)
	vmhost.AddArgument(expectedInput, []byte{byte(vmcommon.UserError)})
	vmhost.AddArgument(expectedInput, []byte(vmOutput.ReturnMessage))
	vmhost.CopyTxHashes(expectedInput, originalVMInput)
	expectedInput.GasProvided = expectedGasProvided
	expectedInput.CallType = vm.AsynchronousCallBack
	expectedInput.ReturnCallAfterError = true
	callbackInput.AsyncArguments = nil
	require.Equal(t, expectedInput, callbackInput)
}

func TestAsyncContext_CreateCallbackInput_NotEnoughGas(t *testing.T) {
	// Due to dynamic gas locking, this situation should never happen
	host, _, originalVMInput := initializeVMAndWasmerAsyncContextWithAliceAndBob(t)
	host.Runtime().InitStateFromContractCallInput(originalVMInput)
	async := makeAsyncContext(t, host, nil)

	asyncCall := defaultAsyncCallAliceToBob()
	asyncCall.Status = vmhost.AsyncCallRejected

	vmOutput := &vmcommon.VMOutput{
		ReturnCode:    vmcommon.UserError,
		ReturnData:    [][]byte{},
		ReturnMessage: "there was a user error",
		GasRemaining:  0,
	}

	destinationErr := vmhost.ErrSignalError
	callbackInput, err := async.createCallbackInput(asyncCall, vmOutput, 0, destinationErr)
	require.Nil(t, callbackInput)
	require.True(t, errors.Is(err, vmhost.ErrNotEnoughGas))
}

func TestAsyncContext_FinishSyncExecution_NilError_NilVMOutput(t *testing.T) {
	host, _, originalVMInput := initializeVMAndWasmerAsyncContextWithAliceAndBob(t)
	host.Runtime().InitStateFromContractCallInput(originalVMInput)
	async := makeAsyncContext(t, host, nil)
	async.finishAsyncLocalCallbackExecution()

	// The expectedOutput must also contain an OutputAccount corresponding to
	// Alice, because of a call to host.Output().GetOutputAccount() in
	// host.Output().GetVMOutput(), which creates and caches an empty account for
	// her.
	expectedOutput := vmhost.MakeEmptyVMOutput()
	expectedOutput.GasRemaining = host.Metering().GasLeft()
	vmhost.AddNewOutputAccount(expectedOutput, Alice, 0, nil)

	host.Output().GetOutputAccount(Alice)
	require.Equal(t, expectedOutput, host.Output().GetVMOutput())
}

func TestAsyncContext_FinishSyncExecution_Error_NilVMOutput(t *testing.T) {
	host, _, originalVMInput := initializeVMAndWasmerAsyncContextWithAliceAndBob(t)
	host.Runtime().InitStateFromContractCallInput(originalVMInput)
	async := makeAsyncContext(t, host, nil)

	async.finishAsyncLocalCallbackExecution()

	expectedOutput := vmhost.MakeEmptyVMOutput()
	expectedOutput.GasRemaining = host.Metering().GasLeft()
	// expectedOutput.ReturnCode = vmcommon.OutOfGas
	// expectedOutput.ReturnMessage = syncExecErr.Error()
	// vmhost.AddFinishData(expectedOutput, []byte(vmcommon.OutOfGas.String()))
	// vmhost.AddFinishData(expectedOutput, originalVMInput.CurrentTxHash)

	// The expectedOutput must also contain an OutputAccount corresponding to
	// Alice, because of a call to host.Output().GetOutputAccount() in
	// host.Output().GetVMOutput(), which creates and caches an empty account for
	// her.
	vmhost.AddNewOutputAccount(expectedOutput, Alice, 0, nil)

	host.Output().GetOutputAccount(Alice)
	require.Equal(t, expectedOutput, host.Output().GetVMOutput())
}

func TestAsyncContext_FinishSyncExecution_ErrorAndVMOutput(t *testing.T) {
	host, _, originalVMInput := initializeVMAndWasmerAsyncContextWithAliceAndBob(t)
	host.Runtime().InitStateFromContractCallInput(originalVMInput)
	async := makeAsyncContext(t, host, nil)

	syncExecOutput := vmhost.MakeEmptyVMOutput()
	syncExecOutput.ReturnCode = vmcommon.UserError
	syncExecOutput.ReturnMessage = "user made an error"
	async.finishAsyncLocalCallbackExecution()

	expectedOutput := vmhost.MakeEmptyVMOutput()
	expectedOutput.GasRemaining = host.Metering().GasLeft()
	// expectedOutput.ReturnCode = vmcommon.UserError
	// expectedOutput.ReturnMessage = "user made an error"
	// vmhost.AddFinishData(expectedOutput, []byte(vmcommon.UserError.String()))
	// vmhost.AddFinishData(expectedOutput, originalVMInput.CurrentTxHash)

	// The expectedOutput must also contain an OutputAccount corresponding to
	// Alice, because of a call to host.Output().GetOutputAccount() in
	// host.Output().GetVMOutput(), which creates and caches an empty account for
	// her.
	vmhost.AddNewOutputAccount(expectedOutput, Alice, 0, nil)

	host.Output().GetOutputAccount(Alice)
	require.Equal(t, expectedOutput, host.Output().GetVMOutput())
}

func defaultAsyncCallAliceToBob() *vmhost.AsyncCall {
	return &vmhost.AsyncCall{
		Destination:     Bob,
		Data:            []byte("function@0A0B0C@03"),
		GasLimit:        0,
		GasLocked:       0,
		ValueBytes:      big.NewInt(88).Bytes(),
		SuccessCallback: "successCallback",
		ErrorCallback:   "errorCallback",
		Status:          vmhost.AsyncCallPending,
	}
}

func defaultCallInputAliceToBob(originalVMInput *vmcommon.ContractCallInput) *vmcommon.ContractCallInput {
	destInput := vmhost.MakeContractCallInput(Alice, Bob, "function", 88)
	vmhost.CopyTxHashes(destInput, originalVMInput)
	vmhost.AddArgument(destInput, []byte{10, 11, 12})
	vmhost.AddArgument(destInput, []byte{3})
	destInput.CallType = vm.AsynchronousCall

	return destInput
}

func defaultDestOutputUserError() *vmcommon.VMOutput {
	return &vmcommon.VMOutput{
		ReturnCode:    vmcommon.UserError,
		ReturnData:    [][]byte{},
		ReturnMessage: "user error occurred",
		GasRemaining:  0,
	}
}

func defaultDestOutputOk() *vmcommon.VMOutput {
	return &vmcommon.VMOutput{
		ReturnCode: vmcommon.Ok,
		ReturnData: [][]byte{
			[]byte("first"),
			[]byte("second"),
			{},
			[]byte("third"),
		},
		ReturnMessage: "a message",
		GasRemaining:  0,
	}
}

func defaultCallbackInputBobToAlice(originalVMInput *vmcommon.ContractCallInput) *vmcommon.ContractCallInput {
	input := vmhost.MakeContractCallInput(Bob, Alice, "successCallback", 0)
	vmhost.AddArgument(input, []byte{0}) // vmcommon.Ok
	vmhost.AddArgument(input, []byte("first"))
	vmhost.AddArgument(input, []byte("second"))
	vmhost.AddArgument(input, []byte{})
	vmhost.AddArgument(input, []byte("third"))
	vmhost.CopyTxHashes(input, originalVMInput)
	input.CallType = vm.AsynchronousCallBack
	return input
}

func defaultCallbackOutputOk() *vmcommon.VMOutput {
	vmOutput := vmhost.MakeEmptyVMOutput()
	vmhost.AddFinishData(vmOutput, []byte("cbFirst"))
	vmhost.AddFinishData(vmOutput, []byte("cbSecond"))

	return vmOutput
}

func defaultOutputDataLengthAsArgs(asyncCall *vmhost.AsyncCall, vmOutput *vmcommon.VMOutput) uint64 {
	separator := 1
	hexSize := 2
	returnCode := 1 * hexSize

	dataLength := 0
	if vmOutput.ReturnCode == vmcommon.Ok {
		dataLength += returnCode + len(asyncCall.SuccessCallback) + separator
		for _, data := range vmOutput.ReturnData {
			dataLength += separator
			dataLength += len(data) * hexSize
		}
	} else {
		dataLength += len(asyncCall.ErrorCallback) + separator + returnCode
		dataLength += separator + len(vmOutput.ReturnMessage)*hexSize
	}

	return uint64(dataLength)
}
