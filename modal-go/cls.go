package modal

import (
	"context"
	"fmt"
	"sort"
	"time"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// ClsService provides Cls related operations.
type ClsService interface {
	FromName(ctx context.Context, appName string, name string, params *ClsFromNameParams) (*Cls, error)
}

type clsServiceImpl struct{ client *Client }

// ClsWithOptionsParams represents runtime options for a Modal Cls.
type ClsWithOptionsParams struct {
	CPU              *float64
	Memory           *int
	GPU              *string
	Env              map[string]string
	Secrets          []*Secret
	Volumes          map[string]*Volume
	Retries          *Retries
	MaxContainers    *int
	BufferContainers *int
	ScaledownWindow  *time.Duration
	Timeout          *time.Duration
}

// ClsWithConcurrencyParams represents concurrency configuration for a Modal Cls.
type ClsWithConcurrencyParams struct {
	MaxInputs    int
	TargetInputs *int
}

// ClsWithBatchingParams represents batching configuration for a Modal Cls.
type ClsWithBatchingParams struct {
	MaxBatchSize int
	Wait         time.Duration
}

type serviceOptions struct {
	cpu                    *float64
	memory                 *int
	gpu                    *string
	env                    *map[string]string
	secrets                *[]*Secret
	volumes                *map[string]*Volume
	retries                *Retries
	maxContainers          *int
	bufferContainers       *int
	scaledownWindow        *time.Duration
	timeout                *time.Duration
	maxConcurrentInputs    *int
	targetConcurrentInputs *int
	batchMaxSize           *int
	batchWait              *time.Duration
}

// Cls represents a Modal class definition that can be instantiated with parameters.
// It contains metadata about the class and its methods.
type Cls struct {
	serviceFunctionID string
	schema            []*pb.ClassParameterSpec
	methodNames       []string
	inputPlaneURL     string // if empty, use control plane
	serviceOptions    *serviceOptions

	client *Client
}

// ClsFromNameParams are options for client.Cls.FromName.
type ClsFromNameParams struct {
	Environment     string
	CreateIfMissing bool
}

// FromName references a Cls from a deployed App by its name.
func (s *clsServiceImpl) FromName(ctx context.Context, appName string, name string, params *ClsFromNameParams) (*Cls, error) {
	if params == nil {
		params = &ClsFromNameParams{}
	}

	cls := Cls{
		methodNames: []string{},
		client:      s.client,
	}

	// Find class service function metadata. Service functions are used to implement class methods,
	// which are invoked using a combination of service function ID and the method name.
	serviceFunctionName := fmt.Sprintf("%s.*", name)
	serviceFunction, err := s.client.cpClient.FunctionGet(ctx, pb.FunctionGetRequest_builder{
		AppName:         appName,
		ObjectTag:       serviceFunctionName,
		EnvironmentName: environmentName(params.Environment, s.client.profile),
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

	cls.serviceFunctionID = serviceFunction.GetFunctionId()

	// Check if we have method metadata on the class service function (v0.67+)
	if serviceFunction.GetHandleMetadata().GetMethodHandleMetadata() != nil {
		for methodName := range serviceFunction.GetHandleMetadata().GetMethodHandleMetadata() {
			cls.methodNames = append(cls.methodNames, methodName)
		}
	} else {
		// Legacy approach not supported
		return nil, fmt.Errorf("Cls requires Modal deployments using client v0.67 or later")
	}

	if inputPlaneURL := serviceFunction.GetHandleMetadata().GetInputPlaneUrl(); inputPlaneURL != "" {
		cls.inputPlaneURL = inputPlaneURL
	}

	return &cls, nil
}

// Instance creates a new instance of the class with the provided parameters.
func (c *Cls) Instance(ctx context.Context, parameters map[string]any) (*ClsInstance, error) {
	var functionID string
	if len(c.schema) == 0 && !hasOptions(c.serviceOptions) {
		functionID = c.serviceFunctionID
	} else {
		opts := c.serviceOptions
		if opts == nil {
			opts = &serviceOptions{}
		}
		boundFunctionID, err := c.bindParameters(ctx, parameters, opts)
		if err != nil {
			return nil, err
		}
		functionID = boundFunctionID
	}

	methods := make(map[string]*Function)
	for _, name := range c.methodNames {
		methods[name] = &Function{
			FunctionID:    functionID,
			MethodName:    &name,
			inputPlaneURL: c.inputPlaneURL,
			client:        c.client,
		}
	}
	return &ClsInstance{methods: methods}, nil
}

// WithOptions overrides the static Function configuration at runtime.
func (c *Cls) WithOptions(params *ClsWithOptionsParams) *Cls {
	if params == nil {
		params = &ClsWithOptionsParams{}
	}

	var secretsPtr *[]*Secret
	if params.Secrets != nil {
		s := params.Secrets
		secretsPtr = &s
	}
	var volumesPtr *map[string]*Volume
	if params.Volumes != nil {
		v := params.Volumes
		volumesPtr = &v
	}
	var envPtr *map[string]string
	if params.Env != nil {
		e := params.Env
		envPtr = &e
	}

	merged := mergeServiceOptions(c.serviceOptions, &serviceOptions{
		cpu:              params.CPU,
		memory:           params.Memory,
		gpu:              params.GPU,
		env:              envPtr,
		secrets:          secretsPtr,
		volumes:          volumesPtr,
		retries:          params.Retries,
		maxContainers:    params.MaxContainers,
		bufferContainers: params.BufferContainers,
		scaledownWindow:  params.ScaledownWindow,
		timeout:          params.Timeout,
	})

	return &Cls{
		serviceFunctionID: c.serviceFunctionID,
		schema:            c.schema,
		methodNames:       c.methodNames,
		inputPlaneURL:     c.inputPlaneURL,
		serviceOptions:    merged,
		client:            c.client,
	}
}

// WithConcurrency creates an instance of the Cls with input concurrency enabled or overridden with new values.
func (c *Cls) WithConcurrency(params *ClsWithConcurrencyParams) *Cls {
	if params == nil {
		params = &ClsWithConcurrencyParams{}
	}

	merged := mergeServiceOptions(c.serviceOptions, &serviceOptions{
		maxConcurrentInputs:    &params.MaxInputs,
		targetConcurrentInputs: params.TargetInputs,
	})

	return &Cls{
		serviceFunctionID: c.serviceFunctionID,
		schema:            c.schema,
		methodNames:       c.methodNames,
		inputPlaneURL:     c.inputPlaneURL,
		serviceOptions:    merged,
		client:            c.client,
	}
}

// WithBatching creates an instance of the Cls with dynamic batching enabled or overridden with new values.
func (c *Cls) WithBatching(params *ClsWithBatchingParams) *Cls {
	if params == nil {
		params = &ClsWithBatchingParams{}
	}

	merged := mergeServiceOptions(c.serviceOptions, &serviceOptions{
		batchMaxSize: &params.MaxBatchSize,
		batchWait:    &params.Wait,
	})

	return &Cls{
		serviceFunctionID: c.serviceFunctionID,
		schema:            c.schema,
		methodNames:       c.methodNames,
		inputPlaneURL:     c.inputPlaneURL,
		serviceOptions:    merged,
		client:            c.client,
	}
}

// bindParameters processes the parameters and binds them to the class function.
func (c *Cls) bindParameters(ctx context.Context, parameters map[string]any, opts *serviceOptions) (string, error) {
	mergedSecrets, err := mergeEnvIntoSecrets(ctx, c.client, opts.env, opts.secrets)
	if err != nil {
		return "", err
	}

	mergedOptions := mergeServiceOptions(opts, &serviceOptions{
		secrets: &mergedSecrets,
		env:     nil,
	})

	serializedParams, err := encodeParameterSet(c.schema, parameters)
	if err != nil {
		return "", fmt.Errorf("failed to serialize parameters: %w", err)
	}
	functionOptions, err := buildFunctionOptionsProto(mergedOptions)
	if err != nil {
		return "", fmt.Errorf("failed to build function options: %w", err)
	}
	bindResp, err := c.client.cpClient.FunctionBindParams(ctx, pb.FunctionBindParamsRequest_builder{
		FunctionId:       c.serviceFunctionID,
		SerializedParams: serializedParams,
		FunctionOptions:  functionOptions,
	}.Build())
	if err != nil {
		return "", fmt.Errorf("failed to bind parameters: %w", err)
	}

	return bindResp.GetBoundFunctionId(), nil
}

// encodeParameterSet encodes the parameter values into a binary format.
func encodeParameterSet(schema []*pb.ClassParameterSpec, parameters map[string]any) ([]byte, error) {
	var encoded []*pb.ClassParameterValue

	for _, parameterSpec := range schema {
		paramValue, err := encodeParameter(parameterSpec, parameters[parameterSpec.GetName()])
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
func encodeParameter(parameterSpec *pb.ClassParameterSpec, value any) (*pb.ClassParameterValue, error) {
	name := parameterSpec.GetName()
	paramType := parameterSpec.GetType()
	paramValue := pb.ClassParameterValue_builder{
		Name: name,
		Type: paramType,
	}.Build()

	switch paramType {
	case pb.ParameterType_PARAM_TYPE_STRING:
		if value == nil && parameterSpec.GetHasDefault() {
			value = parameterSpec.GetStringDefault()
		}
		strValue, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("parameter '%s' must be a string", name)
		}
		paramValue.SetStringValue(strValue)

	case pb.ParameterType_PARAM_TYPE_INT:
		if value == nil && parameterSpec.GetHasDefault() {
			value = parameterSpec.GetIntDefault()
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
		if value == nil && parameterSpec.GetHasDefault() {
			value = parameterSpec.GetBoolDefault()
		}
		boolValue, ok := value.(bool)
		if !ok {
			return nil, fmt.Errorf("parameter '%s' must be a boolean", name)
		}
		paramValue.SetBoolValue(boolValue)

	case pb.ParameterType_PARAM_TYPE_BYTES:
		if value == nil && parameterSpec.GetHasDefault() {
			value = parameterSpec.GetBytesDefault()
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

func hasOptions(o *serviceOptions) bool {
	return o != nil && *o != (serviceOptions{})
}

func mergeServiceOptions(base, new *serviceOptions) *serviceOptions {
	if base == nil {
		return new
	}
	if new == nil {
		return base
	}

	merged := &serviceOptions{
		cpu:                    base.cpu,
		memory:                 base.memory,
		gpu:                    base.gpu,
		env:                    base.env,
		secrets:                base.secrets,
		volumes:                base.volumes,
		retries:                base.retries,
		maxContainers:          base.maxContainers,
		bufferContainers:       base.bufferContainers,
		scaledownWindow:        base.scaledownWindow,
		timeout:                base.timeout,
		maxConcurrentInputs:    base.maxConcurrentInputs,
		targetConcurrentInputs: base.targetConcurrentInputs,
		batchMaxSize:           base.batchMaxSize,
		batchWait:              base.batchWait,
	}

	if new.cpu != nil {
		merged.cpu = new.cpu
	}
	if new.memory != nil {
		merged.memory = new.memory
	}
	if new.gpu != nil {
		merged.gpu = new.gpu
	}
	if new.env != nil {
		merged.env = new.env
	}
	if new.secrets != nil {
		merged.secrets = new.secrets
	}
	if new.volumes != nil {
		merged.volumes = new.volumes
	}
	if new.retries != nil {
		merged.retries = new.retries
	}
	if new.maxContainers != nil {
		merged.maxContainers = new.maxContainers
	}
	if new.bufferContainers != nil {
		merged.bufferContainers = new.bufferContainers
	}
	if new.scaledownWindow != nil {
		merged.scaledownWindow = new.scaledownWindow
	}
	if new.timeout != nil {
		merged.timeout = new.timeout
	}
	if new.maxConcurrentInputs != nil {
		merged.maxConcurrentInputs = new.maxConcurrentInputs
	}
	if new.targetConcurrentInputs != nil {
		merged.targetConcurrentInputs = new.targetConcurrentInputs
	}
	if new.batchMaxSize != nil {
		merged.batchMaxSize = new.batchMaxSize
	}
	if new.batchWait != nil {
		merged.batchWait = new.batchWait
	}

	return merged
}

func buildFunctionOptionsProto(options *serviceOptions) (*pb.FunctionOptions, error) {
	if !hasOptions(options) {
		return nil, nil
	}

	builder := pb.FunctionOptions_builder{}

	if options.cpu != nil || options.memory != nil || options.gpu != nil {
		resBuilder := pb.Resources_builder{}
		if options.cpu != nil {
			resBuilder.MilliCpu = uint32(*options.cpu * 1000)
		}
		if options.memory != nil {
			resBuilder.MemoryMb = uint32(*options.memory)
		}
		if options.gpu != nil {
			gpuConfig, err := parseGPUConfig(*options.gpu)
			if err != nil {
				return nil, err
			}
			resBuilder.GpuConfig = gpuConfig
		}
		builder.Resources = resBuilder.Build()
	}

	secretIds := []string{}
	if options.secrets != nil {
		for _, secret := range *options.secrets {
			if secret != nil {
				secretIds = append(secretIds, secret.SecretID)
			}
		}
	}

	builder.SecretIds = secretIds
	if len(secretIds) > 0 {
		builder.ReplaceSecretIds = true
	}

	if options.volumes != nil {
		volumeMounts := []*pb.VolumeMount{}
		for mountPath, volume := range *options.volumes {
			if volume != nil {
				volumeMounts = append(volumeMounts, pb.VolumeMount_builder{
					VolumeId:               volume.VolumeID,
					MountPath:              mountPath,
					AllowBackgroundCommits: true,
					ReadOnly:               volume.IsReadOnly(),
				}.Build())
			}
		}
		builder.VolumeMounts = volumeMounts
		if len(volumeMounts) > 0 {
			builder.ReplaceVolumeMounts = true
		}
	}

	if options.retries != nil {
		builder.RetryPolicy = pb.FunctionRetryPolicy_builder{
			Retries:            uint32(options.retries.MaxRetries),
			BackoffCoefficient: options.retries.BackoffCoefficient,
			InitialDelayMs:     uint32(options.retries.InitialDelay / time.Millisecond),
			MaxDelayMs:         uint32(options.retries.MaxDelay / time.Millisecond),
		}.Build()
	}

	if options.maxContainers != nil {
		v := uint32(*options.maxContainers)
		builder.ConcurrencyLimit = &v
	}
	if options.bufferContainers != nil {
		v := uint32(*options.bufferContainers)
		builder.BufferContainers = &v
	}

	if options.scaledownWindow != nil {
		if *options.scaledownWindow < time.Second {
			return nil, fmt.Errorf("scaledownWindow must be at least 1 second, got %v", *options.scaledownWindow)
		}
		if (*options.scaledownWindow)%time.Second != 0 {
			return nil, fmt.Errorf("scaledownWindow must be a whole number of seconds, got %v", *options.scaledownWindow)
		}
		v := uint32((*options.scaledownWindow) / time.Second)
		builder.TaskIdleTimeoutSecs = &v
	}
	if options.timeout != nil {
		if *options.timeout < time.Second {
			return nil, fmt.Errorf("timeout must be at least 1 second, got %v", *options.timeout)
		}
		if (*options.timeout)%time.Second != 0 {
			return nil, fmt.Errorf("timeout must be a whole number of seconds, got %v", *options.timeout)
		}
		v := uint32((*options.timeout) / time.Second)
		builder.TimeoutSecs = &v
	}

	if options.maxConcurrentInputs != nil {
		v := uint32(*options.maxConcurrentInputs)
		builder.MaxConcurrentInputs = &v
	}
	if options.targetConcurrentInputs != nil {
		v := uint32(*options.targetConcurrentInputs)
		builder.TargetConcurrentInputs = &v
	}

	if options.batchMaxSize != nil {
		v := uint32(*options.batchMaxSize)
		builder.BatchMaxSize = &v
	}
	if options.batchWait != nil {
		v := uint64((*options.batchWait) / time.Millisecond)
		builder.BatchLingerMs = &v
	}

	return builder.Build(), nil
}
