package vmhookstest

import (
	"math"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-scenario-go/worldmock"
	contextmock "github.com/multiversx/mx-chain-vm-go/mock/context"
	test "github.com/multiversx/mx-chain-vm-go/testcommon"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var repsArgument = []byte{0, 0, 0, byte(numberOfReps)}
var floatArgument1 = []byte{1, 10, 0, 0, 0, 53, 0, 0, 0, 108, 136, 217, 65, 19, 144, 71, 160, 0} // equal to 1.73476272346174595037472187482e+32
var floatArgument2 = []byte{1, 10, 0, 0, 0, 53, 0, 0, 0, 11, 190, 100, 79, 147, 188, 10, 8, 0}

func TestBigFloats_NewFromParts(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatNewFromPartsTest").
			WithArguments(repsArgument).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.Ok().
				ReturnData([]byte{byte(numberOfReps - 1)})
		})
}

func TestBigFloats_NewFromFrac(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatNewFromFracTest").
			WithArguments(repsArgument).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.Ok().
				ReturnData([]byte{byte(numberOfReps - 1)})
		})
}

func TestBigFloats_NewFromSci_Fail(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatNewFromSciTest").
			WithArguments(repsArgument,
				[]byte{0, 0, 1, 100}).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.ReturnCode(10)
		})
}

func TestBigFloats_NewFromSci_Success(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatNewFromSciTest").
			WithArguments(repsArgument,
				[]byte{255, 255, 255, 254}).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			value := float64(-199) * math.Pow10(-2)
			encodedValue, _ := big.NewFloat(value).GobEncode()
			verify.Ok().
				ReturnData(encodedValue)
		})
}

func TestBigFloats_Add(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatAddTest").
			WithArguments(repsArgument,
				floatArgument1, floatArgument1).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			bigFloatValue := new(big.Float).Quo(big.NewFloat(0), big.NewFloat(1))
			_ = bigFloatValue.GobDecode(floatArgument1)
			initialValue := big.NewFloat(0).Set(bigFloatValue)
			resultValue := new(big.Float)
			for i := 0; i < numberOfReps; i++ {
				resultValue.Add(initialValue, bigFloatValue)
				initialValue.Set(resultValue)
			}

			floatBuffer, _ := initialValue.GobEncode()
			verify.Ok().
				ReturnData(floatBuffer)
		})
}

func TestBigFloats_Panic_FailExecution_Add(t *testing.T) {
	floatArgument1 := []byte{1, 10, 0, 0, 0, 53, 0, 0, 0, 58, 0, 31, 28, 26, 150, 254, 14, 45}
	floatArgument2 := []byte{1, 11, 0, 0, 0, 53, 0, 0, 0, 52, 222, 212, 49, 108, 64, 122, 107, 100}

	t.Run("before ValidationOnGobDecodeFlag", func(t *testing.T) {
		test.BuildInstanceCallTest(t).
			WithContracts(
				test.CreateInstanceContract(test.ParentAddress).
					WithCode(test.GetTestSCCode("big-floats", "../../"))).
			WithInput(test.CreateTestContractCallInputBuilder().
				WithGasProvided(100000).
				WithFunction("BigFloatAddTest").
				WithArguments([]byte{0, 0, 0, byte(10)},
					floatArgument1, floatArgument2).
				Build()).
			WithEnableEpochsHandler(&worldmock.EnableEpochsHandlerStub{
				IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					return flag == vmhost.CryptoOpcodesV2Flag
				},
			}).
			AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
				verify.
					ReturnCode(10).
					ReturnMessage("this big Float operation is not permitted while doing float.Add")
			})
	})
	t.Run("after ValidationOnGobDecodeFlag", func(t *testing.T) {
		test.BuildInstanceCallTest(t).
			WithContracts(
				test.CreateInstanceContract(test.ParentAddress).
					WithCode(test.GetTestSCCode("big-floats", "../../"))).
			WithInput(test.CreateTestContractCallInputBuilder().
				WithGasProvided(100000).
				WithFunction("BigFloatAddTest").
				WithArguments([]byte{0, 0, 0, byte(10)},
					floatArgument1, floatArgument2).
				Build()).
			WithEnableEpochsHandler(worldmock.EnableEpochsHandlerStubAllFlags()).
			AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
				verify.
					ReturnCode(10).
					ReturnMessage("Float.GobDecode: msb not set in last word 0x1f1c1a96fe0e2d of 0x.1f1c1a96fe0e2dp+58")
			})
	})
}

