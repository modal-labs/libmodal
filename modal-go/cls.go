package modal

// Cls lookups and Function binding.

import (
	"context"
	"fmt"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/protobuf/proto"
)

type Cls struct {
	methods           map[string]*Function
	schema            []*pb.ClassParameterSpec // schema for parameters used in init
	serviceFunctionId string
	initialized       bool
	ctx               context.Context
}

// Creates parametrized instnace of Cls.
func (c *Cls) Instance(kwargs map[string]any) (*Cls, error) {
	c.initialized = true
	if len(c.schema) == 0 {
		return c, nil
	} else {

		// Create a parameter set with values from kwargs
		paramSet := pb.ClassParameterSet_builder{
			Parameters: []*pb.ClassParameterValue{},
		}.Build()

		// For each parameter in the schema
		for _, paramSpec := range c.schema {
			name := paramSpec.GetName()
			value, provided := kwargs[name]

			// Use default value if value is not provided but avaialble
			defaultValue := paramSpec.GetHasDefault()
			if !provided && paramSpec.GetHasDefault() {
				value = defaultValue
			}

			// Error if required and not provided
			if !provided && !paramSpec.GetHasDefault() {
				return nil, fmt.Errorf("required parameter '%s' not provided", name)
			}

			// Convert value to proto parameter
			paramValue, err := EncodeParameterValue(name, value, paramSpec.GetType())
			if err != nil {
				return nil, err
			}

			newParameters := append(paramSet.GetParameters(), paramValue)
			paramSet.SetParameters(newParameters)
		}

		// Serialize the parameter set
		serializedParams, err := proto.Marshal(paramSet)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize parameters: %w", err)
		}

		// Bind parameters to create a parameterized function
		bindResp, err := client.FunctionBindParams(c.ctx, pb.FunctionBindParamsRequest_builder{
			FunctionId:       c.serviceFunctionId,
			SerializedParams: serializedParams,
		}.Build())
		if err != nil {
			return nil, fmt.Errorf("failed to bind parameters: %w", err)
		}

		// Update all methods to use the bound function ID
		boundFunctionId := bindResp.GetBoundFunctionId()
		for methodName, function := range c.methods {
			function.FunctionId = boundFunctionId
			c.methods[methodName] = function
		}
	}
	return c, nil
}

// Method returns the Function with the given name.
// It returns an error if the Cls is not initialized or if the method doesn't exist.
func (c *Cls) Method(name string) (*Function, error) {
	if !c.initialized {
		return nil, fmt.Errorf("Cls not initialized")
	}
	if c.methods == nil {
		return nil, fmt.Errorf("Cls has no methods")
	}

	method, ok := c.methods[name]
	if !ok {
		return nil, fmt.Errorf("Cls method %s not found", name)
	}

	return method, nil
}

// ClsLookup looks up an existing Cls constructing Function objects for each class method.
func ClsLookup(ctx context.Context, appName string, name string, options LookupOptions) (*Cls, error) {
	ctx = clientContext(ctx)
	cls := Cls{
		methods: make(map[string]*Function),
		ctx:     ctx,
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
	if serviceFunctionHandleMetadata != nil && len(serviceFunctionHandleMetadata.GetMethodHandleMetadata()) > 0 {
		for methodName := range serviceFunctionHandleMetadata.GetMethodHandleMetadata() {
			function := &Function{
				FunctionId: serviceFunction.GetFunctionId(),
				MethodName: &methodName,
				ctx:        ctx,
			}
			cls.methods[methodName] = function
		}
	} else {
		// Legacy approach not supported
		return nil, fmt.Errorf("Go SDK requires Modal deployments using v0.67 or later.\n" +
			"Your deployment uses a legacy class format that is not supported. " +
			"Please update your deployment to make it compatible with this SDK.")

	}

	cls.initialized = false
	return &cls, nil
}

// EncodeParameterValue converts a Go value to a ParameterValue proto message
func EncodeParameterValue(name string, value any, paramType pb.ParameterType) (*pb.ClassParameterValue, error) {
	paramValue := pb.ClassParameterValue_builder{
		Name: name,
		Type: paramType,
	}.Build()

	switch paramType {
	case pb.ParameterType_PARAM_TYPE_STRING:
		strValue, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("parameter '%s' must be a string", name)
		}
		paramValue.SetStringValue(strValue)

	case pb.ParameterType_PARAM_TYPE_INT:
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
		boolValue, ok := value.(bool)
		if !ok {
			return nil, fmt.Errorf("parameter '%s' must be a boolean", name)
		}
		paramValue.SetBoolValue(boolValue)

	case pb.ParameterType_PARAM_TYPE_BYTES:
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
