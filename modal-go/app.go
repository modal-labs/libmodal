package modal

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AppService provides App related operations.
type AppService struct{ client *Client }

// App references a deployed Modal App.
type App struct {
	AppId string
	Name  string
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

// parseGPUConfig parses a GPU configuration string into a GPUConfig object.
// The GPU string format is "type" or "type:count" (e.g. "T4", "A100:2").
// Returns nil if gpu is empty, or an error if the format is invalid.
func parseGPUConfig(gpu string) (*pb.GPUConfig, error) {
	if gpu == "" {
		return nil, nil
	}

	gpuType := gpu
	count := uint32(1)

	if strings.Contains(gpu, ":") {
		parts := strings.SplitN(gpu, ":", 2)
		gpuType = parts[0]
		parsedCount, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil || parsedCount < 1 {
			return nil, fmt.Errorf("invalid GPU count: %s, value must be a positive integer", parts[1])
		}
		count = uint32(parsedCount)
	}

	return pb.GPUConfig_builder{
		Type:    0, // Deprecated field, but required by proto
		Count:   count,
		GpuType: strings.ToUpper(gpuType),
	}.Build(), nil
}

// Lookup looks up an existing App, or creates an empty one.
func (s *AppService) Lookup(ctx context.Context, name string, options *LookupOptions) (*App, error) {
	if options == nil {
		options = &LookupOptions{}
	}

	creationType := pb.ObjectCreationType_OBJECT_CREATION_TYPE_UNSPECIFIED
	if options.CreateIfMissing {
		creationType = pb.ObjectCreationType_OBJECT_CREATION_TYPE_CREATE_IF_MISSING
	}

	resp, err := s.client.cpClient.AppGetOrCreate(ctx, pb.AppGetOrCreateRequest_builder{
		AppName:            name,
		EnvironmentName:    environmentName(options.Environment, s.client.profile),
		ObjectCreationType: creationType,
	}.Build())

	if status, ok := status.FromError(err); ok && status.Code() == codes.NotFound {
		return nil, NotFoundError{fmt.Sprintf("App '%s' not found", name)}
	}
	if err != nil {
		return nil, err
	}

	return &App{AppId: resp.GetAppId(), Name: name}, nil
}
