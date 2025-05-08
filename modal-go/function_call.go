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

// GetOptions are options for getting outputs from Function Calls.
type GetOptions struct {
	Timeout time.Duration
}

// Get waits for the output of a FunctionCall.
// If timeout > 0, the operation will be cancelled after the specified duration.
func (fc *FunctionCall) Get(options GetOptions) (any, error) {
	ctx := fc.ctx
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(fc.ctx, options.Timeout)
		defer cancel()
	}

	return pollFunctionOutput(ctx, fc.FunctionCallId)
}

// CancelOptions are options for cancelling Function Calls.
type CancelOptions struct {
	TerminateContainers bool
}

// Cancel cancels a FunctionCall.
func (fc *FunctionCall) Cancel(options CancelOptions) error {
	_, err := client.FunctionCallCancel(fc.ctx, pb.FunctionCallCancelRequest_builder{
		FunctionCallId:      fc.FunctionCallId,
		TerminateContainers: options.TerminateContainers,
	}.Build())
	if err != nil {
		return fmt.Errorf("FunctionCallCancel failed: %w", err)
	}

	return nil
}
