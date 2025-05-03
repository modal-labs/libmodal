package modal

// Cls lookups and Function binding.

import (
	"context"
	"fmt"
	"sort"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
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

// ClsInstance represents an instantiated Modal class with bound parameters.
// It provides access to the class methods with the bound parameters.
type ClsInstance struct {
	ctx     context.Context
	methods map[string]*Function
}

// Instance creates a new instance of the class with the provided parameters.
func (c *Cls) Instance(kwargs map[string]any) (*ClsInstance, error) {
	instance := &ClsInstance{
		ctx:     c.ctx,
		methods: make(map[string]*Function),
	}
	// Class isn't parametrized, return a simple instance
	if len(c.schema) == 0 {
		for _, name := range c.methodNames {
			function := &Function{
				FunctionId: c.serviceFunctionId,
				MethodName: &name,
				ctx:        c.ctx,
			}
			instance.methods[name] = function
		}
		return instance, nil
	}

	// Class has parameters, bind the parameters to service function
	// and update method references
	boundFunctionId, err := c.bindParameters(kwargs)
	if err != nil {
		return nil, err
	}

	// Update all methods to use the bound function ID
	for _, name := range c.methodNames {
		boundMethod := &Function{
			FunctionId: boundFunctionId,
			MethodName: &name,
			ctx:        c.ctx,
		}
		instance.methods[name] = boundMethod
	}
	return instance, nil
}

// bindParameters processes the parameters and binds them to the class function.
func (c *Cls) bindParameters(kwargs map[string]any) (string, error) {
	// Create a parameter set
	paramSet := pb.ClassParameterSet_builder{
		Parameters: []*pb.ClassParameterValue{},
	}.Build()

	// Process each parameter according to the schema
	for _, paramSpec := range c.schema {
		name := paramSpec.GetName()
		value, provided := kwargs[name]

		// Check if required or use default
		if !provided {
			if paramSpec.GetHasDefault() {
				value = paramSpec.GetHasDefault()
			} else {
				return "", fmt.Errorf("required parameter '%s' not provided", name)
			}
		}

		// Encode the parameter value
		paramValue, err := encodeParameterValue(name, value, paramSpec.GetType())
		if err != nil {
			return "", err
		}

		// Add to the parameter set
		newParameters := append(paramSet.GetParameters(), paramValue)
		paramSet.SetParameters(newParameters)
	}

	// Serialize and bind parameters
	sortParameterSet(paramSet)
	serializedParams, err := proto.Marshal(paramSet)
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

// Method returns the Function with the given name from a ClsInstance.
func (c *ClsInstance) Method(name string) (*Function, error) {
	if c.methods == nil {
		return nil, fmt.Errorf("class instance has no methods")
	}

	method, ok := c.methods[name]
	if !ok {
		return nil, fmt.Errorf("method '%s' not found", name)
	}

	return method, nil
}

// ClsLookup looks up an existing Cls constructing Function objects for each class method.
func ClsLookup(ctx context.Context, appName string, name string, options LookupOptions) (*Cls, error) {
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
			cls.methodNames = append(cls.methodNames, methodName)
		}
	} else {
		// Legacy approach not supported
		return nil, fmt.Errorf("Go SDK requires Modal deployments using v0.67 or later.\n" +
			"Your deployment uses a legacy class format that is not supported. " +
			"Please update your deployment to make it compatible with this SDK.")

	}

	return &cls, nil
}

// Sort ClassParameterSet parameters based on parameter value
func sortParameterSet(paramSet *pb.ClassParameterSet) {
	values := paramSet.GetParameters()
	sort.Slice(values, func(i, j int) bool {
		return values[i] != nil && (values[j] == nil || values[i].GetName() < values[j].GetName())
	})
	paramSet.SetParameters(values)
}

// Helper function to copy methods map
func copyMethods(methods map[string]*Function) map[string]*Function {
	result := make(map[string]*Function)
	for name, method := range methods {
		result[name] = method
	}
	return result
}

// encodeParameterValue converts a Go value to a ParameterValue proto message
func encodeParameterValue(name string, value any, paramType pb.ParameterType) (*pb.ClassParameterValue, error) {
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
