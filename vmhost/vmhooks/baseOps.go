package vmhooks

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	"github.com/multiversx/mx-chain-vm-go/executor"
	"github.com/multiversx/mx-chain-vm-go/math"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
)

const (
	getSCAddressName                 = "getSCAddress"
	getOwnerAddressName              = "getOwnerAddress"
	getShardOfAddressName            = "getShardOfAddress"
	isSmartContractName              = "isSmartContract"
	getExternalBalanceName           = "getExternalBalance"
	blockHashName                    = "blockHash"
	transferValueName                = "transferValue"
	transferESDTExecuteName          = "transferESDTExecute"
	transferESDTNFTExecuteName       = "transferESDTNFTExecute"
	multiTransferESDTNFTExecuteName  = "multiTransferESDTNFTExecute"
	transferValueExecuteName         = "transferValueExecute"
	createAsyncCallName              = "createAsyncCall"
	setAsyncGroupCallbackName        = "setAsyncGroupCallback"
	setAsyncContextCallbackName      = "setAsyncContextCallback"
	getArgumentLengthName            = "getArgumentLength"
	getArgumentName                  = "getArgument"
	getFunctionName                  = "getFunction"
	getNumArgumentsName              = "getNumArguments"
	storageStoreName                 = "storageStore"
	storageLoadLengthName            = "storageLoadLength"
	storageLoadName                  = "storageLoad"
	storageLoadFromAddressName       = "storageLoadFromAddress"
	getCallerName                    = "getCaller"
	checkNoPaymentName               = "checkNoPayment"
	callValueName                    = "callValue"
	getESDTValueName                 = "getESDTValue"
	getESDTTokenNameName             = "getESDTTokenName"
	getESDTTokenNonceName            = "getESDTTokenNonce"
	getESDTTokenTypeName             = "getESDTTokenType"
	getCallValueTokenNameName        = "getCallValueTokenName"
	getESDTValueByIndexName          = "getESDTValueByIndex"
	getESDTTokenNameByIndexName      = "getESDTTokenNameByIndex"
	getESDTTokenNonceByIndexName     = "getESDTTokenNonceByIndex"
	getESDTTokenTypeByIndexName      = "getESDTTokenTypeByIndex"
	getCallValueTokenNameByIndexName = "getCallValueTokenNameByIndex"
	getNumESDTTransfersName          = "getNumESDTTransfers"
	getCurrentESDTNFTNonceName       = "getCurrentESDTNFTNonce"
	writeLogName                     = "writeLog"
	writeEventLogName                = "writeEventLog"
	returnDataName                   = "returnData"
	signalErrorName                  = "signalError"
	getGasLeftName                   = "getGasLeft"
	getESDTBalanceName               = "getESDTBalance"
	getESDTNFTNameLengthName         = "getESDTNFTNameLength"
	getESDTNFTAttributeLengthName    = "getESDTNFTAttributeLength"
	getESDTNFTURILengthName          = "getESDTNFTURILength"
	getESDTTokenDataName             = "getESDTTokenData"
	getESDTLocalRolesName            = "getESDTLocalRoles"
	validateTokenIdentifierName      = "validateTokenIdentifier"
	executeOnDestContextName         = "executeOnDestContext"
	executeOnSameContextName         = "executeOnSameContext"
	executeReadOnlyName              = "executeReadOnly"
	createContractName               = "createContract"
	deployFromSourceContractName     = "deployFromSourceContract"
	upgradeContractName              = "upgradeContract"
	upgradeFromSourceContractName    = "upgradeFromSourceContract"
	deleteContractName               = "deleteContract"
	asyncCallName                    = "asyncCall"
	getNumReturnDataName             = "getNumReturnData"
	getReturnDataSizeName            = "getReturnDataSize"
	getReturnDataName                = "getReturnData"
	cleanReturnDataName              = "cleanReturnData"
	deleteFromReturnDataName         = "deleteFromReturnData"
	setStorageLockName               = "setStorageLock"
	getStorageLockName               = "getStorageLock"
	isStorageLockedName              = "isStorageLocked"
	clearStorageLockName             = "clearStorageLock"
	getBlockTimestampName            = "getBlockTimestamp"
	getBlockNonceName                = "getBlockNonce"
	getBlockRoundName                = "getBlockRound"
	getBlockEpochName                = "getBlockEpoch"
	getBlockRandomSeedName           = "getBlockRandomSeed"
	getStateRootHashName             = "getStateRootHash"
	getPrevBlockTimestampName        = "getPrevBlockTimestamp"
	getPrevBlockNonceName            = "getPrevBlockNonce"
	getPrevBlockRoundName            = "getPrevBlockRound"
	getPrevBlockEpochName            = "getPrevBlockEpoch"
	getPrevBlockRandomSeedName       = "getPrevBlockRandomSeed"
	getOriginalTxHashName            = "getOriginalTxHash"
	getCurrentTxHashName             = "getCurrentTxHash"
	getPrevTxHashName                = "getPrevTxHash"
	getRoundTimeName                 = "getRoundTime"
	epochStartBlockTimeStampName     = "epochStartBlockTimeStamp"
	epochStartBlockNonceName         = "epochStartBlockNonce"
	epochStartBlockRoundName         = "epochStartBlockRound"
)

type CreateContractCallType int

const (
	CreateContract = iota
	DeployContract
)

var logEEI = logger.GetOrCreate("vm/eei")

func getESDTTransferFromInputFailIfWrongIndex(host vmhost.VMHost, index int32) *vmcommon.ESDTTransfer {
	esdtTransfers := host.Runtime().GetVMInput().ESDTTransfers
	if int32(len(esdtTransfers))-1 < index || index < 0 {
		FailExecution(host, vmhost.ErrInvalidTokenIndex)
		return nil
	}
	return esdtTransfers[index]
}

func failIfMoreThanOneESDTTransfer(context *VMHooksImpl) bool {
	runtime := context.GetRuntimeContext()
	if len(runtime.GetVMInput().ESDTTransfers) > 1 {
		FailExecution(context.GetVMHost(), vmhost.ErrTooManyESDTTransfers)
		return true
	}
	return false
}

// GetGasLeft VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetGasLeft() int64 {
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetGasLeft
	err := metering.UseGasBoundedAndAddTracedGas(getGasLeftName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return 0
	}

	return int64(metering.GasLeft())
}

// GetSCAddress VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetSCAddress(resultOffset executor.MemPtr) {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetSCAddress
	err := metering.UseGasBoundedAndAddTracedGas(getSCAddressName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return
	}

	owner := runtime.GetContextAddress()
	err = context.MemStore(resultOffset, owner)
	if err != nil {
		context.FailExecution(err)
		return
	}
}

// GetOwnerAddress VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetOwnerAddress(resultOffset executor.MemPtr) {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetOwnerAddress
	err := metering.UseGasBoundedAndAddTracedGas(getOwnerAddressName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return
	}

	owner, err := blockchain.GetOwnerAddress()
	if err != nil {
		context.FailExecution(err)
		return
	}

	err = context.MemStore(resultOffset, owner)
	if err != nil {
		context.FailExecution(err)
		return
	}
}

// GetShardOfAddress VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetShardOfAddress(addressOffset executor.MemPtr) int32 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetShardOfAddress
	err := metering.UseGasBoundedAndAddTracedGas(getShardOfAddressName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	address, err := context.MemLoad(addressOffset, vmhost.AddressLen)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int32(blockchain.GetShardOfAddress(address))
}

// IsSmartContract VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) IsSmartContract(addressOffset executor.MemPtr) int32 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.IsSmartContract
	err := metering.UseGasBoundedAndAddTracedGas(isSmartContractName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	address, err := context.MemLoad(addressOffset, vmhost.AddressLen)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	isSmartContract := blockchain.IsSmartContract(address)

	return int32(vmhost.BooleanToInt(isSmartContract))
}

// SignalError VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) SignalError(messageOffset executor.MemPtr, messageLength executor.MemLength) {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()
	metering.StartGasTracing(signalErrorName)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.SignalError
	gasToUse += metering.GasSchedule().BaseOperationCost.PersistPerByte * uint64(messageLength)

	err := metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		context.FailExecution(err)
		return
	}

	message, err := context.MemLoad(messageOffset, messageLength)
	if err != nil {
		context.FailExecution(err)
		return
	}

	runtime.SignalUserError(string(message))
}

// GetExternalBalance VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetExternalBalance(addressOffset executor.MemPtr, resultOffset executor.MemPtr) {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetExternalBalance
	err := metering.UseGasBoundedAndAddTracedGas(getExternalBalanceName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return
	}

	address, err := context.MemLoad(addressOffset, vmhost.AddressLen)
	if err != nil {
		context.FailExecution(err)
		return
	}

	balance := blockchain.GetBalance(address)

	err = context.MemStore(resultOffset, balance)
	if err != nil {
		context.FailExecution(err)
		return
	}
}

// GetBlockHash VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetBlockHash(nonce int64, resultOffset executor.MemPtr) int32 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetBlockHash
	err := metering.UseGasBoundedAndAddTracedGas(blockHashName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return 1
	}

	hash := blockchain.BlockHash(uint64(nonce))
	err = context.MemStore(resultOffset, hash)
	if err != nil {
		context.FailExecution(err)
		return 1
	}

	return 0
}

func getESDTDataFromBlockchainHook(
	context *VMHooksImpl,
	addressOffset executor.MemPtr,
	tokenIDOffset executor.MemPtr,
	tokenIDLen executor.MemLength,
	nonce int64,
) (*esdt.ESDigitalToken, error) {
	metering := context.GetMeteringContext()
	blockchain := context.GetBlockchainContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetExternalBalance
	err := metering.UseGasBounded(gasToUse)
	if err != nil {
		return nil, err
	}

	address, err := context.MemLoad(addressOffset, vmhost.AddressLen)
	if err != nil {
		return nil, err
	}

	tokenID, err := context.MemLoad(tokenIDOffset, tokenIDLen)
	if err != nil {
		return nil, err
	}

	esdtToken, err := blockchain.GetESDTToken(address, tokenID, uint64(nonce))
	if err != nil {
		return nil, err
	}

	return esdtToken, nil
}

// GetESDTBalance VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetESDTBalance(
	addressOffset executor.MemPtr,
	tokenIDOffset executor.MemPtr,
	tokenIDLen executor.MemLength,
	nonce int64,
	resultOffset executor.MemPtr,
) int32 {
	metering := context.GetMeteringContext()
	metering.StartGasTracing(getESDTBalanceName)

	esdtData, err := getESDTDataFromBlockchainHook(context, addressOffset, tokenIDOffset, tokenIDLen, nonce)

	if err != nil {
		context.FailExecution(err)
		return -1
	}
	err = context.MemStore(resultOffset, esdtData.Value.Bytes())
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int32(len(esdtData.Value.Bytes()))
}

// GetESDTNFTNameLength VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetESDTNFTNameLength(
	addressOffset executor.MemPtr,
	tokenIDOffset executor.MemPtr,
	tokenIDLen executor.MemLength,
	nonce int64,
) int32 {
	metering := context.GetMeteringContext()
	metering.StartGasTracing(getESDTNFTNameLengthName)

	esdtData, err := getESDTDataFromBlockchainHook(context, addressOffset, tokenIDOffset, tokenIDLen, nonce)

	if err != nil {
		context.FailExecution(err)
		return -1
	}
	if esdtData == nil || esdtData.TokenMetaData == nil {
		FailExecution(context.GetVMHost(), vmhost.ErrNilESDTData)
		return 0
	}

	return int32(len(esdtData.TokenMetaData.Name))
}

