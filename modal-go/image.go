package modal

import (
	"context"
	"fmt"
	"io"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ImageService provides Image related operations.
type ImageService struct{ client *Client }

// Image represents a Modal Image, which can be used to create Sandboxes.
type Image struct {
	ImageId string

	imageRegistryConfig *pb.ImageRegistryConfig
	tag                 string

	client *Client
}

// ImageFromRegistryOptions are options for creating an Image from a registry.
type ImageFromRegistryOptions struct {
	Secret *Secret // Secret for private registry authentication.
}

// FromRegistry builds a Modal Image from a public or private image registry without any changes.
func (s *ImageService) FromRegistry(tag string, options *ImageFromRegistryOptions) *Image {
	if options == nil {
		options = &ImageFromRegistryOptions{}
	}
	var imageRegistryConfig *pb.ImageRegistryConfig
	if options.Secret != nil {
		imageRegistryConfig = pb.ImageRegistryConfig_builder{
			RegistryAuthType: pb.RegistryAuthType_REGISTRY_AUTH_TYPE_STATIC_CREDS,
			SecretId:         options.Secret.SecretId,
		}.Build()
	}

	return &Image{
		ImageId:             "",
		imageRegistryConfig: imageRegistryConfig,
		tag:                 tag,
		client:              s.client,
	}
}

// FromAwsEcr creates an Image from an AWS ECR tag
func (s *ImageService) FromAwsEcr(tag string, secret *Secret) *Image {
	imageRegistryConfig := pb.ImageRegistryConfig_builder{
		RegistryAuthType: pb.RegistryAuthType_REGISTRY_AUTH_TYPE_AWS,
		SecretId:         secret.SecretId,
	}.Build()

	return &Image{
		ImageId:             "",
		imageRegistryConfig: imageRegistryConfig,
		tag:                 tag,
		client:              s.client,
	}
}

// FromGcpArtifactRegistry creates an Image from a GCP Artifact Registry tag.
func (s *ImageService) FromGcpArtifactRegistry(tag string, secret *Secret) *Image {
	imageRegistryConfig := pb.ImageRegistryConfig_builder{
		RegistryAuthType: pb.RegistryAuthType_REGISTRY_AUTH_TYPE_GCP,
		SecretId:         secret.SecretId,
	}.Build()
	return &Image{
		ImageId:             "",
		imageRegistryConfig: imageRegistryConfig,
		tag:                 tag,
		client:              s.client,
	}
}

// FromId looks up an Image from an ID
func (s *ImageService) FromId(ctx context.Context, imageId string) (*Image, error) {
	resp, err := s.client.cpClient.ImageFromId(
		ctx,
		pb.ImageFromIdRequest_builder{
			ImageId: imageId,
		}.Build(),
	)
	if status, ok := status.FromError(err); ok && status.Code() == codes.NotFound {
		return nil, NotFoundError{fmt.Sprintf("Image '%s' not found", imageId)}
	}
	if err != nil {
		return nil, err
	}

	return &Image{ImageId: resp.GetImageId(), client: s.client}, nil
}

// Build eagerly builds an Image on Modal.
func (image *Image) Build(ctx context.Context, app *App) (*Image, error) {
	// Image is already hyrdated
	if image.ImageId != "" {
		return image, nil
	}

	resp, err := image.client.cpClient.ImageGetOrCreate(
		ctx,
		pb.ImageGetOrCreateRequest_builder{
			AppId: app.AppId,
			Image: pb.Image_builder{
				DockerfileCommands:  []string{`FROM ` + image.tag},
				ImageRegistryConfig: image.imageRegistryConfig,
			}.Build(),
			BuilderVersion: imageBuilderVersion("", image.client.profile),
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
			stream, err := image.client.cpClient.ImageJoinStreaming(ctx, pb.ImageJoinStreamingRequest_builder{
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

	image.ImageId = resp.GetImageId()
	return image, nil
}

// ImageDeleteOptions are options for deleting an Image.
type ImageDeleteOptions struct {
}

// Delete deletes an Image by ID. Warning: This removes an *entire Image*, and cannot be undone.
func (s *ImageService) Delete(ctx context.Context, imageId string, options *ImageDeleteOptions) error {
	image, err := s.FromId(ctx, imageId)
	if err != nil {
		return err
	}

	_, err = s.client.cpClient.ImageDelete(ctx, pb.ImageDeleteRequest_builder{ImageId: image.ImageId}.Build())
	return err
}
