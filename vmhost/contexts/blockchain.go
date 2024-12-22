package contexts

import (
	"github.com/multiversx/mx-chain-vm-go/vmhost/vmhooks"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
)

var logBlockchain = logger.GetOrCreate("vm/blockchainContext")

type blockchainContext struct {
	host           vmhost.VMHost
	blockChainHook vmcommon.BlockchainHook
	stateStack     []int
}

// NewBlockchainContext creates a new blockchainContext
func NewBlockchainContext(
	host vmhost.VMHost,
	blockChainHook vmcommon.BlockchainHook,
) (*blockchainContext, error) {
	if check.IfNil(host) {
		return nil, vmhost.ErrNilVMHost
	}

	context := &blockchainContext{
		blockChainHook: blockChainHook,
		host:           host,
	}

	return context, nil
}

// NewAddress returns a new address created using the provided creator address and its nonce.
func (context *blockchainContext) NewAddress(creatorAddress []byte) ([]byte, error) {
	nonce, err := context.GetNonce(creatorAddress)
	if err != nil {
		return nil, err
	}

	isIndirectDeployment := context.IsSmartContract(creatorAddress)
	if !isIndirectDeployment && nonce > 0 {
		nonce--
	}

	vmType := context.host.Runtime().GetVMType()
	return context.blockChainHook.NewAddress(creatorAddress, nonce, vmType)
}

// AccountExists verifies if the provided address exists.
func (context *blockchainContext) AccountExists(address []byte) bool {
	account, err := context.blockChainHook.GetUserAccount(address)
	if err != nil {
		return false
	}

	exists := !vmhost.IfNil(account)
	return exists
}

// GetBalance returns the balance of the account at the given address as a byte array.
// If there is no account at that address, big.NewInt(0).Bytes() will be returned.
func (context *blockchainContext) GetBalance(address []byte) []byte {
	return context.GetBalanceBigInt(address).Bytes()
}

// GetBalanceBigInt returns the balance of the account at the given address as a big.Int.
// If there is no account at that address, 0 will be returned.
func (context *blockchainContext) GetBalanceBigInt(address []byte) *big.Int {
	outputAccount, isNew := context.host.Output().GetOutputAccount(address)
	if !isNew {
		if outputAccount.Balance == nil {
			account, err := context.blockChainHook.GetUserAccount(address)
			if err != nil || vmhost.IfNil(account) {
				return big.NewInt(0)
			}

			outputAccount.Balance = account.GetBalance()
		}

		balance := big.NewInt(0).Add(outputAccount.Balance, outputAccount.BalanceDelta)
		return balance
	}

	account, err := context.blockChainHook.GetUserAccount(address)
	if err != nil || vmhost.IfNil(account) {
		return big.NewInt(0)
	}

	balance := account.GetBalance()
	outputAccount.Balance = balance

	return balance
}

// GetNonce retrieves the nonce of the account at the given address.
func (context *blockchainContext) GetNonce(address []byte) (uint64, error) {
	outputAccount, isNew := context.host.Output().GetOutputAccount(address)

	readNonceFromBlockChain := isNew || outputAccount.Nonce == 0
	if !readNonceFromBlockChain {
		return outputAccount.Nonce, nil
	}

	account, err := context.blockChainHook.GetUserAccount(address)
	if err != nil || vmhost.IfNil(account) {
		return 0, err
	}

	nonce := account.GetNonce()
	outputAccount.Nonce = nonce

	return nonce, nil
}

// IncreaseNonce increments the nonce of the account at the given address.
func (context *blockchainContext) IncreaseNonce(address []byte) {
	nonce, _ := context.GetNonce(address)
	outputAccount, _ := context.host.Output().GetOutputAccount(address)
	outputAccount.Nonce = nonce + 1
}

// GetESDTToken returns the unmarshalled esdt token for the given address and nonce for NFTs
func (context *blockchainContext) GetESDTToken(address []byte, tokenID []byte, nonce uint64) (*esdt.ESDigitalToken, error) {
	return context.blockChainHook.GetESDTToken(address, tokenID, nonce)
}

// GetCodeHash retrieves the hash of the code stored under the given address.
func (context *blockchainContext) GetCodeHash(address []byte) []byte {
	account, err := context.blockChainHook.GetUserAccount(address)
	if err != nil {
		return nil
	}
	if vmhost.IfNil(account) {
		return nil
	}

	codeHash := account.GetCodeHash()
	return codeHash
}

