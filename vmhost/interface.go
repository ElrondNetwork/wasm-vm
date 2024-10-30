package vmhost

import (
	"crypto/elliptic"
	"io"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-go/config"
	"github.com/multiversx/mx-chain-vm-go/crypto"
	"github.com/multiversx/mx-chain-vm-go/executor"
)

// StateStack defines the functionality for working with a state stack
type StateStack interface {
	InitState()
	PushState()
	PopSetActiveState()
	PopDiscard()
	ClearStateStack()
}

// CallArgsParser defines the functionality to parse transaction data for a smart contract call
type CallArgsParser interface {
	ParseData(data string) (string, [][]byte, error)
	IsInterfaceNil() bool
}

// VMHost defines the functionality for working with the VM
type VMHost interface {
	vmcommon.VMExecutionHandler
	Crypto() crypto.VMCrypto
	Blockchain() BlockchainContext
	Runtime() RuntimeContext
	Async() AsyncContext
	ManagedTypes() ManagedTypesContext
	Output() OutputContext
	Metering() MeteringContext
	Storage() StorageContext
	EnableEpochsHandler() EnableEpochsHandler

	ExecuteESDTTransfer(transfersArgs *ESDTTransfersArgs, callType vm.CallType) (*vmcommon.VMOutput, uint64, error)
	CreateNewContract(input *vmcommon.ContractCreateInput, createContractCallType int) ([]byte, error)
	ExecuteOnSameContext(input *vmcommon.ContractCallInput) error
	ExecuteOnDestContext(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, bool, error)
	IsBuiltinFunctionName(functionName string) bool
	IsBuiltinFunctionCall(data []byte) bool
	AreInSameShard(leftAddress []byte, rightAddress []byte) bool
	IsAllowedToExecute(opcode string) bool

	GetGasScheduleMap() config.GasScheduleMap
	GetContexts() (ManagedTypesContext, BlockchainContext, MeteringContext, OutputContext, RuntimeContext, AsyncContext, StorageContext)
	SetRuntimeContext(runtime RuntimeContext)

	SetBuiltInFunctionsContainer(builtInFuncs vmcommon.BuiltInFunctionContainer)
	InitState()

	CompleteLogEntriesWithCallType(vmOutput *vmcommon.VMOutput, callType string)

	Reset()
	SetGasTracing(enableGasTracing bool)
	GetGasTrace() map[string]map[string][]uint64
}

