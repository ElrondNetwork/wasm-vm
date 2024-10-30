package config

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-vm-go/executor"
)

// GasValueForTests defines the gas value for tests
const GasValueForTests = 1

const isNegativeNumber = 1

// AsyncCallbackGasLockForTests defines the gas lock for tests
var AsyncCallbackGasLockForTests = uint64(100_000)

var log = logger.GetOrCreate("vm/config")

// GasScheduleMap (alias) is the map for gas schedule
type GasScheduleMap = map[string]map[string]uint64

// CreateGasConfig will create a gas cost instance
func CreateGasConfig(gasMap GasScheduleMap) (*GasCost, error) {
	baseOps := &BaseOperationCost{}
	err := mapstructure.Decode(gasMap["BaseOperationCost"], baseOps)
	if err != nil {
		return nil, err
	}

	err = checkForZeroUint64Fields(*baseOps)
	if err != nil {
		return nil, err
	}

	baseOpsAPI := &BaseOpsAPICost{}
	err = mapstructure.Decode(gasMap["BaseOpsAPICost"], baseOpsAPI)
	if err != nil {
		return nil, err
	}

	err = checkForZeroUint64Fields(*baseOpsAPI)
	if err != nil {
		return nil, err
	}

	bigFloatOps := &BigFloatAPICost{}
	err = mapstructure.Decode(gasMap["BigFloatAPICost"], bigFloatOps)
	if err != nil {
		return nil, err
	}

	err = checkForZeroUint64Fields(*bigFloatOps)
	if err != nil {
		return nil, err
	}

	bigIntOps := &BigIntAPICost{}
	err = mapstructure.Decode(gasMap["BigIntAPICost"], bigIntOps)
	if err != nil {
		return nil, err
	}

	err = checkForZeroUint64Fields(*bigIntOps)
	if err != nil {
		return nil, err
	}

	cryptOps := &CryptoAPICost{}
	err = mapstructure.Decode(gasMap["CryptoAPICost"], cryptOps)
	if err != nil {
		return nil, err
	}

	err = checkForZeroUint64Fields(*cryptOps)
	if err != nil {
		return nil, err
	}

	MBufferOps := &ManagedBufferAPICost{}
	err = mapstructure.Decode(gasMap["ManagedBufferAPICost"], MBufferOps)
	if err != nil {
		return nil, err
	}

	err = checkForZeroUint64Fields(*MBufferOps)
	if err != nil {
		return nil, err
	}

	wasmOps := &executor.WASMOpcodeCost{}
	err = mapstructure.Decode(gasMap["WASMOpcodeCost"], wasmOps)
	if err != nil {
		return nil, err
	}

	err = checkForZeroUint64Fields(*wasmOps)
	if err != nil {
		return nil, err
	}
	dynamicStorageLoadUnsigned := &DynamicStorageLoadUnsigned{}
	err = mapstructure.Decode(gasMap["DynamicStorageLoad"], dynamicStorageLoadUnsigned)
	if err != nil {
		return nil, err
	}
	dynamicStorageLoadParams := convertFromUnsignedToSigned(dynamicStorageLoadUnsigned)

	isCorrectlyDefined := isDynamicGasComputationFuncCorrectlyDefined(dynamicStorageLoadParams)
	if !isCorrectlyDefined {
		return nil, fmt.Errorf("dynamic gas computation func incorrectly defined, "+
			"quadratic parameter = %d, linear parameter = %d, constant parameter = %d",
			dynamicStorageLoadParams.Quadratic, dynamicStorageLoadParams.Linear, dynamicStorageLoadParams.Constant)
	}

	gasCost := &GasCost{
		BaseOperationCost:    *baseOps,
		BigIntAPICost:        *bigIntOps,
		BigFloatAPICost:      *bigFloatOps,
		BaseOpsAPICost:       *baseOpsAPI,
		CryptoAPICost:        *cryptOps,
		ManagedBufferAPICost: *MBufferOps,
		WASMOpcodeCost:       wasmOps,
		DynamicStorageLoad:   *dynamicStorageLoadParams,
	}

	return gasCost, nil
}

func isDynamicGasComputationFuncCorrectlyDefined(parameters *DynamicStorageLoadCostCoefficients) bool {
	if parameters.Quadratic <= 0 {
		// the "a" from ax^2+bx+c needs to be > 0 in order for the func to be convex
		log.Error("invalid parameters for dynamic gas computation func, the quadratic parameter is not > 0")
		return false
	}
	stationaryPoint := float64(-1*parameters.Linear) / float64(2*parameters.Quadratic) // -b/2a
	if stationaryPoint > 0 {
		// the stationary point should be <= 0 because the func needs to be strictly increasing for x = [0,n)
		log.Error("invalid parameters for dynamic gas computation func, the x of the stationary point is > 0")
		return false
	}
	if parameters.Constant < 0 {
		// f(x) = ax^2+bx+c. f(0) >= 0 only if c >= 0
		log.Error("invalid parameters for dynamic gas computation func, f(x) is not >= 0 for x = [0,n) ")
		return false
	}

	return true
}

func convertFromUnsignedToSigned(dynamicStorageLoadUnsigned *DynamicStorageLoadUnsigned) *DynamicStorageLoadCostCoefficients {
	quadratic := getSignedCoefficient(dynamicStorageLoadUnsigned.QuadraticCoefficient, dynamicStorageLoadUnsigned.SignOfQuadratic)
	linear := getSignedCoefficient(dynamicStorageLoadUnsigned.LinearCoefficient, dynamicStorageLoadUnsigned.SignOfLinear)
	constant := getSignedCoefficient(dynamicStorageLoadUnsigned.ConstantCoefficient, dynamicStorageLoadUnsigned.SignOfConstant)

	return &DynamicStorageLoadCostCoefficients{
		Quadratic:  quadratic,
		Linear:     linear,
		Constant:   constant,
		MinGasCost: dynamicStorageLoadUnsigned.MinimumGasCost,
	}
}