func TestBigFloats_Sub(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatSubTest").
			WithArguments([]byte{0, 0, 0, byte(10)},
				floatArgument1, floatArgument1).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			bigFloatValue := new(big.Float).Quo(big.NewFloat(0), big.NewFloat(1))
			_ = bigFloatValue.GobDecode(floatArgument1)
			initialValue := big.NewFloat(0).Set(bigFloatValue)
			resultValue := new(big.Float)
			for i := 0; i < 10; i++ {
				resultValue.Sub(initialValue, bigFloatValue)
				initialValue.Set(resultValue)
			}

			floatBuffer, _ := initialValue.GobEncode()
			verify.Ok().
				ReturnData(floatBuffer)
		})
}

func TestBigFloats_Panic_FailExecution_Sub(t *testing.T) {
	floatArgument1 := []byte{1, 10, 0, 0, 0, 53, 0, 0, 0, 58, 0, 31, 28, 26, 150, 254, 14, 45}
	floatArgument2 := []byte{1, 10, 0, 0, 0, 53, 0, 0, 0, 52, 222, 212, 49, 108, 64, 122, 107, 100}

	t.Run("before ValidationOnGobDecodeFlag", func(t *testing.T) {
		test.BuildInstanceCallTest(t).
			WithContracts(
				test.CreateInstanceContract(test.ParentAddress).
					WithCode(test.GetTestSCCode("big-floats", "../../"))).
			WithInput(test.CreateTestContractCallInputBuilder().
				WithGasProvided(100000).
				WithFunction("BigFloatSubTest").
				WithArguments([]byte{0, 0, 0, byte(10)},
					floatArgument1, floatArgument2).
				Build()).
			WithEnableEpochsHandler(&worldmock.EnableEpochsHandlerStub{
				IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					return flag == vmhost.CryptoOpcodesV2Flag
				},
			}).
			AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
				verify.
					ReturnCode(10).
					ReturnMessage("this big Float operation is not permitted while doing float.Sub")
			})
	})
	t.Run("after ValidationOnGobDecodeFlag", func(t *testing.T) {
		test.BuildInstanceCallTest(t).
			WithContracts(
				test.CreateInstanceContract(test.ParentAddress).
					WithCode(test.GetTestSCCode("big-floats", "../../"))).
			WithInput(test.CreateTestContractCallInputBuilder().
				WithGasProvided(100000).
				WithFunction("BigFloatSubTest").
				WithArguments([]byte{0, 0, 0, byte(10)},
					floatArgument1, floatArgument2).
				Build()).
			WithEnableEpochsHandler(worldmock.EnableEpochsHandlerStubAllFlags()).
			AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
				verify.
					ReturnCode(10).
					ReturnMessage("Float.GobDecode: msb not set in last word 0x1f1c1a96fe0e2d of 0x.1f1c1a96fe0e2dp+58")
			})
	})
}

func TestBigFloats_Success_Mul(t *testing.T) {
	numberOfReps := 9
	repsArgument := []byte{0, 0, 0, byte(numberOfReps)}
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatMulTest").
			WithArguments(repsArgument,
				floatArgument1).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			bigFloatValue := new(big.Float)
			err := bigFloatValue.GobDecode(floatArgument1)
			require.Nil(t, err)
			for i := 0; i < numberOfReps; i++ {
				resultMul := new(big.Float).Mul(bigFloatValue, bigFloatValue)
				bigFloatValue.Set(resultMul)
			}
			floatBuffer, _ := bigFloatValue.GobEncode()
			verify.Ok().
				ReturnData(floatBuffer)
		})
}

