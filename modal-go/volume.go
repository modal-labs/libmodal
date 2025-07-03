package modal

import (
	"context"
	"fmt"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Volume represents a Modal volume that provides persistent storage.
type Volume struct {
	VolumeId string

	//lint:ignore U1000 may be used in future
	ctx context.Context
}

// VolumeFromNameOptions are options for finding Modal volumes.
type VolumeFromNameOptions struct {
	Environment     string
	CreateIfMissing bool
}

// VolumeFromName references a modal.Volume by its name.
func VolumeFromName(ctx context.Context, name string, options *VolumeFromNameOptions) (*Volume, error) {
	var err error
	ctx, err = clientContext(ctx)
	if err != nil {
		return nil, err
	}

	if options == nil {
		options = &VolumeFromNameOptions{}
	}

	creationType := pb.ObjectCreationType_OBJECT_CREATION_TYPE_UNSPECIFIED
	if options.CreateIfMissing {
		creationType = pb.ObjectCreationType_OBJECT_CREATION_TYPE_CREATE_IF_MISSING
	}

	resp, err := client.VolumeGetOrCreate(ctx, pb.VolumeGetOrCreateRequest_builder{
		DeploymentName:     name,
		EnvironmentName:    environmentName(options.Environment),
		ObjectCreationType: creationType,
	}.Build())

	if status, ok := status.FromError(err); ok && status.Code() == codes.NotFound {
		return nil, NotFoundError{fmt.Sprintf("Volume '%s' not found", name)}
	}
	if err != nil {
		return nil, err
	}

	return &Volume{VolumeId: resp.GetVolumeId(), ctx: ctx}, nil
}