// BlockchainContext defines the functionality needed for interacting with the blockchain context
type BlockchainContext interface {
	StateStack

	NewAddress(creatorAddress []byte) ([]byte, error)
	AccountExists(addr []byte) bool
	GetBalance(addr []byte) []byte
	GetBalanceBigInt(addr []byte) *big.Int
	GetNonce(addr []byte) (uint64, error)
	CurrentEpoch() uint32
	GetStateRootHash() []byte
	LastTimeStamp() uint64
	LastNonce() uint64
	LastRound() uint64
	LastEpoch() uint32
	CurrentRound() uint64
	CurrentNonce() uint64
	CurrentTimeStamp() uint64
	CurrentRandomSeed() []byte
	LastRandomSeed() []byte
	IncreaseNonce(addr []byte)
	GetCodeHash(addr []byte) []byte
	GetCode(addr []byte) ([]byte, error)
	GetCodeSize(addr []byte) (int32, error)
	BlockHash(number uint64) []byte
	GetOwnerAddress() ([]byte, error)
	GetShardOfAddress(addr []byte) uint32
	IsSmartContract(addr []byte) bool
	IsPayable(sndAddress, rcvAddress []byte) (bool, error)
	SaveCompiledCode(codeHash []byte, code []byte)
	GetCompiledCode(codeHash []byte) (bool, []byte)
	GetESDTToken(address []byte, tokenID []byte, nonce uint64) (*esdt.ESDigitalToken, error)
	IsLimitedTransfer(tokenID []byte) bool
	IsPaused(tokenID []byte) bool
	GetUserAccount(address []byte) (vmcommon.UserAccountHandler, error)
	ProcessBuiltInFunction(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
	GetSnapshot() int
	RevertToSnapshot(snapshot int)
	ClearCompiledCodes()
	ExecuteSmartContractCallOnOtherVM(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
}

// RuntimeContext defines the functionality needed for interacting with the runtime context
type RuntimeContext interface {
	StateStack

	GetVMExecutor() executor.Executor
	ReplaceVMExecutor(vmExecutor executor.Executor)
	InitStateFromContractCallInput(input *vmcommon.ContractCallInput)
	SetCustomCallFunction(callFunction string)
	GetVMInput() *vmcommon.ContractCallInput
	SetVMInput(vmInput *vmcommon.ContractCallInput)
	GetContextAddress() []byte
	GetOriginalCallerAddress() []byte
	SetCodeAddress(scAddress []byte)
	GetSCCode() ([]byte, error)
	GetSCCodeSize() uint64
	GetVMType() []byte
	FunctionName() string
	Arguments() [][]byte
	GetCurrentTxHash() []byte
	GetOriginalTxHash() []byte
	ExtractCodeUpgradeFromArgs() ([]byte, []byte, error)
	SignalUserError(message string)
	FailExecution(err error)
	MustVerifyNextContractCode()
	SetRuntimeBreakpointValue(value BreakpointValue)
	GetRuntimeBreakpointValue() BreakpointValue
	GetInstanceStackSize() uint64
	CountSameContractInstancesOnStack(address []byte) uint64
	IsFunctionImported(name string) bool
	ReadOnly() bool
	SetReadOnly(readOnly bool)
	StartWasmerInstance(contract []byte, gasLimit uint64, newCode bool) error
	ClearWarmInstanceCache()
	SetMaxInstanceStackSize(uint64)
	VerifyContractCode() error
	GetInstance() executor.Instance
	GetInstanceTracker() InstanceTracker
	FunctionNameChecked() (string, error)
	CallSCFunction(functionName string) error
	GetPointsUsed() uint64
	SetPointsUsed(gasPoints uint64)
	UseGasBoundedShouldFailExecution() bool
	CleanInstance()

	AddError(err error, otherInfo ...string)
	GetAllErrors() error

	ValidateCallbackName(callbackName string) error

	IsReservedFunctionName(functionName string) bool

	HasFunction(functionName string) bool
	GetPrevTxHash() []byte
	EndExecution()
	ValidateInstances() error
}

// InstanceTracker defines the functionality needed for interacting with the instance tracker
type InstanceTracker interface {
	StateStack

	TrackedInstances() map[string]executor.Instance
}

// ManagedTypesContext defines the functionality needed for interacting with the big int context
type ManagedTypesContext interface {
	StateStack

	GetRandReader() io.Reader
	ConsumeGasForThisBigIntNumberOfBytes(byteLen *big.Int) error
	ConsumeGasForThisIntNumberOfBytes(byteLen int) error
	ConsumeGasForBytes(bytes []byte) error
	ConsumeGasForBigIntCopy(values ...*big.Int) error
	ConsumeGasForBigFloatCopy(values ...*big.Float) error
	NewBigInt(value *big.Int) int32
	NewBigIntFromInt64(int64Value int64) int32
	GetBigIntOrCreate(handle int32) *big.Int
	GetBigInt(id int32) (*big.Int, error)
	GetTwoBigInt(handle1 int32, handle2 int32) (*big.Int, *big.Int, error)
	PutBigFloat(value *big.Float) (int32, error)
	BigFloatPrecIsNotValid(precision uint) bool
	BigFloatExpIsNotValid(exponent int) bool
	EncodedBigFloatIsNotValid(encodedBigFloat []byte) bool
	GetBigFloatOrCreate(handle int32) (*big.Float, error)
	GetBigFloat(handle int32) (*big.Float, error)
	GetTwoBigFloats(handle1 int32, handle2 int32) (*big.Float, *big.Float, error)
	PutEllipticCurve(ec *elliptic.CurveParams) int32
	GetEllipticCurve(handle int32) (*elliptic.CurveParams, error)
	GetEllipticCurveSizeOfField(ecHandle int32) int32
	Get100xCurveGasCostMultiplier(ecHandle int32) int32
	GetScalarMult100xCurveGasCostMultiplier(ecHandle int32) int32
	GetUCompressed100xCurveGasCostMultiplier(ecHandle int32) int32
	GetPrivateKeyByteLengthEC(ecHandle int32) int32
	NewManagedBuffer() int32
	NewManagedBufferFromBytes(bytes []byte) int32
	SetBytes(mBufferHandle int32, bytes []byte)
	GetBytes(mBufferHandle int32) ([]byte, error)
	AppendBytes(mBufferHandle int32, bytes []byte) bool
	GetLength(mBufferHandle int32) int32
	GetSlice(mBufferHandle int32, startPosition int32, lengthOfSlice int32) ([]byte, error)
	DeleteSlice(mBufferHandle int32, startPosition int32, lengthOfSlice int32) ([]byte, error)
	InsertSlice(mBufferHandle int32, startPosition int32, slice []byte) ([]byte, error)
	ReadManagedVecOfManagedBuffers(managedVecHandle int32) ([][]byte, uint64, error)
	WriteManagedVecOfManagedBuffers(data [][]byte, destinationHandle int32) error
	NewManagedMap() int32
	ManagedMapPut(mMapHandle int32, keyHandle int32, valueHandle int32) error
	ManagedMapGet(mMapHandle int32, keyHandle int32, outValueHandle int32) error
	ManagedMapRemove(mMapHandle int32, keyHandle int32, outValueHandle int32) error
	ManagedMapContains(mMapHandle int32, keyHandle int32) (bool, error)
	GetBackTransfers() ([]*vmcommon.ESDTTransfer, *big.Int)
	AddValueOnlyBackTransfer(value *big.Int)
	AddBackTransfers(transfers []*vmcommon.ESDTTransfer)
	PopBackTransferIfAsyncCallBack(vmInput *vmcommon.ContractCallInput)
}

// OutputContext defines the functionality needed for interacting with the output context
type OutputContext interface {
	StateStack
	PopMergeActiveState()
	CensorVMOutput()
	AddToActiveState(rightOutput *vmcommon.VMOutput)

	GetOutputAccount(address []byte) (*vmcommon.OutputAccount, bool)
	GetOutputAccounts() map[string]*vmcommon.OutputAccount
	DeleteOutputAccount(address []byte)
	WriteLog(address []byte, topics [][]byte, data [][]byte)
	WriteLogWithIdentifier(address []byte, topics [][]byte, data [][]byte, identifier []byte)
	TransferValueOnly(destination []byte, sender []byte, value *big.Int, checkPayable bool) error
	Transfer(destination []byte, sender []byte, gasLimit uint64, gasLocked uint64, value *big.Int, asyncData []byte, input []byte, callType vm.CallType) error
	TransferESDT(transfersArgs *ESDTTransfersArgs, callInput *vmcommon.ContractCallInput) (uint64, error)
	GetRefund() uint64
	SetRefund(refund uint64)
	ReturnCode() vmcommon.ReturnCode
	SetReturnCode(returnCode vmcommon.ReturnCode)
	ReturnMessage() string
	SetReturnMessage(message string)
	ReturnData() [][]byte
	ClearReturnData()
	RemoveReturnData(index uint32)
	Finish(data []byte)
	PrependFinish(data []byte)
	DeleteFirstReturnData()
	GetVMOutput() *vmcommon.VMOutput
	RemoveNonUpdatedStorage()
	AddTxValueToAccount(address []byte, value *big.Int)
	DeployCode(input CodeDeployInput)
	CreateVMOutputInCaseOfError(err error) *vmcommon.VMOutput
	NextOutputTransferIndex() uint32
	GetCrtTransferIndex() uint32
	SetCrtTransferIndex(index uint32)
	IsInterfaceNil() bool
}

// MeteringContext defines the functionality needed for interacting with the metering context
type MeteringContext interface {
	StateStack
	PopMergeActiveState()

	InitStateFromContractCallInput(input *vmcommon.VMInput)
	SetGasSchedule(gasMap config.GasScheduleMap)
	GasSchedule() *config.GasCost
	UseGasBoundedAndAddTracedGas(functionName string, gas uint64) error
	FreeGas(gas uint64)
	RestoreGas(gas uint64)
	GasLeft() uint64
	GasUsedForExecution() uint64
	GasSpentByContract() uint64
	GetGasForExecution() uint64
	GetGasProvided() uint64
	GetSCPrepareInitialCost() uint64
	BoundGasLimit(value int64) uint64
	BlockGasLimit() uint64
	DeductInitialGasForExecution(contract []byte) error
	DeductInitialGasForDirectDeployment(input CodeDeployInput) error
	DeductInitialGasForIndirectDeployment(input CodeDeployInput) error
	ComputeExtraGasLockedForAsync() uint64
	UseGasForAsyncStep() error
	UseGasBounded(gasToUse uint64) error
	UseGasForContractInit(gasToUse uint64)
	GetGasLocked() uint64
	UpdateGasStateOnSuccess(vmOutput *vmcommon.VMOutput) error
	UpdateGasStateOnFailure(vmOutput *vmcommon.VMOutput)
	TrackGasUsedByOutOfVMFunction(builtinInput *vmcommon.ContractCallInput, builtinOutput *vmcommon.VMOutput, postBuiltinInput *vmcommon.ContractCallInput)
	DisableRestoreGas()
	EnableRestoreGas()
	StartGasTracing(functionName string)
	SetGasTracing(enableGasTracing bool)
	GetGasTrace() map[string]map[string][]uint64
}

// StorageStatus defines the states the storage can be in
type StorageStatus int

const (
	// StorageUnchanged signals that the storage was not changed
	StorageUnchanged StorageStatus = iota

	// StorageModified signals that the storage has been modified
	StorageModified

	// StorageAdded signals that something was added to storage
	StorageAdded

	// StorageDeleted signals that something was removed from storage
	StorageDeleted
)

// StorageContext defines the functionality needed for interacting with the storage context
type StorageContext interface {
	StateStack

	SetAddress(address []byte)
	GetStorageUpdates(address []byte) map[string]*vmcommon.StorageUpdate
	GetStorageFromAddress(address []byte, key []byte) ([]byte, uint32, bool, error)
	GetStorageFromAddressNoChecks(address []byte, key []byte) ([]byte, uint32, bool, error)
	GetStorage(key []byte) ([]byte, uint32, bool, error)
	GetStorageUnmetered(key []byte) ([]byte, uint32, bool, error)
	SetStorage(key []byte, value []byte) (StorageStatus, error)
	SetProtectedStorage(key []byte, value []byte) (StorageStatus, error)
	SetProtectedStorageToAddress(address []byte, key []byte, value []byte) (StorageStatus, error)
	SetProtectedStorageToAddressUnmetered(address []byte, key []byte, value []byte) (StorageStatus, error)
	UseGasForStorageLoad(tracedFunctionName string, trieDepth int64, blockchainLoadCost uint64, usedCache bool) error
	GetVmProtectedPrefix(prefix string) []byte
}

// AsyncCallInfoHandler defines the functionality for working with AsyncCallInfo
type AsyncCallInfoHandler interface {
	GetDestination() []byte
	GetData() []byte
	GetGasLimit() uint64
	GetGasLocked() uint64
	GetValueBytes() []byte
}

// AsyncContext defines the functionality needed for interacting with the asynchronous execution context
type AsyncContext interface {
	StateStack

	InitStateFromInput(input *vmcommon.ContractCallInput) error
	HasPendingCallGroups() bool
	IsComplete() bool
	GetCallGroup(groupID string) (*AsyncCallGroup, bool)
	SetContextCallback(callbackName string, data []byte, gas uint64) error
	HasCallback() bool
	GetCallerAddress() []byte
	GetParentAddress() []byte
	GetCallerCallID() []byte
	GetReturnData() []byte
	SetReturnData(data []byte)

	Execute() error
	RegisterAsyncCall(groupID string, call *AsyncCall) error
	RegisterLegacyAsyncCall(address []byte, data []byte, value []byte) error

	LoadParentContext() error
	Save() error
	DeleteFromCallID(address []byte) error

	GetCallID() []byte
	GetCallbackAsyncInitiatorCallID() []byte
	IsCrossShard() bool

	Clone() AsyncContext

	UpdateCurrentAsyncCallStatus(
		address []byte,
		callID []byte,
		vmInput *vmcommon.VMInput) (*AsyncCall, bool, error)

	CompleteChildConditional(isChildComplete bool, callID []byte, gasToAccumulate uint64) error
	NotifyChildIsComplete(callID []byte, gasToAccumulate uint64) error

	SetResults(vmOutput *vmcommon.VMOutput)
	GetGasAccumulated() uint64

	SetAsyncArgumentsForCall(input *vmcommon.ContractCallInput)
	SetAsyncArgumentsForCallback(
		input *vmcommon.ContractCallInput,
		asyncCall *AsyncCall,
		gasAccumulated uint64)

	HasLegacyGroup() bool

	SetCallbackParentCall(asyncCall *AsyncCall)
	GetCallbackClosure() ([]byte, error)

	GetAsyncCallByCallID(callID []byte) AsyncCallLocation
	LoadParentContextFromStackOrStorage() (AsyncContext, error)
	ExecuteLocalCallbackAndFinishOutput(
		asyncCall *AsyncCall,
		vmOutput *vmcommon.VMOutput,
		destinationCallInput *vmcommon.ContractCallInput,
		gasAccumulated uint64,
		err error) (bool, *vmcommon.VMOutput)

	/*
		for tests / test framework usage
	*/
	SetCallID(callID []byte)
	SetCallIDForCallInGroup(groupIndex int, callIndex int, callID []byte)
}

// AsyncCallLocation defines the functionality for async calls
type AsyncCallLocation interface {
	GetAsyncCall() *AsyncCall
	GetGroupIndex() int
	GetCallIndex() int
	GetError() error
}

// GasTracing defines the functionality needed for a gas tracing
type GasTracing interface {
	BeginTrace(scAddress string, functionName string)
	AddToCurrentTrace(usedGas uint64)
	AddTracedGas(scAddress string, functionName string, usedGas uint64)
	GetGasTrace() map[string]map[string][]uint64
	IsInterfaceNil() bool
}

// HashComputer provides hash computation
type HashComputer interface {
	Compute(string) []byte
	Size() int
	IsInterfaceNil() bool
}

// EnableEpochsHandler is used to verify which flags are set in a specific epoch based on EnableEpochs config
type EnableEpochsHandler interface {
	IsFlagDefined(flag core.EnableEpochFlag) bool
	IsFlagEnabled(flag core.EnableEpochFlag) bool
	IsFlagEnabledInEpoch(flag core.EnableEpochFlag, epoch uint32) bool
	GetActivationEpoch(flag core.EnableEpochFlag) uint32
	IsInterfaceNil() bool
}
