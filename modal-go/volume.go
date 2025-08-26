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
	Name     string
	readOnly bool
	cancel   context.CancelFunc
	ctx      context.Context
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

	return &Volume{VolumeId: resp.GetVolumeId(), Name: name, readOnly: false, cancel: nil, ctx: ctx}, nil
}

// ReadOnly configures Volume to mount as read-only.
func (v *Volume) ReadOnly() *Volume {
	return &Volume{
		VolumeId: v.VolumeId,
		Name:     v.Name,
		readOnly: true,
		cancel:   v.cancel,
		ctx:      v.ctx,
	}
}

// IsReadOnly returns true if the volume is configured to mount as read-only.
func (v *Volume) IsReadOnly() bool {
	return v.readOnly
}

// VolumeEphemeral creates a nameless, temporary volume. Caller must CloseEphemeral.
func VolumeEphemeral(ctx context.Context, options *EphemeralOptions) (*Volume, error) {
	if options == nil {
		options = &EphemeralOptions{}
	}
	var err error
	ctx, err = clientContext(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.VolumeGetOrCreate(ctx, pb.VolumeGetOrCreateRequest_builder{
		ObjectCreationType: pb.ObjectCreationType_OBJECT_CREATION_TYPE_EPHEMERAL,
		EnvironmentName:    environmentName(options.Environment),
	}.Build())
	if err != nil {
		return nil, err
	}

	ephemeralCtx, cancel := context.WithCancel(ctx)
	startEphemeralHeartbeat(ephemeralCtx, func() error {
		_, err := client.VolumeHeartbeat(ephemeralCtx, pb.VolumeHeartbeatRequest_builder{
			VolumeId: resp.GetVolumeId(),
		}.Build())
		return err
	})

	return &Volume{
		VolumeId: resp.GetVolumeId(),
		readOnly: false,
		cancel:   cancel,
		ctx:      ephemeralCtx,
	}, nil
}

// CloseEphemeral deletes an ephemeral volume, only used with VolumeEphemeral.
func (v *Volume) CloseEphemeral() {
	if v.cancel != nil {
		v.cancel()
	} else {
		// We panic in this case because of invalid usage. In general, methods
		// used with `defer` like CloseEphemeral should not return errors.
		panic(fmt.Sprintf("volume %s is not ephemeral", v.VolumeId))
	}
}
