package vmhooks

import (
	"github.com/multiversx/mx-chain-vm-go/crypto"
	"github.com/multiversx/mx-chain-vm-go/executor"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
)

//go:generate go run generate/cmd/eiGenMain.go

// VMHooksImpl is the VM structure that implements VMHooks,
// with all the hooks (callbacks) from the executor.
type VMHooksImpl struct {
	host vmhost.VMHost
}

// NewVMHooksImpl creates a new VMHooksImpl instance.
func NewVMHooksImpl(host vmhost.VMHost) *VMHooksImpl {
	return &VMHooksImpl{
		host: host,
	}
}

// MemLoad returns the contents from the given offset of the WASM memory.
func (context *VMHooksImpl) MemLoad(memPtr executor.MemPtr, length executor.MemLength) ([]byte, error) {
	return context.host.Runtime().GetInstance().MemLoad(memPtr, length)
}

// MemLoadMultiple returns multiple byte slices loaded from the WASM memory, starting at the given offset and having the provided lengths.
func (context *VMHooksImpl) MemLoadMultiple(memPtr executor.MemPtr, lengths []int32) ([][]byte, error) {
	if len(lengths) == 0 {
		return [][]byte{}, nil
	}

	results := make([][]byte, len(lengths))

	for i, length := range lengths {
		result, err := context.MemLoad(memPtr, length)
		if err != nil {
			return nil, err
		}

		results[i] = result
		memPtr = memPtr.Offset(length)
	}

	return results, nil
}

// MemStore stores the given data in the WASM memory at the given offset.
func (context *VMHooksImpl) MemStore(memPtr executor.MemPtr, data []byte) error {
	return context.host.Runtime().GetInstance().MemStore(memPtr, data)
}

// GetVMHost returns the vm Context from the vm context map
func (context *VMHooksImpl) GetVMHost() vmhost.VMHost {
	return context.host
}

// GetBlockchainContext returns the blockchain context
func (context *VMHooksImpl) GetBlockchainContext() vmhost.BlockchainContext {
	return context.host.Blockchain()
}

// GetRuntimeContext returns the runtime context
func (context *VMHooksImpl) GetRuntimeContext() vmhost.RuntimeContext {
	return context.host.Runtime()
}

// GetCryptoContext returns the crypto context
func (context *VMHooksImpl) GetCryptoContext() crypto.VMCrypto {
	return context.host.Crypto()
}

// GetManagedTypesContext returns the big int context
func (context *VMHooksImpl) GetManagedTypesContext() vmhost.ManagedTypesContext {
	return context.host.ManagedTypes()
}

// GetOutputContext returns the output context
func (context *VMHooksImpl) GetOutputContext() vmhost.OutputContext {
	return context.host.Output()
}

// GetMeteringContext returns the metering context
func (context *VMHooksImpl) GetMeteringContext() vmhost.MeteringContext {
	return context.host.Metering()
}

// GetStorageContext returns the storage context
func (context *VMHooksImpl) GetStorageContext() vmhost.StorageContext {
	return context.host.Storage()
}

// GetEnableEpochsHandler returns the enable epochs handler
func (context *VMHooksImpl) GetEnableEpochsHandler() vmhost.EnableEpochsHandler {
	return context.host.EnableEpochsHandler()
}

// WithFault handles an error, taking into account whether it should completely
// fail the execution of a contract or not.
func (context *VMHooksImpl) WithFault(err error, failExecution bool) bool {
	return WithFaultAndHost(context.host, err, failExecution)
}

// WithFaultAndHost fails the execution with the provided error
func WithFaultAndHost(host vmhost.VMHost, err error, failExecution bool) bool {
	if err == nil {
		return false
	}

	if failExecution {
		runtime := host.Runtime()
		metering := host.Metering()
		metering.UseGas(metering.GasLeft())
		runtime.FailExecution(err)
	}

	return true
}
