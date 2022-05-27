package testcommon

import (
	"testing"

	"github.com/ElrondNetwork/arwen-wasm-vm/v1_5/arwen"
	mock "github.com/ElrondNetwork/arwen-wasm-vm/v1_5/mock/context"
)

// TestConfig is configuration for async call tests
type TestConfig struct {
	ChildAddress      []byte
	ThirdPartyAddress []byte
	VaultAddress      []byte

	GasProvided           uint64
	GasProvidedToChild    uint64
	GasProvidedToCallback uint64
	GasUsedByParent       uint64
	GasUsedByChild        uint64
	GasUsedByCallback     uint64
	GasLockCost           uint64
	GasToLock             uint64

	ParentBalance int64
	ChildBalance  int64

	TransferFromParentToChild int64
	TransferToThirdParty      int64
	TransferToVault           int64
	TransferFromChildToParent int64

	ESDTTokensToTransfer         uint64
	CallbackESDTTokensToTransfer uint64

	ChildCalls          int
	RecursiveChildCalls int

	DeployedContractAddress []byte
	GasUsedByInit           uint64
	GasProvidedForInit      uint64
	AsyncCallStepCost       uint64
	AoTPreparePerByteCost   uint64
	CompilePerByteCost      uint64

	ContractToBeUpdatedAddress []byte
	Owner                      []byte
	IsFlagEnabled              bool
	HasCallback                bool
	CallbackFails              bool

	IsLegacyAsync bool
}

func getAddressOrDefult(address []byte, defaultAddress []byte) []byte {
	if address == nil {
		return defaultAddress
	}
	return address
}

func (config *TestConfig) GetChildAddress() []byte {
	return getAddressOrDefult(config.ChildAddress, ChildAddress)
}

func (config *TestConfig) GetThirdPartyAddress() []byte {
	return getAddressOrDefult(config.ThirdPartyAddress, ThirdPartyAddress)
}

func (config *TestConfig) GetVaultAddress() []byte {
	return getAddressOrDefult(config.VaultAddress, VaultAddress)
}

type testSmartContract struct {
	address      []byte
	balance      int64
	config       *TestConfig
	shardID      uint32
	codeHash     []byte
	codeMetadata []byte
	ownerAddress []byte
}

// MockTestSmartContract represents the config data for the mock smart contract instance to be tested
type MockTestSmartContract struct {
	testSmartContract
	initMethods []func(*mock.InstanceMock, interface{})
	// used only temporarly for call graph building
	tempFunctionsList map[string]bool
}

// CreateMockContract build a contract to be used in a test creted with BuildMockInstanceCallTest
func CreateMockContract(address []byte) *MockTestSmartContract {
	return CreateMockContractOnShard(address, 0)
}

// CreateMockContractOnShard build a contract to be used in a test creted with BuildMockInstanceCallTest
func CreateMockContractOnShard(address []byte, shardID uint32) *MockTestSmartContract {
	return &MockTestSmartContract{
		testSmartContract: testSmartContract{
			address: address,
			shardID: shardID,
		},
		tempFunctionsList: make(map[string]bool, 0),
	}
}

// WithBalance provides the balance for the MockTestSmartContract
func (mockSC *MockTestSmartContract) WithBalance(balance int64) *MockTestSmartContract {
	mockSC.balance = balance
	return mockSC
}

// WithShardID provides the shardID for the MockTestSmartContract
func (mockSC *MockTestSmartContract) WithShardID(shardID uint32) *MockTestSmartContract {
	mockSC.shardID = shardID
	return mockSC
}

// WithConfig provides the config object for the MockTestSmartContract
func (mockSC *MockTestSmartContract) WithConfig(config *TestConfig) *MockTestSmartContract {
	mockSC.config = config
	return mockSC
}

// WithCodeMetadata provides the code metadata for the MockTestSmartContract
func (mockSC *MockTestSmartContract) WithCodeMetadata(codeMetadata []byte) *MockTestSmartContract {
	mockSC.codeMetadata = codeMetadata
	return mockSC
}

// WithCodeHash provides the code hash for the MockTestSmartContract
func (mockSC *MockTestSmartContract) WithCodeHash(codeHash []byte) *MockTestSmartContract {
	mockSC.codeHash = codeHash
	return mockSC
}

// WithOwnerAddress provides the owner address for the MockTestSmartContract
func (mockSC *MockTestSmartContract) WithOwnerAddress(ownerAddress []byte) *MockTestSmartContract {
	mockSC.ownerAddress = ownerAddress
	return mockSC
}

// WithMethods provides the methods for the MockTestSmartContract
func (mockSC *MockTestSmartContract) WithMethods(initMethods ...func(*mock.InstanceMock, interface{})) MockTestSmartContract {
	mockSC.initMethods = initMethods
	return *mockSC
}

func (mockSC *MockTestSmartContract) GetShardID() uint32 {
	return mockSC.shardID
}

func (mockSC MockTestSmartContract) Initialize(
	t testing.TB,
	host arwen.VMHost,
	imb *mock.InstanceBuilderMock,
	createContractAccounts bool,
) {
	instance := imb.CreateAndStoreInstanceMock(t, host, mockSC.address, mockSC.codeHash, mockSC.codeMetadata, mockSC.ownerAddress, mockSC.shardID, mockSC.balance, createContractAccounts)
	for _, initMethod := range mockSC.initMethods {
		initMethod(instance, mockSC.config)
	}
}
