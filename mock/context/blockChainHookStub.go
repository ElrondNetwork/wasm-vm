package mock

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data/esdt"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var _ vmcommon.BlockchainHook = (*BlockchainHookStub)(nil)

// BlockchainHookStub is used in tests to check that interface methods were called
type BlockchainHookStub struct {
	NewAddressCalled               func(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error)
	GetStorageDataCalled           func(accountsAddress []byte, index []byte) ([]byte, uint32, error)
	GetBlockHashCalled             func(nonce uint64) ([]byte, error)
	LastNonceCalled                func() uint64
	LastRoundCalled                func() uint64
	LastTimeStampCalled            func() uint64
	LastRandomSeedCalled           func() []byte
	LastEpochCalled                func() uint32
	GetStateRootHashCalled         func() []byte
	CurrentNonceCalled             func() uint64
	CurrentRoundCalled             func() uint64
	CurrentTimeStampCalled         func() uint64
	CurrentRandomSeedCalled        func() []byte
	CurrentEpochCalled             func() uint32
	RoundTimeCalled                func() uint64
	EpochStartBlockTimeStampCalled func() uint64
	EpochStartBlockNonceCalled     func() uint64
	EpochStartBlockRoundCalled     func() uint64

	ProcessBuiltInFunctionCalled            func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
	GetBuiltinFunctionNamesCalled           func() vmcommon.FunctionNames
	GetAllStateCalled                       func(address []byte) (map[string][]byte, error)
	GetUserAccountCalled                    func(address []byte) (vmcommon.UserAccountHandler, error)
	GetShardOfAddressCalled                 func(address []byte) uint32
	IsSmartContractCalled                   func(address []byte) bool
	IsPayableCalled                         func(address []byte) (bool, error)
	GetCompiledCodeCalled                   func(codeHash []byte) (bool, []byte)
	SaveCompiledCodeCalled                  func(codeHash []byte, code []byte)
	GetCodeCalled                           func(account vmcommon.UserAccountHandler) []byte
	GetESDTTokenCalled                      func(address []byte, tokenID []byte, nonce uint64) (*esdt.ESDigitalToken, error)
	GetSnapshotCalled                       func() int
	RevertToSnapshotCalled                  func(snapshot int) error
	ExecuteSmartContractCallOnOtherVMCalled func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
}

// NewAddress mocked method
func (b *BlockchainHookStub) NewAddress(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
	if b.NewAddressCalled != nil {
		return b.NewAddressCalled(creatorAddress, creatorNonce, vmType)
	}
	return []byte("newAddress"), nil
}

// GetStorageData mocked method
func (b *BlockchainHookStub) GetStorageData(accountAddress []byte, index []byte) ([]byte, uint32, error) {
	if b.GetStorageDataCalled != nil {
		return b.GetStorageDataCalled(accountAddress, index)
	}
	return nil, 0, nil
}

// GetBlockhash mocked method
func (b *BlockchainHookStub) GetBlockhash(nonce uint64) ([]byte, error) {
	if b.GetBlockHashCalled != nil {
		return b.GetBlockHashCalled(nonce)
	}
	return []byte("roothash"), nil
}

// LastNonce mocked method
func (b *BlockchainHookStub) LastNonce() uint64 {
	if b.LastNonceCalled != nil {
		return b.LastNonceCalled()
	}
	return 0
}

// LastRound mocked method
func (b *BlockchainHookStub) LastRound() uint64 {
	if b.LastRoundCalled != nil {
		return b.LastRoundCalled()
	}
	return 0
}

// LastTimeStamp mocked method
func (b *BlockchainHookStub) LastTimeStamp() uint64 {
	if b.LastTimeStampCalled != nil {
		return b.LastTimeStampCalled()
	}
	return 0
}

// LastRandomSeed mocked method
func (b *BlockchainHookStub) LastRandomSeed() []byte {
	if b.LastRandomSeedCalled != nil {
		return b.LastRandomSeedCalled()
	}
	return []byte("seed")
}

// LastEpoch mocked method
func (b *BlockchainHookStub) LastEpoch() uint32 {
	if b.LastEpochCalled != nil {
		return b.LastEpochCalled()
	}
	return 0
}

// GetStateRootHash mocked method
func (b *BlockchainHookStub) GetStateRootHash() []byte {
	if b.GetStateRootHashCalled != nil {
		return b.GetStateRootHashCalled()
	}
	return []byte("roothash")
}

// CurrentNonce mocked method
func (b *BlockchainHookStub) CurrentNonce() uint64 {
	if b.CurrentNonceCalled != nil {
		return b.CurrentNonceCalled()
	}
	return 0
}

// CurrentRound mocked method
func (b *BlockchainHookStub) CurrentRound() uint64 {
	if b.CurrentRoundCalled != nil {
		return b.CurrentRoundCalled()
	}
	return 0
}

// CurrentTimeStamp mocked method
func (b *BlockchainHookStub) CurrentTimeStamp() uint64 {
	if b.CurrentTimeStampCalled != nil {
		return b.CurrentTimeStampCalled()
	}
	return 0
}

// CurrentRandomSeed mocked method
func (b *BlockchainHookStub) CurrentRandomSeed() []byte {
	if b.CurrentRandomSeedCalled != nil {
		return b.CurrentRandomSeedCalled()
	}
	return []byte("seed")
}

