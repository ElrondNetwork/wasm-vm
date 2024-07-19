package vmhost

import "github.com/multiversx/mx-chain-core-go/core"

const (
	// CryptoOpcodesV2Flag defines the flag that activates the new crypto APIs for RC1.7
	CryptoOpcodesV2Flag core.EnableEpochFlag = "CryptoOpcodesV2Flag"
	// ValidationOnGobDecodeFlag defines the flag that allows the GobDecode validation added on go1.21
	ValidationOnGobDecodeFlag core.EnableEpochFlag = "ValidationOnGobDecodeFlag"
	// all new flags must be added to allFlags slice from hostCore/host
)
