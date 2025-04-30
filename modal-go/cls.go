package modal

// Cls lookups and Function binding.

import (
	"context"
	"fmt"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

type Cls struct {
	ClassId string
	Methods map[string]*Function
	ctx     context.Context
}

// ClsLookup looks up an existing Cls constructing Function objects for each class method.
func ClsLookup(ctx context.Context, appName string, name string, options LookupOptions) (*Cls, error) {
	ctx = clientContext(ctx)
	cls := Cls{
		Methods: make(map[string]*Function),
		ctx:     context.Background(),
	}

	resp, err := client.ClassGet(ctx, pb.ClassGetRequest_builder{
		AppName:           appName,
		ObjectTag:         name,
		Namespace:         pb.DeploymentNamespace_DEPLOYMENT_NAMESPACE_WORKSPACE,
		EnvironmentName:   environmentName(options.Environment),
		LookupPublished:   false,
		OnlyClassFunction: true,
	}.Build())

	if err != nil {
		return nil, err
	}

	cls.ClassId = resp.GetClassId()

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

	// Check if we have method metadata on the class service function (v0.67+)
	serviceFunctionHandleMetadata := serviceFunction.GetHandleMetadata()
	if serviceFunctionHandleMetadata != nil && len(serviceFunctionHandleMetadata.GetMethodHandleMetadata()) > 0 {
		for methodName := range serviceFunctionHandleMetadata.GetMethodHandleMetadata() {
			function := &Function{
				FunctionId: serviceFunction.GetFunctionId(),
				MethodName: methodName,
				ctx:        ctx,
			}
			cls.Methods[methodName] = function
		}
	} else {
		// Legacy approach not supported
		return nil, fmt.Errorf("Go SDK requires Modal deployments using v0.67 or later.\n" +
			"Your deployment uses a legacy class format that is not supported. " +
			"Please update your deployment to make it compatible with this SDK.")

	}

	return &cls, nil
}
