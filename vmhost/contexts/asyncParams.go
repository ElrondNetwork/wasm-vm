//nolint:all
package contexts

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/txDataBuilder"
	"github.com/multiversx/mx-chain-vm-go/crypto"
)

// AddAsyncArgumentsToOutputTransfers
// Called to process OutputTransfers created by a
// direct call (on dest) builtin function call by the VM
func AddAsyncArgumentsToOutputTransfers(
	asyncParams *vmcommon.AsyncArguments,
	callType vm.CallType,
	vmOutput *vmcommon.VMOutput,
) error {
	if asyncParams == nil {
		return nil
	}

	for _, outAcc := range vmOutput.OutputAccounts {

		for t, outTransfer := range outAcc.OutputTransfers {
			if outTransfer.CallType != callType {
				continue
			}

			asyncData, err := createDataFromAsyncParams(
				asyncParams,
				callType)

			if err != nil {
				return err
			}

			outAcc.OutputTransfers[t] = vmcommon.OutputTransfer{
				Index:         outTransfer.Index,
				Value:         outTransfer.Value,
				GasLimit:      outTransfer.GasLimit,
				GasLocked:     outTransfer.GasLocked,
				AsyncData:     asyncData,
				Data:          outTransfer.Data,
				CallType:      outTransfer.CallType,
				SenderAddress: outTransfer.SenderAddress,
			}
		}
	}

	return nil
}

func createDataFromAsyncParams(
	asyncParams *vmcommon.AsyncArguments,
	callType vm.CallType,
) ([]byte, error) {
	if asyncParams == nil {
		if callType == vm.AsynchronousCall || callType == vm.AsynchronousCallBack {
			return nil, vmcommon.ErrAsyncParams
		} else {
			return nil, nil
		}
	}

	callData := txDataBuilder.NewBuilder()
	callData.Bytes(asyncParams.CallID)
	callData.Bytes(asyncParams.CallerCallID)
	if callType == vm.AsynchronousCallBack {
		callData.Bytes(asyncParams.CallbackAsyncInitiatorCallID)
		callData.Bytes(big.NewInt(int64(asyncParams.GasAccumulated)).Bytes())
	}

	return callData.ToBytes(), nil
}

// GenerateNewCallID will generate a new call ID as byte slice
func GenerateNewCallID(hasher crypto.Hasher, parentCallID []byte, suffix []byte) []byte {
	newCallID := append(parentCallID, suffix...)
	newCallID, err := hasher.Sha256(newCallID)
	if err != nil {
		return []byte{}
	}
	return newCallID
}