func TestBigFloats_FailExponentTooBig_Mul(t *testing.T) {
	numberOfReps := 10
	repsArgument := []byte{0, 0, 0, byte(numberOfReps)}
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatMulTest").
			WithArguments(repsArgument,
				floatArgument1).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			bigFloatValue := new(big.Float)
			err := bigFloatValue.GobDecode(floatArgument1)
			require.Nil(t, err)
			for i := 0; i < numberOfReps; i++ {
				resultMul := new(big.Float).Mul(bigFloatValue, bigFloatValue)
				bigFloatValue.Set(resultMul)
			}
			verify.ReturnCode(10).
				ReturnMessage("exponent is either too small or too big")
		})
}

func TestBigFloats_FailExecution_Mul(t *testing.T) {
	numberOfReps := 30
	repsArgument := []byte{0, 0, 0, byte(numberOfReps)}
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatMulTest").
			WithArguments(repsArgument,
				floatArgument1).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.
				ReturnCode(10)
		})
}

func TestBigFloats_Div(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatDivTest").
			WithArguments(repsArgument,
				floatArgument1,
				floatArgument2).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			numerator := new(big.Float)
			_ = numerator.GobDecode(floatArgument1)
			denominator := new(big.Float)
			_ = denominator.GobDecode(floatArgument2)
			for i := 0; i < numberOfReps; i++ {
				resultMul := new(big.Float).Quo(numerator, denominator)
				numerator.Set(resultMul)
			}
			floatBuffer, _ := numerator.GobEncode()
			verify.Ok().
				ReturnData(floatBuffer)
		})
}

func TestBigFloats_Truncate(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatTruncateTest").
			WithArguments(repsArgument,
				floatArgument1,
				floatArgument2).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			value1 := big.NewFloat(0)
			_ = value1.GobDecode(floatArgument1)
			value2 := big.NewFloat(0)
			_ = value2.GobDecode(floatArgument2)
			for i := 0; i < numberOfReps; i++ {
				rDiv := big.NewInt(0)
				value1.Int(rDiv)
				result := big.NewFloat(0).Sub(value1, value2)
				value1.Set(result)
			}
			floatBuffer, _ := value1.GobEncode()
			verify.Ok().
				ReturnData(floatBuffer)
		})
}

func TestBigFloats_Abs(t *testing.T) {
	bigFloatArguments := make([][]byte, numberOfReps+1)
	bigFloatArguments[0] = []byte{0, 0, 0, 1}
	encodedFloat, _ := big.NewFloat(-1623).GobEncode()
	bigFloatArguments[1] = encodedFloat

	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatAbsTest").
			WithArguments(bigFloatArguments...).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			encodedAbsFloat, _ := big.NewFloat(1623).GobEncode()
			verify.Ok().
				ReturnData(encodedAbsFloat)
		})
}

func TestBigFloats_Neg(t *testing.T) {
	bigFloatArguments := make([][]byte, numberOfReps+1)
	bigFloatArguments[0] = []byte{0, 0, 0, 1}
	encodedFloat, _ := big.NewFloat(-1623).GobEncode()
	bigFloatArguments[1] = encodedFloat

	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatNegTest").
			WithArguments(bigFloatArguments...).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			encodedFloatValue, _ := big.NewFloat(-1623).GobEncode()
			floatValue := new(big.Float)
			err := floatValue.GobDecode(encodedFloatValue)
			assert.Nil(t, err)
			floatValue.Neg(floatValue)
			encodedNegFloat, _ := floatValue.GobEncode()
			verify.Ok().
				ReturnData(encodedNegFloat)
		})
}

func TestBigFloats_Cmp(t *testing.T) {
	bigFloatArguments := make([][]byte, 2*numberOfReps+1)
	bigFloatArguments[0] = repsArgument
	argsCounter := 1
	for i := 0; i < numberOfReps; i++ {
		floatValue := big.NewFloat(-float64(i) - 1)
		encodedFloat, _ := floatValue.GobEncode()
		bigFloatArguments[argsCounter] = encodedFloat
		absFloatValue := new(big.Float).Neg(floatValue)
		encodedAbsFloat, _ := absFloatValue.GobEncode()
		bigFloatArguments[argsCounter+1] = encodedAbsFloat
		argsCounter += 2
	}

	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatCmpTest").
			WithArguments(bigFloatArguments...).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			encodedLastFloat := bigFloatArguments[numberOfReps*2]
			lastFloat := new(big.Float)
			_ = lastFloat.GobDecode(encodedLastFloat)
			encodedPreviousLastFloat := bigFloatArguments[numberOfReps*2-1]
			previousLastFloat := new(big.Float)
			_ = previousLastFloat.GobDecode(encodedPreviousLastFloat)
			cmpResult := previousLastFloat.Cmp(lastFloat)
			verify.Ok().
				ReturnData([]byte{byte(cmpResult)})
		})
}

