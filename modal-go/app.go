package modal

import (
	"context"
	"fmt"
	"time"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// App references a deployed Modal App.
type App struct {
	AppId string
	ctx   context.Context
}

// LookupOptions are options for finding deployed Modal objects.
type LookupOptions struct {
	Environment     string
	CreateIfMissing bool
}

// DeleteOptions are options for deleting a named object.
type DeleteOptions struct {
	Environment string // Environment to delete the object from.
}

// EphemeralOptions are options for creating a temporary, nameless object.
type EphemeralOptions struct {
	Environment string // Environment to create the object in.
}

// SandboxOptions are options for creating a Modal Sandbox.
type SandboxOptions struct {
	CPU     float64       // CPU request in physical cores.
	Memory  int           // Memory request in MiB.
	Timeout time.Duration // Maximum duration for the Sandbox.
	Command []string      // Command to run in the Sandbox on startup.
}

// AppLookup looks up an existing App, or creates an empty one.
func AppLookup(ctx context.Context, name string, options *LookupOptions) (*App, error) {
	if options == nil {
		options = &LookupOptions{}
	}
	var err error
	ctx, err = clientContext(ctx)
	if err != nil {
		return nil, err
	}

	creationType := pb.ObjectCreationType_OBJECT_CREATION_TYPE_UNSPECIFIED
	if options.CreateIfMissing {
		creationType = pb.ObjectCreationType_OBJECT_CREATION_TYPE_CREATE_IF_MISSING
	}

	resp, err := client.AppGetOrCreate(ctx, pb.AppGetOrCreateRequest_builder{
		AppName:            name,
		EnvironmentName:    environmentName(options.Environment),
		ObjectCreationType: creationType,
	}.Build())

	if status, ok := status.FromError(err); ok && status.Code() == codes.NotFound {
		return nil, NotFoundError{fmt.Sprintf("app '%s' not found", name)}
	}
	if err != nil {
		return nil, err
	}

	return &App{AppId: resp.GetAppId(), ctx: ctx}, nil
}

// CreateSandbox creates a new Sandbox in the App with the specified image and options.
func (app *App) CreateSandbox(image *Image, options *SandboxOptions) (*Sandbox, error) {
	if options == nil {
		options = &SandboxOptions{}
	}
	createResp, err := client.SandboxCreate(app.ctx, pb.SandboxCreateRequest_builder{
		AppId: app.AppId,
		Definition: pb.Sandbox_builder{
			EntrypointArgs: options.Command,
			ImageId:        image.ImageId,
			TimeoutSecs:    uint32(options.Timeout.Seconds()),
			NetworkAccess: pb.NetworkAccess_builder{
				NetworkAccessType: pb.NetworkAccess_OPEN,
			}.Build(),
			Resources: pb.Resources_builder{
				MilliCpu: uint32(1000 * options.CPU),
				MemoryMb: uint32(options.Memory),
			}.Build(),
		}.Build(),
	}.Build())

	if err != nil {
		return nil, err
	}

	return newSandbox(app.ctx, createResp.GetSandboxId()), nil
}

// ImageFromRegistry creates an Image from a registry tag.
func (app *App) ImageFromRegistry(tag string) (*Image, error) {
	return fromRegistryInternal(app, tag, nil)
}

// ImageFromAwsEcr creates an Image from an AWS ECR tag, and secret for auth.
func (app *App) ImageFromAwsEcr(tag string, secret *Secret) (*Image, error) {
	imageRegistryConfig := pb.ImageRegistryConfig_builder{
		RegistryAuthType: pb.RegistryAuthType_REGISTRY_AUTH_TYPE_AWS,
		SecretId:         secret.SecretId,
	}.Build()
	return fromRegistryInternal(app, tag, imageRegistryConfig)
}
