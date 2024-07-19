package testcommon

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-scenario-go/worldmock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	"github.com/multiversx/mx-chain-vm-go/config"
	"github.com/multiversx/mx-chain-vm-go/executor"
	executorwrapper "github.com/multiversx/mx-chain-vm-go/executor/wrapper"
	"github.com/multiversx/mx-chain-vm-go/testcommon/testexecutor"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
	"github.com/multiversx/mx-chain-vm-go/vmhost/hostCore"
	"github.com/multiversx/mx-chain-vm-go/vmhost/mock"
	"github.com/stretchr/testify/require"
)

// TestHostBuilder allows tests to configure and initialize the VM host and blockhain mock on which they operate.
type TestHostBuilder struct {
	tb               testing.TB
	blockchainHook   vmcommon.BlockchainHook
	vmHostParameters *vmhost.VMHostParameters
	host             vmhost.VMHost
}

// NewTestHostBuilder commences a test host builder pattern.
func NewTestHostBuilder(tb testing.TB) *TestHostBuilder {
	esdtTransferParser, _ := parsers.NewESDTTransferParser(worldmock.WorldMarshalizer)
	return &TestHostBuilder{
		tb: tb,
		vmHostParameters: &vmhost.VMHostParameters{
			VMType:                    DefaultVMType,
			BlockGasLimit:             uint64(1000),
			GasSchedule:               nil,
			BuiltInFuncContainer:      nil,
			ProtectedKeyPrefix:        []byte("E" + "L" + "R" + "O" + "N" + "D"),
			ESDTTransferParser:        esdtTransferParser,
			EpochNotifier:             &mock.EpochNotifierStub{},
			EnableEpochsHandler:       worldmock.EnableEpochsHandlerStubAllFlags(),
			OverrideVMExecutor:        nil,
			WasmerSIGSEGVPassthrough:  false,
			Hasher:                    defaultHasher,
			MapOpcodeAddressIsAllowed: map[string]map[string]struct{}{},
		},
	}
}

// Ensures gas costs are initialized.
func (thb *TestHostBuilder) initializeGasCosts() {
	if thb.vmHostParameters.GasSchedule == nil {
		thb.vmHostParameters.GasSchedule = config.MakeGasMapForTests()
	}
}

// Ensures the built-in function container is initialized.
func (thb *TestHostBuilder) initializeBuiltInFuncContainer() {
	if thb.vmHostParameters.BuiltInFuncContainer == nil {
		thb.vmHostParameters.BuiltInFuncContainer = builtInFunctions.NewBuiltInFunctionContainer()
	}

}

// WithBlockchainHook sets a pre-built blockchain hook for the VM to work with.
func (thb *TestHostBuilder) WithBlockchainHook(blockchainHook vmcommon.BlockchainHook) *TestHostBuilder {
	thb.blockchainHook = blockchainHook
	return thb
}

// WithBuiltinFunctions sets up builtin functions in the blockchain hook.
// Only works if the blockchain hook is of type worldmock.MockWorld.
func (thb *TestHostBuilder) WithBuiltinFunctions() *TestHostBuilder {
	thb.initializeGasCosts()
	mockWorld, ok := thb.blockchainHook.(*worldmock.MockWorld)
	require.True(thb.tb, ok, "builtin functions can only be injected into blockchain hooks of type MockWorld")
	err := mockWorld.InitBuiltinFunctions(thb.vmHostParameters.GasSchedule)
	require.Nil(thb.tb, err)
	thb.vmHostParameters.BuiltInFuncContainer = mockWorld.BuiltinFuncs.Container
	return thb
}

// WithExecutorFactory allows tests to choose what executor to use.
func (thb *TestHostBuilder) WithExecutorFactory(executorFactory executor.ExecutorAbstractFactory) *TestHostBuilder {
	thb.vmHostParameters.OverrideVMExecutor = executorFactory
	return thb
}

// WithExecutorLogs sets an ExecutorLogger, which wraps the existing OverrideVMExecutor
func (thb *TestHostBuilder) WithExecutorLogs(executorLogger executorwrapper.ExecutorLogger) *TestHostBuilder {
	if thb.vmHostParameters.OverrideVMExecutor == nil {
		thb.tb.Fatal("WithExecutorLogs() requires WithExecutorFactory()")
	}

	wrapper := executorwrapper.NewWrappedExecutorFactory(
		executorLogger,
		thb.vmHostParameters.OverrideVMExecutor)

	return thb.WithExecutorFactory(wrapper)
}

// WithWasmerSIGSEGVPassthrough allows tests to configure the WasmerSIGSEGVPassthrough flag.
func (thb *TestHostBuilder) WithWasmerSIGSEGVPassthrough(wasmerSIGSEGVPassthrough bool) *TestHostBuilder {
	thb.vmHostParameters.WasmerSIGSEGVPassthrough = wasmerSIGSEGVPassthrough
	return thb
}

// WithGasSchedule allows tests to use the gas costs. The default is config.MakeGasMapForTests().
func (thb *TestHostBuilder) WithGasSchedule(gasSchedule config.GasScheduleMap) *TestHostBuilder {
	thb.vmHostParameters.GasSchedule = gasSchedule
	return thb
}

// WithEnableEpochsHandler sets an enable epochs handler for the VM to work with.
func (thb *TestHostBuilder) WithEnableEpochsHandler(handler vmcommon.EnableEpochsHandler) *TestHostBuilder {
	thb.vmHostParameters.EnableEpochsHandler = handler
	return thb
}

// Build initializes the VM host with all configured options.
func (thb *TestHostBuilder) Build() vmhost.VMHost {
	thb.initializeHost()
	return thb.host
}

func (thb *TestHostBuilder) initializeHost() {
	thb.initializeGasCosts()
	if thb.host == nil {
		thb.host = thb.newHost()
	}
}

func (thb *TestHostBuilder) newHost() vmhost.VMHost {
	if check.IfNil(thb.vmHostParameters.OverrideVMExecutor) {
		exec := testexecutor.NewDefaultTestExecutorFactory(thb.tb)
		thb.vmHostParameters.OverrideVMExecutor = exec
	}

	thb.initializeBuiltInFuncContainer()
	host, err := hostCore.NewVMHost(
		thb.blockchainHook,
		thb.vmHostParameters,
	)
	require.Nil(thb.tb, err)
	require.NotNil(thb.tb, host)

	return host
}