func TestBigFloats_Sign(t *testing.T) {
	bigFloatArguments := make([][]byte, numberOfReps+1)
	bigFloatArguments[0] = []byte{0, 0, 0, 1}
	encodedFloat, _ := big.NewFloat(-1623).GobEncode()
	bigFloatArguments[1] = encodedFloat

	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatSignTest").
			WithArguments(bigFloatArguments...).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			negativeSignFloat := big.NewFloat(-1)
			verify.Ok().
				ReturnData([]byte{byte(negativeSignFloat.Sign())})
		})
}

func TestBigFloats_Clone(t *testing.T) {
	bigFloatArguments := make([][]byte, numberOfReps+1)
	bigFloatArguments[0] = repsArgument
	for i := 0; i < numberOfReps; i++ {
		floatValue := big.NewFloat(-float64(i) - 1)
		encodedFloat, _ := floatValue.GobEncode()
		bigFloatArguments[i+1] = encodedFloat
	}

	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatCloneTest").
			WithArguments(bigFloatArguments...).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			encodedLastFloat := bigFloatArguments[numberOfReps]
			lastFloat := new(big.Float)
			_ = lastFloat.GobDecode(encodedLastFloat)
			verify.Ok().
				ReturnData(encodedLastFloat)
		})
}

func TestBigFloats_Sqrt(t *testing.T) {
	bigFloatArguments := make([][]byte, numberOfReps+1)
	bigFloatArguments[0] = repsArgument
	for i := 0; i < numberOfReps; i++ {
		floatValue := big.NewFloat(float64(i) + 1)
		encodedFloat, _ := floatValue.GobEncode()
		bigFloatArguments[i+1] = encodedFloat
	}

	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatSqrtTest").
			WithArguments(bigFloatArguments...).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			encodedLastFloat := bigFloatArguments[numberOfReps]
			lastFloat := new(big.Float)
			_ = lastFloat.GobDecode(encodedLastFloat)
			sqrtFloat := new(big.Float).Sqrt(lastFloat)
			encodedSqrtFloat, _ := sqrtFloat.GobEncode()
			verify.Ok().
				ReturnData(encodedSqrtFloat)
		})
}

func TestBigFloats_Pow(t *testing.T) {
	bigFloatArguments := make([][]byte, numberOfReps+1)
	bigFloatArguments[0] = []byte{0, 0, 0, byte(3)}
	for i := 0; i < numberOfReps; i++ {
		floatValue := big.NewFloat(1.6)
		encodedFloat, _ := floatValue.GobEncode()
		bigFloatArguments[i+1] = encodedFloat
	}

	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatPowTest").
			WithArguments(bigFloatArguments...).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			resultFloat := big.NewFloat(1.6)
			intermediaryFloat := new(big.Float).Mul(resultFloat, resultFloat)
			resultFloat.Set(intermediaryFloat)
			encodedResult, _ := resultFloat.GobEncode()
			verify.Ok().
				ReturnData(encodedResult)
		})
}

func TestBigFloats_Floor(t *testing.T) {
	bigFloatArguments := make([][]byte, numberOfReps+1)
	bigFloatArguments[0] = repsArgument
	for i := 0; i < numberOfReps; i++ {
		floatValue := big.NewFloat((float64(i) + 2) / (float64(i) + 1))
		encodedFloat, _ := floatValue.GobEncode()
		bigFloatArguments[i+1] = encodedFloat
	}

	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatFloorTest").
			WithArguments(bigFloatArguments...).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			encodedLastFloat := bigFloatArguments[numberOfReps]
			lastFloat := new(big.Float)
			_ = lastFloat.GobDecode(encodedLastFloat)
			bigIntOp := new(big.Int)
			lastFloat.Int(bigIntOp)
			verify.Ok().
				ReturnData(bigIntOp.Bytes())
		})
}

