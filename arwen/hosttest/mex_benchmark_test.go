package hosttest

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/wasm-vm/arwen"
	worldmock "github.com/ElrondNetwork/wasm-vm/mock/world"
	test "github.com/ElrondNetwork/wasm-vm/testcommon"
	"github.com/stretchr/testify/require"
)

func Test_RunDEXPairBenchmark(t *testing.T) {
	owner := arwen.MakeTestWalletAddress("owner")
	user := arwen.MakeTestWalletAddress("user")

	_, host, mex := setupMEXPair(t, owner, user)

	mex.AddLiquidity(user, 1_000_000, 1, 1_000_000, 1)

	for i := 0; i < 10_000; i++ {
		mex.Swap(mex.WEGLDToken, 100, mex.MEXToken, 1)
		mex.Swap(mex.MEXToken, 100, mex.WEGLDToken, 1)
	}

	defer func() {
		host.Reset()
	}()
}

func setupMEXPair(t *testing.T, owner Address, user Address) (*worldmock.MockWorld, arwen.VMHost, *MEXSetup) {
	world, ownerAccount, host, err := prepare(t, owner)
	require.Nil(t, err)

	userAccount := world.AcctMap.CreateAccount(user, world)
	userAccount.Balance = big.NewInt(100)
	mex := NewMEXSetup(t, host, world, ownerAccount, userAccount)
	mex.Deploy()

	mex.ApplyInitialSetup()

	return world, host, mex
}

type MEXSetup struct {
	WEGLDToken               []byte
	MEXToken                 []byte
	LPToken                  []byte
	OwnerAccount             *worldmock.Account
	OwnerAddress             Address
	RouterAddress            Address
	PairAddress              Address
	TotalFeePercent          uint64
	SpecialFeePercent        uint64
	MaxObservationsPerRecord int
	Code                     []byte
	UserAccount              *worldmock.Account
	UserWEGLDBalance         uint64
	UserMEXBalance           uint64

	T     *testing.T
	Host  arwen.VMHost
	World *worldmock.MockWorld
}

func NewMEXSetup(
	t *testing.T,
	host arwen.VMHost,
	world *worldmock.MockWorld,
	ownerAccount *worldmock.Account,
	userAccount *worldmock.Account,
) *MEXSetup {
	return &MEXSetup{
		WEGLDToken:               []byte("WEGLD-abcdef"),
		MEXToken:                 []byte("MEX-abcdef"),
		LPToken:                  []byte("LPTOK-abcdef"),
		OwnerAccount:             ownerAccount,
		OwnerAddress:             ownerAccount.Address,
		RouterAddress:            ownerAccount.Address,
		PairAddress:              test.MakeTestSCAddress("pairSC"),
		TotalFeePercent:          300,
		SpecialFeePercent:        50,
		MaxObservationsPerRecord: 10,
		Code:                     test.GetTestSCCode("pair", "../../"),
		UserAccount:              userAccount,
		UserWEGLDBalance:         5_000_000_000,
		UserMEXBalance:           5_000_000_000,

		T:     t,
		Host:  host,
		World: world,
	}
}

func (mex *MEXSetup) Deploy() {
	t := mex.T
	host := mex.Host
	world := mex.World

	vmInput := test.CreateTestContractCreateInputBuilder().
		WithCallerAddr(mex.OwnerAddress).
		WithContractCode(mex.Code).
		WithArguments(
			mex.WEGLDToken,
			mex.MEXToken,
			mex.OwnerAddress,
			mex.RouterAddress,
			big.NewInt(int64(mex.TotalFeePercent)).Bytes(),
			big.NewInt(int64(mex.SpecialFeePercent)).Bytes(),
		).
		WithGasProvided(0xFFFFFFFFFFFFFFFF).
		Build()

	world.NewAddressMocks = append(world.NewAddressMocks, &worldmock.NewAddressMock{
		CreatorAddress: mex.OwnerAddress,
		CreatorNonce:   mex.OwnerAccount.Nonce,
		NewAddress:     mex.PairAddress,
	})

	mex.OwnerAccount.Nonce++ // nonce increases before deploy
	vmOutput, err := host.RunSmartContractCreate(vmInput)
	require.Nil(t, err)
	require.NotNil(t, vmOutput)
	require.Equal(t, "", vmOutput.ReturnMessage)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	_ = world.UpdateAccounts(vmOutput.OutputAccounts, nil)
}

