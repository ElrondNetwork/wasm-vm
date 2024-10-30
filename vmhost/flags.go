package vmhost

import "github.com/multiversx/mx-chain-core-go/core"

const (
	// MultiESDTTransferFixOnCallBackFlag defines the flag that activates the multi esdt transfer fix on callback
	MultiESDTTransferFixOnCallBackFlag core.EnableEpochFlag = "MultiESDTTransferFixOnCallBackFlag"

	// RemoveNonUpdatedStorageFlag defines the flag that activates the remove non updated storage fix
	RemoveNonUpdatedStorageFlag core.EnableEpochFlag = "RemoveNonUpdatedStorageFlag"

	// CreateNFTThroughExecByCallerFlag defines the flag that activates the create nft through exec by caller fix
	CreateNFTThroughExecByCallerFlag core.EnableEpochFlag = "CreateNFTThroughExecByCallerFlag"

	// StorageAPICostOptimizationFlag defines the flag that activates the storage api cost optimization
	StorageAPICostOptimizationFlag core.EnableEpochFlag = "StorageAPICostOptimizationFlag"

	// CheckExecuteOnReadOnlyFlag defines the flag that activates the check execute on read only
	CheckExecuteOnReadOnlyFlag core.EnableEpochFlag = "CheckExecuteOnReadOnlyFlag"

	// FailExecutionOnEveryAPIErrorFlag defines the flag that activates the fail execution on every api error
	FailExecutionOnEveryAPIErrorFlag core.EnableEpochFlag = "FailExecutionOnEveryAPIErrorFlag"

	// ManagedCryptoAPIsFlag defines the flag that activates the manage crypto apis
	ManagedCryptoAPIsFlag core.EnableEpochFlag = "ManagedCryptoAPIsFlag"

	// DisableExecByCallerFlag defines the flag that activates disable exec by caller
	DisableExecByCallerFlag core.EnableEpochFlag = "DisableExecByCallerFlag"

	// RefactorContextFlag defines the flag that activates the refactor context
	RefactorContextFlag core.EnableEpochFlag = "RefactorContextFlag"

	// RuntimeMemStoreLimitFlag defines the flag that activates the runtime mem store limit
	RuntimeMemStoreLimitFlag core.EnableEpochFlag = "RuntimeMemStoreLimitFlag"

	// RuntimeCodeSizeFixFlag defines the flag that activates the runtime code size fix
	RuntimeCodeSizeFixFlag core.EnableEpochFlag = "RuntimeCodeSizeFixFlag"

	// FixOOGReturnCodeFlag defines the flag that activates the fix oog return code
	FixOOGReturnCodeFlag core.EnableEpochFlag = "FixOOGReturnCodeFlag"

	// DynamicGasCostForDataTrieStorageLoadFlag defines the flag that activates the dynamic gas cost for data trie storage load
	DynamicGasCostForDataTrieStorageLoadFlag core.EnableEpochFlag = "DynamicGasCostForDataTrieStorageLoadFlag"

	// CryptoOpcodesV2Flag defines the flag that activates the new crypto APIs for RC1.7
	CryptoOpcodesV2Flag core.EnableEpochFlag = "CryptoOpcodesV2Flag"

	// MultiESDTNFTTransferAndExecuteByUserFlag defines the flag that activates the enshrined sovereign functions
	MultiESDTNFTTransferAndExecuteByUserFlag core.EnableEpochFlag = "MultiESDTNFTTransferAndExecuteByUserFlag"

	// UseGasBoundedShouldFailExecutionFlag defines the flag that activates failing of execution if gas bounded check fails
	UseGasBoundedShouldFailExecutionFlag core.EnableEpochFlag = "UseGasBoundedShouldFailExecutionFlag"

	// AsyncV3Flag defines the flag that activates async v3
	AsyncV3Flag core.EnableEpochFlag = "AsyncV3Flag"
)