// GetCode retrieves the code stored under the given address.
func (context *blockchainContext) GetCode(address []byte) ([]byte, error) {
	outputAccount, isNew := context.host.Output().GetOutputAccount(address)
	hasCode := !isNew && len(outputAccount.Code) > 0
	if hasCode {
		return outputAccount.Code, nil
	}

	account, err := context.blockChainHook.GetUserAccount(address)
	if err != nil {
		return nil, err
	}
	if vmhost.IfNil(account) {
		return nil, vmhost.ErrInvalidAccount
	}

	code := context.blockChainHook.GetCode(account)
	if len(code) == 0 {
		return nil, vmhost.ErrContractNotFound
	}

	outputAccount.Code = code

	return code, nil
}

// GetCodeSize returns the size of the code stored under the given address.
func (context *blockchainContext) GetCodeSize(address []byte) (int32, error) {
	account, err := context.blockChainHook.GetUserAccount(address)
	if err != nil || vmhost.IfNil(account) {
		return 0, err
	}

	code := context.blockChainHook.GetCode(account)
	result := int32(len(code))
	return result, nil
}

// BlockHash returns the hash of the block that has the given nonce.
func (context *blockchainContext) BlockHash(number uint64) []byte {
	block, err := context.blockChainHook.GetBlockhash(number)
	if err != nil {
		return nil
	}

	return block
}

// CurrentEpoch returns the number of the current epoch.
func (context *blockchainContext) CurrentEpoch() uint32 {
	return context.blockChainHook.CurrentEpoch()
}

// CurrentNonce returns the nonce of the block currently being built.
func (context *blockchainContext) CurrentNonce() uint64 {
	return context.blockChainHook.CurrentNonce()
}

// GetStateRootHash returns the root hash of the entire state.
func (context *blockchainContext) GetStateRootHash() []byte {
	return context.blockChainHook.GetStateRootHash()
}

// LastTimeStamp returns the timestamp of the last commited block
func (context *blockchainContext) LastTimeStamp() uint64 {
	return context.blockChainHook.LastTimeStamp()
}

// LastNonce returns the nonce of the last commited block.
func (context *blockchainContext) LastNonce() uint64 {
	return context.blockChainHook.LastNonce()
}

// LastRound returns the round of the last commited block.
func (context *blockchainContext) LastRound() uint64 {
	return context.blockChainHook.LastRound()
}

// LastEpoch returns the epoch number of the last commited block.
func (context *blockchainContext) LastEpoch() uint32 {
	return context.blockChainHook.LastEpoch()
}

// CurrentRound returns the round of the block currently being built.
func (context *blockchainContext) CurrentRound() uint64 {
	return context.blockChainHook.CurrentRound()
}

// CurrentTimeStamp returns the timestamp of the block currently being built.
func (context *blockchainContext) CurrentTimeStamp() uint64 {
	return context.blockChainHook.CurrentTimeStamp()
}

// LastRandomSeed returns the randomness seed of the last commited block.
func (context *blockchainContext) LastRandomSeed() []byte {
	return context.blockChainHook.LastRandomSeed()
}

// CurrentRandomSeed returns the random seed from header of the block being built.
func (context *blockchainContext) CurrentRandomSeed() []byte {
	return context.blockChainHook.CurrentRandomSeed()
}

// RoundTime returns the duruation of a round
func (context *blockchainContext) RoundTime() uint64 {
	return context.blockChainHook.RoundTime()
}

// EpochStartBlockTimeStamp returns the timestamp of the first block of the current epoch.
func (context *blockchainContext) EpochStartBlockTimeStamp() uint64 {
	return context.blockChainHook.EpochStartBlockTimeStamp()
}

// EpochStartBlockNonce returns the nonce of the first block of the current epoch.
func (context *blockchainContext) EpochStartBlockNonce() uint64 {
	return context.blockChainHook.EpochStartBlockNonce()
}

// EpochStartBlockRound returns the round of the first block of the current epoch.
func (context *blockchainContext) EpochStartBlockRound() uint64 {
	return context.blockChainHook.EpochStartBlockRound()
}

// GetOwnerAddress returns the owner address of the contract being executed.
func (context *blockchainContext) GetOwnerAddress() ([]byte, error) {
	scAddress := context.host.Runtime().GetContextAddress()
	scAccount, err := context.blockChainHook.GetUserAccount(scAddress)
	if err != nil || vmhost.IfNil(scAccount) {
		return nil, err
	}

	return scAccount.GetOwnerAddress(), nil
}