func getSignedCoefficient(coefficient uint64, sign uint64) int64 {
	if sign == isNegativeNumber {
		return int64(coefficient) * -1
	}

	return int64(coefficient)
}

func checkForZeroUint64Fields(arg interface{}) error {
	v := reflect.ValueOf(arg)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() != reflect.Uint64 && field.Kind() != reflect.Uint32 {
			continue
		}
		if field.Uint() == 0 {
			name := v.Type().Field(i).Name
			return fmt.Errorf("gas cost for operation %s has been set to 0 or is not set", name)
		}
	}

	return nil
}

// MakeGasMap creates a new GasScheduleMap instance
func MakeGasMap(value, asyncCallbackGasLock uint64) GasScheduleMap {
	gasMap := make(GasScheduleMap)
	gasMap = FillGasMap(gasMap, value, asyncCallbackGasLock)
	return gasMap
}

// FillGasMap will fill the gas map with the provided value
func FillGasMap(gasMap GasScheduleMap, value, asyncCallbackGasLock uint64) GasScheduleMap {
	gasMap["BuiltInCost"] = FillGasMapBuiltInCosts(value)
	gasMap["BaseOperationCost"] = FillGasMapBaseOperationCosts(value)
	gasMap["BaseOpsAPICost"] = FillGasMapBaseOpsAPICosts(value, asyncCallbackGasLock)
	// EthAPICost needed on node for VM tests < 1.5
	gasMap["EthAPICost"] = FillGasMapEthereumAPICosts(value)
	gasMap["BigIntAPICost"] = FillGasMapBigIntAPICosts(value)
	gasMap["BigFloatAPICost"] = FillGasMapBigFloatAPICosts(value)
	gasMap["CryptoAPICost"] = FillGasMapCryptoAPICosts(value)
	gasMap["ManagedBufferAPICost"] = FillGasMapManagedBufferAPICosts(value)
	gasMap["WASMOpcodeCost"] = FillGasMapWASMOpcodeValues(value)
	gasMap["DynamicStorageLoad"] = FillGasMapDynamicStorageLoad()

	customFillGasMapWASMOpcodeCosts(gasMap["WASMOpcodeCost"])

	return gasMap
}

// FillGasMapBuiltInCosts fills the builtin costs
func FillGasMapBuiltInCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["ChangeOwnerAddress"] = value
	gasMap["ClaimDeveloperRewards"] = value
	gasMap["SaveUserName"] = value
	gasMap["SaveKeyValue"] = value
	gasMap["ESDTTransfer"] = value
	gasMap["ESDTBurn"] = value
	gasMap["ESDTLocalMint"] = value
	gasMap["ESDTLocalBurn"] = value
	gasMap["ESDTNFTCreate"] = value
	gasMap["ESDTNFTAddQuantity"] = value
	gasMap["ESDTNFTBurn"] = value
	gasMap["ESDTNFTTransfer"] = value
	gasMap["ESDTNFTChangeCreateOwner"] = value
	gasMap["ESDTNFTAddUri"] = value
	gasMap["ESDTNFTUpdateAttributes"] = value
	gasMap["ESDTNFTMultiTransfer"] = value
	gasMap["ESDTModifyRoyalties"] = value
	gasMap["ESDTModifyCreator"] = value
	gasMap["ESDTNFTRecreate"] = value
	gasMap["ESDTNFTUpdate"] = value
	gasMap["ESDTNFTSetNewURIs"] = value
	gasMap["SetGuardian"] = value
	gasMap["GuardAccount"] = value
	gasMap["UnGuardAccount"] = value
	gasMap["TrieLoadPerNode"] = value
	gasMap["TrieStorePerNode"] = value

	return gasMap
}

// FillGasMapBaseOperationCosts fills the base operation costs
func FillGasMapBaseOperationCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["StorePerByte"] = value
	gasMap["DataCopyPerByte"] = value
	gasMap["ReleasePerByte"] = value
	gasMap["PersistPerByte"] = value
	gasMap["CompilePerByte"] = value
	gasMap["AoTPreparePerByte"] = value
	gasMap["GetCode"] = value

	return gasMap
}

// FillGasMapBaseOpsAPICosts fills the API calls costs
func FillGasMapBaseOpsAPICosts(value, asyncCallbackGasLock uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["GetSCAddress"] = value
	gasMap["GetOwnerAddress"] = value
	gasMap["IsSmartContract"] = value
	gasMap["GetShardOfAddress"] = value
	gasMap["GetExternalBalance"] = value
	gasMap["GetBlockHash"] = value
	gasMap["GetOriginalTxHash"] = value
	gasMap["GetCurrentTxHash"] = value
	gasMap["GetPrevTxHash"] = value
	gasMap["TransferValue"] = value
	gasMap["GetArgument"] = value
	gasMap["GetFunction"] = value
	gasMap["GetNumArguments"] = value
	gasMap["StorageStore"] = value
	gasMap["StorageLoad"] = value
	gasMap["CachedStorageLoad"] = value
	gasMap["GetCaller"] = value
	gasMap["GetCallValue"] = value
	gasMap["Log"] = value
	gasMap["Finish"] = value
	gasMap["SignalError"] = value
	gasMap["GetBlockTimeStamp"] = value
	gasMap["GetGasLeft"] = value
	gasMap["Int64GetArgument"] = value
	gasMap["Int64StorageStore"] = value
	gasMap["Int64StorageLoad"] = value
	gasMap["Int64Finish"] = value
	gasMap["GetStateRootHash"] = value
	gasMap["GetBlockNonce"] = value
	gasMap["GetBlockEpoch"] = value
	gasMap["GetBlockRound"] = value
	gasMap["GetBlockRandomSeed"] = value
	gasMap["ExecuteOnSameContext"] = value
	gasMap["ExecuteOnDestContext"] = value
	gasMap["DelegateExecution"] = value
	gasMap["ExecuteReadOnly"] = value
	gasMap["AsyncCallStep"] = value
	gasMap["AsyncCallbackGasLock"] = asyncCallbackGasLock
	gasMap["CreateAsyncCall"] = value
	gasMap["CreateAsyncV3Call"] = value
	gasMap["SetAsyncCallback"] = value
	gasMap["SetAsyncGroupCallback"] = value
	gasMap["SetAsyncContextCallback"] = value
	gasMap["GetCallbackClosure"] = value
	gasMap["CreateContract"] = value
	gasMap["GetReturnData"] = value
	gasMap["GetNumReturnData"] = value
	gasMap["GetReturnDataSize"] = value
	gasMap["CleanReturnData"] = value
	gasMap["DeleteFromReturnData"] = value
	gasMap["GetCodeMetadata"] = value
	gasMap["IsBuiltinFunction"] = value
	gasMap["IsReservedFunctionName"] = value

	return gasMap
}