// GetESDTNFTAttributeLength VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetESDTNFTAttributeLength(
	addressOffset executor.MemPtr,
	tokenIDOffset executor.MemPtr,
	tokenIDLen executor.MemLength,
	nonce int64,
) int32 {
	metering := context.GetMeteringContext()
	metering.StartGasTracing(getESDTNFTAttributeLengthName)

	esdtData, err := getESDTDataFromBlockchainHook(context, addressOffset, tokenIDOffset, tokenIDLen, nonce)

	if err != nil {
		context.FailExecution(err)
		return -1
	}
	if esdtData == nil || esdtData.TokenMetaData == nil {
		FailExecution(context.GetVMHost(), vmhost.ErrNilESDTData)
		return 0
	}

	return int32(len(esdtData.TokenMetaData.Attributes))
}

// GetESDTNFTURILength VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetESDTNFTURILength(
	addressOffset executor.MemPtr,
	tokenIDOffset executor.MemPtr,
	tokenIDLen executor.MemLength,
	nonce int64,
) int32 {
	metering := context.GetMeteringContext()
	metering.StartGasTracing(getESDTNFTURILengthName)

	esdtData, err := getESDTDataFromBlockchainHook(context, addressOffset, tokenIDOffset, tokenIDLen, nonce)

	if err != nil {
		context.FailExecution(err)
		return -1
	}
	if esdtData == nil || esdtData.TokenMetaData == nil {
		FailExecution(context.GetVMHost(), vmhost.ErrNilESDTData)
		return 0
	}
	if len(esdtData.TokenMetaData.URIs) == 0 {
		return 0
	}

	return int32(len(esdtData.TokenMetaData.URIs[0]))
}

// GetESDTTokenData VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetESDTTokenData(
	addressOffset executor.MemPtr,
	tokenIDOffset executor.MemPtr,
	tokenIDLen executor.MemLength,
	nonce int64,
	valueHandle int32,
	propertiesOffset executor.MemPtr,
	hashOffset executor.MemPtr,
	nameOffset executor.MemPtr,
	attributesOffset executor.MemPtr,
	creatorOffset executor.MemPtr,
	royaltiesHandle int32,
	urisOffset executor.MemPtr,
) int32 {
	managedType := context.GetManagedTypesContext()
	metering := context.GetMeteringContext()
	metering.StartGasTracing(getESDTTokenDataName)

	esdtData, err := getESDTDataFromBlockchainHook(context, addressOffset, tokenIDOffset, tokenIDLen, nonce)

	if err != nil {
		context.FailExecution(err)
		return -1
	}

	value := managedType.GetBigIntOrCreate(valueHandle)
	value.Set(esdtData.Value)

	err = context.MemStore(propertiesOffset, esdtData.Properties)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	if esdtData.TokenMetaData != nil {
		err = context.MemStore(hashOffset, esdtData.TokenMetaData.Hash)
		if err != nil {
			context.FailExecution(err)
			return -1
		}
		err = context.MemStore(nameOffset, esdtData.TokenMetaData.Name)
		if err != nil {
			context.FailExecution(err)
			return -1
		}
		err = context.MemStore(attributesOffset, esdtData.TokenMetaData.Attributes)
		if err != nil {
			context.FailExecution(err)
			return -1
		}
		err = context.MemStore(creatorOffset, esdtData.TokenMetaData.Creator)
		if err != nil {
			context.FailExecution(err)
			return -1
		}

		royalties := managedType.GetBigIntOrCreate(royaltiesHandle)
		royalties.SetUint64(uint64(esdtData.TokenMetaData.Royalties))

		if len(esdtData.TokenMetaData.URIs) > 0 {
			err = context.MemStore(urisOffset, esdtData.TokenMetaData.URIs[0])
			if err != nil {
				context.FailExecution(err)
				return -1
			}
		}
	}
	return int32(len(esdtData.Value.Bytes()))
}

// GetESDTLocalRoles VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetESDTLocalRoles(tokenIdHandle int32) int64 {
	managedType := context.GetManagedTypesContext()
	storage := context.GetStorageContext()
	metering := context.GetMeteringContext()

	tokenID, err := managedType.GetBytes(tokenIdHandle)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	esdtRoleKeyPrefix := []byte(core.ProtectedKeyPrefix + core.ESDTRoleIdentifier + core.ESDTKeyIdentifier)
	key := []byte(string(esdtRoleKeyPrefix) + string(tokenID))

	data, trieDepth, usedCache, err := storage.GetStorage(key)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	err = storage.UseGasForStorageLoad(
		storageLoadName,
		int64(trieDepth),
		metering.GasSchedule().BaseOpsAPICost.StorageLoad,
		usedCache)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	enableEpochsHandler := context.host.EnableEpochsHandler()
	return getESDTRoles(data, enableEpochsHandler.IsFlagEnabled(vmhost.CryptoOpcodesV2Flag))
}

// ValidateTokenIdentifier VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) ValidateTokenIdentifier(
	tokenIdHandle int32,
) int32 {
	managedType := context.GetManagedTypesContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetArgument
	err := metering.UseGasBoundedAndAddTracedGas(validateTokenIdentifierName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	tokenID, err := managedType.GetBytes(tokenIdHandle)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	if ValidateToken(tokenID) {
		return 1
	} else {
		return 0
	}

}

// TransferValue VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) TransferValue(
	destOffset executor.MemPtr,
	valueOffset executor.MemPtr,
	dataOffset executor.MemPtr,
	length executor.MemLength) int32 {

	host := context.GetVMHost()
	runtime := host.Runtime()
	metering := host.Metering()
	output := host.Output()
	metering.StartGasTracing(transferValueName)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.TransferValue
	err := metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		context.FailExecution(err)
		return 1
	}

	sender := runtime.GetContextAddress()
	dest, err := context.MemLoad(destOffset, vmhost.AddressLen)
	if err != nil {
		context.FailExecution(err)
		return 1
	}

	valueBytes, err := context.MemLoad(valueOffset, vmhost.BalanceLen)
	if err != nil {
		context.FailExecution(err)
		return 1
	}

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.PersistPerByte, uint64(length))
	err = metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		context.FailExecution(err)
		return 1
	}

	data, err := context.MemLoad(dataOffset, length)
	if err != nil {
		context.FailExecution(err)
		return 1
	}

	if host.IsBuiltinFunctionCall(data) {
		context.FailExecution(vmhost.ErrTransferValueOnESDTCall)
		return 1
	}

	err = output.Transfer(dest, sender, 0, 0, big.NewInt(0).SetBytes(valueBytes), nil, data, vm.DirectCall)
	if err != nil {
		context.FailExecution(err)
		return 1
	}

	return 0
}

type indirectContractCallArguments struct {
	dest      []byte
	value     *big.Int
	function  []byte
	args      [][]byte
	actualLen int32
}

func (context *VMHooksImpl) extractIndirectContractCallArgumentsWithValue(
	host vmhost.VMHost,
	destOffset executor.MemPtr,
	valueOffset executor.MemPtr,
	functionOffset executor.MemPtr,
	functionLength executor.MemLength,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) (*indirectContractCallArguments, error) {
	return context.extractIndirectContractCallArguments(
		host,
		destOffset,
		valueOffset,
		true,
		functionOffset,
		functionLength,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
}

func (context *VMHooksImpl) extractIndirectContractCallArgumentsWithoutValue(
	host vmhost.VMHost,
	destOffset executor.MemPtr,
	functionOffset executor.MemPtr,
	functionLength executor.MemLength,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) (*indirectContractCallArguments, error) {
	return context.extractIndirectContractCallArguments(
		host,
		destOffset,
		0,
		false,
		functionOffset,
		functionLength,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
}

func (context *VMHooksImpl) extractIndirectContractCallArguments(
	host vmhost.VMHost,
	destOffset executor.MemPtr,
	valueOffset executor.MemPtr,
	hasValueOffset bool,
	functionOffset executor.MemPtr,
	functionLength executor.MemLength,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) (*indirectContractCallArguments, error) {
	metering := host.Metering()

	dest, err := context.MemLoad(destOffset, vmhost.AddressLen)
	if err != nil {
		return nil, err
	}

	var value *big.Int

	if hasValueOffset {
		valueBytes, err := context.MemLoad(valueOffset, vmhost.BalanceLen)
		if err != nil {
			return nil, err
		}
		value = big.NewInt(0).SetBytes(valueBytes)
	}

	function, err := context.MemLoad(functionOffset, functionLength)
	if err != nil {
		return nil, err
	}

	args, actualLen, err := context.getArgumentsFromMemory(
		host,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
	if err != nil {
		return nil, err
	}

	gasToUse := math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(actualLen))
	err = metering.UseGasBounded(gasToUse)
	if err != nil && host.Runtime().UseGasBoundedShouldFailExecution() {
		return nil, err
	}

	return &indirectContractCallArguments{
		dest:      dest,
		value:     value,
		function:  function,
		args:      args,
		actualLen: actualLen,
	}, nil
}

// TransferValueExecute VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) TransferValueExecute(
	destOffset executor.MemPtr,
	valueOffset executor.MemPtr,
	gasLimit int64,
	functionOffset executor.MemPtr,
	functionLength executor.MemLength,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) int32 {
	host := context.GetVMHost()
	return context.TransferValueExecuteWithHost(
		host,
		destOffset,
		valueOffset,
		gasLimit,
		functionOffset,
		functionLength,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
}

// TransferValueExecuteWithHost - transferValueExecute with host instead of pointer context
func (context *VMHooksImpl) TransferValueExecuteWithHost(
	host vmhost.VMHost,
	destOffset executor.MemPtr,
	valueOffset executor.MemPtr,
	gasLimit int64,
	functionOffset executor.MemPtr,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) int32 {
	runtime := host.Runtime()
	metering := host.Metering()
	metering.StartGasTracing(transferValueExecuteName)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.TransferValue
	err := metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		context.FailExecution(err)
		return 1
	}

	callArgs, err := context.extractIndirectContractCallArgumentsWithValue(
		host, destOffset, valueOffset, functionOffset, functionLength, numArguments, argumentsLengthOffset, dataOffset)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	return TransferValueExecuteWithTypedArgs(
		host,
		callArgs.dest,
		callArgs.value,
		gasLimit,
		callArgs.function,
		callArgs.args,
	)
}

// TransferValueExecuteWithTypedArgs - transferValueExecute with args already read from memory
func TransferValueExecuteWithTypedArgs(
	host vmhost.VMHost,
	dest []byte,
	value *big.Int,
	gasLimit int64,
	function []byte,
	args [][]byte,
) int32 {
	runtime := host.Runtime()
	metering := host.Metering()
	output := host.Output()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.TransferValue
	err := metering.UseGasBounded(gasToUse)

	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return 1
	}

	sender := runtime.GetContextAddress()

	var contractCallInput *vmcommon.ContractCallInput
	if len(function) > 0 {
		contractCallInput, err = prepareIndirectContractCallInput(
			host,
			sender,
			value,
			gasLimit,
			dest,
			function,
			args,
			gasToUse,
			false,
		)
		if err != nil {
			FailExecution(host, err)
			return 1
		}
	}

	if contractCallInput != nil {
		if host.IsBuiltinFunctionName(contractCallInput.Function) {
			FailExecution(host, vmhost.ErrNilESDTData)
			return 1
		}
	}

	if host.AreInSameShard(sender, dest) && contractCallInput != nil && host.Blockchain().IsSmartContract(dest) {
		logEEI.Trace("eGLD pre-transfer execution begin")
		vmOutput, err := executeOnDestContextFromAPI(host, contractCallInput)
		if err != nil {
			logEEI.Trace("eGLD pre-transfer execution failed", "error", err)
			FailExecution(host, err)
			return 1
		}
		host.CompleteLogEntriesWithCallType(vmOutput, vmhost.TransferAndExecuteString)

		return 0
	}

	data := ""
	if contractCallInput != nil {
		data = makeCrossShardCallFromInput(contractCallInput.Function, contractCallInput.Arguments)
	}

	err = metering.UseGasBounded(uint64(gasLimit))
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return 1
	}

	err = output.Transfer(dest, sender, uint64(gasLimit), 0, value, nil, []byte(data), vm.DirectCall)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	return 0
}