// CurrentEpoch mocked method
func (b *BlockchainHookStub) CurrentEpoch() uint32 {
	if b.CurrentEpochCalled != nil {
		return b.CurrentEpochCalled()
	}
	return 0
}

// RoundTime mocked method
func (b *BlockchainHookStub) RoundTime() uint64 {
	if b.RoundTimeCalled != nil {
		return b.RoundTimeCalled()
	}
	return 0
}

// EpochStartBlockTimeStamp mocked method
func (b *BlockchainHookStub) EpochStartBlockTimeStamp() uint64 {
	if b.EpochStartBlockTimeStampCalled != nil {
		return b.EpochStartBlockTimeStampCalled()
	}
	return 0
}

// EpochStartBlockNonce mocked method
func (b *BlockchainHookStub) EpochStartBlockNonce() uint64 {
	if b.EpochStartBlockNonceCalled != nil {
		return b.EpochStartBlockNonceCalled()
	}
	return 0
}

// EpochStartBlockRound VMHooks implementation.
func (b *BlockchainHookStub) EpochStartBlockRound() uint64 {
	if b.EpochStartBlockRoundCalled != nil {
		return b.EpochStartBlockRoundCalled()
	}
	return 0
}

// ProcessBuiltInFunction mocked method
func (b *BlockchainHookStub) ProcessBuiltInFunction(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	if b.ProcessBuiltInFunctionCalled != nil {
		return b.ProcessBuiltInFunctionCalled(input)
	}
	return &vmcommon.VMOutput{}, nil
}

// GetBuiltinFunctionNames mocked method
func (b *BlockchainHookStub) GetBuiltinFunctionNames() vmcommon.FunctionNames {
	if b.GetBuiltinFunctionNamesCalled != nil {
		return b.GetBuiltinFunctionNamesCalled()
	}
	return make(vmcommon.FunctionNames)
}

// GetAllState mocked method
func (b *BlockchainHookStub) GetAllState(address []byte) (map[string][]byte, error) {
	if b.GetAllStateCalled != nil {
		return b.GetAllStateCalled(address)
	}
	return nil, nil
}

// GetUserAccount mocked method
func (b *BlockchainHookStub) GetUserAccount(address []byte) (vmcommon.UserAccountHandler, error) {
	if b.GetUserAccountCalled != nil {
		return b.GetUserAccountCalled(address)
	}
	return nil, nil
}

// GetESDTToken mocked method
func (b *BlockchainHookStub) GetESDTToken(address []byte, tokenID []byte, nonce uint64) (*esdt.ESDigitalToken, error) {
	if b.GetESDTTokenCalled != nil {
		return b.GetESDTTokenCalled(address, tokenID, nonce)
	}
	return &esdt.ESDigitalToken{Value: big.NewInt(0)}, nil
}

// GetCode mocked method
func (b *BlockchainHookStub) GetCode(account vmcommon.UserAccountHandler) []byte {
	if b.GetCodeCalled != nil {
		return b.GetCodeCalled(account)
	}
	return nil
}

// GetShardOfAddress mocked method
func (b *BlockchainHookStub) GetShardOfAddress(address []byte) uint32 {
	if b.GetShardOfAddressCalled != nil {
		return b.GetShardOfAddressCalled(address)
	}
	return 0
}

// IsSmartContract mocked method
func (b *BlockchainHookStub) IsSmartContract(address []byte) bool {
	if b.IsSmartContractCalled != nil {
		return b.IsSmartContractCalled(address)
	}
	return false
}

// IsPayable mocked method
func (b *BlockchainHookStub) IsPayable(_, address []byte) (bool, error) {
	if b.IsPayableCalled != nil {
		return b.IsPayableCalled(address)
	}
	return true, nil
}

// SaveCompiledCode mocked method
func (b *BlockchainHookStub) SaveCompiledCode(codeHash []byte, code []byte) {
	if b.SaveCompiledCodeCalled != nil {
		b.SaveCompiledCodeCalled(codeHash, code)
	}
}

// GetCompiledCode mocked method
func (b *BlockchainHookStub) GetCompiledCode(codeHash []byte) (bool, []byte) {
	if b.GetCompiledCodeCalled != nil {
		return b.GetCompiledCodeCalled(codeHash)
	}
	return false, nil
}

// ClearCompiledCodes mocked method
func (b *BlockchainHookStub) ClearCompiledCodes() {
}

// GetSnapshot mocked method
func (b *BlockchainHookStub) GetSnapshot() int {
	if b.GetSnapshotCalled != nil {
		return b.GetSnapshotCalled()
	}
	return 1
}

// RevertToSnapshot mocked method
func (b *BlockchainHookStub) RevertToSnapshot(snapshot int) error {
	if b.RevertToSnapshotCalled != nil {
		return b.RevertToSnapshotCalled(snapshot)
	}
	return nil
}

// IsPaused -
func (b *BlockchainHookStub) IsPaused(_ []byte) bool {
	return false
}

// IsLimitedTransfer -
func (b *BlockchainHookStub) IsLimitedTransfer(_ []byte) bool {
	return false
}

// ExecuteSmartContractCallOnOtherVM -
func (b *BlockchainHookStub) ExecuteSmartContractCallOnOtherVM(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	if b.ExecuteSmartContractCallOnOtherVMCalled != nil {
		return b.ExecuteSmartContractCallOnOtherVMCalled(input)
	}
	return nil, nil
}

// IsInterfaceNil mocked method
func (b *BlockchainHookStub) IsInterfaceNil() bool {
	return b == nil
}
