package vmhost

import (
	"errors"
	"fmt"

	"github.com/multiversx/mx-chain-vm-go/executor"
)

// ErrNilVMType signals that the provided VMType is nil
var ErrNilVMType = errors.New("nil VMType")

// ErrNilVMHost signals that the provided VMHost is nil
var ErrNilVMHost = errors.New("nil VMHost")

// ErrNilExecutor signals that the provided Executor is nil
var ErrNilExecutor = errors.New("nil Executor")

// ErrNilHasher signals that the provided Hasher is nil
var ErrNilHasher = errors.New("nil Hasher")

// ErrReturnCodeNotOk signals that the returned code is different than vmcommon.Ok
var ErrReturnCodeNotOk = errors.New("return code is not ok")

// ErrInvalidCallOnReadOnlyMode signals that an operation is not permitted due to read only mode
var ErrInvalidCallOnReadOnlyMode = errors.New("operation not permitted in read only mode")

// ErrNotEnoughGas signals that there is not enough gas for the operation
var ErrNotEnoughGas = errors.New("not enough gas")

// ErrUnhandledRuntimeBreakpoint signals that the runtime breakpoint is unhandled
var ErrUnhandledRuntimeBreakpoint = errors.New("unhandled runtime breakpoint")

// ErrSignalError is given when the smart contract signals an error
var ErrSignalError = errors.New("error signalled by smartcontract")

// ErrExecutionFailed signals that the execution failed
var ErrExecutionFailed = errors.New("execution failed")

// ErrExecutionPanicked signals that the execution failed irrecoverably
var ErrExecutionPanicked = errors.New("VM execution panicked")

// ErrExecutionFailedWithTimeout signals that the execution failed with timeout
var ErrExecutionFailedWithTimeout = errors.New("execution failed with timeout")

// ErrMemoryLimit signals that too much memory was allocated by the contract
var ErrMemoryLimit = errors.New("memory limit reached")

// ErrBadBounds signals that a certain variable is out of bounds
var ErrBadBounds = errors.New("bad bounds")

// ErrBadLowerBounds signals that a certain variable is lower than allowed
var ErrBadLowerBounds = fmt.Errorf("%w (lower)", ErrBadBounds)

// ErrBadUpperBounds signals that a certain variable is higher than allowed
var ErrBadUpperBounds = fmt.Errorf("%w (upper)", ErrBadBounds)

// ErrNegativeLength signals that the given length is less than 0
var ErrNegativeLength = errors.New("negative length")

// ErrFailedTransfer signals that the transfer operation has failed
var ErrFailedTransfer = errors.New("failed transfer")

// ErrTransferInsufficientFunds signals that the transfer has failed due to insufficient funds
var ErrTransferInsufficientFunds = fmt.Errorf("%w (insufficient funds)", ErrFailedTransfer)

// ErrTransferNegativeValue signals that the transfer has failed due to the fact that the value is less than 0
var ErrTransferNegativeValue = fmt.Errorf("%w (negative value)", ErrFailedTransfer)

// ErrUpgradeFailed signals that the upgrade encountered an error
var ErrUpgradeFailed = errors.New("upgrade failed")

// ErrInvalidUpgradeArguments signals that the upgrade process failed due to invalid arguments
var ErrInvalidUpgradeArguments = fmt.Errorf("%w (invalid arguments)", ErrUpgradeFailed)

// ErrInitFuncCalledInRun signals that the init func was called directly, which is forbidden
var ErrInitFuncCalledInRun = fmt.Errorf("%w (calling init() directly is forbidden)", executor.ErrInvalidFunction)

// ErrCallBackFuncCalledInRun signals that a callback func was called directly, which is forbidden
var ErrCallBackFuncCalledInRun = fmt.Errorf("%w (calling callBack() directly is forbidden)", executor.ErrInvalidFunction)

// ErrInvalidFunctionName signals that the function name is invalid
var ErrInvalidFunctionName = fmt.Errorf("%w (invalid name)", executor.ErrInvalidFunction)

// ErrContractInvalid signals that the contract code is invalid
var ErrContractInvalid = fmt.Errorf("invalid contract code")

// ErrContractNotFound signals that the contract was not found
var ErrContractNotFound = fmt.Errorf("%w (not found)", ErrContractInvalid)

