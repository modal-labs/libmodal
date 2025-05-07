package modal

import (
	"context"
	"fmt"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

// FunctionCall references a Modal Function Call. Function Calls are
// Function invocations with a given input. They can be consumed
// asynchronously (see Get()) or cancelled (see Cancel()).
type FunctionCall struct {
	FunctionCallId string
	ctx            context.Context
}

// Gets the ouptut for a FunctionCall
func (fc *FunctionCall) Get() (any, error) {
	return pollFunctionOutput(fc.ctx, &fc.FunctionCallId)
}

// Lookup a FunctionCall
func FunctionCallLookup(ctx context.Context, functionCallId string) (*FunctionCall, error) {
	ctx = clientContext(ctx)
	functionCall := FunctionCall{
		FunctionCallId: functionCallId,
		ctx:            ctx,
	}
	return &functionCall, nil
}

// Cancel a FunctionCall
func (fc *FunctionCall) Cancel(terminateContainers bool) error {
	_, err := client.FunctionCallCancel(fc.ctx, pb.FunctionCallCancelRequest_builder{
		FunctionCallId:      fc.FunctionCallId,
		TerminateContainers: terminateContainers,
	}.Build())
	if err != nil {
		return fmt.Errorf("FunctionCallCancel failed: %v", err)
	}

	return nil
}