func makeCrossShardCallFromInput(function string, arguments [][]byte) string {
	txData := function
	for _, arg := range arguments {
		txData += "@" + hex.EncodeToString(arg)
	}

	return txData
}

// TransferESDTExecute VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) TransferESDTExecute(
	destOffset executor.MemPtr,
	tokenIDOffset executor.MemPtr,
	tokenIDLen executor.MemLength,
	valueOffset executor.MemPtr,
	gasLimit int64,
	functionOffset executor.MemPtr,
	functionLength executor.MemLength,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) int32 {

	return context.TransferESDTNFTExecute(destOffset, tokenIDOffset, tokenIDLen, valueOffset, 0,
		gasLimit, functionOffset, functionLength, numArguments, argumentsLengthOffset, dataOffset)
}

// TransferESDTNFTExecute VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) TransferESDTNFTExecute(
	destOffset executor.MemPtr,
	tokenIDOffset executor.MemPtr,
	tokenIDLen executor.MemLength,
	valueOffset executor.MemPtr,
	nonce int64,
	gasLimit int64,
	functionOffset executor.MemPtr,
	functionLength executor.MemLength,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) int32 {
	host := context.GetVMHost()
	metering := host.Metering()
	metering.StartGasTracing(transferESDTNFTExecuteName)
	return context.TransferESDTNFTExecuteWithHost(
		host,
		destOffset,
		tokenIDOffset,
		tokenIDLen,
		valueOffset,
		nonce,
		gasLimit,
		functionOffset,
		functionLength,
		numArguments,
		argumentsLengthOffset,
		dataOffset)
}

// MultiTransferESDTNFTExecute VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) MultiTransferESDTNFTExecute(
	destOffset executor.MemPtr,
	numTokenTransfers int32,
	tokenTransfersArgsLengthOffset executor.MemPtr,
	tokenTransferDataOffset executor.MemPtr,
	gasLimit int64,
	functionOffset executor.MemPtr,
	functionLength executor.MemLength,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) int32 {
	host := context.GetVMHost()
	runtime := host.Runtime()
	metering := host.Metering()
	metering.StartGasTracing(multiTransferESDTNFTExecuteName)

	if numTokenTransfers == 0 {
		FailExecution(host, vmhost.ErrFailedTransfer)
		return 1
	}

	callArgs, err := context.extractIndirectContractCallArgumentsWithoutValue(
		host, destOffset, functionOffset, functionLength, numArguments, argumentsLengthOffset, dataOffset)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	gasToUse := math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(callArgs.actualLen))
	err = metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return 1
	}

	transferArgs, actualLen, err := context.getArgumentsFromMemory(
		host,
		numTokenTransfers*parsers.ArgsPerTransfer,
		tokenTransfersArgsLengthOffset,
		tokenTransferDataOffset,
	)

	if err != nil {
		FailExecution(host, err)
		return 1
	}

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(actualLen))
	err = metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return 1
	}

	transfers := make([]*vmcommon.ESDTTransfer, numTokenTransfers)
	for i := int32(0); i < numTokenTransfers; i++ {
		tokenStartIndex := i * parsers.ArgsPerTransfer
		transfer := &vmcommon.ESDTTransfer{
			ESDTTokenName:  transferArgs[tokenStartIndex],
			ESDTTokenNonce: big.NewInt(0).SetBytes(transferArgs[tokenStartIndex+1]).Uint64(),
			ESDTValue:      big.NewInt(0).SetBytes(transferArgs[tokenStartIndex+2]),
			ESDTTokenType:  uint32(core.Fungible),
		}
		if transfer.ESDTTokenNonce > 0 {
			transfer.ESDTTokenType = uint32(core.NonFungible)
		}
		transfers[i] = transfer
	}

	return TransferESDTNFTExecuteWithTypedArgs(
		host,
		callArgs.dest,
		transfers,
		gasLimit,
		callArgs.function,
		callArgs.args,
	)
}

// TransferESDTNFTExecuteWithHost contains only memory reading of arguments
func (context *VMHooksImpl) TransferESDTNFTExecuteWithHost(
	host vmhost.VMHost,
	destOffset executor.MemPtr,
	tokenIDOffset executor.MemPtr,
	tokenIDLen executor.MemLength,
	valueOffset executor.MemPtr,
	nonce int64,
	gasLimit int64,
	functionOffset executor.MemPtr,
	functionLength executor.MemLength,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) int32 {
	runtime := host.Runtime()
	metering := host.Metering()

	tokenIdentifier, executeErr := context.MemLoad(tokenIDOffset, tokenIDLen)
	if executeErr != nil {
		FailExecution(host, executeErr)
		return 1
	}

	callArgs, err := context.extractIndirectContractCallArgumentsWithValue(
		host, destOffset, valueOffset, functionOffset, functionLength, numArguments, argumentsLengthOffset, dataOffset)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	gasToUse := math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(callArgs.actualLen))
	err = metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return 1
	}

	transfer := &vmcommon.ESDTTransfer{
		ESDTValue:      callArgs.value,
		ESDTTokenName:  tokenIdentifier,
		ESDTTokenNonce: uint64(nonce),
		ESDTTokenType:  uint32(core.Fungible),
	}
	if nonce > 0 {
		transfer.ESDTTokenType = uint32(core.NonFungible)
	}
	return TransferESDTNFTExecuteWithTypedArgs(
		host,
		callArgs.dest,
		[]*vmcommon.ESDTTransfer{transfer},
		gasLimit,
		callArgs.function,
		callArgs.args,
	)
}

// TransferESDTNFTExecuteWithTypedArgs defines the actual transfer ESDT execute logic
func TransferESDTNFTExecuteWithTypedArgs(
	host vmhost.VMHost,
	dest []byte,
	transfers []*vmcommon.ESDTTransfer,
	gasLimit int64,
	function []byte,
	data [][]byte,
) int32 {
	var executeErr error

	runtime := host.Runtime()
	metering := host.Metering()

	output := host.Output()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.TransferValue * uint64(len(transfers))
	err := metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return 1
	}

	sender := runtime.GetContextAddress()

	var contractCallInput *vmcommon.ContractCallInput
	if len(function) > 0 {
		contractCallInput, executeErr = prepareIndirectContractCallInput(
			host,
			sender,
			big.NewInt(0),
			gasLimit,
			dest,
			function,
			data,
			gasToUse,
			false,
		)
		if executeErr != nil {
			FailExecution(host, executeErr)
			return 1
		}

		contractCallInput.ESDTTransfers = transfers
	}

	snapshotBeforeTransfer := host.Blockchain().GetSnapshot()

	originalCaller := host.Runtime().GetOriginalCallerAddress()
	transfersArgs := &vmhost.ESDTTransfersArgs{
		Destination:    dest,
		OriginalCaller: originalCaller,
		Sender:         sender,
		Transfers:      transfers,
		SenderForExec:  sender,
	}
	gasLimitForExec, executeErr := output.TransferESDT(transfersArgs, contractCallInput)
	if executeErr != nil {
		FailExecution(host, executeErr)
		return 1
	}

	if host.AreInSameShard(sender, dest) && contractCallInput != nil && host.Blockchain().IsSmartContract(dest) {
		contractCallInput.GasProvided = gasLimitForExec
		contractCallInput.CallerAddr = sender
		logEEI.Trace("ESDT post-transfer execution begin")
		_, executeErr := executeOnDestContextFromAPI(host, contractCallInput)
		if executeErr != nil {
			logEEI.Trace("ESDT post-transfer execution failed", "error", executeErr)
			host.Blockchain().RevertToSnapshot(snapshotBeforeTransfer)
			FailExecution(host, executeErr)
			return 1
		}

		return 0
	}

	return 0

}

// TransferESDTNFTExecuteByUserWithTypedArgs defines the actual transfer ESDT execute logic and execution
func TransferESDTNFTExecuteByUserWithTypedArgs(
	host vmhost.VMHost,
	callerForExecution []byte,
	dest []byte,
	transfers []*vmcommon.ESDTTransfer,
	gasLimit int64,
	function []byte,
	data [][]byte,
) int32 {
	var executeErr error

	runtime := host.Runtime()
	metering := host.Metering()

	output := host.Output()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.TransferValue * uint64(len(transfers))
	err := metering.UseGasBounded(gasToUse)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	sender := runtime.GetContextAddress()

	var contractCallInput *vmcommon.ContractCallInput
	if len(function) > 0 {
		contractCallInput, executeErr = prepareIndirectContractCallInput(
			host,
			sender,
			big.NewInt(0),
			gasLimit,
			dest,
			function,
			data,
			gasToUse,
			false,
		)
		if executeErr != nil {
			FailExecution(host, executeErr)
			return 1
		}

		contractCallInput.ESDTTransfers = transfers
	}

	originalCaller := host.Runtime().GetOriginalCallerAddress()
	transfersArgs := &vmhost.ESDTTransfersArgs{
		Destination:    dest,
		OriginalCaller: originalCaller,
		Sender:         sender,
		Transfers:      transfers,
		SenderForExec:  callerForExecution,
	}
	gasLimitForExec, executeErr := output.TransferESDT(transfersArgs, contractCallInput)
	if err != nil {
		FailExecution(host, executeErr)
		return 1
	}

	if host.AreInSameShard(sender, dest) && contractCallInput != nil && host.Blockchain().IsSmartContract(dest) {
		contractCallInput.GasProvided = gasLimitForExec
		contractCallInput.CallerAddr = callerForExecution
		logEEI.Trace("ESDT post-transfer execution begin")
		_, executeErr = executeOnDestContextFromAPI(host, contractCallInput)
		if executeErr != nil {
			logEEI.Trace("ESDT post-transfer execution failed, started transfer to user", "error", executeErr)

			// in case of failed execution, the funds have to be moved to the user
			returnTransferArgs := &vmhost.ESDTTransfersArgs{
				Destination:      callerForExecution,
				OriginalCaller:   originalCaller,
				Sender:           dest,
				Transfers:        transfers,
				SenderForExec:    dest,
				ReturnAfterError: true,
			}
			_, executeErr = output.TransferESDT(returnTransferArgs, nil)
			if err != nil {
				FailExecution(host, executeErr)
				return 1
			}

			return 1
		}

		return 0
	}

	return 0
}

