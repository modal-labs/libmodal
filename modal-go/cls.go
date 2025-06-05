package modal

// Cls lookups and Function binding.

import (
	"context"
	"fmt"
	"sort"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Cls represents a Modal class definition that can be instantiated with parameters.
// It contains metadata about the class and its methods.
type Cls struct {
	ctx               context.Context
	serviceFunctionId string
	schema            []*pb.ClassParameterSpec
	methodNames       []string
}

// ClsLookup looks up an existing Cls on a deployed App.
func ClsLookup(ctx context.Context, appName string, name string, options *LookupOptions) (*Cls, error) {
	if options == nil {
		options = &LookupOptions{}
	}
	ctx = clientContext(ctx)
	cls := Cls{
		methodNames: []string{},
		ctx:         ctx,
	}

	// Find class service function metadata. Service functions are used to implement class methods,
	// which are invoked using a combination of service function ID and the method name.
	serviceFunctionName := fmt.Sprintf("%s.*", name)
	serviceFunction, err := client.FunctionGet(ctx, pb.FunctionGetRequest_builder{
		AppName:         appName,
		ObjectTag:       serviceFunctionName,
		Namespace:       pb.DeploymentNamespace_DEPLOYMENT_NAMESPACE_WORKSPACE,
		EnvironmentName: environmentName(options.Environment),
	}.Build())

	if status, ok := status.FromError(err); ok && status.Code() == codes.NotFound {
		return nil, NotFoundError{fmt.Sprintf("class '%s/%s' not found", appName, name)}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to look up class service function: %w", err)
	}

	// Validate that we only support parameter serialization format PROTO.
	parameterInfo := serviceFunction.GetHandleMetadata().GetClassParameterInfo()
	schema := parameterInfo.GetSchema()
	if len(schema) > 0 && parameterInfo.GetFormat() != pb.ClassParameterInfo_PARAM_SERIALIZATION_FORMAT_PROTO {
		return nil, fmt.Errorf("unsupported parameter format: %v", parameterInfo.GetFormat())
	} else {
		cls.schema = schema
	}

	cls.serviceFunctionId = serviceFunction.GetFunctionId()

	// Check if we have method metadata on the class service function (v0.67+)
	serviceFunctionHandleMetadata := serviceFunction.GetHandleMetadata()
	if serviceFunctionHandleMetadata != nil && serviceFunctionHandleMetadata.GetMethodHandleMetadata() != nil {
		for methodName := range serviceFunctionHandleMetadata.GetMethodHandleMetadata() {
			cls.methodNames = append(cls.methodNames, methodName)
		}
	} else {
		// Legacy approach not supported
		return nil, fmt.Errorf("Cls requires Modal deployments using client v0.67 or later")
	}

	return &cls, nil
}

// Instance creates a new instance of the class with the provided parameters.
func (c *Cls) Instance(params map[string]any) (*ClsInstance, error) {
	var functionId string
	if len(c.schema) == 0 {
		// Class isn't parametrized, return a simple instance.
		functionId = c.serviceFunctionId
	} else {
		// Class has parameters, bind the parameters to service function
		// and update method references.
		boundFunctionId, err := c.bindParameters(params)
		if err != nil {
			return nil, err
		}
		functionId = boundFunctionId
	}

	methods := make(map[string]*Function)
	for _, name := range c.methodNames {
		methods[name] = &Function{
			FunctionId: functionId,
			MethodName: &name,
			ctx:        c.ctx,
		}
	}
	return &ClsInstance{methods: methods}, nil
}

// bindParameters processes the parameters and binds them to the class function.
func (c *Cls) bindParameters(params map[string]any) (string, error) {
	serializedParams, err := encodeParameterSet(c.schema, params)
	if err != nil {
		return "", fmt.Errorf("failed to serialize parameters: %w", err)
	}

	// Bind parameters to create a parameterized function
	bindResp, err := client.FunctionBindParams(c.ctx, pb.FunctionBindParamsRequest_builder{
		FunctionId:       c.serviceFunctionId,
		SerializedParams: serializedParams,
	}.Build())
	if err != nil {
		return "", fmt.Errorf("failed to bind parameters: %w", err)
	}

	return bindResp.GetBoundFunctionId(), nil
}

// encodeParameterSet encodes the parameter values into a binary format.
func encodeParameterSet(schema []*pb.ClassParameterSpec, params map[string]any) ([]byte, error) {
	var encoded []*pb.ClassParameterValue

	for _, paramSpec := range schema {
		paramValue, err := encodeParameter(paramSpec, params[paramSpec.GetName()])
		if err != nil {
			return nil, err
		}
		encoded = append(encoded, paramValue)
	}

	// Sort keys, identical to Python `SerializeToString(deterministic=True)`.
	sort.Slice(encoded, func(i, j int) bool {
		return encoded[i].GetName() < encoded[j].GetName()
	})
	return proto.Marshal(pb.ClassParameterSet_builder{Parameters: encoded}.Build())
}

// encodeParameter converts a Go value to a ParameterValue proto message
func encodeParameter(paramSpec *pb.ClassParameterSpec, value any) (*pb.ClassParameterValue, error) {
	name := paramSpec.GetName()
	paramType := paramSpec.GetType()
	paramValue := pb.ClassParameterValue_builder{
		Name: name,
		Type: paramType,
	}.Build()

	switch paramType {
	case pb.ParameterType_PARAM_TYPE_STRING:
		if value == nil && paramSpec.GetHasDefault() {
			value = paramSpec.GetStringDefault()
		}
		strValue, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("parameter '%s' must be a string", name)
		}
		paramValue.SetStringValue(strValue)

	case pb.ParameterType_PARAM_TYPE_INT:
		if value == nil && paramSpec.GetHasDefault() {
			value = paramSpec.GetIntDefault()
		}
		var intValue int64
		switch v := value.(type) {
		case int:
			intValue = int64(v)
		case int64:
			intValue = v
		case int32:
			intValue = int64(v)
		default:
			return nil, fmt.Errorf("parameter '%s' must be an integer", name)
		}
		paramValue.SetIntValue(intValue)

	case pb.ParameterType_PARAM_TYPE_BOOL:
		if value == nil && paramSpec.GetHasDefault() {
			value = paramSpec.GetBoolDefault()
		}
		boolValue, ok := value.(bool)
		if !ok {
			return nil, fmt.Errorf("parameter '%s' must be a boolean", name)
		}
		paramValue.SetBoolValue(boolValue)

	case pb.ParameterType_PARAM_TYPE_BYTES:
		if value == nil && paramSpec.GetHasDefault() {
			value = paramSpec.GetBytesDefault()
		}
		bytesValue, ok := value.([]byte)
		if !ok {
			return nil, fmt.Errorf("parameter '%s' must be a byte slice", name)
		}
		paramValue.SetBytesValue(bytesValue)

	default:
		return nil, fmt.Errorf("unsupported parameter type: %v", paramType)
	}

	return paramValue, nil
}

// ClsInstance represents an instantiated Modal class with bound parameters.
// It provides access to the class methods with the bound parameters.
type ClsInstance struct {
	methods map[string]*Function
}

// Method returns the Function with the given name from a ClsInstance.
func (c *ClsInstance) Method(name string) (*Function, error) {
	method, ok := c.methods[name]
	if !ok {
		return nil, NotFoundError{fmt.Sprintf("method '%s' not found on class", name)}
	}
	return method, nil
}
