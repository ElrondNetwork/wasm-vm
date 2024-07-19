package testcommon

import (
	"testing"

	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-go/config"
	"github.com/multiversx/mx-chain-vm-go/executor"
	executorwrapper "github.com/multiversx/mx-chain-vm-go/executor/wrapper"
	contextmock "github.com/multiversx/mx-chain-vm-go/mock/context"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
	"github.com/stretchr/testify/require"
)

// InstanceTestSmartContract represents the config data for the smart contract instance to be tested
type InstanceTestSmartContract struct {
	testSmartContract
	code []byte
}

// CreateInstanceContract build a contract to be used in a test creted with BuildInstanceCallTest
func CreateInstanceContract(address []byte) *InstanceTestSmartContract {
	return &InstanceTestSmartContract{
		testSmartContract: testSmartContract{
			address:      address,
			ownerAddress: UserAddress,
		},
	}
}

// WithBalance provides the balance for the InstanceTestSmartContract
func (mockSC *InstanceTestSmartContract) WithBalance(balance int64) *InstanceTestSmartContract {
	mockSC.balance = balance
	return mockSC
}

// WithConfig provides the config object for the InstanceTestSmartContract
func (mockSC *InstanceTestSmartContract) WithConfig(testConfig *TestConfig) *InstanceTestSmartContract {
	mockSC.config = testConfig
	return mockSC
}

// WithCode provides the code for the InstanceTestSmartContract
func (mockSC *InstanceTestSmartContract) WithCode(code []byte) *InstanceTestSmartContract {
	mockSC.code = code
	return mockSC
}

// WithOwner provides the owner for the InstanceTestSmartContract
func (mockSC *InstanceTestSmartContract) WithOwner(owner []byte) *InstanceTestSmartContract {
	mockSC.ownerAddress = owner
	return mockSC
}

// WithCodeMetadata provides the owner for the InstanceTestSmartContract
func (mockSC *InstanceTestSmartContract) WithCodeMetadata(metadata []byte) *InstanceTestSmartContract {
	mockSC.codeMetadata = metadata
	return mockSC
}

// InstanceCallTestTemplate holds the data to build a contract call test
type InstanceCallTestTemplate struct {
	testTemplateConfig
	contracts     []*InstanceTestSmartContract
	setup         func(vmhost.VMHost, *contextmock.BlockchainHookStub)
	assertResults func(vmhost.VMHost, *contextmock.BlockchainHookStub, *VMOutputVerifier)
	host          vmhost.VMHost
	hostBuilder   *TestHostBuilder
}

// BuildInstanceCallTest starts the building process for a contract call test
func BuildInstanceCallTest(tb testing.TB) *InstanceCallTestTemplate {
	return &InstanceCallTestTemplate{
		testTemplateConfig: testTemplateConfig{
			tb:                       tb,
			useMocks:                 false,
			wasmerSIGSEGVPassthrough: false,
		},
		hostBuilder: NewTestHostBuilder(tb),
		setup:       func(vmhost.VMHost, *contextmock.BlockchainHookStub) {},
	}
}

// WithContracts provides the contracts to be used by the contract call test
func (template *InstanceCallTestTemplate) WithContracts(usedContracts ...*InstanceTestSmartContract) *InstanceCallTestTemplate {
	template.contracts = usedContracts
	return template
}

// WithInput provides the ContractCallInput to be used by the contract call test
func (template *InstanceCallTestTemplate) WithInput(input *vmcommon.ContractCallInput) *InstanceCallTestTemplate {
	template.input = input
	return template
}

// WithSetup provides the setup function to be used by the contract call test
func (template *InstanceCallTestTemplate) WithSetup(setup func(vmhost.VMHost, *contextmock.BlockchainHookStub)) *InstanceCallTestTemplate {
	template.setup = setup
	return template
}

// WithGasSchedule provides gas schedule for the test
func (template *InstanceCallTestTemplate) WithGasSchedule(gasSchedule config.GasScheduleMap) *InstanceCallTestTemplate {
	template.hostBuilder.WithGasSchedule(gasSchedule)
	return template
}

// WithExecutorFactory provides the wasmer executor for the test
func (template *InstanceCallTestTemplate) WithExecutorFactory(executorFactory executor.ExecutorAbstractFactory) *InstanceCallTestTemplate {
	template.hostBuilder.WithExecutorFactory(executorFactory)
	return template
}

// WithExecutorLogs sets an ExecutorLogger
func (template *InstanceCallTestTemplate) WithExecutorLogs(executorLogger executorwrapper.ExecutorLogger) *InstanceCallTestTemplate {
	template.hostBuilder.WithExecutorLogs(executorLogger)
	return template
}

// WithWasmerSIGSEGVPassthrough sets the wasmerSIGSEGVPassthrough flag
func (template *InstanceCallTestTemplate) WithWasmerSIGSEGVPassthrough(passthrough bool) *InstanceCallTestTemplate {
	template.hostBuilder.WithWasmerSIGSEGVPassthrough(passthrough)
	return template
}

// WithEnableEpochsHandler provides the enable epochs handler to be used by the mock contract call test
func (template *InstanceCallTestTemplate) WithEnableEpochsHandler(handler vmhost.EnableEpochsHandler) *InstanceCallTestTemplate {
	template.hostBuilder.WithEnableEpochsHandler(handler)
	return template
}

// GetVMHost returns the inner VMHost
func (template *InstanceCallTestTemplate) GetVMHost() vmhost.VMHost {
	return template.host
}

// AndAssertResults starts the test and asserts the results
func (template *InstanceCallTestTemplate) AndAssertResults(assertResults func(vmhost.VMHost, *contextmock.BlockchainHookStub, *VMOutputVerifier)) {
	template.assertResults = assertResults
	runTestWithInstances(template, true)
}

// AndAssertResultsWithoutReset starts the test and asserts the results
func (template *InstanceCallTestTemplate) AndAssertResultsWithoutReset(assertResults func(vmhost.VMHost, *contextmock.BlockchainHookStub, *VMOutputVerifier)) {
	template.assertResults = assertResults
	runTestWithInstances(template, false)
}

func runTestWithInstances(template *InstanceCallTestTemplate, reset bool) {
	var blhookStub *contextmock.BlockchainHookStub
	if template.host == nil {
		blhookStub = BlockchainHookStubForContracts(template.contracts)
		template.hostBuilder.WithBlockchainHook(blhookStub)
		template.host = template.hostBuilder.Build()
		template.setup(template.host, blhookStub)
	}

	defer func() {
		if reset {
			template.host.Reset()
		}

		// Extra verification for instance leaks
		err := template.host.Runtime().ValidateInstances()
		require.Nil(template.tb, err)
	}()

	vmOutput, err := template.host.RunSmartContractCall(template.input)

	if template.assertResults != nil {
		allErrors := template.host.Runtime().GetAllErrors()
		verify := NewVMOutputVerifierWithAllErrors(template.tb, vmOutput, err, allErrors)
		template.assertResults(template.host, blhookStub, verify)
	}
}
