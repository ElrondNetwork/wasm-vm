package arwen

import (
	"crypto/elliptic"
	"math/big"

	"github.com/ElrondNetwork/arwen-wasm-vm/v1_4/config"
	"github.com/ElrondNetwork/arwen-wasm-vm/v1_4/crypto"
	"github.com/ElrondNetwork/arwen-wasm-vm/v1_4/wasmer"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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

	ExecuteESDTTransfer(destination []byte, sender []byte, esdtTransfers []*vmcommon.ESDTTransfer, callType vm.CallType) (*vmcommon.VMOutput, uint64, error)
	CreateNewContract(input *vmcommon.ContractCreateInput) ([]byte, error)
	ExecuteOnSameContext(input *vmcommon.ContractCallInput) error
	ExecuteOnDestContext(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, bool, error)
	GetAPIMethods() *wasmer.Imports
	IsBuiltinFunctionName(functionName string) bool
	IsBuiltinFunctionCall(data []byte) bool
	AreInSameShard(leftAddress []byte, rightAddress []byte) bool

	GetGasScheduleMap() config.GasScheduleMap
	GetContexts() (ManagedTypesContext, BlockchainContext, MeteringContext, OutputContext, RuntimeContext, AsyncContext, StorageContext)
	SetRuntimeContext(runtime RuntimeContext)

	SetBuiltInFunctionsContainer(builtInFuncs vmcommon.BuiltInFunctionContainer)
	InitState()

	SetGasFlag(flag bool)
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
	IsPayable(address []byte) (bool, error)
	SaveCompiledCode(codeHash []byte, code []byte)
	GetCompiledCode(codeHash []byte) (bool, []byte)
	GetESDTToken(address []byte, tokenID []byte, nonce uint64) (*esdt.ESDigitalToken, error)
	GetUserAccount(address []byte) (vmcommon.UserAccountHandler, error)
	ProcessBuiltInFunction(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
	GetSnapshot() int
	RevertToSnapshot(snapshot int)
}

// RuntimeContext defines the functionality needed for interacting with the runtime context
type RuntimeContext interface {
	StateStack

	InitStateFromContractCallInput(input *vmcommon.ContractCallInput)
	SetCustomCallFunction(callFunction string)
	GetVMInput() *vmcommon.VMInput
	SetVMInput(vmInput *vmcommon.VMInput)
	GetSCAddress() []byte
	SetSCAddress(scAddress []byte)
	GetSCCode() ([]byte, error)
	GetSCCodeSize() uint64
	GetVMType() []byte
	Function() string
	Arguments() [][]byte
	GetCurrentTxHash() []byte
	GetOriginalTxHash() []byte
	ExtractCodeUpgradeFromArgs() ([]byte, []byte, error)
	SignalUserError(message string)
	FailExecution(err error)
	MustVerifyNextContractCode()
	SetRuntimeBreakpointValue(value BreakpointValue)
	GetRuntimeBreakpointValue() BreakpointValue
	RunningInstancesCount() uint64
	CountSameContractInstancesOnStack(address []byte) uint64
	IsFunctionImported(name string) bool
	ReadOnly() bool
	SetReadOnly(readOnly bool)
	StartWasmerInstance(contract []byte, gasLimit uint64, newCode bool) error
	CleanWasmerInstance()
	SetMaxInstanceCount(uint64)
	VerifyContractCode() error
	GetInstance() wasmer.InstanceHandler
	GetInstanceExports() wasmer.ExportsMap
	GetInitFunction() wasmer.ExportedFunctionCallback
	GetFunctionToCall() (wasmer.ExportedFunctionCallback, error)
	GetPointsUsed() uint64
	SetPointsUsed(gasPoints uint64)
	MemStore(offset int32, data []byte) error
	MemLoad(offset int32, length int32) ([]byte, error)
	MemLoadMultiple(offset int32, lengths []int32) ([][]byte, error)
	ElrondAPIErrorShouldFailExecution() bool
	ElrondSyncExecAPIErrorShouldFailExecution() bool
	CryptoAPIErrorShouldFailExecution() bool
	BigIntAPIErrorShouldFailExecution() bool
	ManagedBufferAPIErrorShouldFailExecution() bool

	AddError(err error, otherInfo ...string)
	GetAllErrors() error

	ValidateCallbackName(callbackName string) error
	HasFunction(functionName string) bool
	GetPrevTxHash() []byte
	GetAndEliminateFirstArgumentFromList() []byte

	// TODO remove after implementing proper mocking of Wasmer instances; this is
	// used for tests only
	ReplaceInstanceBuilder(builder InstanceBuilder)

	IsFirstCallACallback() bool
}

// AddressAndCallID holds info from a runtime stack
type AddressAndCallID struct {
	Address      []byte
	CallID       []byte
	IndexOnStack int
}

// ManagedTypesContext defines the functionality needed for interacting with the big int context
type ManagedTypesContext interface {
	StateStack

	ConsumeGasForThisBigIntNumberOfBytes(byteLen *big.Int)
	ConsumeGasForThisIntNumberOfBytes(byteLen int)
	ConsumeGasForBigIntCopy(values ...*big.Int)
	PutBigInt(value int64) int32
	GetBigIntOrCreate(handle int32) *big.Int
	GetBigInt(id int32) (*big.Int, error)
	GetTwoBigInt(handle1 int32, handle2 int32) (*big.Int, *big.Int, error)
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
	WriteLog(address []byte, topics [][]byte, data []byte)
	TransferValueOnly(destination []byte, sender []byte, value *big.Int, checkPayable bool) error
	Transfer(destination []byte, sender []byte, gasLimit uint64, gasLocked uint64, value *big.Int, input []byte, callType vm.CallType) error
	TransferESDT(destination []byte, sender []byte, transfers []*vmcommon.ESDTTransfer, callInput *vmcommon.ContractCallInput) (uint64, error)
	SelfDestruct(address []byte, beneficiary []byte)
	GetRefund() uint64
	SetRefund(refund uint64)
	ReturnCode() vmcommon.ReturnCode
	SetReturnCode(returnCode vmcommon.ReturnCode)
	ReturnMessage() string
	SetReturnMessage(message string)
	ReturnData() [][]byte
	ClearReturnData()
	Finish(data []byte)
	PrependFinish(data []byte)
	GetVMOutput() *vmcommon.VMOutput
	AddTxValueToAccount(address []byte, value *big.Int)
	DeployCode(input CodeDeployInput)
	CreateVMOutputInCaseOfError(err error) *vmcommon.VMOutput
}

// MeteringContext defines the functionality needed for interacting with the metering context
type MeteringContext interface {
	StateStack
	PopMergeActiveState()

	InitStateFromContractCallInput(input *vmcommon.VMInput)
	SetGasSchedule(gasMap config.GasScheduleMap)
	GasSchedule() *config.GasCost
	UseGas(gas uint64)
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
	ComputeGasLockedForAsync() uint64
	UseGasForAsyncStep() error
	UseGasBounded(gasToUse uint64) error
	GetGasLocked() uint64
	UpdateGasStateOnSuccess(vmOutput *vmcommon.VMOutput) error
	UpdateGasStateOnFailure(vmOutput *vmcommon.VMOutput)
	TrackGasUsedByBuiltinFunction(builtinInput *vmcommon.ContractCallInput, builtinOutput *vmcommon.VMOutput, postBuiltinInput *vmcommon.ContractCallInput)
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
	GetAddress() []byte
	GetStorageUpdates(address []byte) map[string]*vmcommon.StorageUpdate
	GetStorageFromAddress(address []byte, key []byte) []byte
	GetStorageFromAddressNoChecks(address []byte, key []byte) []byte
	GetStorage(key []byte) []byte
	GetStorageUnmetered(key []byte) []byte
	SetStorage(key []byte, value []byte) (StorageStatus, error)
	SetProtectedStorage(key []byte, value []byte) (StorageStatus, error)
	SetProtectedStorageToAddress(address []byte, key []byte, value []byte) (StorageStatus, error)
}

// AsyncCallInfoHandler defines the functionality for working with AsyncCallInfo
type AsyncCallInfoHandler interface {
	GetDestination() []byte
	GetData() []byte
	GetGasLimit() uint64
	GetGasLocked() uint64
	GetValueBytes() []byte
}

// InstanceBuilder defines the functionality needed to create Wasmer instances
type InstanceBuilder interface {
	NewInstanceWithOptions(contractCode []byte, options wasmer.CompilationOptions) (wasmer.InstanceHandler, error)
	NewInstanceFromCompiledCodeWithOptions(compiledCode []byte, options wasmer.CompilationOptions) (wasmer.InstanceHandler, error)
}

// AsyncContext defines the functionality needed for interacting with the asynchronous execution context
type AsyncContext interface {
	StateStack

	InitStateFromInput(input *vmcommon.ContractCallInput)
	HasPendingCallGroups() bool
	IsComplete() bool
	GetCallGroup(groupID string) (*AsyncCallGroup, bool)
	SetGroupCallback(groupID string, callbackName string, data []byte, gas uint64) error
	SetContextCallback(callbackName string, data []byte, gas uint64) error
	HasCallback() bool
	GetAddress() []byte
	GetCallerAddress() []byte
	GetCallerCallID() []byte
	GetReturnData() []byte
	SetReturnData(data []byte)
	GetGasPrice() uint64

	Execute() error
	RegisterAsyncCall(groupID string, call *AsyncCall) error
	RegisterLegacyAsyncCall(address []byte, data []byte, value []byte) error

	LoadParentContext() error
	Save() error
	SaveAsyncContextsFromStack() error
	Delete() error

	GetCallID() []byte
	GetCallbackAsyncInitiatorCallID() []byte
	IsCrossShard() bool
	GenerateNewCallIDAndIncrementCounter() []byte
	GenerateNewCallbackID() []byte

	Clone() AsyncContext

	UpdateCurrentAsyncCallStatus(address []byte, callID []byte, asyncCallIdentifier []byte, vmInput *vmcommon.VMInput) (*AsyncCall, error)
	ExecuteCrossShardCallback() error

	CompleteChild(asyncCallIdentifier []byte, gasToAccumulate uint64) error
	NotifyChildIsComplete(asyncCallIdentifier []byte, gasToAccumulate uint64, gasToRestore uint64) (AsyncContext, error)

	AccumulateGasFromPreviousState()
	SetResults(vmOutput *vmcommon.VMOutput)
	GetGasAccumulated() uint64

	PrependArgumentsForAsyncContext(args [][]byte) ([]byte, [][]byte)

	// for tests usage
	SetCallID(callID []byte)
	SetCallIDForCallInGroup(groupIndex int, callIndex int, callID []byte)
}
