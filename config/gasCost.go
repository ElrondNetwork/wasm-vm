// Package config contains structures that configure the VM
package config

import (
	"github.com/multiversx/mx-chain-vm-go/executor"
)

// GasCost defines the gas cost config structure
type GasCost struct {
	BaseOperationCost    BaseOperationCost
	BigIntAPICost        BigIntAPICost
	BigFloatAPICost      BigFloatAPICost
	BaseOpsAPICost       BaseOpsAPICost
	ManagedBufferAPICost ManagedBufferAPICost
	ManagedMapAPICost    ManagedMapAPICost
	CryptoAPICost        CryptoAPICost
	WASMOpcodeCost       *executor.WASMOpcodeCost
	DynamicStorageLoad   DynamicStorageLoadCostCoefficients
}

// BaseOperationCost defines the base operations gas cost config structure
type BaseOperationCost struct {
	StorePerByte      uint64
	ReleasePerByte    uint64
	DataCopyPerByte   uint64
	PersistPerByte    uint64
	CompilePerByte    uint64
	AoTPreparePerByte uint64
	GetCode           uint64
}

// BaseOpsAPICost defines the API operations gas cost config structure
type BaseOpsAPICost struct {
	GetSCAddress             uint64
	GetOwnerAddress          uint64
	IsSmartContract          uint64
	GetShardOfAddress        uint64
	GetExternalBalance       uint64
	GetBlockHash             uint64
	GetOriginalTxHash        uint64
	GetCurrentTxHash         uint64
	GetPrevTxHash            uint64
	TransferValue            uint64
	GetArgument              uint64
	GetFunction              uint64
	GetNumArguments          uint64
	StorageStore             uint64
	StorageLoad              uint64
	CachedStorageLoad        uint64
	GetCaller                uint64
	GetCallValue             uint64
	Log                      uint64
	Finish                   uint64
	SignalError              uint64
	GetBlockTimeStamp        uint64
	GetGasLeft               uint64
	Int64GetArgument         uint64
	Int64StorageStore        uint64
	Int64StorageLoad         uint64
	Int64Finish              uint64
	GetStateRootHash         uint64
	GetBlockNonce            uint64
	GetBlockEpoch            uint64
	GetBlockRound            uint64
	GetBlockRandomSeed       uint64
	ExecuteOnSameContext     uint64
	ExecuteOnDestContext     uint64
	DelegateExecution        uint64
	ExecuteReadOnly          uint64
	AsyncCallStep            uint64
	AsyncCallbackGasLock     uint64
	CreateAsyncCall          uint64
	SetAsyncCallback         uint64
	SetAsyncGroupCallback    uint64
	SetAsyncContextCallback  uint64
	GetCallbackClosure       uint64
	CreateContract           uint64
	GetReturnData            uint64
	GetNumReturnData         uint64
	GetReturnDataSize        uint64
	CleanReturnData          uint64
	DeleteFromReturnData     uint64
	GetCodeMetadata          uint64
	IsBuiltinFunction        uint64
	IsReservedFunctionName   uint64
	GetRoundTime             uint64
	EpochStartBlockTimeStamp uint64
	EpochStartBlockNonce     uint64
	EpochStartBlockRound     uint64
}

// DynamicStorageLoadCostCoefficients holds the signed coefficients of the func that will compute the gas cost
// based on the trie depth.
type DynamicStorageLoadCostCoefficients struct {
	Quadratic int64
	Linear    int64
	Constant  int64

	MinGasCost uint64
}

// DynamicStorageLoadUnsigned is used to store the coefficients for the func that will compute the gas cost
// based on the trie depth. The coefficients are unsigned.
type DynamicStorageLoadUnsigned struct {
	QuadraticCoefficient uint64
	SignOfQuadratic      uint64
	LinearCoefficient    uint64
	SignOfLinear         uint64
	ConstantCoefficient  uint64
	SignOfConstant       uint64
	MinimumGasCost       uint64
}

