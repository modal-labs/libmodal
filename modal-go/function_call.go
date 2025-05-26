package modal

import (
	"context"
	"fmt"
	"time"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

// FunctionCall references a Modal Function Call. Function Calls are
// Function invocations with a given input. They can be consumed
// asynchronously (see Get()) or cancelled (see Cancel()).
type FunctionCall struct {
	FunctionCallId string
	ctx            context.Context
}

// FunctionCallFromId looks up a FunctionCall.
func FunctionCallFromId(ctx context.Context, functionCallId string) (*FunctionCall, error) {
	ctx = clientContext(ctx)
	functionCall := FunctionCall{
		FunctionCallId: functionCallId,
		ctx:            ctx,
	}
	return &functionCall, nil
}

// FunctionCallGetOptions are options for getting outputs from Function Calls.
type FunctionCallGetOptions struct {
	Timeout time.Duration
}

// Get waits for the output of a FunctionCall.
// If timeout > 0, the operation will be cancelled after the specified duration.
func (fc *FunctionCall) Get(options FunctionCallGetOptions) (any, error) {
	ctx := fc.ctx

	// Use default if not specified.
	timeoutSeconds := options.Timeout
	if options.Timeout == 0 {
		timeoutSeconds = OutputsTimeout
	}
	return pollFunctionOutput(ctx, fc.FunctionCallId, timeoutSeconds)
}

// Helper function to find the minimum of two float32 values
func minTimeout(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// FunctionCallCancelOptions are options for cancelling Function Calls.
type FunctionCallCancelOptions struct {
	TerminateContainers bool
}

// Cancel cancels a FunctionCall.
func (fc *FunctionCall) Cancel(options FunctionCallCancelOptions) error {
	_, err := client.FunctionCallCancel(fc.ctx, pb.FunctionCallCancelRequest_builder{
		FunctionCallId:      fc.FunctionCallId,
		TerminateContainers: options.TerminateContainers,
	}.Build())
	if err != nil {
		return fmt.Errorf("FunctionCallCancel failed: %w", err)
	}

	return nil
}