// FillGasMapEthereumAPICosts is needed for EthAPICost, which is needed on node for VM tests < 1.5
func FillGasMapEthereumAPICosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["UseGas"] = value
	gasMap["GetAddress"] = value
	gasMap["GetExternalBalance"] = value
	gasMap["GetBlockHash"] = value
	gasMap["Call"] = value
	gasMap["CallDataCopy"] = value
	gasMap["GetCallDataSize"] = value
	gasMap["CallCode"] = value
	gasMap["CallDelegate"] = value
	gasMap["CallStatic"] = value
	gasMap["StorageStore"] = value
	gasMap["StorageLoad"] = value
	gasMap["GetCaller"] = value
	gasMap["GetCallValue"] = value
	gasMap["CodeCopy"] = value
	gasMap["GetCodeSize"] = value
	gasMap["GetBlockCoinbase"] = value
	gasMap["Create"] = value
	gasMap["GetBlockDifficulty"] = value
	gasMap["ExternalCodeCopy"] = value
	gasMap["GetExternalCodeSize"] = value
	gasMap["GetGasLeft"] = value
	gasMap["GetBlockGasLimit"] = value
	gasMap["GetTxGasPrice"] = value
	gasMap["Log"] = value
	gasMap["GetBlockNumber"] = value
	gasMap["GetTxOrigin"] = value
	gasMap["Finish"] = value
	gasMap["Revert"] = value
	gasMap["GetReturnDataSize"] = value
	gasMap["ReturnDataCopy"] = value
	gasMap["SelfDestruct"] = value
	gasMap["GetBlockTimeStamp"] = value

	return gasMap
}

// FillGasMapBigIntAPICosts fills the big int costs
func FillGasMapBigIntAPICosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["BigIntNew"] = value
	gasMap["BigIntUnsignedByteLength"] = value
	gasMap["BigIntSignedByteLength"] = value
	gasMap["BigIntGetUnsignedBytes"] = value
	gasMap["BigIntGetSignedBytes"] = value
	gasMap["BigIntSetUnsignedBytes"] = value
	gasMap["BigIntSetSignedBytes"] = value
	gasMap["BigIntIsInt64"] = value
	gasMap["BigIntGetInt64"] = value
	gasMap["BigIntSetInt64"] = value
	gasMap["BigIntAdd"] = value
	gasMap["BigIntSub"] = value
	gasMap["BigIntMul"] = value
	gasMap["BigIntSqrt"] = value
	gasMap["BigIntPow"] = value
	gasMap["BigIntLog"] = value
	gasMap["BigIntTDiv"] = value
	gasMap["BigIntTMod"] = value
	gasMap["BigIntEDiv"] = value
	gasMap["BigIntEMod"] = value
	gasMap["BigIntAbs"] = value
	gasMap["BigIntNeg"] = value
	gasMap["BigIntSign"] = value
	gasMap["BigIntCmp"] = value
	gasMap["BigIntNot"] = value
	gasMap["BigIntAnd"] = value
	gasMap["BigIntOr"] = value
	gasMap["BigIntXor"] = value
	gasMap["BigIntShr"] = value
	gasMap["BigIntShl"] = value
	gasMap["BigIntFinishUnsigned"] = value
	gasMap["BigIntFinishSigned"] = value
	gasMap["BigIntStorageLoadUnsigned"] = value
	gasMap["BigIntStorageStoreUnsigned"] = value
	gasMap["BigIntGetUnsignedArgument"] = value
	gasMap["BigIntGetSignedArgument"] = value
	gasMap["BigIntGetCallValue"] = value
	gasMap["BigIntGetExternalBalance"] = value
	gasMap["CopyPerByteForTooBig"] = value

	return gasMap
}

// FillGasMapBigFloatAPICosts fills the big floats costs
func FillGasMapBigFloatAPICosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["BigFloatNewFromParts"] = value
	gasMap["BigFloatAdd"] = value
	gasMap["BigFloatSub"] = value
	gasMap["BigFloatMul"] = value
	gasMap["BigFloatDiv"] = value
	gasMap["BigFloatTruncate"] = value
	gasMap["BigFloatNeg"] = value
	gasMap["BigFloatClone"] = value
	gasMap["BigFloatCmp"] = value
	gasMap["BigFloatAbs"] = value
	gasMap["BigFloatSqrt"] = value
	gasMap["BigFloatPow"] = value
	gasMap["BigFloatFloor"] = value
	gasMap["BigFloatCeil"] = value
	gasMap["BigFloatIsInt"] = value
	gasMap["BigFloatSetBigInt"] = value
	gasMap["BigFloatSetInt64"] = value
	gasMap["BigFloatGetConst"] = value

	return gasMap
}