// CreateAsyncCall VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) CreateAsyncCall(
	destOffset executor.MemPtr,
	valueOffset executor.MemPtr,
	dataOffset executor.MemPtr,
	dataLength executor.MemLength,
	successOffset executor.MemPtr,
	successLength executor.MemLength,
	errorOffset executor.MemPtr,
	errorLength executor.MemLength,
	gas int64,
	extraGasForCallback int64,
) int32 {
	host := context.GetVMHost()
	return context.CreateAsyncCallWithHost(
		host,
		destOffset,
		valueOffset,
		dataOffset,
		dataLength,
		successOffset,
		successLength,
		errorOffset,
		errorLength,
		gas,
		extraGasForCallback)
}

// CreateAsyncCallWithHost - createAsyncCall with host instead of pointer
func (context *VMHooksImpl) CreateAsyncCallWithHost(host vmhost.VMHost,
	destOffset executor.MemPtr,
	valueOffset executor.MemPtr,
	dataOffset executor.MemPtr,
	dataLength executor.MemLength,
	successOffset executor.MemPtr,
	successLength executor.MemLength,
	errorOffset executor.MemPtr,
	errorLength executor.MemLength,
	gas int64,
	extraGasForCallback int64,
) int32 {

	calledSCAddress, err := context.MemLoad(destOffset, vmhost.AddressLen)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	value, err := context.MemLoad(valueOffset, vmhost.BalanceLen)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	data, err := context.MemLoad(dataOffset, dataLength)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	successFunc, err := context.MemLoad(successOffset, successLength)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	errorFunc, err := context.MemLoad(errorOffset, errorLength)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	return CreateAsyncCallWithTypedArgs(host,
		calledSCAddress,
		value,
		data,
		successFunc,
		errorFunc,
		gas,
		extraGasForCallback,
		nil)
}

// CreateAsyncCallWithTypedArgs - createAsyncCall with arguments already read from memory
func CreateAsyncCallWithTypedArgs(host vmhost.VMHost,
	calledSCAddress []byte,
	value []byte,
	data []byte,
	successFunc []byte,
	errorFunc []byte,
	gas int64,
	extraGasForCallback int64,
	callbackClosure []byte) int32 {

	metering := host.Metering()
	runtime := host.Runtime()
	async := host.Async()

	metering.StartGasTracing(createAsyncCallName)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.CreateAsyncCall
	err := metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return 1
	}

	asyncCall := &vmhost.AsyncCall{
		Status:          vmhost.AsyncCallPending,
		Destination:     calledSCAddress,
		Data:            data,
		ValueBytes:      value,
		GasLimit:        uint64(gas),
		SuccessCallback: string(successFunc),
		ErrorCallback:   string(errorFunc),
		GasLocked:       uint64(extraGasForCallback),
		CallbackClosure: callbackClosure,
	}

	if asyncCall.HasDefinedAnyCallback() {
		gasToUse = metering.GasSchedule().BaseOpsAPICost.SetAsyncCallback
		err = metering.UseGasBounded(gasToUse)
		if err != nil && runtime.UseGasBoundedShouldFailExecution() {
			FailExecution(host, err)
			return 1
		}
	}

	err = async.RegisterAsyncCall("", asyncCall)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	return 0
}

// SetAsyncContextCallback VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) SetAsyncContextCallback(
	callback executor.MemPtr,
	callbackLength executor.MemLength,
	data executor.MemPtr,
	dataLength executor.MemLength,
	gas int64,
) int32 {
	host := context.GetVMHost()
	runtime := host.Runtime()
	metering := host.Metering()
	async := host.Async()
	metering.StartGasTracing(setAsyncContextCallbackName)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.SetAsyncContextCallback
	err := metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return 1
	}

	callbackNameBytes, err := context.MemLoad(callback, callbackLength)
	if err != nil {
		context.FailExecution(err)
		return 1
	}

	dataBytes, err := context.MemLoad(data, dataLength)
	if err != nil {
		context.FailExecution(err)
		return 1
	}

	err = async.SetContextCallback(
		string(callbackNameBytes),
		dataBytes,
		uint64(gas))
	if err != nil {
		context.FailExecution(err)
		return 1
	}

	return 0
}

// UpgradeContract VMHooks implementation.
// @autogenerate(VMHooks)
// @autogenerate(VMHooks)
func (context *VMHooksImpl) UpgradeContract(
	destOffset executor.MemPtr,
	gasLimit int64,
	valueOffset executor.MemPtr,
	codeOffset executor.MemPtr,
	codeMetadataOffset executor.MemPtr,
	length executor.MemLength,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) {
	host := context.GetVMHost()
	runtime := host.Runtime()
	metering := host.Metering()
	metering.StartGasTracing(upgradeContractName)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.CreateContract
	err := metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return
	}

	value, err := context.MemLoad(valueOffset, vmhost.BalanceLen)
	if err != nil {
		context.FailExecution(err)
		return
	}

	code, err := context.MemLoad(codeOffset, length)
	if err != nil {
		context.FailExecution(err)
		return
	}

	codeMetadata, err := context.MemLoad(codeMetadataOffset, vmhost.CodeMetadataLen)
	if err != nil {
		context.FailExecution(err)
		return
	}

	data, actualLen, err := context.getArgumentsFromMemory(
		host,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return
	}

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(actualLen))
	err = metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return
	}

	if err != nil {
		context.FailExecution(err)
		return
	}

	calledSCAddress, err := context.MemLoad(destOffset, vmhost.AddressLen)
	if err != nil {
		context.FailExecution(err)
		return
	}

	gasSchedule := metering.GasSchedule()
	gasToUse = math.MulUint64(gasSchedule.BaseOperationCost.DataCopyPerByte, uint64(length))
	err = metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return
	}

	upgradeContract(host, calledSCAddress, code, codeMetadata, value, data, gasLimit)
}

// UpgradeFromSourceContract VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) UpgradeFromSourceContract(
	destOffset executor.MemPtr,
	gasLimit int64,
	valueOffset executor.MemPtr,
	sourceContractAddressOffset executor.MemPtr,
	codeMetadataOffset executor.MemPtr,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) {
	host := context.GetVMHost()
	runtime := host.Runtime()
	metering := host.Metering()
	metering.StartGasTracing(upgradeFromSourceContractName)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.CreateContract
	err := metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return
	}

	value, err := context.MemLoad(valueOffset, vmhost.BalanceLen)
	if err != nil {
		FailExecution(host, err)
		return
	}

	sourceContractAddress, err := context.MemLoad(sourceContractAddressOffset, vmhost.AddressLen)
	if err != nil {
		FailExecution(host, err)
		return
	}

	codeMetadata, err := context.MemLoad(codeMetadataOffset, vmhost.CodeMetadataLen)
	if err != nil {
		context.FailExecution(err)
		return
	}

	data, actualLen, err := context.getArgumentsFromMemory(
		host,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return
	}

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(actualLen))
	err = metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return
	}

	if err != nil {
		FailExecution(host, err)
		return
	}

	calledSCAddress, err := context.MemLoad(destOffset, vmhost.AddressLen)
	if err != nil {
		FailExecution(host, err)
		return
	}

	UpgradeFromSourceContractWithTypedArgs(
		host,
		sourceContractAddress,
		calledSCAddress,
		value,
		data,
		gasLimit,
		codeMetadata,
	)
}

// UpgradeFromSourceContractWithTypedArgs - upgradeFromSourceContract with args already read from memory
func UpgradeFromSourceContractWithTypedArgs(
	host vmhost.VMHost,
	sourceContractAddress []byte,
	destContractAddress []byte,
	value []byte,
	data [][]byte,
	gasLimit int64,
	codeMetadata []byte,
) {
	blockchain := host.Blockchain()

	code, err := blockchain.GetCode(sourceContractAddress)
	if err != nil {
		FailExecution(host, err)
		return
	}

	upgradeContract(host, destContractAddress, code, codeMetadata, value, data, gasLimit)
}

func upgradeContract(
	host vmhost.VMHost,
	destContractAddress []byte,
	code []byte,
	codeMetadata []byte,
	value []byte,
	data [][]byte,
	gasLimit int64,
) {
	runtime := host.Runtime()
	metering := host.Metering()
	gasSchedule := metering.GasSchedule()
	minAsyncCallCost := math.AddUint64(
		math.MulUint64(2, gasSchedule.BaseOpsAPICost.AsyncCallStep),
		gasSchedule.BaseOpsAPICost.AsyncCallbackGasLock)
	if uint64(gasLimit) < minAsyncCallCost {
		runtime.SetRuntimeBreakpointValue(vmhost.BreakpointOutOfGas)
		return
	}

	// Set up the async call as if it is not known whether the called SC
	// is in the same shard with the caller or not. This will be later resolved
	// by runtime.ExecuteAsyncCall().
	callData := vmhost.UpgradeFunctionName + "@" + hex.EncodeToString(code) + "@" + hex.EncodeToString(codeMetadata)
	for _, arg := range data {
		callData += "@" + hex.EncodeToString(arg)
	}

	async := host.Async()
	err := async.RegisterLegacyAsyncCall(
		destContractAddress,
		[]byte(callData),
		value,
	)
	logEEI.Trace("upgradeContract", "error", err)

	if errors.Is(err, vmhost.ErrNotEnoughGas) {
		runtime.SetRuntimeBreakpointValue(vmhost.BreakpointOutOfGas)
		return
	}
	if err != nil {
		FailExecution(host, err)
		return
	}
}

// DeleteContract VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) DeleteContract(
	destOffset executor.MemPtr,
	gasLimit int64,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) {
	host := context.GetVMHost()
	runtime := host.Runtime()
	metering := host.Metering()
	metering.StartGasTracing(deleteContractName)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.CreateContract
	err := metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return
	}

	data, actualLen, err := context.getArgumentsFromMemory(
		host,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return
	}

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(actualLen))
	err = metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return
	}

	if err != nil {
		FailExecution(host, err)
		return
	}

	calledSCAddress, err := context.MemLoad(destOffset, vmhost.AddressLen)
	if err != nil {
		FailExecution(host, err)
		return
	}

	deleteContract(
		host,
		calledSCAddress,
		data,
		gasLimit,
	)
}

func deleteContract(
	host vmhost.VMHost,
	dest []byte,
	data [][]byte,
	gasLimit int64,
) {
	runtime := host.Runtime()
	metering := host.Metering()
	gasSchedule := metering.GasSchedule()
	minAsyncCallCost := math.AddUint64(
		math.MulUint64(2, gasSchedule.BaseOpsAPICost.AsyncCallStep),
		gasSchedule.BaseOpsAPICost.AsyncCallbackGasLock)
	if uint64(gasLimit) < minAsyncCallCost {
		runtime.SetRuntimeBreakpointValue(vmhost.BreakpointOutOfGas)
		return
	}

	callData := vmhost.DeleteFunctionName
	for _, arg := range data {
		callData += "@" + hex.EncodeToString(arg)
	}

	async := host.Async()
	err := async.RegisterLegacyAsyncCall(
		dest,
		[]byte(callData),
		big.NewInt(0).Bytes(),
	)
	logEEI.Trace("deleteContract", "error", err)

	if errors.Is(err, vmhost.ErrNotEnoughGas) {
		runtime.SetRuntimeBreakpointValue(vmhost.BreakpointOutOfGas)
		return
	}
	if err != nil {
		FailExecution(host, err)
		return
	}
}