func (mex *MEXSetup) ApplyInitialSetup() {
	mex.setLPToken()
	mex.setActiveState()
	mex.setMaxObservationsPerRecord()
	mex.setRequiredTokenRoles()
	mex.setESDTBalances()
}

func (mex *MEXSetup) setLPToken() {
	t := mex.T
	host := mex.Host
	world := mex.World

	vmInput := test.CreateTestContractCallInputBuilder().
		WithCallerAddr(mex.OwnerAddress).
		WithRecipientAddr(mex.PairAddress).
		WithFunction("setLpTokenIdentifier").
		WithArguments(mex.LPToken).
		WithGasProvided(0xFFFFFFFFFFFFFFFF).
		Build()

	vmOutput, err := host.RunSmartContractCall(vmInput)
	require.Nil(t, err)
	require.NotNil(t, vmOutput)
	require.Equal(t, "", vmOutput.ReturnMessage)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	_ = world.UpdateAccounts(vmOutput.OutputAccounts, nil)
}

func (mex *MEXSetup) setActiveState() {
	t := mex.T
	host := mex.Host
	world := mex.World

	vmInput := test.CreateTestContractCallInputBuilder().
		WithCallerAddr(mex.OwnerAddress).
		WithRecipientAddr(mex.PairAddress).
		WithFunction("resume").
		WithGasProvided(0xFFFFFFFFFFFFFFFF).
		Build()

	vmOutput, err := host.RunSmartContractCall(vmInput)
	require.Nil(t, err)
	require.NotNil(t, vmOutput)
	require.Equal(t, "", vmOutput.ReturnMessage)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	_ = world.UpdateAccounts(vmOutput.OutputAccounts, nil)
}

func (mex *MEXSetup) setMaxObservationsPerRecord() {
	t := mex.T
	host := mex.Host
	world := mex.World

	vmInput := test.CreateTestContractCallInputBuilder().
		WithCallerAddr(mex.OwnerAddress).
		WithRecipientAddr(mex.PairAddress).
		WithFunction("setMaxObservationsPerRecord").
		WithArguments(big.NewInt(int64(mex.MaxObservationsPerRecord)).Bytes()).
		WithGasProvided(0xFFFFFFFFFFFFFFFF).
		Build()

	vmOutput, err := host.RunSmartContractCall(vmInput)
	require.Nil(t, err)
	require.NotNil(t, vmOutput)
	require.Equal(t, "", vmOutput.ReturnMessage)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	_ = world.UpdateAccounts(vmOutput.OutputAccounts, nil)
}

func (mex *MEXSetup) setRequiredTokenRoles() {
	world := mex.World
	pairAccount := world.AcctMap.GetAccount(mex.PairAddress)

	roles := []string{core.ESDTRoleLocalMint, core.ESDTRoleLocalBurn}
	pairAccount.SetTokenRolesAsStrings(mex.LPToken, roles)
}

func (mex *MEXSetup) setESDTBalances() {
	mex.UserAccount.SetTokenBalanceUint64(mex.WEGLDToken, 0, mex.UserWEGLDBalance)
	mex.UserAccount.SetTokenBalanceUint64(mex.MEXToken, 0, mex.UserMEXBalance)
}