// ErrMemoryDeclarationMissing signals that a memory declaration is missing
var ErrMemoryDeclarationMissing = fmt.Errorf("%w (missing memory declaration)", ErrContractInvalid)

// ErrMaxInstancesReached signals that the max number of Wasmer instances has been reached.
var ErrMaxInstancesReached = fmt.Errorf("%w (max instances reached)", ErrExecutionFailed)

// ErrStoreReservedKey signals that an attempt to write under an reserved key has been made
var ErrStoreReservedKey = errors.New("cannot write to storage under reserved key")

// ErrCannotWriteProtectedKey signals an attempt to write to a protected key, while storage protection is enforced
var ErrCannotWriteProtectedKey = errors.New("cannot write to protected key")

// ErrNonPayableFunctionEgld signals that a non-payable function received non-zero call value
var ErrNonPayableFunctionEgld = errors.New("function does not accept EGLD payment")

// ErrNonPayableFunctionEsdt signals that a non-payable function received non-zero ESDT call value
var ErrNonPayableFunctionEsdt = errors.New("function does not accept ESDT payment")

// ErrArgIndexOutOfRange signals that the argument index is out of range
var ErrArgIndexOutOfRange = errors.New("argument index out of range")

// ErrArgOutOfRange signals that the argument is out of range
var ErrArgOutOfRange = errors.New("argument out of range")

// ErrStorageValueOutOfRange signals that the storage value is out of range
var ErrStorageValueOutOfRange = errors.New("storage value out of range")

// ErrDivZero signals that an attempt to divide by 0 has been made
var ErrDivZero = errors.New("division by 0")

// ErrBigIntCannotBeRepresentedAsInt64 signals that a big in conversion to int64 was attempted, but with a too large input
var ErrBigIntCannotBeRepresentedAsInt64 = errors.New("big int cannot be represented as int64")

// ErrBytesExceedInt64 signals that managed buffer cannot be converted to int64
var ErrBytesExceedInt64 = errors.New("bytes cannot be parsed as int64")

// ErrBytesExceedUint64 signals that managed buffer cannot be converted to unsigned int64
var ErrBytesExceedUint64 = errors.New("bytes cannot be parsed as uint64")

// ErrBitwiseNegative signals that an attempt to apply a bitwise operation on negative numbers has been made
var ErrBitwiseNegative = errors.New("bitwise operations only allowed on positive integers")

// ErrShiftNegative signals that an attempt to apply a bitwise shift operation on negative numbers has been made
var ErrShiftNegative = errors.New("bitwise shift operations only allowed on positive integers and by a positive amount")

// ErrAsyncCallGroupExistsAlready signals that an AsyncCallGroup with the same name already exists
var ErrAsyncCallGroupExistsAlready = errors.New("async call group exists already")

// ErrNilDestinationCallVMOutput signals that the destination call execution returned a nil VMOutput
var ErrNilDestinationCallVMOutput = errors.New("nil destination call VMOutput")

// ErrAsyncCallNotFound signals that the requested AsyncCall was not found
var ErrAsyncCallNotFound = errors.New("async call not found")

// ErrAsyncNotAllowed signals that the requested AsyncCall is not allowed
var ErrAsyncNotAllowed = errors.New("async call is not allowed at this location")

// ErrUnknownCallType signals that the call type is not recognized
var ErrUnknownCallType = errors.New("unknown call type")

// ErrCannotUseBuiltinAsCallback signals that the specified callback was set to a built-in function, which is forbidden
var ErrCannotUseBuiltinAsCallback = errors.New("cannot use built-in function as callback")

// ErrOnlyOneLegacyAsyncCallAllowed signals that there was an attempt to create more than one legacy async calls, which is forbidden
var ErrOnlyOneLegacyAsyncCallAllowed = errors.New("only one legacy async call allowed")

// ErrLegacyAsyncCallNotFound signals that a legacy async call was expected, but is missing
var ErrLegacyAsyncCallNotFound = errors.New("legacy async call not found")

// ErrLegacyAsyncCallInvalid signals that the legacy async call is invalid
var ErrLegacyAsyncCallInvalid = errors.New("legacy async call invalid")

// ErrNoStoredAsyncContextFound signals that no persisted data was found for the AsyncContext to load
var ErrNoStoredAsyncContextFound = errors.New("no async context found in storage")