// AsyncCall VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) AsyncCall(
	destOffset executor.MemPtr,
	valueOffset executor.MemPtr,
	dataOffset executor.MemPtr,
	length executor.MemLength) {

	host := context.GetVMHost()
	runtime := host.Runtime()
	async := host.Async()
	metering := host.Metering()
	metering.StartGasTracing(asyncCallName)

	gasSchedule := metering.GasSchedule()
	gasToUse := gasSchedule.BaseOpsAPICost.AsyncCallStep
	err := metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return
	}

	calledSCAddress, err := context.MemLoad(destOffset, vmhost.AddressLen)
	if err != nil {
		context.FailExecution(err)
		return
	}

	value, err := context.MemLoad(valueOffset, vmhost.BalanceLen)
	if err != nil {
		context.FailExecution(err)
		return
	}

	gasToUse = math.MulUint64(gasSchedule.BaseOperationCost.DataCopyPerByte, uint64(length))
	err = metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return
	}

	data, err := context.MemLoad(dataOffset, length)
	if err != nil {
		context.FailExecution(err)
		return
	}

	err = async.RegisterLegacyAsyncCall(calledSCAddress, data, value)
	if errors.Is(err, vmhost.ErrNotEnoughGas) {
		runtime.SetRuntimeBreakpointValue(vmhost.BreakpointOutOfGas)
		return
	}
	if err != nil {
		context.FailExecution(err)
		return
	}
}

// GetArgumentLength VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetArgumentLength(id int32) int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetArgument
	err := metering.UseGasBoundedAndAddTracedGas(getArgumentLengthName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	args := runtime.Arguments()
	if id < 0 || int32(len(args)) <= id {
		context.FailExecution(vmhost.ErrInvalidArgument)
		return -1
	}

	return int32(len(args[id]))
}

// GetArgument VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetArgument(id int32, argOffset executor.MemPtr) int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetArgument
	err := metering.UseGasBoundedAndAddTracedGas(getArgumentName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	args := runtime.Arguments()
	if id < 0 || int32(len(args)) <= id {
		context.FailExecution(vmhost.ErrInvalidArgument)
		return -1
	}

	err = context.MemStore(argOffset, args[id])
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int32(len(args[id]))
}

// GetFunction VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetFunction(functionOffset executor.MemPtr) int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetFunction
	err := metering.UseGasBoundedAndAddTracedGas(getFunctionName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	function := runtime.FunctionName()
	err = context.MemStore(functionOffset, []byte(function))
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int32(len(function))
}

// GetNumArguments VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetNumArguments() int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetNumArguments
	err := metering.UseGasBoundedAndAddTracedGas(getNumArgumentsName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	args := runtime.Arguments()
	return int32(len(args))
}

// StorageStore VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) StorageStore(
	keyOffset executor.MemPtr,
	keyLength executor.MemLength,
	dataOffset executor.MemPtr,
	dataLength executor.MemLength) int32 {

	host := context.GetVMHost()
	return context.StorageStoreWithHost(
		host,
		keyOffset,
		keyLength,
		dataOffset,
		dataLength,
	)
}

// StorageStoreWithHost - storageStore with host instead of pointer context
func (context *VMHooksImpl) StorageStoreWithHost(
	host vmhost.VMHost,
	keyOffset executor.MemPtr,
	keyLength executor.MemLength,
	dataOffset executor.MemPtr,
	dataLength executor.MemLength) int32 {

	key, err := context.MemLoad(keyOffset, keyLength)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	data, err := context.MemLoad(dataOffset, dataLength)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	return StorageStoreWithTypedArgs(host, key, data)
}

// StorageStoreWithTypedArgs - storageStore with args already read from memory
func StorageStoreWithTypedArgs(host vmhost.VMHost, key []byte, data []byte) int32 {
	storage := host.Storage()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.StorageStore
	err := metering.UseGasBoundedAndAddTracedGas(storageStoreName, gasToUse)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	storageStatus, err := storage.SetStorage(key, data)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	return int32(storageStatus)
}

// StorageLoadLength VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) StorageLoadLength(keyOffset executor.MemPtr, keyLength executor.MemLength) int32 {
	storage := context.GetStorageContext()
	metering := context.GetMeteringContext()

	key, err := context.MemLoad(keyOffset, keyLength)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	data, trieDepth, usedCache, err := storage.GetStorageUnmetered(key)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	err = storage.UseGasForStorageLoad(
		storageLoadLengthName,
		int64(trieDepth),
		metering.GasSchedule().BaseOpsAPICost.StorageLoad,
		usedCache)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int32(len(data))
}

// StorageLoadFromAddress VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) StorageLoadFromAddress(
	addressOffset executor.MemPtr,
	keyOffset executor.MemPtr,
	keyLength executor.MemLength,
	dataOffset executor.MemPtr) int32 {

	host := context.GetVMHost()
	return context.StorageLoadFromAddressWithHost(
		host,
		addressOffset,
		keyOffset,
		keyLength,
		dataOffset,
	)
}

// StorageLoadFromAddressWithHost - storageLoadFromAddress with host instead of pointer context
func (context *VMHooksImpl) StorageLoadFromAddressWithHost(
	host vmhost.VMHost,
	addressOffset executor.MemPtr,
	keyOffset executor.MemPtr,
	keyLength executor.MemLength,
	dataOffset executor.MemPtr) int32 {

	key, err := context.MemLoad(keyOffset, keyLength)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	address, err := context.MemLoad(addressOffset, vmhost.AddressLen)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	data, err := StorageLoadFromAddressWithTypedArgs(host, address, key)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	err = context.MemStore(dataOffset, data)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	return int32(len(data))
}

// StorageLoadFromAddressWithTypedArgs - storageLoadFromAddress with args already read from memory
func StorageLoadFromAddressWithTypedArgs(host vmhost.VMHost, address []byte, key []byte) ([]byte, error) {
	storage := host.Storage()
	metering := host.Metering()
	data, trieDepth, usedCache, err := storage.GetStorageFromAddress(address, key)
	if err != nil {
		return nil, err
	}
	err = storage.UseGasForStorageLoad(
		storageLoadFromAddressName,
		int64(trieDepth),
		metering.GasSchedule().BaseOpsAPICost.StorageLoad,
		usedCache)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// StorageLoad VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) StorageLoad(keyOffset executor.MemPtr, keyLength executor.MemLength, dataOffset executor.MemPtr) int32 {
	host := context.GetVMHost()
	return context.StorageLoadWithHost(
		host,
		keyOffset,
		keyLength,
		dataOffset,
	)
}

// StorageLoadWithHost - storageLoad with host instead of pointer context
func (context *VMHooksImpl) StorageLoadWithHost(host vmhost.VMHost, keyOffset executor.MemPtr, keyLength executor.MemLength, dataOffset executor.MemPtr) int32 {

	key, err := context.MemLoad(keyOffset, keyLength)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	data, err := StorageLoadWithWithTypedArgs(host, key)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	err = context.MemStore(dataOffset, data)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	return int32(len(data))
}

// StorageLoadWithWithTypedArgs - storageLoad with args already read from memory
func StorageLoadWithWithTypedArgs(host vmhost.VMHost, key []byte) ([]byte, error) {
	storage := host.Storage()
	metering := host.Metering()
	data, trieDepth, usedCache, err := storage.GetStorage(key)
	if err != nil {
		return nil, err
	}

	err = storage.UseGasForStorageLoad(
		storageLoadName,
		int64(trieDepth),
		metering.GasSchedule().BaseOpsAPICost.StorageLoad,
		usedCache)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// SetStorageLock VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) SetStorageLock(keyOffset executor.MemPtr, keyLength executor.MemLength, lockTimestamp int64) int32 {
	host := context.GetVMHost()
	return context.SetStorageLockWithHost(
		host,
		keyOffset,
		keyLength,
		lockTimestamp,
	)
}

// SetStorageLockWithHost - setStorageLock with host instead of pointer context
func (context *VMHooksImpl) SetStorageLockWithHost(host vmhost.VMHost, keyOffset executor.MemPtr, keyLength executor.MemLength, lockTimestamp int64) int32 {
	metering := host.Metering()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.Int64StorageStore
	err := metering.UseGasBoundedAndAddTracedGas(setStorageLockName, gasToUse)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	key, err := context.MemLoad(keyOffset, keyLength)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	return SetStorageLockWithTypedArgs(host, key, lockTimestamp)
}

// SetStorageLockWithTypedArgs - setStorageLock with args already read from memory
func SetStorageLockWithTypedArgs(host vmhost.VMHost, key []byte, lockTimestamp int64) int32 {
	storage := host.Storage()
	timeLockKeyPrefix := string(storage.GetVmProtectedPrefix(vmhost.TimeLockKeyPrefix))
	timeLockKey := vmhost.CustomStorageKey(timeLockKeyPrefix, key)
	bigTimestamp := big.NewInt(0).SetInt64(lockTimestamp)
	storageStatus, err := storage.SetProtectedStorage(timeLockKey, bigTimestamp.Bytes())
	if err != nil {
		FailExecution(host, err)
		return -1
	}
	return int32(storageStatus)
}

// GetStorageLock VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetStorageLock(keyOffset executor.MemPtr, keyLength executor.MemLength) int64 {
	metering := context.GetMeteringContext()
	storage := context.GetStorageContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.StorageLoad
	err := metering.UseGasBoundedAndAddTracedGas(getStorageLockName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	key, err := context.MemLoad(keyOffset, keyLength)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	timeLockKeyPrefix := string(storage.GetVmProtectedPrefix(vmhost.TimeLockKeyPrefix))
	timeLockKey := vmhost.CustomStorageKey(timeLockKeyPrefix, key)

	data, trieDepth, usedCache, err := storage.GetStorage(timeLockKey)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	err = storage.UseGasForStorageLoad(
		getStorageLockName,
		int64(trieDepth),
		metering.GasSchedule().BaseOpsAPICost.StorageLoad,
		usedCache)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	timeLock := big.NewInt(0).SetBytes(data).Int64()

	// TODO if timelock <= currentTimeStamp { fail somehow }

	return timeLock
}

// IsStorageLocked VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) IsStorageLocked(keyOffset executor.MemPtr, keyLength executor.MemLength) int32 {
	timeLock := context.GetStorageLock(keyOffset, keyLength)
	if timeLock < 0 {
		return -1
	}

	currentTimestamp := context.GetBlockTimestamp()
	if timeLock <= currentTimestamp {
		return 0
	}

	return 1
}

// ClearStorageLock VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) ClearStorageLock(keyOffset executor.MemPtr, keyLength executor.MemLength) int32 {
	return context.SetStorageLock(keyOffset, keyLength, 0)
}