// GetShardOfAddress returns the number of the shard containing the given address.
func (context *blockchainContext) GetShardOfAddress(addr []byte) uint32 {
	return context.blockChainHook.GetShardOfAddress(addr)
}

// IsSmartContract verifies whether the provided address is a smart contract or not.
func (context *blockchainContext) IsSmartContract(addr []byte) bool {
	return context.blockChainHook.IsSmartContract(addr)
}

// IsPayable returns true if the SC at the given address is payable
func (context *blockchainContext) IsPayable(sndAddress []byte, rcvAddress []byte) (bool, error) {
	return context.blockChainHook.IsPayable(sndAddress, rcvAddress)
}

// SaveCompiledCode stores the provided precompiled binary code under the specified hash.
func (context *blockchainContext) SaveCompiledCode(codeHash []byte, code []byte) {
	context.blockChainHook.SaveCompiledCode(codeHash, code)
}

// GetCompiledCode retrieves the precompiled binary code stored under the specified hash.
func (context *blockchainContext) GetCompiledCode(codeHash []byte) (bool, []byte) {
	return context.blockChainHook.GetCompiledCode(codeHash)
}

// GetUserAccount returns a user account
func (context *blockchainContext) GetUserAccount(address []byte) (vmcommon.UserAccountHandler, error) {
	return context.blockChainHook.GetUserAccount(address)
}

// ProcessBuiltInFunction will process the builtIn function for the created input
func (context *blockchainContext) ProcessBuiltInFunction(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	return context.blockChainHook.ProcessBuiltInFunction(input)
}

// InitState does nothing
func (context *blockchainContext) InitState() {
}

// ClearStateStack clears the state stack from the current context.
func (context *blockchainContext) ClearStateStack() {
	context.stateStack = make([]int, 0)
}

// PushState appends the current snapshot to the state stack.
func (context *blockchainContext) PushState() {
	snapshot := context.blockChainHook.GetSnapshot()
	context.stateStack = append(context.stateStack, snapshot)
}

// PopSetActiveState removes the latest entry from the state stack and reverts to that snapshot
func (context *blockchainContext) PopSetActiveState() {
	stateStackLen := len(context.stateStack)
	if stateStackLen == 0 {
		return
	}

	prevSnapshot := context.stateStack[stateStackLen-1]
	err := context.blockChainHook.RevertToSnapshot(prevSnapshot)
	if err != nil {
		vmhooks.FailExecution(context.host, err)
		context.host.Runtime().AddError(err, "RevertToSnapshot")
		logBlockchain.Error("PopSetActiveState RevertToSnapshot", "error", err)
		return
	}

	context.stateStack = context.stateStack[:stateStackLen-1]
}

// PopDiscard removes the latest entry from the state stack
func (context *blockchainContext) PopDiscard() {
	stateStackLen := len(context.stateStack)
	if stateStackLen == 0 {
		return
	}

	context.stateStack = context.stateStack[:stateStackLen-1]
}

// GetSnapshot - gets the latest snapshot via blockchain hook
func (context *blockchainContext) GetSnapshot() int {
	return context.blockChainHook.GetSnapshot()
}

// RevertToSnapshot - reverts to the specified snapshot via blockchain hook
func (context *blockchainContext) RevertToSnapshot(snapshot int) {
	_ = context.blockChainHook.RevertToSnapshot(snapshot)
}

// IsLimitedTransfer returns true if token transfers are limited
func (context *blockchainContext) IsLimitedTransfer(tokenID []byte) bool {
	return context.blockChainHook.IsLimitedTransfer(tokenID)
}

// IsPaused returns true if tokenID is paused
func (context *blockchainContext) IsPaused(tokenID []byte) bool {
	return context.blockChainHook.IsPaused(tokenID)
}

// ClearCompiledCodes cleans the compiled codes cache
func (context *blockchainContext) ClearCompiledCodes() {
	context.blockChainHook.ClearCompiledCodes()
}

// ExecuteSmartContractCallOnOtherVM runs contract on another VM
func (context *blockchainContext) ExecuteSmartContractCallOnOtherVM(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	return context.blockChainHook.ExecuteSmartContractCallOnOtherVM(input)
}
