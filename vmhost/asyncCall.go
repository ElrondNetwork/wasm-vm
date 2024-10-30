package vmhost

import (
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-go/math"
)

// AsyncCall holds the information about an individual async call
type AsyncCall struct {
	CallID        []byte
	Status        AsyncCallStatus
	ExecutionMode AsyncCallExecutionMode

	Destination []byte
	Data        []byte
	GasLimit    uint64
	GasLocked   uint64

	ValueBytes      []byte
	SuccessCallback string
	ErrorCallback   string

	CallbackClosure []byte

	IsBuiltinFunctionCall bool
	IsAsyncV3             bool

	HasPendingCallback     bool
	PendingCallbackGasLocked uint64
}

// Clone creates a deep clone of the AsyncCall
func (ac *AsyncCall) Clone() *AsyncCall {
	clone := &AsyncCall{
		CallID:            ac.CallID,
		Status:            ac.Status,
		ExecutionMode:     ac.ExecutionMode,
		Destination:       make([]byte, len(ac.Destination)),
		Data:              make([]byte, len(ac.Data)),
		GasLimit:          ac.GasLimit,
		GasLocked:         ac.GasLocked,
		ValueBytes:        make([]byte, len(ac.ValueBytes)),
		SuccessCallback:   ac.SuccessCallback,
		ErrorCallback:     ac.ErrorCallback,
		IsAsyncV3:         ac.IsAsyncV3,
		HasPendingCallback: ac.HasPendingCallback,
		PendingCallbackGasLocked: ac.PendingCallbackGasLocked,
	}

	copy(clone.Destination, ac.Destination)
	copy(clone.Data, ac.Data)
	copy(clone.ValueBytes, ac.ValueBytes)

	return clone
}

// GetIdentifier returns the identifier of an async call
func (ac *AsyncCall) GetIdentifier() []byte {
	return ac.CallID
}

// GetDestination returns the destination of an async call
func (ac *AsyncCall) GetDestination() []byte {
	return ac.Destination
}

// GetData returns the transaction data of the async call
func (ac *AsyncCall) GetData() []byte {
	return ac.Data
}

// GetGasLimit returns the gas limit of the current async call
func (ac *AsyncCall) GetGasLimit() uint64 {
	return ac.GasLimit
}

// GetGasLocked returns the gas locked for the async callback
func (ac *AsyncCall) GetGasLocked() uint64 {
	return ac.GasLocked
}

// GetTotalGas returns the sum of the gas limit and gas locked
func (ac *AsyncCall) GetTotalGas() uint64 {
	return math.AddUint64(ac.GasLimit, ac.GasLocked)
}

// GetValue returns the byte representation of the value of the async call
func (ac *AsyncCall) GetValue() []byte {
	return ac.ValueBytes
}

// IsLocal returns true if the async call allows for local execution
func (ac *AsyncCall) IsLocal() bool {
	return !ac.IsRemote()
}

// IsRemote returns true if the async call must be sent remotely
func (ac *AsyncCall) IsRemote() bool {
	remote := (ac.ExecutionMode == AsyncUnknown) || (ac.ExecutionMode == AsyncBuiltinFuncCrossShard)
	return remote
}

// HasCallback returns true if there is a callback to execute, depending on the status of the async call
func (ac *AsyncCall) HasCallback() bool {
	callback := ac.GetCallbackName()

	return len(callback) != 0
}

// HasDefinedAnyCallback returns true if this AsyncCall defines at least one non-empty callback name
func (ac *AsyncCall) HasDefinedAnyCallback() bool {
	return len(ac.SuccessCallback) > 0 || len(ac.ErrorCallback) > 0
}

// MarkSkippedCallback this async call has skipped calling its callback
func (ac *AsyncCall) MarkSkippedCallback(pendingGasLock uint64) {
	ac.HasPendingCallback = true
	ac.PendingCallbackGasLocked = pendingGasLock
}

// UpdateStatus sets the status of the async call depending on the provided ReturnCode
func (ac *AsyncCall) UpdateStatus(returnCode vmcommon.ReturnCode) {
	ac.Status = AsyncCallResolved
	if returnCode != vmcommon.Ok {
		ac.Status = AsyncCallRejected
	}
}

// Reject sets the rejected status for this async call
func (ac *AsyncCall) Reject() {
	ac.Status = AsyncCallRejected
}

// GetCallbackName returns the name of the callback to execute, depending on
// the status of the async call
func (ac *AsyncCall) GetCallbackName() string {
	if ac.Status == AsyncCallResolved {
		return ac.SuccessCallback
	}

	return ac.ErrorCallback
}

// IsInterfaceNil returns true if there is no value under the interface
func (ac *AsyncCall) IsInterfaceNil() bool {
	return ac == nil
}

func (ac *AsyncCall) toSerializable() *SerializableAsyncCall {
	return &SerializableAsyncCall{
		CallID:            ac.CallID,
		Status:            SerializableAsyncCallStatus(ac.Status),
		ExecutionMode:     SerializableAsyncCallExecutionMode(ac.ExecutionMode),
		Destination:       ac.Destination,
		Data:              ac.Data,
		GasLimit:          ac.GasLimit,
		GasLocked:         ac.GasLocked,
		ValueBytes:        ac.ValueBytes,
		SuccessCallback:   ac.SuccessCallback,
		ErrorCallback:     ac.ErrorCallback,
		CallbackClosure:   ac.CallbackClosure,
		IsAsyncV3:         ac.IsAsyncV3,
		HasPendingCallback: ac.HasPendingCallback,
		PendingCallbackGasLocked: ac.PendingCallbackGasLocked,
	}
}

func fromSerializableAsyncCalls(serializableAsyncCalls []*SerializableAsyncCall) []*AsyncCall {
	var asyncCalls = make([]*AsyncCall, len(serializableAsyncCalls))
	for i, serAsyncCall := range serializableAsyncCalls {
		asyncCalls[i] = serAsyncCall.fromSerializable()
	}
	return asyncCalls
}

func (serAsyncCall *SerializableAsyncCall) fromSerializable() *AsyncCall {
	return &AsyncCall{
		CallID:            serAsyncCall.CallID,
		Status:            AsyncCallStatus(serAsyncCall.Status),
		ExecutionMode:     AsyncCallExecutionMode(serAsyncCall.ExecutionMode),
		Destination:       serAsyncCall.Destination,
		Data:              serAsyncCall.Data,
		GasLimit:          serAsyncCall.GasLimit,
		GasLocked:         serAsyncCall.GasLocked,
		ValueBytes:        serAsyncCall.ValueBytes,
		SuccessCallback:   serAsyncCall.SuccessCallback,
		ErrorCallback:     serAsyncCall.ErrorCallback,
		CallbackClosure:   serAsyncCall.CallbackClosure,
		IsAsyncV3:         serAsyncCall.IsAsyncV3,
		HasPendingCallback: serAsyncCall.HasPendingCallback,
		PendingCallbackGasLocked: serAsyncCall.PendingCallbackGasLocked,
	}
}