// GetCaller VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetCaller(resultOffset executor.MemPtr) {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetCaller
	err := metering.UseGasBoundedAndAddTracedGas(getCallerName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return
	}

	caller := runtime.GetVMInput().CallerAddr

	err = context.MemStore(resultOffset, caller)
	if err != nil {
		context.FailExecution(err)
		return
	}
}

// CheckNoPayment VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) CheckNoPayment() {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetCallValue
	err := metering.UseGasBoundedAndAddTracedGas(checkNoPaymentName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return
	}

	vmInput := runtime.GetVMInput()
	if vmInput.CallValue.Sign() > 0 {
		context.FailExecution(vmhost.ErrNonPayableFunctionEgld)
		return
	}
	if len(vmInput.ESDTTransfers) > 0 {
		context.FailExecution(vmhost.ErrNonPayableFunctionEsdt)
		return
	}
}

// GetCallValue VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetCallValue(resultOffset executor.MemPtr) int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetCallValue
	err := metering.UseGasBoundedAndAddTracedGas(callValueName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	value := runtime.GetVMInput().CallValue.Bytes()
	value = vmhost.PadBytesLeft(value, vmhost.BalanceLen)

	err = context.MemStore(resultOffset, value)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int32(len(value))
}

// GetESDTValue VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetESDTValue(resultOffset executor.MemPtr) int32 {
	isFail := failIfMoreThanOneESDTTransfer(context)
	if isFail {
		return -1
	}
	return context.GetESDTValueByIndex(resultOffset, 0)
}

// GetESDTValueByIndex VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetESDTValueByIndex(resultOffset executor.MemPtr, index int32) int32 {
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetCallValue
	err := metering.UseGasBoundedAndAddTracedGas(getESDTValueByIndexName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	var value []byte

	esdtTransfer := getESDTTransferFromInputFailIfWrongIndex(context.GetVMHost(), index)
	if esdtTransfer != nil && esdtTransfer.ESDTValue.Cmp(vmhost.Zero) > 0 {
		value = esdtTransfer.ESDTValue.Bytes()
		value = vmhost.PadBytesLeft(value, vmhost.BalanceLen)
	}

	err = context.MemStore(resultOffset, value)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int32(len(value))
}

// GetESDTTokenName VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetESDTTokenName(resultOffset executor.MemPtr) int32 {
	isFail := failIfMoreThanOneESDTTransfer(context)
	if isFail {
		return -1
	}
	return context.GetESDTTokenNameByIndex(resultOffset, 0)
}

// GetESDTTokenNameByIndex VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetESDTTokenNameByIndex(resultOffset executor.MemPtr, index int32) int32 {
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetCallValue
	err := metering.UseGasBoundedAndAddTracedGas(getESDTTokenNameByIndexName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	esdtTransfer := getESDTTransferFromInputFailIfWrongIndex(context.GetVMHost(), index)
	var tokenName []byte
	if esdtTransfer != nil {
		tokenName = esdtTransfer.ESDTTokenName
	}

	err = context.MemStore(resultOffset, tokenName)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int32(len(tokenName))
}

// GetESDTTokenNonce VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetESDTTokenNonce() int64 {
	isFail := failIfMoreThanOneESDTTransfer(context)
	if isFail {
		return -1
	}
	return context.GetESDTTokenNonceByIndex(0)
}

// GetESDTTokenNonceByIndex VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetESDTTokenNonceByIndex(index int32) int64 {
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetCallValue
	err := metering.UseGasBoundedAndAddTracedGas(getESDTTokenNonceByIndexName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	esdtTransfer := getESDTTransferFromInputFailIfWrongIndex(context.GetVMHost(), index)
	nonce := uint64(0)
	if esdtTransfer != nil {
		nonce = esdtTransfer.ESDTTokenNonce
	}
	return int64(nonce)
}

// GetCurrentESDTNFTNonce VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetCurrentESDTNFTNonce(
	addressOffset executor.MemPtr,
	tokenIDOffset executor.MemPtr,
	tokenIDLen executor.MemLength) int64 {

	metering := context.GetMeteringContext()
	storage := context.GetStorageContext()

	destination, err := context.MemLoad(addressOffset, vmhost.AddressLen)
	if err != nil {
		context.FailExecution(err)
		return 0
	}

	tokenID, err := context.MemLoad(tokenIDOffset, tokenIDLen)
	if err != nil {
		context.FailExecution(err)
		return 0
	}

	key := []byte(core.ProtectedKeyPrefix + core.ESDTNFTLatestNonceIdentifier + string(tokenID))
	data, trieDepth, _, err := storage.GetStorageFromAddress(destination, key)
	if err != nil {
		context.FailExecution(err)
		return 0
	}

	err = storage.UseGasForStorageLoad(
		getCurrentESDTNFTNonceName,
		int64(trieDepth),
		metering.GasSchedule().BaseOpsAPICost.StorageLoad,
		false)
	if err != nil {
		context.FailExecution(err)
		return 0
	}

	nonce := big.NewInt(0).SetBytes(data).Uint64()
	return int64(nonce)
}

// GetESDTTokenType VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetESDTTokenType() int32 {
	isFail := failIfMoreThanOneESDTTransfer(context)
	if isFail {
		return -1
	}
	return context.GetESDTTokenTypeByIndex(0)
}

// GetESDTTokenTypeByIndex VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetESDTTokenTypeByIndex(index int32) int32 {
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetCallValue
	err := metering.UseGasBoundedAndAddTracedGas(getESDTTokenTypeByIndexName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	esdtTransfer := getESDTTransferFromInputFailIfWrongIndex(context.GetVMHost(), index)
	if esdtTransfer != nil {
		return int32(esdtTransfer.ESDTTokenType)
	}
	return 0
}

// GetNumESDTTransfers VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetNumESDTTransfers() int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetCallValue
	err := metering.UseGasBoundedAndAddTracedGas(getNumESDTTransfersName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int32(len(runtime.GetVMInput().ESDTTransfers))
}

// GetCallValueTokenName VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetCallValueTokenName(callValueOffset executor.MemPtr, tokenNameOffset executor.MemPtr) int32 {
	isFail := failIfMoreThanOneESDTTransfer(context)
	if isFail {
		return -1
	}
	return context.GetCallValueTokenNameByIndex(callValueOffset, tokenNameOffset, 0)
}

// GetCallValueTokenNameByIndex VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetCallValueTokenNameByIndex(
	callValueOffset executor.MemPtr,
	tokenNameOffset executor.MemPtr,
	index int32) int32 {

	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetCallValue
	err := metering.UseGasBoundedAndAddTracedGas(getCallValueTokenNameByIndexName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	callValue := runtime.GetVMInput().CallValue.Bytes()
	tokenName := make([]byte, 0)
	esdtTransfer := getESDTTransferFromInputFailIfWrongIndex(context.GetVMHost(), index)

	if esdtTransfer != nil {
		tokenName = make([]byte, len(esdtTransfer.ESDTTokenName))
		copy(tokenName, esdtTransfer.ESDTTokenName)
		callValue = esdtTransfer.ESDTValue.Bytes()
	}
	callValue = vmhost.PadBytesLeft(callValue, vmhost.BalanceLen)

	err = context.MemStore(tokenNameOffset, tokenName)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	err = context.MemStore(callValueOffset, callValue)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int32(len(tokenName))
}

// IsReservedFunctionName VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) IsReservedFunctionName(nameHandle int32) int32 {
	host := context.host
	managedTypes := context.GetManagedTypesContext()
	runtime := host.Runtime()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.IsReservedFunctionName
	err := metering.UseGasBounded(gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	name, err := managedTypes.GetBytes(nameHandle)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	if runtime.IsReservedFunctionName(string(name)) {
		return 1
	}

	return 0
}

// WriteLog VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) WriteLog(
	dataPointer executor.MemPtr,
	dataLength executor.MemLength,
	topicPtr executor.MemPtr,
	numTopics int32) {

	// note: deprecated
	runtime := context.GetRuntimeContext()
	output := context.GetOutputContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.Log
	gas := math.MulUint64(metering.GasSchedule().BaseOperationCost.PersistPerByte, uint64(numTopics*vmhost.HashLen+dataLength))
	gasToUse = math.AddUint64(gasToUse, gas)

	if numTopics < 0 || dataLength < 0 {
		err := vmhost.ErrNegativeLength
		context.FailExecution(err)
		return
	}

	err := metering.UseGasBoundedAndAddTracedGas(writeLogName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return
	}

	log, err := context.MemLoad(dataPointer, dataLength)
	if err != nil {
		context.FailExecution(err)
		return
	}

	topics := make([][]byte, numTopics)
	for i := int32(0); i < numTopics; i++ {
		topics[i], err = context.MemLoad(topicPtr.Offset(i*vmhost.HashLen), vmhost.HashLen)
		if err != nil {
			context.FailExecution(err)
			return
		}
	}

	output.WriteLog(runtime.GetContextAddress(), topics, [][]byte{log})
}

// WriteEventLog VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) WriteEventLog(
	numTopics int32,
	topicLengthsOffset executor.MemPtr,
	topicOffset executor.MemPtr,
	dataOffset executor.MemPtr,
	dataLength executor.MemLength,
) {

	host := context.GetVMHost()
	runtime := context.GetRuntimeContext()
	output := context.GetOutputContext()
	metering := context.GetMeteringContext()

	topics, topicDataTotalLen, err := context.getArgumentsFromMemory(
		host,
		numTopics,
		topicLengthsOffset,
		topicOffset,
	)
	if err != nil {
		context.FailExecution(err)
		return
	}

	data, err := context.MemLoad(dataOffset, dataLength)
	if err != nil {
		context.FailExecution(err)
		return
	}

	gasToUse := metering.GasSchedule().BaseOpsAPICost.Log
	gasForData := math.MulUint64(
		metering.GasSchedule().BaseOperationCost.DataCopyPerByte,
		uint64(topicDataTotalLen+dataLength))
	gasToUse = math.AddUint64(gasToUse, gasForData)
	err = metering.UseGasBoundedAndAddTracedGas(writeEventLogName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return
	}

	output.WriteLog(runtime.GetContextAddress(), topics, [][]byte{data})
}

// GetBlockTimestamp VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetBlockTimestamp() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetBlockTimeStamp
	err := metering.UseGasBoundedAndAddTracedGas(getBlockTimestampName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int64(blockchain.CurrentTimeStamp())
}

// GetBlockNonce VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetBlockNonce() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetBlockNonce
	err := metering.UseGasBoundedAndAddTracedGas(getBlockNonceName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int64(blockchain.CurrentNonce())
}

// GetBlockRound VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetBlockRound() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetBlockRound
	err := metering.UseGasBoundedAndAddTracedGas(getBlockRoundName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int64(blockchain.CurrentRound())
}

// GetBlockEpoch VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetBlockEpoch() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetBlockEpoch
	err := metering.UseGasBoundedAndAddTracedGas(getBlockEpochName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int64(blockchain.CurrentEpoch())
}