func (mex *MEXSetup) AddLiquidity(
	userAddress Address,
	WEGLDAmount uint64,
	minWEGLDAmount uint64,
	MEXAmount uint64,
	minMEXAmount uint64,
) {
	t := mex.T
	host := mex.Host
	world := mex.World

	vmInputBuiler := test.CreateTestContractCallInputBuilder().
		WithCallerAddr(mex.UserAccount.Address).
		WithRecipientAddr(mex.PairAddress).
		WithFunction("addLiquidity").
		WithArguments(
			big.NewInt(int64(minWEGLDAmount)).Bytes(),
			big.NewInt(int64(minMEXAmount)).Bytes(),
		).
		WithGasProvided(0xFFFFFFFFFFFFFFFF)

	vmInputBuiler.
		WithESDTTokenName(mex.WEGLDToken).
		WithESDTValue(big.NewInt(int64(WEGLDAmount))).
		NextESDTTransfer().
		WithESDTTokenName(mex.MEXToken).
		WithESDTValue(big.NewInt(int64(MEXAmount)))

	vmInput := vmInputBuiler.Build()

	mex.performESDTTransfer(
		vmInput.CallerAddr,
		vmInput.RecipientAddr,
		vmInput.ESDTTransfers,
	)

	vmOutput, err := host.RunSmartContractCall(vmInput)
	require.Nil(t, err)
	require.NotNil(t, vmOutput)
	require.Equal(t, "", vmOutput.ReturnMessage)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	_ = world.UpdateAccounts(vmOutput.OutputAccounts, nil)
}

func (mex *MEXSetup) Swap(
	leftToken []byte,
	leftAmount uint64,
	rightToken []byte,
	rightAmount uint64,
) {
	t := mex.T
	host := mex.Host
	world := mex.World

	vmInputBuiler := test.CreateTestContractCallInputBuilder().
		WithCallerAddr(mex.UserAccount.Address).
		WithRecipientAddr(mex.PairAddress)

	vmInputBuiler.
		WithESDTTokenName(leftToken).
		WithESDTValue(big.NewInt(int64(leftAmount))).
		WithFunction("swapTokensFixedInput").
		WithArguments(
			rightToken,
			big.NewInt(int64(rightAmount)).Bytes(),
		).
		WithGasProvided(0xFFFFFFFFFFF)

	vmInput := vmInputBuiler.Build()

	mex.performESDTTransfer(
		vmInput.CallerAddr,
		vmInput.RecipientAddr,
		vmInput.ESDTTransfers,
	)

	vmOutput, err := host.RunSmartContractCall(vmInput)
	require.Nil(t, err)
	require.NotNil(t, vmOutput)
	require.Equal(t, "", vmOutput.ReturnMessage)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	_ = world.UpdateAccounts(vmOutput.OutputAccounts, nil)
}

func (mex *MEXSetup) performESDTTransfer(
	sender Address,
	receiver Address,
	esdtTransfers []*vmcommon.ESDTTransfer,
) {
	t := mex.T
	world := mex.World

	nrTransfers := len(esdtTransfers)
	nrTransfersAsBytes := big.NewInt(0).SetUint64(uint64(nrTransfers)).Bytes()

	multiTransferInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  sender,
			Arguments:   make([][]byte, 0),
			CallValue:   big.NewInt(0),
			CallType:    vm.DirectCall,
			GasPrice:    1,
			GasProvided: 0xFFFFFFFF,
			GasLocked:   0,
		},
		RecipientAddr:     sender,
		Function:          core.BuiltInFunctionMultiESDTNFTTransfer,
		AllowInitFunction: false,
	}
	multiTransferInput.Arguments = append(multiTransferInput.Arguments, receiver, nrTransfersAsBytes)

	for i := 0; i < nrTransfers; i++ {
		token := esdtTransfers[i].ESDTTokenName
		nonceAsBytes := big.NewInt(0).SetUint64(esdtTransfers[i].ESDTTokenNonce).Bytes()
		value := esdtTransfers[i].ESDTValue.Bytes()

		multiTransferInput.Arguments = append(multiTransferInput.Arguments, token, nonceAsBytes, value)
	}

	vmOutput, err := world.BuiltinFuncs.ProcessBuiltInFunction(multiTransferInput)
	require.Nil(t, err)
	require.NotNil(t, vmOutput)
	require.Equal(t, "", vmOutput.ReturnMessage)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	_ = world.UpdateAccounts(vmOutput.OutputAccounts, nil)
}
