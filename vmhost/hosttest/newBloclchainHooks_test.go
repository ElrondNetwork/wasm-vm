package hostCoretest

import (
	"math/big"
	"testing"

	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-go/testcommon"

	"github.com/stretchr/testify/require"
)

func TestNewBlockchainHooks_GetRoundTime(t *testing.T) {
	code := testcommon.GetTestSCCode("new-blockchain-hooks", "../../")

	blockchainHook := testcommon.BlockchainHookStubForCall(code, nil)
	blockchainHook.RoundTimeCalled = func() uint64 {
		return 32
	}

	host := testcommon.NewTestHostBuilder(t).
		WithBlockchainHook(blockchainHook).
		Build()
	defer func() {
		host.Reset()
	}()

	input := testcommon.DefaultTestContractCallInput()
	input.GasProvided = 100000
	input.Function = "test_round_time"
	input.CallValue = big.NewInt(64)

	vmOutput, err := host.RunSmartContractCall(input)
	require.Nil(t, err)

	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	require.Len(t, vmOutput.ReturnData, 1)
	require.Equal(t, "", vmOutput.ReturnMessage)

	require.Equal(t, []byte{32}, vmOutput.ReturnData[0])
}

func TestNewBlockchainHooks_EpochStartBlockTimeStamp(t *testing.T) {
	code := testcommon.GetTestSCCode("new-blockchain-hooks", "../../")

	blockchainHook := testcommon.BlockchainHookStubForCall(code, nil)
	blockchainHook.EpochStartBlockTimeStampCalled = func() uint64 {
		return 31
	}

	host := testcommon.NewTestHostBuilder(t).
		WithBlockchainHook(blockchainHook).
		Build()
	defer func() {
		host.Reset()
	}()

	input := testcommon.DefaultTestContractCallInput()
	input.GasProvided = 100000
	input.Function = "test_epoch_start_block_time_stamp"
	input.CallValue = big.NewInt(64)

	vmOutput, err := host.RunSmartContractCall(input)
	require.Nil(t, err)

	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	require.Len(t, vmOutput.ReturnData, 1)
	require.Equal(t, "", vmOutput.ReturnMessage)

	require.Equal(t, []byte{31}, vmOutput.ReturnData[0])
}

func TestNewBlockchainHooks_EpochStartBlockNonce(t *testing.T) {
	code := testcommon.GetTestSCCode("new-blockchain-hooks", "../../")

	blockchainHook := testcommon.BlockchainHookStubForCall(code, nil)
	blockchainHook.EpochStartBlockNonceCalled = func() uint64 {
		return 30
	}

	host := testcommon.NewTestHostBuilder(t).
		WithBlockchainHook(blockchainHook).
		Build()
	defer func() {
		host.Reset()
	}()

	input := testcommon.DefaultTestContractCallInput()
	input.GasProvided = 100000
	input.Function = "test_epoch_start_block_nonce"
	input.CallValue = big.NewInt(64)

	vmOutput, err := host.RunSmartContractCall(input)
	require.Nil(t, err)

	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	require.Len(t, vmOutput.ReturnData, 1)
	require.Equal(t, "", vmOutput.ReturnMessage)

	require.Equal(t, []byte{30}, vmOutput.ReturnData[0])
}

func TestNewBlockchainHooks_EpochStartBlockRound(t *testing.T) {
	code := testcommon.GetTestSCCode("new-blockchain-hooks", "../../")

	blockchainHook := testcommon.BlockchainHookStubForCall(code, nil)
	blockchainHook.EpochStartBlockRoundCalled = func() uint64 {
		return 29
	}

	host := testcommon.NewTestHostBuilder(t).
		WithBlockchainHook(blockchainHook).
		Build()
	defer func() {
		host.Reset()
	}()

	input := testcommon.DefaultTestContractCallInput()
	input.GasProvided = 100000
	input.Function = "test_epoch_start_block_round"
	input.CallValue = big.NewInt(64)

	vmOutput, err := host.RunSmartContractCall(input)
	require.Nil(t, err)

	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	require.Len(t, vmOutput.ReturnData, 1)
	require.Equal(t, "", vmOutput.ReturnMessage)

	require.Equal(t, []byte{29}, vmOutput.ReturnData[0])
}