// BigIntAPICost defines the big int operations gas cost config structure
type BigIntAPICost struct {
	BigIntNew                  uint64
	BigIntUnsignedByteLength   uint64
	BigIntSignedByteLength     uint64
	BigIntGetUnsignedBytes     uint64
	BigIntGetSignedBytes       uint64
	BigIntSetUnsignedBytes     uint64
	BigIntSetSignedBytes       uint64
	BigIntIsInt64              uint64
	BigIntGetInt64             uint64
	BigIntSetInt64             uint64
	BigIntAdd                  uint64
	BigIntSub                  uint64
	BigIntMul                  uint64
	BigIntSqrt                 uint64
	BigIntPow                  uint64
	BigIntLog                  uint64
	BigIntTDiv                 uint64
	BigIntTMod                 uint64
	BigIntEDiv                 uint64
	BigIntEMod                 uint64
	BigIntAbs                  uint64
	BigIntNeg                  uint64
	BigIntSign                 uint64
	BigIntCmp                  uint64
	BigIntNot                  uint64
	BigIntAnd                  uint64
	BigIntOr                   uint64
	BigIntXor                  uint64
	BigIntShr                  uint64
	BigIntShl                  uint64
	BigIntFinishUnsigned       uint64
	BigIntFinishSigned         uint64
	BigIntStorageLoadUnsigned  uint64
	BigIntStorageStoreUnsigned uint64
	BigIntGetUnsignedArgument  uint64
	BigIntGetSignedArgument    uint64
	BigIntGetCallValue         uint64
	BigIntGetExternalBalance   uint64
	CopyPerByteForTooBig       uint64
}

// BigFloatAPICost defines the big float operations gas cost config structure
type BigFloatAPICost struct {
	BigFloatNewFromParts uint64
	BigFloatAdd          uint64
	BigFloatSub          uint64
	BigFloatMul          uint64
	BigFloatDiv          uint64
	BigFloatTruncate     uint64
	BigFloatNeg          uint64
	BigFloatClone        uint64
	BigFloatCmp          uint64
	BigFloatAbs          uint64
	BigFloatSqrt         uint64
	BigFloatPow          uint64
	BigFloatFloor        uint64
	BigFloatCeil         uint64
	BigFloatIsInt        uint64
	BigFloatSetBigInt    uint64
	BigFloatSetInt64     uint64
	BigFloatGetConst     uint64
}

// CryptoAPICost defines the crypto operations gas cost config structure
type CryptoAPICost struct {
	SHA256                  uint64
	Keccak256               uint64
	Ripemd160               uint64
	VerifyBLS               uint64
	VerifyEd25519           uint64
	VerifySecp256k1         uint64
	EllipticCurveNew        uint64
	AddECC                  uint64
	DoubleECC               uint64
	IsOnCurveECC            uint64
	ScalarMultECC           uint64
	MarshalECC              uint64
	MarshalCompressedECC    uint64
	UnmarshalECC            uint64
	UnmarshalCompressedECC  uint64
	GenerateKeyECC          uint64
	EncodeDERSig            uint64
	VerifySecp256r1         uint64
	VerifyBLSSignatureShare uint64
	VerifyBLSMultiSig       uint64
}

// ManagedBufferAPICost defines the managed buffer operations gas cost config structure
type ManagedBufferAPICost struct {
	MBufferNew                  uint64
	MBufferNewFromBytes         uint64
	MBufferGetLength            uint64
	MBufferGetBytes             uint64
	MBufferGetByteSlice         uint64
	MBufferCopyByteSlice        uint64
	MBufferSetBytes             uint64
	MBufferAppend               uint64
	MBufferAppendBytes          uint64
	MBufferToBigIntUnsigned     uint64
	MBufferToBigIntSigned       uint64
	MBufferFromBigIntUnsigned   uint64
	MBufferFromBigIntSigned     uint64
	MBufferToSmallIntUnsigned   uint64
	MBufferToSmallIntSigned     uint64
	MBufferFromSmallIntUnsigned uint64
	MBufferFromSmallIntSigned   uint64
	MBufferToBigFloat           uint64
	MBufferFromBigFloat         uint64
	MBufferStorageStore         uint64
	MBufferStorageLoad          uint64
	MBufferGetArgument          uint64
	MBufferFinish               uint64
	MBufferSetRandom            uint64
}

// ManagedMapAPICost defines the managed map operations gas cost config structure
type ManagedMapAPICost struct {
	ManagedMapNew      uint64
	ManagedMapPut      uint64
	ManagedMapGet      uint64
	ManagedMapRemove   uint64
	ManagedMapContains uint64
}
