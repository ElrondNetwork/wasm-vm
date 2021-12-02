package contexts

import (
	"github.com/ElrondNetwork/elrond-go-core/data/vm"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

func (context *asyncContext) NotifyChildIsComplete(callID []byte, gasToAccumulate uint64) error {
	if logAsync.GetLevel() == logger.LogTrace {
		logAsync.Trace("NofityChildIsComplete")
		logAsync.Trace("", "address", string(context.address))
		logAsync.Trace("", "callID", context.callID) // DebugCallIDAsString
		logAsync.Trace("", "callerAddr", string(context.callerAddr))
		logAsync.Trace("", "callerCallID", context.callerCallID)
		logAsync.Trace("", "notifier callID", callID)
		logAsync.Trace("", "gasToAccumulate", gasToAccumulate)
	}

	err := context.completeChild(callID, gasToAccumulate)
	if err != nil {
		return err
	}

	if !context.IsComplete() {
		return context.Save()
	}

	return context.complete()
}

func (context *asyncContext) completeChild(callID []byte, gasToAccumulate uint64) error {
	return context.CompleteChildConditional(true, callID, gasToAccumulate)
}

func (context *asyncContext) CompleteChildConditional(isChildComplete bool, callID []byte, gasToAccumulate uint64) error {
	if !isChildComplete {
		return nil
	}
	context.decrementCallsCounter()
	context.accumulateGas(gasToAccumulate)
	if callID != nil {
		err := context.DeleteAsyncCallAndCleanGroup(callID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (context *asyncContext) complete() error {
	// There are no more callbacks to return from other shards. The context can
	// be deleted from storage.
	err := context.DeleteFromAddress(context.address)
	if err != nil {
		return err
	}

	// if we reached first call, stop notification chain
	if context.IsFirstCall() {
		return nil
	}

	currentCallID := context.GetCallID()
	if context.callType == vm.AsynchronousCall {
		vmOutput := context.childResults
		isCallbackComplete, _, err := context.callCallback(currentCallID, vmOutput, nil)
		if err != nil {
			return err
		}
		if isCallbackComplete {
			return context.NotifyChildIsComplete(currentCallID, 0)
		}
	} else if context.callType == vm.AsynchronousCallBack {
		err = context.LoadParentContext()
		if err != nil {
			return err
		}

		currentCallID := context.GetCallerCallID()
		return context.NotifyChildIsComplete(currentCallID, context.gasAccumulated)
	} else if context.callType == vm.DirectCall {
		err = context.LoadParentContext()
		if err != nil {
			return err
		}

		return context.NotifyChildIsComplete(nil, context.gasAccumulated)
	}

	return nil
}
