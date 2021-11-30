package mandostestcli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	am "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/arwenmandos"
	mc "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/mandos-go/controller"
)

func resolveArgument(exeDir string, arg string) (string, bool, error) {
	fi, err := os.Stat(arg)
	if os.IsNotExist(err) {
		arg = filepath.Join(exeDir, arg)
		fmt.Println(arg)
		fi, err = os.Stat(arg)
	}
	if err != nil {
		return "", false, err
	}
	return arg, fi.IsDir(), nil
}

func parseOptionFlags() *mc.RunScenarioOptions {
	forceTraceGas := flag.Bool("force-trace-gas", false, "overrides the traceGas option in the scenarios")
	flag.Parse()

	return &mc.RunScenarioOptions{
		ForceTraceGas: *forceTraceGas,
	}
}

// MandosTestCLI provides the functionality for any mandos-go test executor.
func MandosTestCLI() {
	options := parseOptionFlags()

	// directory of this executable
	exeDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// argument
	if len(os.Args) != 2 {
		panic("One argument expected - the path to the json test.")
	}
	jsonFilePath, isDir, err := resolveArgument(exeDir, os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// init
	executor, err := am.NewArwenTestExecutor()
	if err != nil {
		panic("Could not instantiate Arwen VM")
	}

	// execute
	switch {
	case isDir:
		runner := mc.NewScenarioRunner(
			executor,
			mc.NewDefaultFileResolver(),
		)
		err = runner.RunAllJSONScenariosInDirectory(
			jsonFilePath,
			"",
			".scen.json",
			[]string{},
			options)
	case strings.HasSuffix(jsonFilePath, ".scen.json"):
		runner := mc.NewScenarioRunner(
			executor,
			mc.NewDefaultFileResolver(),
		)
		err = runner.RunSingleJSONScenario(jsonFilePath, options)
	default:
		runner := mc.NewTestRunner(
			executor,
			mc.NewDefaultFileResolver(),
		)
		err = runner.RunSingleJSONTest(jsonFilePath)
	}

	// print result
	if err == nil {
		fmt.Println("SUCCESS")
	} else {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}