// GetBlockRandomSeed VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetBlockRandomSeed(pointer executor.MemPtr) {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetBlockRandomSeed
	err := metering.UseGasBoundedAndAddTracedGas(getBlockRandomSeedName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return
	}

	randomSeed := blockchain.CurrentRandomSeed()
	err = context.MemStore(pointer, randomSeed)
	if err != nil {
		context.FailExecution(err)
	}
}

// GetStateRootHash VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetStateRootHash(pointer executor.MemPtr) {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetStateRootHash
	err := metering.UseGasBoundedAndAddTracedGas(getStateRootHashName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return
	}

	stateRootHash := blockchain.GetStateRootHash()
	err = context.MemStore(pointer, stateRootHash)
	if err != nil {
		context.FailExecution(err)
	}
}

// GetPrevBlockTimestamp VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetPrevBlockTimestamp() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetBlockTimeStamp
	err := metering.UseGasBoundedAndAddTracedGas(getPrevBlockTimestampName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int64(blockchain.LastTimeStamp())
}

// GetPrevBlockNonce VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetPrevBlockNonce() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetBlockNonce
	err := metering.UseGasBoundedAndAddTracedGas(getPrevBlockNonceName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int64(blockchain.LastNonce())
}

// GetPrevBlockRound VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetPrevBlockRound() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetBlockRound
	err := metering.UseGasBoundedAndAddTracedGas(getPrevBlockRoundName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int64(blockchain.LastRound())
}

// GetPrevBlockEpoch VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetPrevBlockEpoch() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetBlockEpoch
	err := metering.UseGasBoundedAndAddTracedGas(getPrevBlockEpochName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int64(blockchain.LastEpoch())
}

// GetPrevBlockRandomSeed VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetPrevBlockRandomSeed(pointer executor.MemPtr) {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetBlockRandomSeed
	err := metering.UseGasBoundedAndAddTracedGas(getPrevBlockRandomSeedName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return
	}

	randomSeed := blockchain.LastRandomSeed()
	err = context.MemStore(pointer, randomSeed)
	if err != nil {
		context.FailExecution(err)
	}
}

// GetRoundTime VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetRoundTime() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetRoundTime
	err := metering.UseGasBoundedAndAddTracedGas(getRoundTimeName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int64(blockchain.RoundTime())
}

// EpochStartBlockTimeStamp VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) EpochStartBlockTimeStamp() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.EpochStartBlockTimeStamp
	err := metering.UseGasBoundedAndAddTracedGas(epochStartBlockTimeStampName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int64(blockchain.EpochStartBlockTimeStamp())
}

// EpochStartBlockNonce VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) EpochStartBlockNonce() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.EpochStartBlockNonce
	err := metering.UseGasBoundedAndAddTracedGas(epochStartBlockNonceName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int64(blockchain.EpochStartBlockNonce())
}

// EpochStartBlockRound VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) EpochStartBlockRound() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.EpochStartBlockRound

	err := metering.UseGasBoundedAndAddTracedGas(epochStartBlockRoundName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	return int64(blockchain.EpochStartBlockRound())
}

// Finish VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) Finish(pointer executor.MemPtr, length executor.MemLength) {
	output := context.GetOutputContext()
	metering := context.GetMeteringContext()
	metering.StartGasTracing(returnDataName)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.Finish
	gas := math.MulUint64(metering.GasSchedule().BaseOperationCost.PersistPerByte, uint64(length))
	gasToUse = math.AddUint64(gasToUse, gas)
	err := metering.UseGasBounded(gasToUse)
	if err != nil {
		context.FailExecution(err)
		return
	}

	data, err := context.MemLoad(pointer, length)
	if err != nil {
		context.FailExecution(err)
		return
	}

	output.Finish(data)
}

// ExecuteOnSameContext VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) ExecuteOnSameContext(
	gasLimit int64,
	addressOffset executor.MemPtr,
	valueOffset executor.MemPtr,
	functionOffset executor.MemPtr,
	functionLength executor.MemLength,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) int32 {
	host := context.GetVMHost()
	metering := host.Metering()
	metering.StartGasTracing(executeOnSameContextName)

	return context.ExecuteOnSameContextWithHost(
		host,
		gasLimit,
		addressOffset,
		valueOffset,
		functionOffset,
		functionLength,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
}

// ExecuteOnSameContextWithHost - executeOnSameContext with host instead of pointer context
func (context *VMHooksImpl) ExecuteOnSameContextWithHost(
	host vmhost.VMHost,
	gasLimit int64,
	addressOffset executor.MemPtr,
	valueOffset executor.MemPtr,
	functionOffset executor.MemPtr,
	functionLength executor.MemLength,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) int32 {

	callArgs, err := context.extractIndirectContractCallArgumentsWithValue(
		host, addressOffset, valueOffset, functionOffset, functionLength, numArguments, argumentsLengthOffset, dataOffset)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	return ExecuteOnSameContextWithTypedArgs(
		host,
		gasLimit,
		callArgs.value,
		callArgs.function,
		callArgs.dest,
		callArgs.args,
	)
}

// ExecuteOnSameContextWithTypedArgs - executeOnSameContext with args already read from memory
func ExecuteOnSameContextWithTypedArgs(
	host vmhost.VMHost,
	gasLimit int64,
	value *big.Int,
	function []byte,
	dest []byte,
	args [][]byte,
) int32 {
	runtime := host.Runtime()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.ExecuteOnSameContext
	err := metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return -1
	}

	sender := runtime.GetContextAddress()

	contractCallInput, err := prepareIndirectContractCallInput(
		host,
		sender,
		value,
		gasLimit,
		dest,
		function,
		args,
		gasToUse,
		true,
	)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	if host.IsBuiltinFunctionName(contractCallInput.Function) {
		FailExecution(host, vmhost.ErrInvalidBuiltInFunctionCall)
		return 1
	}

	err = host.ExecuteOnSameContext(contractCallInput)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	return 0
}

// ExecuteOnDestContext VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) ExecuteOnDestContext(
	gasLimit int64,
	addressOffset executor.MemPtr,
	valueOffset executor.MemPtr,
	functionOffset executor.MemPtr,
	functionLength executor.MemLength,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) int32 {
	host := context.GetVMHost()
	metering := host.Metering()
	metering.StartGasTracing(executeOnDestContextName)

	return context.ExecuteOnDestContextWithHost(
		host,
		gasLimit,
		addressOffset,
		valueOffset,
		functionOffset,
		functionLength,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
}

// ExecuteOnDestContextWithHost - executeOnDestContext with host instead of pointer context
func (context *VMHooksImpl) ExecuteOnDestContextWithHost(
	host vmhost.VMHost,
	gasLimit int64,
	addressOffset executor.MemPtr,
	valueOffset executor.MemPtr,
	functionOffset executor.MemPtr,
	functionLength executor.MemLength,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) int32 {

	callArgs, err := context.extractIndirectContractCallArgumentsWithValue(
		host, addressOffset, valueOffset, functionOffset, functionLength, numArguments, argumentsLengthOffset, dataOffset)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	return ExecuteOnDestContextWithTypedArgs(
		host,
		gasLimit,
		callArgs.value,
		callArgs.function,
		callArgs.dest,
		callArgs.args,
	)
}

// ExecuteOnDestContextWithTypedArgs - executeOnDestContext with args already read from memory
func ExecuteOnDestContextWithTypedArgs(
	host vmhost.VMHost,
	gasLimit int64,
	value *big.Int,
	function []byte,
	dest []byte,
	args [][]byte,
) int32 {
	runtime := host.Runtime()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.ExecuteOnDestContext
	err := metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return -1
	}

	sender := runtime.GetContextAddress()

	contractCallInput, err := prepareIndirectContractCallInput(
		host,
		sender,
		value,
		gasLimit,
		dest,
		function,
		args,
		gasToUse,
		true,
	)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	vmOutput, err := executeOnDestContextFromAPI(host, contractCallInput)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	host.CompleteLogEntriesWithCallType(vmOutput, vmhost.ExecuteOnDestContextString)

	return 0
}

// ExecuteReadOnly VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) ExecuteReadOnly(
	gasLimit int64,
	addressOffset executor.MemPtr,
	functionOffset executor.MemPtr,
	functionLength executor.MemLength,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) int32 {
	host := context.GetVMHost()
	metering := host.Metering()
	metering.StartGasTracing(executeReadOnlyName)

	return context.ExecuteReadOnlyWithHost(
		host,
		gasLimit,
		addressOffset,
		functionOffset,
		functionLength,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
}

// ExecuteReadOnlyWithHost - executeReadOnly with host instead of pointer context
func (context *VMHooksImpl) ExecuteReadOnlyWithHost(
	host vmhost.VMHost,
	gasLimit int64,
	addressOffset executor.MemPtr,
	functionOffset executor.MemPtr,
	functionLength executor.MemLength,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) int32 {

	callArgs, err := context.extractIndirectContractCallArgumentsWithoutValue(
		host, addressOffset, functionOffset, functionLength, numArguments, argumentsLengthOffset, dataOffset)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	return ExecuteReadOnlyWithTypedArguments(
		host,
		gasLimit,
		callArgs.function,
		callArgs.dest,
		callArgs.args,
	)
}

// ExecuteReadOnlyWithTypedArguments - executeReadOnly with args already read from memory
func ExecuteReadOnlyWithTypedArguments(
	host vmhost.VMHost,
	gasLimit int64,
	function []byte,
	dest []byte,
	args [][]byte,
) int32 {
	runtime := host.Runtime()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.ExecuteReadOnly
	err := metering.UseGasBounded(gasToUse)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	sender := runtime.GetContextAddress()

	contractCallInput, err := prepareIndirectContractCallInput(
		host,
		sender,
		big.NewInt(0),
		gasLimit,
		dest,
		function,
		args,
		gasToUse,
		true,
	)
	if err != nil {
		FailExecution(host, err)
		return -1
	}

	if host.IsBuiltinFunctionName(contractCallInput.Function) {
		FailExecution(host, vmhost.ErrInvalidBuiltInFunctionCall)
		return 1
	}

	wasReadOnly := runtime.ReadOnly()
	runtime.SetReadOnly(true)
	_, err = executeOnDestContextFromAPI(host, contractCallInput)
	runtime.SetReadOnly(wasReadOnly)

	if err != nil {
		FailExecution(host, err)
		return -1
	}

	return 0
}

// CreateContract VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) CreateContract(
	gasLimit int64,
	valueOffset executor.MemPtr,
	codeOffset executor.MemPtr,
	codeMetadataOffset executor.MemPtr,
	length executor.MemLength,
	resultOffset executor.MemPtr,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) int32 {
	host := context.GetVMHost()
	return context.createContractWithHost(
		host,
		gasLimit,
		valueOffset,
		codeOffset,
		codeMetadataOffset,
		length,
		resultOffset,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
}

func (context *VMHooksImpl) createContractWithHost(
	host vmhost.VMHost,
	gasLimit int64,
	valueOffset executor.MemPtr,
	codeOffset executor.MemPtr,
	codeMetadataOffset executor.MemPtr,
	length executor.MemLength,
	resultOffset executor.MemPtr,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) int32 {
	runtime := host.Runtime()

	metering := host.Metering()
	metering.StartGasTracing(createContractName)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.CreateContract
	err := metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return -1
	}

	sender := runtime.GetContextAddress()
	value, err := context.MemLoad(valueOffset, vmhost.BalanceLen)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	code, err := context.MemLoad(codeOffset, length)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	codeMetadata, err := context.MemLoad(codeMetadataOffset, vmhost.CodeMetadataLen)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	data, actualLen, err := context.getArgumentsFromMemory(
		host,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(actualLen))
	err = metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return 1
	}

	valueAsInt := big.NewInt(0).SetBytes(value)
	newAddress, err := createContract(sender, data, valueAsInt, gasLimit, code, codeMetadata, host, CreateContract)

	if err != nil {
		FailExecution(host, err)
		return 1
	}

	err = context.MemStore(resultOffset, newAddress)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	return 0
}