// FillGasMapCryptoAPICosts fills the crypto costs
func FillGasMapCryptoAPICosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["SHA256"] = value
	gasMap["Keccak256"] = value
	gasMap["Ripemd160"] = value
	gasMap["VerifyBLS"] = value
	gasMap["VerifyEd25519"] = value
	gasMap["VerifySecp256k1"] = value
	gasMap["EllipticCurveNew"] = value
	gasMap["AddECC"] = value
	gasMap["DoubleECC"] = value
	gasMap["IsOnCurveECC"] = value
	gasMap["ScalarMultECC"] = value
	gasMap["MarshalECC"] = value
	gasMap["MarshalCompressedECC"] = value
	gasMap["UnmarshalECC"] = value
	gasMap["UnmarshalCompressedECC"] = value
	gasMap["GenerateKeyECC"] = value
	gasMap["EncodeDERSig"] = value
	gasMap["VerifySecp256r1"] = value
	gasMap["VerifyBLSSignatureShare"] = value
	gasMap["VerifyBLSMultiSig"] = value

	return gasMap
}

// FillGasMapManagedBufferAPICosts fills the managed buffer costs
func FillGasMapManagedBufferAPICosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["MBufferNew"] = value
	gasMap["MBufferNewFromBytes"] = value
	gasMap["MBufferGetLength"] = value
	gasMap["MBufferGetBytes"] = value
	gasMap["MBufferGetByteSlice"] = value
	gasMap["MBufferCopyByteSlice"] = value
	gasMap["MBufferSetBytes"] = value
	gasMap["MBufferAppend"] = value
	gasMap["MBufferAppendBytes"] = value
	gasMap["MBufferToBigIntUnsigned"] = value
	gasMap["MBufferToSmallIntUnsigned"] = value
	gasMap["MBufferToSmallIntSigned"] = value
	gasMap["MBufferFromSmallIntUnsigned"] = value
	gasMap["MBufferFromSmallIntSigned"] = value
	gasMap["MBufferToBigIntSigned"] = value
	gasMap["MBufferFromBigIntUnsigned"] = value
	gasMap["MBufferFromBigIntSigned"] = value
	gasMap["MBufferToBigFloat"] = value
	gasMap["MBufferFromBigFloat"] = value
	gasMap["MBufferStorageStore"] = value
	gasMap["MBufferStorageLoad"] = value
	gasMap["MBufferGetArgument"] = value
	gasMap["MBufferFinish"] = value
	gasMap["MBufferSetRandom"] = value

	return gasMap
}