func TestBigFloats_Ceil(t *testing.T) {
	bigFloatArguments := make([][]byte, numberOfReps+1)
	bigFloatArguments[0] = repsArgument
	for i := 0; i < numberOfReps; i++ {
		floatValue := big.NewFloat((float64(i) + 2) / (float64(i) + 1))
		encodedFloat, _ := floatValue.GobEncode()
		bigFloatArguments[i+1] = encodedFloat
	}

	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatCeilTest").
			WithArguments(bigFloatArguments...).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			encodedLastFloat := bigFloatArguments[numberOfReps]
			lastFloat := new(big.Float)
			_ = lastFloat.GobDecode(encodedLastFloat)
			bigIntOp := new(big.Int)
			lastFloat.Int(bigIntOp)
			bigIntOp.Add(bigIntOp, big.NewInt(1))
			verify.Ok().
				ReturnData(bigIntOp.Bytes())
		})
}

func TestBigFloats_IsInt(t *testing.T) {
	bigFloatArguments := make([][]byte, numberOfReps+1)
	bigFloatArguments[0] = repsArgument
	for i := 0; i < numberOfReps; i++ {
		floatValue := big.NewFloat(float64(i) + 2)
		encodedFloat, _ := floatValue.GobEncode()
		bigFloatArguments[i+1] = encodedFloat
	}

	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatIsIntTest").
			WithArguments(bigFloatArguments...).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			encodedLastFloat := bigFloatArguments[numberOfReps]
			lastFloat := new(big.Float)
			_ = lastFloat.GobDecode(encodedLastFloat)
			var isInt byte
			if lastFloat.IsInt() {
				isInt = 1
			} else {
				isInt = 0
			}
			verify.Ok().
				ReturnData([]byte{isInt})
		})
}

func TestBigFloats_SetInt64(t *testing.T) {
	bigFloatArguments := make([][]byte, numberOfReps+1)
	bigFloatArguments[0] = repsArgument
	for i := 0; i < numberOfReps; i++ {
		bigFloatArguments[i+1] = []byte{0, 0, 0, byte(i)}
	}

	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatSetInt64Test").
			WithArguments(bigFloatArguments...).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			floatValue := big.NewFloat(0)
			floatValue.SetInt64(int64(numberOfReps - 1))
			encodedFloatValue, _ := floatValue.GobEncode()
			verify.Ok().
				ReturnData(encodedFloatValue)
		})
}

func TestBigFloats_SetBigInt(t *testing.T) {
	bigFloatArguments := make([][]byte, numberOfReps+1)
	bigFloatArguments[0] = repsArgument
	for i := 0; i < numberOfReps; i++ {
		bigIntValue := big.NewInt(int64(i))
		bigFloatArguments[i+1] = bigIntValue.Bytes()
	}

	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatSetBigIntTest").
			WithArguments(bigFloatArguments...).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			floatValue := big.NewFloat(0)
			floatValue.SetInt(big.NewInt(int64(numberOfReps) - 1))
			encodedFloatValue, _ := floatValue.GobEncode()
			verify.Ok().
				ReturnData(encodedFloatValue)
		})
}

func TestBigFloats_GetConstPi(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatGetConstPiTest").
			WithArguments([]byte{0, 0, 0, byte(numberOfReps)}).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			piValue := math.Pi
			bigFloatValue := big.NewFloat(0).SetFloat64(piValue)
			encodedFloat, _ := bigFloatValue.GobEncode()
			verify.Ok().
				ReturnData(encodedFloat)
		})
}

func TestBigFloats_GetConstE(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("big-floats", "../../"))).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithGasProvided(100000).
			WithFunction("BigFloatGetConstETest").
			WithArguments([]byte{0, 0, 0, byte(numberOfReps)}).
			Build()).
		AndAssertResults(func(host vmhost.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			piValue := math.E
			bigFloatValue := big.NewFloat(0).SetFloat64(piValue)
			encodedFloat, _ := bigFloatValue.GobEncode()
			verify.Ok().
				ReturnData(encodedFloat)
		})
}
