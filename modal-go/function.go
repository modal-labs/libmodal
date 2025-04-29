package modal

import (
	"bytes"
	"context"
	"fmt"
	"time"

	pickle "github.com/kisielk/og-rek"
	proto "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

func timeNow() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}

// function references a deployed Modal Function.
type Function struct {
	FunctionId string
	ctx        context.Context
}

// FunctionLookup looks up an existing Function.
func FunctionLookup(ctx context.Context, appName string, name string, options LookupOptions) (*Function, error) {
	ctx = clientContext(ctx)

	resp, err := client.FunctionGet(ctx, proto.FunctionGetRequest_builder{
		AppName:         appName,
		ObjectTag:       name,
		Namespace:       proto.DeploymentNamespace_DEPLOYMENT_NAMESPACE_WORKSPACE,
		EnvironmentName: environmentName(options.Environment),
	}.Build())

	if err != nil {
		return nil, err
	}

	return &Function{FunctionId: resp.GetFunctionId(), ctx: ctx}, nil
}

// Serialize function inputs to the Python pickle format.
func (function *Function) serialize(args []interface{}, kwargs map[string]interface{}) (bytes.Buffer, error) {
	var inputBuffer bytes.Buffer

	e := pickle.NewEncoder(&inputBuffer)
	err := e.Encode(pickle.Tuple{
		args,
		kwargs,
	})

	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("error pickling data: %w", err)
	}
	return inputBuffer, nil
}

// Deserialize from Python pickle into Go basic types.
func (function *Function) deserialize(buffer []byte) (interface{}, error) {
	decoder := pickle.NewDecoder(bytes.NewReader(buffer))
	result, err := decoder.Decode()
	if err != nil {
		return nil, fmt.Errorf("error unpickling data: %w", err)
	}

	return result, nil
}

// Execute a single input into a remote Function.
func (function *Function) Remote(ctx context.Context, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	ctx = clientContext(ctx)

	payload, _ := function.serialize(args, kwargs)

	// Single input sync invocation
	var functionInputs []*proto.FunctionPutInputsItem
	functionInputItem := proto.FunctionPutInputsItem_builder{
		Idx: 0,
		Input: proto.FunctionInput_builder{
			Args:       payload.Bytes(),
			FinalInput: true,
		}.Build(),
	}.Build()
	functionInputs = append(functionInputs, functionInputItem)

	functionMapResponse, err := client.FunctionMap(ctx, proto.FunctionMapRequest_builder{
		FunctionId:                 function.FunctionId,
		FunctionCallType:           proto.FunctionCallType_FUNCTION_CALL_TYPE_UNARY,
		FunctionCallInvocationType: proto.FunctionCallInvocationType_FUNCTION_CALL_INVOCATION_TYPE_SYNC,
		PipelinedInputs:            functionInputs,
	}.Build())
	if err != nil {
		return nil, fmt.Errorf("FunctionMap errro: %v", err)
	}

	response, err := client.FunctionGetOutputs(ctx, proto.FunctionGetOutputsRequest_builder{
		FunctionCallId: functionMapResponse.GetFunctionCallId(),
		MaxValues:      1,
		Timeout:        55,
		LastEntryId:    "0-0",
		ClearOnSuccess: true,
		RequestedAt:    timeNow(),
	}.Build())
	if err != nil {
		return nil, fmt.Errorf("FunctionGetOutputs failed: %v", err)
	}

	// Output serialization may fail if any of the output items can't be deserialized
	// into a supported Go type. Users are expected to serialize outputs correctly.
	outputs := response.GetOutputs()
	if len(outputs) > 0 {
		data := outputs[0].GetResult().GetData()
		resultPayload, err := function.deserialize(data)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize output: %v", err)
		}

		return resultPayload, nil
	}
	return nil, nil
}
