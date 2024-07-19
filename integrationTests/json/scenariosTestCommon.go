package vmjsonintegrationtest

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	scenexec "github.com/multiversx/mx-chain-scenario-go/scenario/executor"
	scenio "github.com/multiversx/mx-chain-scenario-go/scenario/io"
	"github.com/multiversx/mx-chain-scenario-go/worldmock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-go/executor"
	executorwrapper "github.com/multiversx/mx-chain-vm-go/executor/wrapper"
	vmscenario "github.com/multiversx/mx-chain-vm-go/scenario"
	"github.com/multiversx/mx-chain-vm-go/testcommon/testexecutor"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = logger.SetLogLevel("*:NONE")
}

func getTestRoot() string {
	exePath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	vmTestRoot := filepath.Join(exePath, "../../test")
	return vmTestRoot
}

// ScenariosTestBuilder defines the Scenarios builder component
type ScenariosTestBuilder struct {
	t                   *testing.T
	folder              string
	singleFile          string
	exclusions          []string
	executorLogger      executorwrapper.ExecutorLogger
	executorFactory     executor.ExecutorAbstractFactory
	enableEpochsHandler vmcommon.EnableEpochsHandler
	currentError        error
	overrideVMType      []byte
}

// ScenariosTest will create a new ScenariosTestBuilder instance
func ScenariosTest(t *testing.T) *ScenariosTestBuilder {
	return &ScenariosTestBuilder{
		t:               t,
		folder:          "",
		singleFile:      "",
		executorLogger:  nil,
		executorFactory: nil,
		enableEpochsHandler: &worldmock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag != vmhost.ValidationOnGobDecodeFlag // relevant for big float tests. TODO: remove this when completely migrate to go1.22
			},
		},
	}
}

// Folder sets the folder
func (mtb *ScenariosTestBuilder) Folder(folder string) *ScenariosTestBuilder {
	mtb.folder = folder
	return mtb
}

// File sets the file
func (mtb *ScenariosTestBuilder) File(fileName string) *ScenariosTestBuilder {
	mtb.singleFile = fileName
	return mtb
}

// Exclude sets the exclusion path
func (mtb *ScenariosTestBuilder) Exclude(path string) *ScenariosTestBuilder {
	mtb.exclusions = append(mtb.exclusions, path)
	return mtb
}

// WithExecutorLogs sets a StringLogger
func (mtb *ScenariosTestBuilder) WithExecutorLogs() *ScenariosTestBuilder {
	mtb.executorLogger = executorwrapper.NewStringLogger()
	return mtb
}

// WithConsoleExecutorLogs sets a custom logger
func (mtb *ScenariosTestBuilder) WithConsoleExecutorLogs() *ScenariosTestBuilder {
	mtb.executorLogger = executorwrapper.NewConsoleLogger()
	return mtb
}

// WithExecutorFactory sets an executor factory
func (mtb *ScenariosTestBuilder) WithExecutorFactory(executorFactory executor.ExecutorAbstractFactory) *ScenariosTestBuilder {
	mtb.executorFactory = executorFactory
	return mtb
}

// WithEnableEpochsHandler overrides the epoch flags
func (mtb *ScenariosTestBuilder) WithEnableEpochsHandler(enableEpochsHandler vmcommon.EnableEpochsHandler) *ScenariosTestBuilder {
	mtb.enableEpochsHandler = enableEpochsHandler
	return mtb
}

// WithVMType overrides the default VM type
func (mtb *ScenariosTestBuilder) WithVMType(overrideVMType []byte) *ScenariosTestBuilder {
	mtb.overrideVMType = overrideVMType
	return mtb
}

// Run will start the testing process
func (mtb *ScenariosTestBuilder) Run() *ScenariosTestBuilder {
	if check.IfNil(mtb.executorFactory) {
		mtb.executorFactory = testexecutor.NewDefaultTestExecutorFactory(mtb.t)
	}

	vmBuilder := vmscenario.NewScenarioVMHostBuilder()
	vmBuilder.OverrideVMExecutor = mtb.executorFactory
	if mtb.overrideVMType != nil {
		vmBuilder.VMType = mtb.overrideVMType
	}
	if mtb.executorLogger != nil {
		vmBuilder.OverrideVMExecutor = executorwrapper.NewWrappedExecutorFactory(
			mtb.executorLogger,
			mtb.executorFactory)
	}

	executor := scenexec.NewScenarioExecutor(vmBuilder)
	defer executor.Close()

	executor.World.EnableEpochsHandler = mtb.enableEpochsHandler

	runner := scenio.NewScenarioController(
		executor,
		scenio.NewDefaultFileResolver(),
		vmBuilder.GetVMType(),
	)

	if len(mtb.singleFile) > 0 {
		fullPath := path.Join(getTestRoot(), mtb.folder)
		fullPath = path.Join(fullPath, mtb.singleFile)

		mtb.currentError = runner.RunSingleJSONScenario(
			fullPath,
			scenio.DefaultRunScenarioOptions())
	} else {
		mtb.currentError = runner.RunAllJSONScenariosInDirectory(
			getTestRoot(),
			mtb.folder,
			".scen.json",
			mtb.exclusions,
			scenio.DefaultRunScenarioOptions())
	}

	return mtb
}

// CheckNoError does an assert for the containing error
func (mtb *ScenariosTestBuilder) CheckNoError() *ScenariosTestBuilder {
	if mtb.currentError != nil {
		mtb.t.Error(mtb.currentError)
	}
	return mtb
}

// RequireError does an assert for the containing error
func (mtb *ScenariosTestBuilder) RequireError(expectedErrorMsg string) *ScenariosTestBuilder {
	require.EqualError(mtb.t, mtb.currentError, expectedErrorMsg)
	return mtb
}

// CheckLog will check the containing error
func (mtb *ScenariosTestBuilder) CheckLog(expectedLogs string) *ScenariosTestBuilder {
	stringLogger, ok := (mtb.executorLogger).(*executorwrapper.StringLogger)
	require.True(mtb.t, ok)
	require.NotNil(mtb.t, mtb.executorLogger)
	actualLog := stringLogger.String()
	if actualLog != expectedLogs {
		timestampStr := time.Now().Format("2006_01_02_15_04_05")
		fileExpected, err := os.Create(fmt.Sprintf("executorLog_%s_expected.txt", timestampStr))
		require.Nil(mtb.t, err)
		_, _ = fileExpected.WriteString(expectedLogs)
		err = fileExpected.Close()
		require.Nil(mtb.t, err)
		fileActual, err := os.Create(fmt.Sprintf("executorLog_%s_actual.txt", timestampStr))
		require.Nil(mtb.t, err)
		_, _ = fileActual.WriteString(actualLog)
		err = fileActual.Close()
		require.Nil(mtb.t, err)
		mtb.t.Error("log mismatch, see saved logs")
	}
	return mtb
}

// ExtractLog returns the string generated by the logger
func (mtb *ScenariosTestBuilder) ExtractLog() string {
	stringLogger, ok := (mtb.executorLogger).(*executorwrapper.StringLogger)
	require.True(mtb.t, ok, "executor logger must be StringLogger")
	return stringLogger.String()
}