// FillGasMapWASMOpcodeValues fills the wasm opcodes costs
func FillGasMapWASMOpcodeValues(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)

	gasMap["AtomicFence"] = value
	gasMap["AtomicNotify"] = value
	gasMap["Block"] = value
	gasMap["Br"] = value
	gasMap["BrIf"] = value
	gasMap["BrTable"] = value
	gasMap["Call"] = value
	gasMap["CallIndirect"] = value
	gasMap["Catch"] = value
	gasMap["CatchAll"] = value
	gasMap["DataDrop"] = value
	gasMap["Delegate"] = value
	gasMap["Drop"] = value
	gasMap["ElemDrop"] = value
	gasMap["Else"] = value
	gasMap["End"] = value
	gasMap["F32Abs"] = value
	gasMap["F32Add"] = value
	gasMap["F32Ceil"] = value
	gasMap["F32Const"] = value
	gasMap["F32ConvertI32S"] = value
	gasMap["F32ConvertI32U"] = value
	gasMap["F32ConvertI64S"] = value
	gasMap["F32ConvertI64U"] = value
	gasMap["F32Copysign"] = value
	gasMap["F32DemoteF64"] = value
	gasMap["F32Div"] = value
	gasMap["F32Eq"] = value
	gasMap["F32Floor"] = value
	gasMap["F32Ge"] = value
	gasMap["F32Gt"] = value
	gasMap["F32Le"] = value
	gasMap["F32Load"] = value
	gasMap["F32Lt"] = value
	gasMap["F32Max"] = value
	gasMap["F32Min"] = value
	gasMap["F32Mul"] = value
	gasMap["F32Ne"] = value
	gasMap["F32Nearest"] = value
	gasMap["F32Neg"] = value
	gasMap["F32ReinterpretI32"] = value
	gasMap["F32Sqrt"] = value
	gasMap["F32Store"] = value
	gasMap["F32Sub"] = value
	gasMap["F32Trunc"] = value
	gasMap["F32x4Abs"] = value
	gasMap["F32x4Add"] = value
	gasMap["F32x4Ceil"] = value
	gasMap["F32x4ConvertI32x4S"] = value
	gasMap["F32x4ConvertI32x4U"] = value
	gasMap["F32x4DemoteF64x2Zero"] = value
	gasMap["F32x4Div"] = value
	gasMap["F32x4Eq"] = value
	gasMap["F32x4ExtractLane"] = value
	gasMap["F32x4Floor"] = value
	gasMap["F32x4Ge"] = value
	gasMap["F32x4Gt"] = value
	gasMap["F32x4Le"] = value
	gasMap["F32x4Lt"] = value
	gasMap["F32x4Max"] = value
	gasMap["F32x4Min"] = value
	gasMap["F32x4Mul"] = value
	gasMap["F32x4Ne"] = value
	gasMap["F32x4Nearest"] = value
	gasMap["F32x4Neg"] = value
	gasMap["F32x4PMax"] = value
	gasMap["F32x4PMin"] = value
	gasMap["F32x4ReplaceLane"] = value
	gasMap["F32x4Splat"] = value
	gasMap["F32x4Sqrt"] = value
	gasMap["F32x4Sub"] = value
	gasMap["F32x4Trunc"] = value
	gasMap["F64Abs"] = value
	gasMap["F64Add"] = value
	gasMap["F64Ceil"] = value
	gasMap["F64Const"] = value
	gasMap["F64ConvertI32S"] = value
	gasMap["F64ConvertI32U"] = value
	gasMap["F64ConvertI64S"] = value
	gasMap["F64ConvertI64U"] = value
	gasMap["F64Copysign"] = value
	gasMap["F64Div"] = value
	gasMap["F64Eq"] = value
	gasMap["F64Floor"] = value
	gasMap["F64Ge"] = value
	gasMap["F64Gt"] = value
	gasMap["F64Le"] = value
	gasMap["F64Load"] = value
	gasMap["F64Lt"] = value
	gasMap["F64Max"] = value
	gasMap["F64Min"] = value
	gasMap["F64Mul"] = value
	gasMap["F64Ne"] = value
	gasMap["F64Nearest"] = value
	gasMap["F64Neg"] = value
	gasMap["F64PromoteF32"] = value
	gasMap["F64ReinterpretI64"] = value
	gasMap["F64Sqrt"] = value
	gasMap["F64Store"] = value
	gasMap["F64Sub"] = value
	gasMap["F64Trunc"] = value
	gasMap["F64x2Abs"] = value
	gasMap["F64x2Add"] = value
	gasMap["F64x2Ceil"] = value
	gasMap["F64x2ConvertI64x2S"] = value
	gasMap["F64x2ConvertI64x2U"] = value
	gasMap["F64x2ConvertLowI32x4S"] = value
	gasMap["F64x2ConvertLowI32x4U"] = value
	gasMap["F64x2Div"] = value
	gasMap["F64x2Eq"] = value
	gasMap["F64x2ExtractLane"] = value
	gasMap["F64x2Floor"] = value
	gasMap["F64x2Ge"] = value
	gasMap["F64x2Gt"] = value
	gasMap["F64x2Le"] = value
	gasMap["F64x2Lt"] = value
	gasMap["F64x2Max"] = value
	gasMap["F64x2Min"] = value
	gasMap["F64x2Mul"] = value
	gasMap["F64x2Ne"] = value
	gasMap["F64x2Nearest"] = value
	gasMap["F64x2Neg"] = value
	gasMap["F64x2PMax"] = value
	gasMap["F64x2PMin"] = value
	gasMap["F64x2PromoteLowF32x4"] = value
	gasMap["F64x2ReplaceLane"] = value
	gasMap["F64x2Splat"] = value
	gasMap["F64x2Sqrt"] = value
	gasMap["F64x2Sub"] = value
	gasMap["F64x2Trunc"] = value
	gasMap["GlobalGet"] = value
	gasMap["GlobalSet"] = value
	gasMap["I16x8Abs"] = value
	gasMap["I16x8Add"] = value
	gasMap["I16x8AddSatS"] = value
	gasMap["I16x8AddSatU"] = value
	gasMap["I16x8AddSaturateS"] = value
	gasMap["I16x8AddSaturateU"] = value
	gasMap["I16x8AllTrue"] = value
	gasMap["I16x8AnyTrue"] = value
	gasMap["I16x8Bitmask"] = value
	gasMap["I16x8Eq"] = value
	gasMap["I16x8ExtAddPairwiseI8x16S"] = value
	gasMap["I16x8ExtAddPairwiseI8x16U"] = value
	gasMap["I16x8ExtMulHighI8x16S"] = value
	gasMap["I16x8ExtMulHighI8x16U"] = value
	gasMap["I16x8ExtMulLowI8x16S"] = value
	gasMap["I16x8ExtMulLowI8x16U"] = value
	gasMap["I16x8ExtendHighI8x16S"] = value
	gasMap["I16x8ExtendHighI8x16U"] = value
	gasMap["I16x8ExtendLowI8x16S"] = value
	gasMap["I16x8ExtendLowI8x16U"] = value
	gasMap["I16x8ExtractLaneS"] = value
	gasMap["I16x8ExtractLaneU"] = value
	gasMap["I16x8GeS"] = value
	gasMap["I16x8GeU"] = value
	gasMap["I16x8GtS"] = value
	gasMap["I16x8GtU"] = value
	gasMap["I16x8LeS"] = value
	gasMap["I16x8LeU"] = value
	gasMap["I16x8Load8x8S"] = value
	gasMap["I16x8Load8x8U"] = value
	gasMap["I16x8LtS"] = value
	gasMap["I16x8LtU"] = value
	gasMap["I16x8MaxS"] = value
	gasMap["I16x8MaxU"] = value
	gasMap["I16x8MinS"] = value
	gasMap["I16x8MinU"] = value
	gasMap["I16x8Mul"] = value
	gasMap["I16x8NarrowI32x4S"] = value
	gasMap["I16x8NarrowI32x4U"] = value
	gasMap["I16x8Ne"] = value
	gasMap["I16x8Neg"] = value
	gasMap["I16x8Q15MulrSatS"] = value
	gasMap["I16x8ReplaceLane"] = value
	gasMap["I16x8RoundingAverageU"] = value
	gasMap["I16x8Shl"] = value
	gasMap["I16x8ShrS"] = value
	gasMap["I16x8ShrU"] = value
	gasMap["I16x8Splat"] = value
	gasMap["I16x8Sub"] = value
	gasMap["I16x8SubSatS"] = value
	gasMap["I16x8SubSatU"] = value
	gasMap["I16x8SubSaturateS"] = value
	gasMap["I16x8SubSaturateU"] = value
	gasMap["I16x8WidenHighI8x16S"] = value
	gasMap["I16x8WidenHighI8x16U"] = value
	gasMap["I16x8WidenLowI8x16S"] = value
	gasMap["I16x8WidenLowI8x16U"] = value
	gasMap["I32Add"] = value
	gasMap["I32And"] = value
	gasMap["I32AtomicLoad"] = value
	gasMap["I32AtomicLoad16U"] = value
	gasMap["I32AtomicLoad8U"] = value
	gasMap["I32AtomicRmw16AddU"] = value
	gasMap["I32AtomicRmw16AndU"] = value
	gasMap["I32AtomicRmw16CmpxchgU"] = value
	gasMap["I32AtomicRmw16OrU"] = value
	gasMap["I32AtomicRmw16SubU"] = value
	gasMap["I32AtomicRmw16XchgU"] = value
	gasMap["I32AtomicRmw16XorU"] = value
	gasMap["I32AtomicRmw8AddU"] = value
	gasMap["I32AtomicRmw8AndU"] = value
	gasMap["I32AtomicRmw8CmpxchgU"] = value
	gasMap["I32AtomicRmw8OrU"] = value
	gasMap["I32AtomicRmw8SubU"] = value
	gasMap["I32AtomicRmw8XchgU"] = value
	gasMap["I32AtomicRmw8XorU"] = value
	gasMap["I32AtomicRmwAdd"] = value
	gasMap["I32AtomicRmwAnd"] = value
	gasMap["I32AtomicRmwCmpxchg"] = value
	gasMap["I32AtomicRmwOr"] = value
	gasMap["I32AtomicRmwSub"] = value
	gasMap["I32AtomicRmwXchg"] = value
	gasMap["I32AtomicRmwXor"] = value
	gasMap["I32AtomicStore"] = value
	gasMap["I32AtomicStore16"] = value
	gasMap["I32AtomicStore8"] = value
	gasMap["I32AtomicWait"] = value
	gasMap["I32Clz"] = value
	gasMap["I32Const"] = value
	gasMap["I32Ctz"] = value
	gasMap["I32DivS"] = value
	gasMap["I32DivU"] = value
	gasMap["I32Eq"] = value
	gasMap["I32Eqz"] = value
	gasMap["I32Extend16S"] = value
	gasMap["I32Extend8S"] = value
	gasMap["I32GeS"] = value
	gasMap["I32GeU"] = value
	gasMap["I32GtS"] = value
	gasMap["I32GtU"] = value
	gasMap["I32LeS"] = value
	gasMap["I32LeU"] = value
	gasMap["I32Load"] = value
	gasMap["I32Load16S"] = value
	gasMap["I32Load16U"] = value
	gasMap["I32Load8S"] = value
	gasMap["I32Load8U"] = value
	gasMap["I32LtS"] = value
	gasMap["I32LtU"] = value
	gasMap["I32Mul"] = value
	gasMap["I32Ne"] = value
	gasMap["I32Or"] = value
	gasMap["I32Popcnt"] = value
	gasMap["I32ReinterpretF32"] = value
	gasMap["I32RemS"] = value
	gasMap["I32RemU"] = value
	gasMap["I32Rotl"] = value
	gasMap["I32Rotr"] = value
	gasMap["I32Shl"] = value
	gasMap["I32ShrS"] = value
	gasMap["I32ShrU"] = value
	gasMap["I32Store"] = value
	gasMap["I32Store16"] = value
	gasMap["I32Store8"] = value
	gasMap["I32Sub"] = value
	gasMap["I32TruncF32S"] = value
	gasMap["I32TruncF32U"] = value
	gasMap["I32TruncF64S"] = value
	gasMap["I32TruncF64U"] = value
	gasMap["I32TruncSatF32S"] = value
	gasMap["I32TruncSatF32U"] = value
	gasMap["I32TruncSatF64S"] = value
	gasMap["I32TruncSatF64U"] = value
	gasMap["I32WrapI64"] = value
	gasMap["I32Xor"] = value
	gasMap["I32x4Abs"] = value
	gasMap["I32x4Add"] = value
	gasMap["I32x4AllTrue"] = value
	gasMap["I32x4AnyTrue"] = value
	gasMap["I32x4Bitmask"] = value
	gasMap["I32x4DotI16x8S"] = value
	gasMap["I32x4Eq"] = value
	gasMap["I32x4ExtAddPairwiseI16x8S"] = value
	gasMap["I32x4ExtAddPairwiseI16x8U"] = value
	gasMap["I32x4ExtMulHighI16x8S"] = value
	gasMap["I32x4ExtMulHighI16x8U"] = value
	gasMap["I32x4ExtMulLowI16x8S"] = value
	gasMap["I32x4ExtMulLowI16x8U"] = value
	gasMap["I32x4ExtendHighI16x8S"] = value
	gasMap["I32x4ExtendHighI16x8U"] = value
	gasMap["I32x4ExtendLowI16x8S"] = value
	gasMap["I32x4ExtendLowI16x8U"] = value
	gasMap["I32x4ExtractLane"] = value
	gasMap["I32x4GeS"] = value
	gasMap["I32x4GeU"] = value
	gasMap["I32x4GtS"] = value
	gasMap["I32x4GtU"] = value
	gasMap["I32x4LeS"] = value
	gasMap["I32x4LeU"] = value
	gasMap["I32x4Load16x4S"] = value
	gasMap["I32x4Load16x4U"] = value
	gasMap["I32x4LtS"] = value
	gasMap["I32x4LtU"] = value
	gasMap["I32x4MaxS"] = value
	gasMap["I32x4MaxU"] = value
	gasMap["I32x4MinS"] = value
	gasMap["I32x4MinU"] = value
	gasMap["I32x4Mul"] = value
	gasMap["I32x4Ne"] = value
	gasMap["I32x4Neg"] = value
	gasMap["I32x4ReplaceLane"] = value
	gasMap["I32x4Shl"] = value
	gasMap["I32x4ShrS"] = value
	gasMap["I32x4ShrU"] = value
	gasMap["I32x4Splat"] = value
	gasMap["I32x4Sub"] = value
	gasMap["I32x4TruncSatF32x4S"] = value
	gasMap["I32x4TruncSatF32x4U"] = value
	gasMap["I32x4TruncSatF64x2SZero"] = value
	gasMap["I32x4TruncSatF64x2UZero"] = value
	gasMap["I32x4WidenHighI16x8S"] = value
	gasMap["I32x4WidenHighI16x8U"] = value
	gasMap["I32x4WidenLowI16x8S"] = value
	gasMap["I32x4WidenLowI16x8U"] = value
	gasMap["I64Add"] = value
	gasMap["I64And"] = value
	gasMap["I64AtomicLoad"] = value
	gasMap["I64AtomicLoad16U"] = value
	gasMap["I64AtomicLoad32U"] = value
	gasMap["I64AtomicLoad8U"] = value
	gasMap["I64AtomicRmw16AddU"] = value
	gasMap["I64AtomicRmw16AndU"] = value
	gasMap["I64AtomicRmw16CmpxchgU"] = value
	gasMap["I64AtomicRmw16OrU"] = value
	gasMap["I64AtomicRmw16SubU"] = value
	gasMap["I64AtomicRmw16XchgU"] = value
	gasMap["I64AtomicRmw16XorU"] = value
	gasMap["I64AtomicRmw32AddU"] = value
	gasMap["I64AtomicRmw32AndU"] = value
	gasMap["I64AtomicRmw32CmpxchgU"] = value
	gasMap["I64AtomicRmw32OrU"] = value
	gasMap["I64AtomicRmw32SubU"] = value
	gasMap["I64AtomicRmw32XchgU"] = value
	gasMap["I64AtomicRmw32XorU"] = value
	gasMap["I64AtomicRmw8AddU"] = value
	gasMap["I64AtomicRmw8AndU"] = value
	gasMap["I64AtomicRmw8CmpxchgU"] = value
	gasMap["I64AtomicRmw8OrU"] = value
	gasMap["I64AtomicRmw8SubU"] = value
	gasMap["I64AtomicRmw8XchgU"] = value
	gasMap["I64AtomicRmw8XorU"] = value
	gasMap["I64AtomicRmwAdd"] = value
	gasMap["I64AtomicRmwAnd"] = value
	gasMap["I64AtomicRmwCmpxchg"] = value
	gasMap["I64AtomicRmwOr"] = value
	gasMap["I64AtomicRmwSub"] = value
	gasMap["I64AtomicRmwXchg"] = value
	gasMap["I64AtomicRmwXor"] = value
	gasMap["I64AtomicStore"] = value
	gasMap["I64AtomicStore16"] = value
	gasMap["I64AtomicStore32"] = value
	gasMap["I64AtomicStore8"] = value
	gasMap["I64AtomicWait"] = value
	gasMap["I64Clz"] = value
	gasMap["I64Const"] = value
	gasMap["I64Ctz"] = value
	gasMap["I64DivS"] = value
	gasMap["I64DivU"] = value
	gasMap["I64Eq"] = value
	gasMap["I64Eqz"] = value
	gasMap["I64Extend16S"] = value
	gasMap["I64Extend32S"] = value
	gasMap["I64Extend8S"] = value
	gasMap["I64ExtendI32S"] = value
	gasMap["I64ExtendI32U"] = value
	gasMap["I64GeS"] = value
	gasMap["I64GeU"] = value
	gasMap["I64GtS"] = value
	gasMap["I64GtU"] = value
	gasMap["I64LeS"] = value
	gasMap["I64LeU"] = value
	gasMap["I64Load"] = value
	gasMap["I64Load16S"] = value
	gasMap["I64Load16U"] = value
	gasMap["I64Load32S"] = value
	gasMap["I64Load32U"] = value
	gasMap["I64Load8S"] = value
	gasMap["I64Load8U"] = value
	gasMap["I64LtS"] = value
	gasMap["I64LtU"] = value
	gasMap["I64Mul"] = value
	gasMap["I64Ne"] = value
	gasMap["I64Or"] = value
	gasMap["I64Popcnt"] = value
	gasMap["I64ReinterpretF64"] = value
	gasMap["I64RemS"] = value
	gasMap["I64RemU"] = value
	gasMap["I64Rotl"] = value
	gasMap["I64Rotr"] = value
	gasMap["I64Shl"] = value
	gasMap["I64ShrS"] = value
	gasMap["I64ShrU"] = value
	gasMap["I64Store"] = value
	gasMap["I64Store16"] = value
	gasMap["I64Store32"] = value
	gasMap["I64Store8"] = value
	gasMap["I64Sub"] = value
	gasMap["I64TruncF32S"] = value
	gasMap["I64TruncF32U"] = value
	gasMap["I64TruncF64S"] = value
	gasMap["I64TruncF64U"] = value
	gasMap["I64TruncSatF32S"] = value
	gasMap["I64TruncSatF32U"] = value
	gasMap["I64TruncSatF64S"] = value
	gasMap["I64TruncSatF64U"] = value
	gasMap["I64Xor"] = value
	gasMap["I64x2Abs"] = value
	gasMap["I64x2Add"] = value
	gasMap["I64x2AllTrue"] = value
	gasMap["I64x2AnyTrue"] = value
	gasMap["I64x2Bitmask"] = value
	gasMap["I64x2Eq"] = value
	gasMap["I64x2ExtMulHighI32x4S"] = value
	gasMap["I64x2ExtMulHighI32x4U"] = value
	gasMap["I64x2ExtMulLowI32x4S"] = value
	gasMap["I64x2ExtMulLowI32x4U"] = value
	gasMap["I64x2ExtendHighI32x4S"] = value
	gasMap["I64x2ExtendHighI32x4U"] = value
	gasMap["I64x2ExtendLowI32x4S"] = value
	gasMap["I64x2ExtendLowI32x4U"] = value
	gasMap["I64x2ExtractLane"] = value
	gasMap["I64x2GeS"] = value
	gasMap["I64x2GtS"] = value
	gasMap["I64x2LeS"] = value
	gasMap["I64x2Load32x2S"] = value
	gasMap["I64x2Load32x2U"] = value
	gasMap["I64x2LtS"] = value
	gasMap["I64x2Mul"] = value
	gasMap["I64x2Ne"] = value
	gasMap["I64x2Neg"] = value
	gasMap["I64x2ReplaceLane"] = value
	gasMap["I64x2Shl"] = value
	gasMap["I64x2ShrS"] = value
	gasMap["I64x2ShrU"] = value
	gasMap["I64x2Splat"] = value
	gasMap["I64x2Sub"] = value
	gasMap["I64x2TruncSatF64x2S"] = value
	gasMap["I64x2TruncSatF64x2U"] = value
	gasMap["I8x16Abs"] = value
	gasMap["I8x16Add"] = value
	gasMap["I8x16AddSatS"] = value
	gasMap["I8x16AddSatU"] = value
	gasMap["I8x16AddSaturateS"] = value
	gasMap["I8x16AddSaturateU"] = value
	gasMap["I8x16AllTrue"] = value
	gasMap["I8x16AnyTrue"] = value
	gasMap["I8x16Bitmask"] = value
	gasMap["I8x16Eq"] = value
	gasMap["I8x16ExtractLaneS"] = value
	gasMap["I8x16ExtractLaneU"] = value
	gasMap["I8x16GeS"] = value
	gasMap["I8x16GeU"] = value
	gasMap["I8x16GtS"] = value
	gasMap["I8x16GtU"] = value
	gasMap["I8x16LeS"] = value
	gasMap["I8x16LeU"] = value
	gasMap["I8x16LtS"] = value
	gasMap["I8x16LtU"] = value
	gasMap["I8x16MaxS"] = value
	gasMap["I8x16MaxU"] = value
	gasMap["I8x16MinS"] = value
	gasMap["I8x16MinU"] = value
	gasMap["I8x16Mul"] = value
	gasMap["I8x16NarrowI16x8S"] = value
	gasMap["I8x16NarrowI16x8U"] = value
	gasMap["I8x16Ne"] = value
	gasMap["I8x16Neg"] = value
	gasMap["I8x16Popcnt"] = value
	gasMap["I8x16ReplaceLane"] = value
	gasMap["I8x16RoundingAverageU"] = value
	gasMap["I8x16Shl"] = value
	gasMap["I8x16ShrS"] = value
	gasMap["I8x16ShrU"] = value
	gasMap["I8x16Shuffle"] = value
	gasMap["I8x16Splat"] = value
	gasMap["I8x16Sub"] = value
	gasMap["I8x16SubSatS"] = value
	gasMap["I8x16SubSatU"] = value
	gasMap["I8x16SubSaturateS"] = value
	gasMap["I8x16SubSaturateU"] = value
	gasMap["I8x16Swizzle"] = value
	gasMap["If"] = value
	gasMap["LocalGet"] = value
	gasMap["LocalSet"] = value
	gasMap["LocalTee"] = value
	gasMap["LocalAllocate"] = value
	gasMap["LocalsUnmetered"] = value
	gasMap["Loop"] = value
	gasMap["MaxMemoryGrow"] = value
	gasMap["MaxMemoryGrowDelta"] = value
	gasMap["MemoryAtomicNotify"] = value
	gasMap["MemoryAtomicWait32"] = value
	gasMap["MemoryAtomicWait64"] = value
	gasMap["MemoryCopy"] = value
	gasMap["MemoryFill"] = value
	gasMap["MemoryGrow"] = value
	gasMap["MemoryInit"] = value
	gasMap["MemorySize"] = value
	gasMap["Nop"] = value
	gasMap["RefFunc"] = value
	gasMap["RefIsNull"] = value
	gasMap["RefNull"] = value
	gasMap["Rethrow"] = value
	gasMap["Return"] = value
	gasMap["ReturnCall"] = value
	gasMap["ReturnCallIndirect"] = value
	gasMap["Select"] = value
	gasMap["TableCopy"] = value
	gasMap["TableFill"] = value
	gasMap["TableGet"] = value
	gasMap["TableGrow"] = value
	gasMap["TableInit"] = value
	gasMap["TableSet"] = value
	gasMap["TableSize"] = value
	gasMap["Throw"] = value
	gasMap["Try"] = value
	gasMap["TypedSelect"] = value
	gasMap["Unreachable"] = value
	gasMap["Unwind"] = value
	gasMap["V128And"] = value
	gasMap["V128AndNot"] = value
	gasMap["V128AnyTrue"] = value
	gasMap["V128Bitselect"] = value
	gasMap["V128Const"] = value
	gasMap["V128Load"] = value
	gasMap["V128Load16Lane"] = value
	gasMap["V128Load16Splat"] = value
	gasMap["V128Load16x4S"] = value
	gasMap["V128Load16x4U"] = value
	gasMap["V128Load32Lane"] = value
	gasMap["V128Load32Splat"] = value
	gasMap["V128Load32Zero"] = value
	gasMap["V128Load32x2S"] = value
	gasMap["V128Load32x2U"] = value
	gasMap["V128Load64Lane"] = value
	gasMap["V128Load64Splat"] = value
	gasMap["V128Load64Zero"] = value
	gasMap["V128Load8Lane"] = value
	gasMap["V128Load8Splat"] = value
	gasMap["V128Load8x8S"] = value
	gasMap["V128Load8x8U"] = value
	gasMap["V128Not"] = value
	gasMap["V128Or"] = value
	gasMap["V128Store"] = value
	gasMap["V128Store16Lane"] = value
	gasMap["V128Store32Lane"] = value
	gasMap["V128Store64Lane"] = value
	gasMap["V128Store8Lane"] = value
	gasMap["V128Xor"] = value
	gasMap["V16x8LoadSplat"] = value
	gasMap["V32x4LoadSplat"] = value
	gasMap["V64x2LoadSplat"] = value
	gasMap["V8x16LoadSplat"] = value
	gasMap["V8x16Shuffle"] = value
	gasMap["V8x16Swizzle"] = value

	return gasMap
}

// FillGasMapDynamicStorageLoad populates the gas map with the coefficients needed for dynamic storage load
func FillGasMapDynamicStorageLoad() map[string]uint64 {
	gasMap := make(map[string]uint64)

	gasMap["QuadraticCoefficient"] = 688
	gasMap["SignOfQuadratic"] = 0
	gasMap["LinearCoefficient"] = 31858
	gasMap["SignOfLinear"] = 0
	gasMap["ConstantCoefficient"] = 15287
	gasMap["SignOfConstant"] = 0

	return gasMap
}

// MakeGasMapForTests creates a gas schedule map for tests
func MakeGasMapForTests() GasScheduleMap {
	return MakeGasMap(GasValueForTests, AsyncCallbackGasLockForTests)
}

func customFillGasMapWASMOpcodeCosts(gasMap map[string]uint64) {
	gasMap["LocalsUnmetered"] = 100
	gasMap["MaxMemoryGrow"] = 8
	gasMap["MaxMemoryGrowDelta"] = 10
}