// ErrCannotInterpretCallbackArgs signals that the cross-shard callback arguments are invalid
var ErrCannotInterpretCallbackArgs = errors.New("cannot interpret callback args")

// ErrContextCallbackDisabled signals that group callbacks cannot be set nor executed
var ErrContextCallbackDisabled = errors.New("context callback disabled")

// ErrInvalidAccount signals that a certain account does not exist
var ErrInvalidAccount = errors.New("account does not exist")

// ErrDeploymentOverExistingAccount signals that an attempt to deploy a new SC over an already existing account has been made
var ErrDeploymentOverExistingAccount = errors.New("cannot deploy over existing account")

// ErrAccountNotPayable signals that the value transfer to a non payable contract is not possible
var ErrAccountNotPayable = errors.New("sending value to non payable contract")

// ErrInvalidPublicKeySize signals that the public key size is invalid
var ErrInvalidPublicKeySize = errors.New("invalid public key size")

// ErrNilCallbackFunction signals that a nil callback function has been provided
var ErrNilCallbackFunction = errors.New("nil callback function")

// ErrUpgradeNotAllowed signals that an upgrade is not allowed
var ErrUpgradeNotAllowed = errors.New("upgrade not allowed")

// ErrNilContract signals that the contract is nil
var ErrNilContract = errors.New("nil contract")

// ErrBuiltinCallOnSameContextDisallowed signals that calling a built-in function on the same context is not allowed
var ErrBuiltinCallOnSameContextDisallowed = errors.New("calling built-in function on the same context is disallowed")

// ErrSyncExecutionNotInSameShard signals that the sync execution request is not in the same shard
var ErrSyncExecutionNotInSameShard = errors.New("sync execution request is not in the same shard")

// ErrInputAndOutputGasDoesNotMatch is raised when the output gas (gas used + gas locked + gas remaining)
// is not equal to the input gas
var ErrInputAndOutputGasDoesNotMatch = errors.New("input and output gas does not match")

// ErrTransferValueOnESDTCall signals that balance transfer was given in esdt call
var ErrTransferValueOnESDTCall = errors.New("transfer value on esdt call")

// ErrNoBigIntUnderThisHandle signals that there is no bigInt for the given handle
var ErrNoBigIntUnderThisHandle = errors.New("no bigInt under the given handle")

// ErrNoBigFloatUnderThisHandle signals that there is no bigInt for the given handle
var ErrNoBigFloatUnderThisHandle = errors.New("no bigFloat under the given handle")

// ErrPositiveExponent signals that the exponent is greater or equal to 0
var ErrPositiveExponent = errors.New("exponent must be negative")

// ErrLengthOfBufferNotCorrect signals that length of the buffer is not correct
var ErrLengthOfBufferNotCorrect = errors.New("length of buffer is not correct")

// ErrNoEllipticCurveUnderThisHandle singals that there is no elliptic curve for the given handle
var ErrNoEllipticCurveUnderThisHandle = errors.New("no elliptic curve under the given handle")

// ErrPointNotOnCurve signals that the point to be used is not on curve
var ErrPointNotOnCurve = errors.New("point is not on curve")

// ErrNoManagedBufferUnderThisHandle signals that there is no buffer for the given handle
var ErrNoManagedBufferUnderThisHandle = errors.New("no managed buffer under the given handle")

// ErrNoManagedMapUnderThisHandle signals that there is no buffer for the given handle
var ErrNoManagedMapUnderThisHandle = errors.New("no managed map under the given handle")

// ErrNilHostParameters signals that nil host parameters was provided
var ErrNilHostParameters = errors.New("nil host parameters")

// ErrNilESDTTransferParser signals that nil esdt transfer parser was provided
var ErrNilESDTTransferParser = errors.New("nil esdt transfer parser")

// ErrNilCallArgsParser signals that nil call arguments parser was provided
var ErrNilCallArgsParser = errors.New("nil call args parser")

// ErrNilBuiltInFunctionsContainer signals that nil built in functions container was provided
var ErrNilBuiltInFunctionsContainer = errors.New("nil built in functions container")

// ErrNilBlockChainHook signals that nil blockchain hook was provided
var ErrNilBlockChainHook = errors.New("nil blockchain hook")

// ErrTooManyESDTTransfers signals that too many ESDT transfers are in sc call
var ErrTooManyESDTTransfers = errors.New("too many ESDT transfers")

