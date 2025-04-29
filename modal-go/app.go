package modal

import (
	"context"
	"time"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
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

// SandboxOptions are options for creating a Modal Sandbox.
type SandboxOptions struct {
	CPU     float64       // CPU request in physical cores.
	Memory  int           // Memory request in MiB.
	Timeout time.Duration // Maximum duration for the Sandbox.
	Command []string      // Command to run in the Sandbox on startup.
}

// AppLookup looks up an existing App, or creates an empty one.
func AppLookup(ctx context.Context, name string, options LookupOptions) (*App, error) {
	ctx = clientContext(ctx)

	creationType := pb.ObjectCreationType_OBJECT_CREATION_TYPE_UNSPECIFIED
	if options.CreateIfMissing {
		creationType = pb.ObjectCreationType_OBJECT_CREATION_TYPE_CREATE_IF_MISSING
	}

	resp, err := client.AppGetOrCreate(ctx, pb.AppGetOrCreateRequest_builder{
		AppName:            name,
		EnvironmentName:    environmentName(options.Environment),
		ObjectCreationType: creationType,
	}.Build())

	if err != nil {
		return nil, err
	}

	return &App{AppId: resp.GetAppId(), ctx: ctx}, nil
}

// CreateSandbox creates a new Sandbox in the App with the specified image and options.
func (app *App) CreateSandbox(image *Image, options SandboxOptions) (*Sandbox, error) {
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
	return fromRegistryInternal(app, tag)
}
