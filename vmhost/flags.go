package vmhost

import "github.com/multiversx/mx-chain-core-go/core"

const (
	// CryptoOpcodesV2Flag defines the flag that activates the new crypto APIs for RC1.7
	CryptoOpcodesV2Flag core.EnableEpochFlag = "CryptoOpcodesV2Flag"

	// MultiESDTNFTTransferAndExecuteByUserFlag defines the flag that activates the enshrined sovereign functions
	MultiESDTNFTTransferAndExecuteByUserFlag core.EnableEpochFlag = "MultiESDTNFTTransferAndExecuteByUserFlag"

	// UseGasBoundedShouldFailExecutionFlag defines the flag that activates failing of execution if gas bounded check fails
	UseGasBoundedShouldFailExecutionFlag core.EnableEpochFlag = "UseGasBoundedShouldFailExecutionFlag"

	// MaskInternalDependenciesErrorsFlag defines the flag that activates masking of internal dependencies errors
	MaskInternalDependenciesErrorsFlag core.EnableEpochFlag = "MaskInternalDependenciesErrorsFlag"
)