// DeployFromSourceContract VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) DeployFromSourceContract(
	gasLimit int64,
	valueOffset executor.MemPtr,
	sourceContractAddressOffset executor.MemPtr,
	codeMetadataOffset executor.MemPtr,
	resultAddressOffset executor.MemPtr,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) int32 {
	host := context.GetVMHost()
	runtime := host.Runtime()
	metering := host.Metering()
	metering.StartGasTracing(deployFromSourceContractName)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.CreateContract
	err := metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return 1
	}

	value, err := context.MemLoad(valueOffset, vmhost.BalanceLen)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	sourceContractAddress, err := context.MemLoad(sourceContractAddressOffset, vmhost.AddressLen)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	codeMetadata, err := context.MemLoad(codeMetadataOffset, vmhost.CodeMetadataLen)
	if err != nil {
		context.FailExecution(err)
		return 1
	}

	data, actualLen, err := context.getArgumentsFromMemory(
		host,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
	if err != nil {
		FailExecution(host, err)
		return 1
	}

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(actualLen))
	err = metering.UseGasBounded(gasToUse)
	if err != nil && runtime.UseGasBoundedShouldFailExecution() {
		FailExecution(host, err)
		return 1
	}

	newAddress, err := DeployFromSourceContractWithTypedArgs(
		host,
		sourceContractAddress,
		codeMetadata,
		big.NewInt(0).SetBytes(value),
		data,
		gasLimit,
	)

	if err != nil {
		context.FailExecution(err)
		return 1
	}

	err = context.MemStore(resultAddressOffset, newAddress)
	if err != nil {
		context.FailExecution(err)
		return 1
	}

	return 0
}

// DeployFromSourceContractWithTypedArgs - deployFromSourceContract with args already read from memory
func DeployFromSourceContractWithTypedArgs(
	host vmhost.VMHost,
	sourceContractAddress []byte,
	codeMetadata []byte,
	value *big.Int,
	data [][]byte,
	gasLimit int64,
) ([]byte, error) {
	runtime := host.Runtime()
	sender := runtime.GetContextAddress()

	blockchain := host.Blockchain()
	code, err := blockchain.GetCode(sourceContractAddress)
	if err != nil {
		FailExecution(host, err)
		return nil, err
	}

	return createContract(sender, data, value, gasLimit, code, codeMetadata, host, DeployContract)
}

func createContract(
	sender []byte,
	data [][]byte,
	value *big.Int,
	gasLimit int64,
	code []byte,
	codeMetadata []byte,
	host vmhost.VMHost,
	createContractCallType CreateContractCallType,
) ([]byte, error) {
	originalCaller := host.Runtime().GetOriginalCallerAddress()
	metering := host.Metering()
	contractCreate := &vmcommon.ContractCreateInput{
		VMInput: vmcommon.VMInput{
			OriginalCallerAddr: originalCaller,
			CallerAddr:         sender,
			Arguments:          data,
			CallValue:          value,
			GasPrice:           0,
			GasProvided:        metering.BoundGasLimit(gasLimit),
		},
		ContractCode:         code,
		ContractCodeMetadata: codeMetadata,
	}

	currentVMInput := host.Runtime().GetVMInput()
	if len(currentVMInput.RelayerAddr) > 0 {
		contractCreate.RelayerAddr = make([]byte, len(currentVMInput.RelayerAddr))
		copy(contractCreate.RelayerAddr, currentVMInput.RelayerAddr)
	}

	return host.CreateNewContract(contractCreate, int(createContractCallType))
}

// GetNumReturnData VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetNumReturnData() int32 {
	output := context.GetOutputContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetNumReturnData
	err := metering.UseGasBoundedAndAddTracedGas(getNumReturnDataName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	returnData := output.ReturnData()
	return int32(len(returnData))
}

// GetReturnDataSize VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetReturnDataSize(resultID int32) int32 {
	output := context.GetOutputContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetReturnDataSize
	err := metering.UseGasBoundedAndAddTracedGas(getReturnDataSizeName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return -1
	}

	returnData := output.ReturnData()
	if resultID >= int32(len(returnData)) || resultID < 0 {
		context.FailExecution(vmhost.ErrInvalidArgument)
		return 0
	}

	return int32(len(returnData[resultID]))
}

// GetReturnData VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetReturnData(resultID int32, dataOffset executor.MemPtr) int32 {
	host := context.GetVMHost()

	result := GetReturnDataWithHostAndTypedArgs(host, resultID)
	if result == nil {
		return 0
	}

	err := context.MemStore(dataOffset, result)
	if err != nil {
		context.FailExecution(err)
		return 0
	}

	return int32(len(result))
}

func GetReturnDataWithHostAndTypedArgs(host vmhost.VMHost, resultID int32) []byte {
	output := host.Output()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetReturnData
	err := metering.UseGasBoundedAndAddTracedGas(getReturnDataName, gasToUse)
	if err != nil {
		FailExecution(host, err)
		return nil
	}

	returnData := output.ReturnData()
	if resultID >= int32(len(returnData)) || resultID < 0 {
		FailExecution(host, vmhost.ErrInvalidArgument)
		return nil
	}

	return returnData[resultID]
}

// CleanReturnData VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) CleanReturnData() {
	host := context.GetVMHost()
	CleanReturnDataWithHost(host)
}

// CleanReturnDataWithHost - exposed version of v1_5_deleteFromReturnData for tests
func CleanReturnDataWithHost(host vmhost.VMHost) {
	output := host.Output()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.CleanReturnData
	err := metering.UseGasBoundedAndAddTracedGas(cleanReturnDataName, gasToUse)
	if err != nil {
		FailExecution(host, err)
		return
	}

	output.ClearReturnData()
}

// DeleteFromReturnData VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) DeleteFromReturnData(resultID int32) {
	host := context.GetVMHost()
	DeleteFromReturnDataWithHost(host, resultID)
}

// DeleteFromReturnDataWithHost - exposed version of v1_5_deleteFromReturnData for tests
func DeleteFromReturnDataWithHost(host vmhost.VMHost, resultID int32) {
	output := host.Output()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.DeleteFromReturnData
	err := metering.UseGasBoundedAndAddTracedGas(deleteFromReturnDataName, gasToUse)
	if err != nil {
		FailExecution(host, err)
		return
	}

	returnData := output.ReturnData()
	if resultID < int32(len(returnData)) {
		output.RemoveReturnData(uint32(resultID))
	}
}

// GetOriginalTxHash VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetOriginalTxHash(dataOffset executor.MemPtr) {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetOriginalTxHash
	err := metering.UseGasBoundedAndAddTracedGas(getOriginalTxHashName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return
	}

	err = context.MemStore(dataOffset, runtime.GetOriginalTxHash())
	if err != nil {
		context.FailExecution(err)
	}
}

// GetCurrentTxHash VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetCurrentTxHash(dataOffset executor.MemPtr) {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetCurrentTxHash
	err := metering.UseGasBoundedAndAddTracedGas(getCurrentTxHashName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return
	}

	err = context.MemStore(dataOffset, runtime.GetCurrentTxHash())
	if err != nil {
		context.FailExecution(err)
	}
}

// GetPrevTxHash VMHooks implementation.
// @autogenerate(VMHooks)
func (context *VMHooksImpl) GetPrevTxHash(dataOffset executor.MemPtr) {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().BaseOpsAPICost.GetPrevTxHash
	err := metering.UseGasBoundedAndAddTracedGas(getPrevTxHashName, gasToUse)
	if err != nil {
		context.FailExecution(err)
		return
	}

	err = context.MemStore(dataOffset, runtime.GetPrevTxHash())
	if err != nil {
		context.FailExecution(err)
	}
}

func prepareIndirectContractCallInput(
	host vmhost.VMHost,
	sender []byte,
	value *big.Int,
	gasLimit int64,
	destination []byte,
	function []byte,
	data [][]byte,
	_ uint64,
	syncExecutionRequired bool,
) (*vmcommon.ContractCallInput, error) {
	runtime := host.Runtime()
	metering := host.Metering()

	if syncExecutionRequired && !host.AreInSameShard(runtime.GetContextAddress(), destination) {
		return nil, vmhost.ErrSyncExecutionNotInSameShard
	}

	contractCallInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			OriginalCallerAddr: host.Runtime().GetOriginalCallerAddress(),
			CallerAddr:         sender,
			Arguments:          data,
			CallValue:          value,
			GasPrice:           0,
			GasProvided:        metering.BoundGasLimit(gasLimit),
			CallType:           vm.DirectCall,
		},
		RecipientAddr: destination,
		Function:      string(function),
	}

	currentVMInput := runtime.GetVMInput()
	if len(currentVMInput.RelayerAddr) > 0 {
		contractCallInput.RelayerAddr = make([]byte, len(currentVMInput.RelayerAddr))
		copy(contractCallInput.RelayerAddr, currentVMInput.RelayerAddr)
	}

	return contractCallInput, nil
}

func (context *VMHooksImpl) getArgumentsFromMemory(
	_ vmhost.VMHost,
	numArguments int32,
	argumentsLengthOffset executor.MemPtr,
	dataOffset executor.MemPtr,
) ([][]byte, int32, error) {
	if numArguments < 0 {
		return nil, 0, fmt.Errorf("negative numArguments (%d)", numArguments)
	}

	argumentsLengthData, err := context.MemLoad(argumentsLengthOffset, numArguments*4)
	if err != nil {
		return nil, 0, err
	}

	argumentLengths := createInt32Array(argumentsLengthData, numArguments)
	data, err := context.MemLoadMultiple(dataOffset, argumentLengths)
	if err != nil {
		return nil, 0, err
	}

	totalArgumentBytes := int32(0)
	for _, length := range argumentLengths {
		totalArgumentBytes += length
	}

	return data, totalArgumentBytes, nil
}

func createInt32Array(rawData []byte, numIntegers int32) []int32 {
	integers := make([]int32, numIntegers)
	index := 0
	for cursor := 0; cursor < len(rawData); cursor += 4 {
		rawInt := rawData[cursor : cursor+4]
		actualInt := binary.LittleEndian.Uint32(rawInt)
		integers[index] = int32(actualInt)
		index++
	}
	return integers
}

func executeOnDestContextFromAPI(host vmhost.VMHost, input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	host.Async().SetAsyncArgumentsForCall(input)
	vmOutput, isChildComplete, err := host.ExecuteOnDestContext(input)
	if err != nil {
		return nil, err
	}

	err = host.Async().CompleteChildConditional(isChildComplete, nil, 0)
	if err != nil {
		return nil, err
	}
	return vmOutput, err
}
