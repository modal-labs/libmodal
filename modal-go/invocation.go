package modal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/protobuf/proto"
)

// controlPlaneInvocation implements the invocation interface.
type controlPlaneInvocation struct {
	FunctionCallId  string
	input           *pb.FunctionInput
	functionCallJwt string
	inputJwt        string
	ctx             context.Context
}

// createControlPlaneInvocation executes a function call and returns a new controlPlaneInvocation.
func createControlPlaneInvocation(ctx context.Context, functionId string, input *pb.FunctionInput, invocationType pb.FunctionCallInvocationType) (*controlPlaneInvocation, error) {
	functionPutInputsItem := pb.FunctionPutInputsItem_builder{
		Idx:   0,
		Input: input,
	}.Build()

	functionMapResponse, err := client.FunctionMap(ctx, pb.FunctionMapRequest_builder{
		FunctionId:                 functionId,
		FunctionCallType:           pb.FunctionCallType_FUNCTION_CALL_TYPE_UNARY,
		FunctionCallInvocationType: invocationType,
		PipelinedInputs:            []*pb.FunctionPutInputsItem{functionPutInputsItem},
	}.Build())
	if err != nil {
		return nil, err
	}

	return &controlPlaneInvocation{
		FunctionCallId:  functionMapResponse.GetFunctionCallId(),
		input:           input,
		functionCallJwt: functionMapResponse.GetFunctionCallJwt(),
		inputJwt:        functionMapResponse.GetPipelinedInputs()[0].GetInputJwt(),
		ctx:             ctx,
	}, nil
}

// controlPlaneInvocationFromFunctionCallId creates a controlPlaneInvocation from a function call ID.
func controlPlaneInvocationFromFunctionCallId(ctx context.Context, functionCallId string) *controlPlaneInvocation {
	return &controlPlaneInvocation{FunctionCallId: functionCallId, ctx: ctx}
}

func (c *controlPlaneInvocation) awaitOutput(timeout *time.Duration) (any, error) {
	return pollFunctionOutput(c.ctx, c.FunctionCallId, timeout)
}

func (c *controlPlaneInvocation) retry(retryCount uint32) error {
	if c.input == nil {
		return fmt.Errorf("cannot retry function invocation - input missing")
	}
	retryItem := pb.FunctionRetryInputsItem_builder{
		InputJwt:   c.inputJwt,
		Input:      c.input,
		RetryCount: retryCount,
	}.Build()
	functionRetryResponse, err := client.FunctionRetryInputs(c.ctx, pb.FunctionRetryInputsRequest_builder{
		FunctionCallJwt: c.functionCallJwt,
		Inputs:          []*pb.FunctionRetryInputsItem{retryItem},
	}.Build())
	if err != nil {
		return err
	}
	c.inputJwt = functionRetryResponse.GetInputJwts()[0]
	return nil
}

// Poll for outputs for a given FunctionCall ID.
func pollFunctionOutput(ctx context.Context, functionCallId string, timeout *time.Duration) (any, error) {
	startTime := time.Now()
	pollTimeout := outputsTimeout
	if timeout != nil {
		// Refresh backend call once per outputsTimeout.
		pollTimeout = min(*timeout, outputsTimeout)
	}

	for {
		response, err := client.FunctionGetOutputs(ctx, pb.FunctionGetOutputsRequest_builder{
			FunctionCallId: functionCallId,
			MaxValues:      1,
			Timeout:        float32(pollTimeout.Seconds()),
			LastEntryId:    "0-0",
			ClearOnSuccess: true,
			RequestedAt:    timeNowSeconds(),
		}.Build())
		if err != nil {
			return nil, fmt.Errorf("FunctionGetOutputs failed: %w", err)
		}

		// Output serialization may fail if any of the output items can't be deserialized
		// into a supported Go type. Users are expected to serialize outputs correctly.
		outputs := response.GetOutputs()
		if len(outputs) > 0 {
			return processResult(ctx, outputs[0].GetResult(), outputs[0].GetDataFormat())
		}

		if timeout != nil {
			remainingTime := *timeout - time.Since(startTime)
			if remainingTime <= 0 {
				message := fmt.Sprintf("Timeout exceeded: %.1fs", timeout.Seconds())
				return nil, FunctionTimeoutError{message}
			}
			pollTimeout = min(outputsTimeout, remainingTime)
		}
	}
}

// processResult processes the result from an invocation.
func processResult(ctx context.Context, result *pb.GenericResult, dataFormat pb.DataFormat) (any, error) {
	if result == nil {
		return nil, RemoteError{"Received null result from invocation"}
	}

	var data []byte
	var err error
	switch result.WhichDataOneof() {
	case pb.GenericResult_Data_case:
		data = result.GetData()
	case pb.GenericResult_DataBlobId_case:
		data, err = blobDownload(ctx, result.GetDataBlobId())
		if err != nil {
			return nil, err
		}
	case pb.GenericResult_DataOneof_not_set_case:
		data = nil
	}

	switch result.GetStatus() {
	case pb.GenericResult_GENERIC_STATUS_TIMEOUT:
		return nil, FunctionTimeoutError{result.GetException()}
	case pb.GenericResult_GENERIC_STATUS_INTERNAL_FAILURE:
		return nil, InternalFailure{result.GetException()}
	case pb.GenericResult_GENERIC_STATUS_SUCCESS:
		// Proceed to the block below this switch statement.
	default:
		// In this case, `result.GetData()` may have a pickled user code exception with traceback
		// from Python. We ignore this and only take the string representation.
		return nil, RemoteError{result.GetException()}
	}

	return deserializeDataFormat(data, dataFormat)
}

// blobDownload downloads a blob by its ID.
func blobDownload(ctx context.Context, blobId string) ([]byte, error) {
	resp, err := client.BlobGet(ctx, pb.BlobGetRequest_builder{
		BlobId: blobId,
	}.Build())
	if err != nil {
		return nil, err
	}
	s3resp, err := http.Get(resp.GetDownloadUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to download blob: %w", err)
	}
	defer s3resp.Body.Close()
	buf, err := io.ReadAll(s3resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob data: %w", err)
	}
	return buf, nil
}

func deserializeDataFormat(data []byte, dataFormat pb.DataFormat) (any, error) {
	switch dataFormat {
	case pb.DataFormat_DATA_FORMAT_PICKLE:
		return pickleDeserialize(data)
	case pb.DataFormat_DATA_FORMAT_ASGI:
		return nil, fmt.Errorf("ASGI data format is not supported in Go")
	case pb.DataFormat_DATA_FORMAT_GENERATOR_DONE:
		var done pb.GeneratorDone
		if err := proto.Unmarshal(data, &done); err != nil {
			return nil, fmt.Errorf("failed to unmarshal GeneratorDone: %w", err)
		}
		return &done, nil
	default:
		return nil, fmt.Errorf("unsupported data format: %s", dataFormat.String())
	}
}