// ErrInfinityFloatOperation signals that operations with infinity are not allowed
var ErrInfinityFloatOperation = errors.New("infinity operations are not allowed")

// ErrBigFloatWrongPrecision signals that the precision has a wrong value
var ErrBigFloatWrongPrecision = errors.New("precision of the big float must be 53")

// ErrBigFloatDecode signals that big float parse error
var ErrBigFloatDecode = errors.New("big float decode error")

// ErrBigFloatEncode signals that big float parse error
var ErrBigFloatEncode = errors.New("big float encode error")

// ErrSha256Hash signals a sha256 hash error
var ErrSha256Hash = errors.New("sha256 hash error")

// ErrKeccak256Hash signals a keccak256 hash error
var ErrKeccak256Hash = errors.New("keccak256 hash error")

// ErrRipemd160Hash signals a ripemd160 hash error
var ErrRipemd160Hash = errors.New("ripemd160 hash error")

// ErrBlsVerify signals a bls verify error
var ErrBlsVerify = errors.New("bls verify error")

// ErrEd25519Verify signals a ed25519 verify error
var ErrEd25519Verify = errors.New("ed25519 verify error")

// ErrSecp256k1Verify signals a secp256k1 verify error
var ErrSecp256k1Verify = errors.New("secp256k1 verify error")

// ErrAllOperandsAreEqualToZero signals that all operands are equal to 0
var ErrAllOperandsAreEqualToZero = errors.New("all operands are equal to 0")

// ErrExponentTooBigOrTooSmall signals that the exponent is too big or too small
var ErrExponentTooBigOrTooSmall = errors.New("exponent is either too small or too big")

// ErrNilEpochNotifier signals that epoch notifier is nil
var ErrNilEpochNotifier = errors.New("nil epoch notifier")

// ErrNilEnableEpochsHandler signals that enable epochs handler is nil
var ErrNilEnableEpochsHandler = errors.New("nil enable epochs handler")

// ErrNoAsyncParentContext signals that load parent was called for an async call
var ErrNoAsyncParentContext = errors.New("this should not be called for async calls (only callbacks and direct calls)")

// ErrAsyncInit signals an async context initialization error
var ErrAsyncInit = errors.New("async context initialization error")

// ErrAsyncNoOutputFromCallback signals that an error happen while producing the output of a callback
var ErrAsyncNoOutputFromCallback = errors.New("callback VMOutput should not be nil")

// ErrAsyncNoMultiLevel signals that no multi-level async calls are allowed
var ErrAsyncNoMultiLevel = errors.New("multi-level async calls are not allowed yet")

// ErrAsyncNoCallbackForClosure signals that closure can't be obtained
var ErrAsyncNoCallbackForClosure = errors.New("no callback for closure, cannot call callback directly")

// ErrVMIsClosing signals that vm is closing
var ErrVMIsClosing = errors.New("vm is closing")

// ErrNilESDTData is given when ESDT data is missing
var ErrNilESDTData = errors.New("nil esdt data")

// ErrInvalidArgument is given when argument is invalid
var ErrInvalidArgument = errors.New("invalid argument")

// ErrInvalidTokenIndex is given when argument is invalid
var ErrInvalidTokenIndex = errors.New("invalid token index")

// ErrInvalidBuiltInFunctionCall signals that built in function was used in the wrong context
var ErrInvalidBuiltInFunctionCall = errors.New("invalid built in function call")

// ErrCannotWriteOnReadOnly signals that write operation on read only is not allowed
var ErrCannotWriteOnReadOnly = errors.New("cannot write on read only mode")

// ErrEmptyProtectedKeyPrefix signals that the protected key prefix is empty or nil
var ErrEmptyProtectedKeyPrefix = errors.New("protectedKeyPrefix is empty or nil")

// ErrInvalidGasProvided signals that an unacceptable GasProvided value was specified
var ErrInvalidGasProvided = errors.New("invalid gas provided")

// ErrNilMapOpcodeAddress signals that nil map of opcodes and addresses was provided
var ErrNilMapOpcodeAddress = errors.New("nil map opcode address")

// ErrOpcodeIsNotAllowed signals that opcode is not allowed for the address
var ErrOpcodeIsNotAllowed = errors.New("opcode is not allowed")
