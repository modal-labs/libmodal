package modal

// Cls lookups and Function binding.

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

// ClsOptions represents runtime options for a Modal Cls.
type ClsOptions struct {
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

// ClsConcurrencyOptions represents concurrency configuration for a Modal Cls.
type ClsConcurrencyOptions struct {
	MaxInputs    int
	TargetInputs *int
}

// ClsBatchingOptions represents batching configuration for a Modal Cls.
type ClsBatchingOptions struct {
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
	ctx               context.Context
	serviceFunctionId string
	schema            []*pb.ClassParameterSpec
	methodNames       []string
	inputPlaneUrl     string // if empty, use control plane
	options           *serviceOptions
}

// ClsLookup looks up an existing Cls on a deployed App.
func ClsLookup(ctx context.Context, appName string, name string, options *LookupOptions) (*Cls, error) {
	if options == nil {
		options = &LookupOptions{}
	}

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
	if serviceFunction.GetHandleMetadata().GetMethodHandleMetadata() != nil {
		for methodName := range serviceFunction.GetHandleMetadata().GetMethodHandleMetadata() {
			cls.methodNames = append(cls.methodNames, methodName)
		}
	} else {
		// Legacy approach not supported
		return nil, fmt.Errorf("Cls requires Modal deployments using client v0.67 or later")
	}

	if inputPlaneUrl := serviceFunction.GetHandleMetadata().GetInputPlaneUrl(); inputPlaneUrl != "" {
		cls.inputPlaneUrl = inputPlaneUrl
	}

	return &cls, nil
}

// Instance creates a new instance of the class with the provided parameters.
func (c *Cls) Instance(params map[string]any) (*ClsInstance, error) {
	var functionId string
	if len(c.schema) == 0 && !hasOptions(c.options) {
		functionId = c.serviceFunctionId
	} else {
		boundFunctionId, err := c.bindParameters(params)
		if err != nil {
			return nil, err
		}
		functionId = boundFunctionId
	}

	methods := make(map[string]*Function)
	for _, name := range c.methodNames {
		methods[name] = &Function{
			FunctionId:    functionId,
			MethodName:    &name,
			inputPlaneUrl: c.inputPlaneUrl,
			ctx:           c.ctx,
		}
	}
	return &ClsInstance{methods: methods}, nil
}

// WithOptions overrides the static Function configuration at runtime.
func (c *Cls) WithOptions(opts ClsOptions) *Cls {
	var secretsPtr *[]*Secret
	if opts.Secrets != nil {
		s := opts.Secrets
		secretsPtr = &s
	}
	var volumesPtr *map[string]*Volume
	if opts.Volumes != nil {
		v := opts.Volumes
		volumesPtr = &v
	}
	var envPtr *map[string]string
	if opts.Env != nil {
		e := opts.Env
		envPtr = &e
	}

	merged := mergeServiceOptions(c.options, &serviceOptions{
		cpu:              opts.CPU,
		memory:           opts.Memory,
		gpu:              opts.GPU,
		env:              envPtr,
		secrets:          secretsPtr,
		volumes:          volumesPtr,
		retries:          opts.Retries,
		maxContainers:    opts.MaxContainers,
		bufferContainers: opts.BufferContainers,
		scaledownWindow:  opts.ScaledownWindow,
		timeout:          opts.Timeout,
	})

	return &Cls{
		ctx:               c.ctx,
		serviceFunctionId: c.serviceFunctionId,
		schema:            c.schema,
		methodNames:       c.methodNames,
		inputPlaneUrl:     c.inputPlaneUrl,
		options:           merged,
	}
}

// WithConcurrency creates an instance of the Cls with input concurrency enabled or overridden with new values.
func (c *Cls) WithConcurrency(opts ClsConcurrencyOptions) *Cls {
	merged := mergeServiceOptions(c.options, &serviceOptions{
		maxConcurrentInputs:    &opts.MaxInputs,
		targetConcurrentInputs: opts.TargetInputs,
	})

	return &Cls{
		ctx:               c.ctx,
		serviceFunctionId: c.serviceFunctionId,
		schema:            c.schema,
		methodNames:       c.methodNames,
		inputPlaneUrl:     c.inputPlaneUrl,
		options:           merged,
	}
}

// WithBatching creates an instance of the Cls with dynamic batching enabled or overridden with new values.
func (c *Cls) WithBatching(opts ClsBatchingOptions) *Cls {
	merged := mergeServiceOptions(c.options, &serviceOptions{
		batchMaxSize: &opts.MaxBatchSize,
		batchWait:    &opts.Wait,
	})

	return &Cls{
		ctx:               c.ctx,
		serviceFunctionId: c.serviceFunctionId,
		schema:            c.schema,
		methodNames:       c.methodNames,
		inputPlaneUrl:     c.inputPlaneUrl,
		options:           merged,
	}
}

// bindParameters processes the parameters and binds them to the class function.
func (c *Cls) bindParameters(params map[string]any) (string, error) {
	serializedParams, err := encodeParameterSet(c.schema, params)
	if err != nil {
		return "", fmt.Errorf("failed to serialize parameters: %w", err)
	}

	var envSecret *Secret
	if c.options != nil && c.options.env != nil && len(*c.options.env) > 0 {
		envSecret, err = SecretFromMap(c.ctx, *c.options.env, nil)
		if err != nil {
			return "", err
		}
	}

	functionOptions, err := buildFunctionOptionsProto(c.options, envSecret)
	if err != nil {
		return "", fmt.Errorf("failed to build function options: %w", err)
	}

	// Bind parameters to create a parameterized function
	bindResp, err := client.FunctionBindParams(c.ctx, pb.FunctionBindParamsRequest_builder{
		FunctionId:       c.serviceFunctionId,
		SerializedParams: serializedParams,
		FunctionOptions:  functionOptions,
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

func buildFunctionOptionsProto(opts *serviceOptions, envSecret *Secret) (*pb.FunctionOptions, error) {
	if !hasOptions(opts) {
		return nil, nil
	}

	builder := pb.FunctionOptions_builder{}

	if opts.cpu != nil || opts.memory != nil || opts.gpu != nil {
		resBuilder := pb.Resources_builder{}
		if opts.cpu != nil {
			resBuilder.MilliCpu = uint32(*opts.cpu * 1000)
		}
		if opts.memory != nil {
			resBuilder.MemoryMb = uint32(*opts.memory)
		}
		if opts.gpu != nil {
			gpuConfig, err := parseGPUConfig(*opts.gpu)
			if err != nil {
				return nil, err
			}
			resBuilder.GpuConfig = gpuConfig
		}
		builder.Resources = resBuilder.Build()
	}

	secretIds := []string{}
	if opts.secrets != nil {
		for _, secret := range *opts.secrets {
			if secret != nil {
				secretIds = append(secretIds, secret.SecretId)
			}
		}
	}
	if (opts.env != nil && len(*opts.env) > 0) != (envSecret != nil) {
		return nil, fmt.Errorf("internal error: env and envSecret must both be provided or neither be provided")
	}
	if envSecret != nil {
		secretIds = append(secretIds, envSecret.SecretId)
	}

	builder.SecretIds = secretIds
	if len(secretIds) > 0 {
		builder.ReplaceSecretIds = true
	}

	if opts.volumes != nil {
		volumeMounts := []*pb.VolumeMount{}
		for mountPath, volume := range *opts.volumes {
			if volume != nil {
				volumeMounts = append(volumeMounts, pb.VolumeMount_builder{
					VolumeId:               volume.VolumeId,
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

	if opts.retries != nil {
		builder.RetryPolicy = pb.FunctionRetryPolicy_builder{
			Retries:            uint32(opts.retries.MaxRetries),
			BackoffCoefficient: opts.retries.BackoffCoefficient,
			InitialDelayMs:     uint32(opts.retries.InitialDelay / time.Millisecond),
			MaxDelayMs:         uint32(opts.retries.MaxDelay / time.Millisecond),
		}.Build()
	}

	if opts.maxContainers != nil {
		v := uint32(*opts.maxContainers)
		builder.ConcurrencyLimit = &v
	}
	if opts.bufferContainers != nil {
		v := uint32(*opts.bufferContainers)
		builder.BufferContainers = &v
	}

	if opts.scaledownWindow != nil {
		if *opts.scaledownWindow < time.Second {
			return nil, fmt.Errorf("scaledownWindow must be at least 1 second, got %v", *opts.scaledownWindow)
		}
		if (*opts.scaledownWindow)%time.Second != 0 {
			return nil, fmt.Errorf("scaledownWindow must be a whole number of seconds, got %v", *opts.scaledownWindow)
		}
		v := uint32((*opts.scaledownWindow) / time.Second)
		builder.TaskIdleTimeoutSecs = &v
	}
	if opts.timeout != nil {
		if *opts.timeout < time.Second {
			return nil, fmt.Errorf("timeout must be at least 1 second, got %v", *opts.timeout)
		}
		if (*opts.timeout)%time.Second != 0 {
			return nil, fmt.Errorf("timeout must be a whole number of seconds, got %v", *opts.timeout)
		}
		v := uint32((*opts.timeout) / time.Second)
		builder.TimeoutSecs = &v
	}

	if opts.maxConcurrentInputs != nil {
		v := uint32(*opts.maxConcurrentInputs)
		builder.MaxConcurrentInputs = &v
	}
	if opts.targetConcurrentInputs != nil {
		v := uint32(*opts.targetConcurrentInputs)
		builder.TargetConcurrentInputs = &v
	}

	if opts.batchMaxSize != nil {
		v := uint32(*opts.batchMaxSize)
		builder.BatchMaxSize = &v
	}
	if opts.batchWait != nil {
		v := uint64((*opts.batchWait) / time.Millisecond)
		builder.BatchLingerMs = &v
	}

	return builder.Build(), nil
}
