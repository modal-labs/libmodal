package modal

import (
	"context"
	"fmt"
	"io"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

// Image represents a Modal image, which can be used to create sandboxes.
type Image struct {
	ImageId string

	//lint:ignore U1000 may be used in future
	ctx context.Context
}

func fromRegistryInternal(app *App, tag string) (*Image, error) {
	resp, err := client.ImageGetOrCreate(
		app.ctx,
		pb.ImageGetOrCreateRequest_builder{
			AppId: app.AppId,
			Image: pb.Image_builder{
				DockerfileCommands: []string{`FROM ` + tag},
			}.Build(),
			Namespace:      pb.DeploymentNamespace_DEPLOYMENT_NAMESPACE_WORKSPACE,
			BuilderVersion: "2024.10", // TODO: make this configurable
		}.Build(),
	)
	if err != nil {
		return nil, err
	}

	result := resp.GetResult()
	var metadata *pb.ImageMetadata

	if result != nil && result.GetStatus() != pb.GenericResult_GENERIC_STATUS_UNSPECIFIED {
		// Image has already been built
		metadata = resp.GetMetadata()
	} else {
		// Not built or in the process of building - wait for build
		lastEntryId := ""
		for result == nil {
			stream, err := client.ImageJoinStreaming(app.ctx, pb.ImageJoinStreamingRequest_builder{
				ImageId:     resp.GetImageId(),
				Timeout:     55,
				LastEntryId: lastEntryId,
			}.Build())
			if err != nil {
				return nil, err
			}
			for {
				item, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					return nil, err
				}
				if item.GetEntryId() != "" {
					lastEntryId = item.GetEntryId()
				}
				if item.GetResult() != nil && item.GetResult().GetStatus() != pb.GenericResult_GENERIC_STATUS_UNSPECIFIED {
					result = item.GetResult()
					metadata = item.GetMetadata()
					break
				}
				// Ignore all log lines and progress updates.
			}
		}
	}

	_ = metadata

	switch result.GetStatus() {
	case pb.GenericResult_GENERIC_STATUS_FAILURE:
		return nil, RemoteError{fmt.Sprintf("Image build for %s failed with the exception:\n%s", resp.GetImageId(), result.GetException())}
	case pb.GenericResult_GENERIC_STATUS_TERMINATED:
		return nil, RemoteError{fmt.Sprintf("Image build for %s terminated due to external shut-down, please try again", resp.GetImageId())}
	case pb.GenericResult_GENERIC_STATUS_TIMEOUT:
		return nil, RemoteError{fmt.Sprintf("Image build for %s timed out, please try again with a larger timeout parameter", resp.GetImageId())}
	case pb.GenericResult_GENERIC_STATUS_SUCCESS:
		// Success, do nothing
	default:
		return nil, RemoteError{fmt.Sprintf("Image build for %s failed with unknown status: %s", resp.GetImageId(), result.GetStatus())}
	}

	img := &Image{
		ImageId: resp.GetImageId(),
		ctx:     app.ctx,
	}
	return img, nil
}
