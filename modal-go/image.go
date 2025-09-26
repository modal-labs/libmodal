package modal

import (
	"context"
	"fmt"
	"io"
	"strings"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ImageService provides Image related operations.
type ImageService interface {
	FromRegistry(tag string, params *ImageFromRegistryParams) *Image
	FromAwsEcr(tag string, secret *Secret) *Image
	FromGcpArtifactRegistry(tag string, secret *Secret) *Image
	FromID(ctx context.Context, imageID string) (*Image, error)
	Delete(ctx context.Context, imageID string, params *ImageDeleteParams) error
}

type imageServiceImpl struct{ client *Client }

// ImageDockerfileCommandsParams are options for Image.DockerfileCommands().
type ImageDockerfileCommandsParams struct {
	// Environment variables to set in the build environment.
	Env map[string]string

	// Secrets that will be made available as environment variables to this layer's build environment.
	Secrets []*Secret

	// GPU reservation for this layer's build environment (e.g. "A100", "T4:2", "A100-80GB:4").
	GPU string

	// Ignore cached builds for this layer, similar to 'docker build --no-cache'.
	ForceBuild bool
}

// layer represents a single image layer with its build configuration.
type layer struct {
	commands   []string
	env        map[string]string
	secrets    []*Secret
	gpu        string
	forceBuild bool
}

// Image represents a Modal Image, which can be used to create Sandboxes.
type Image struct {
	ImageID string

	imageRegistryConfig *pb.ImageRegistryConfig
	tag                 string
	layers              []layer

	client *Client
}

// ImageFromRegistryParams are options for creating an Image from a registry.
type ImageFromRegistryParams struct {
	Secret *Secret // Secret for private registry authentication.
}

// FromRegistry builds a Modal Image from a public or private image registry without any changes.
func (s *imageServiceImpl) FromRegistry(tag string, params *ImageFromRegistryParams) *Image {
	if params == nil {
		params = &ImageFromRegistryParams{}
	}
	var imageRegistryConfig *pb.ImageRegistryConfig
	if params.Secret != nil {
		imageRegistryConfig = pb.ImageRegistryConfig_builder{
			RegistryAuthType: pb.RegistryAuthType_REGISTRY_AUTH_TYPE_STATIC_CREDS,
			SecretId:         params.Secret.SecretID,
		}.Build()
	}

	return &Image{
		ImageID:             "",
		imageRegistryConfig: imageRegistryConfig,
		tag:                 tag,
		layers:              []layer{{}},
		client:              s.client,
	}
}

// FromAwsEcr creates an Image from an AWS ECR tag
func (s *imageServiceImpl) FromAwsEcr(tag string, secret *Secret) *Image {
	imageRegistryConfig := pb.ImageRegistryConfig_builder{
		RegistryAuthType: pb.RegistryAuthType_REGISTRY_AUTH_TYPE_AWS,
		SecretId:         secret.SecretID,
	}.Build()

	return &Image{
		ImageID:             "",
		imageRegistryConfig: imageRegistryConfig,
		tag:                 tag,
		layers:              []layer{{}},
		client:              s.client,
	}
}

// FromGcpArtifactRegistry creates an Image from a GCP Artifact Registry tag.
func (s *imageServiceImpl) FromGcpArtifactRegistry(tag string, secret *Secret) *Image {
	imageRegistryConfig := pb.ImageRegistryConfig_builder{
		RegistryAuthType: pb.RegistryAuthType_REGISTRY_AUTH_TYPE_GCP,
		SecretId:         secret.SecretID,
	}.Build()
	return &Image{
		ImageID:             "",
		imageRegistryConfig: imageRegistryConfig,
		tag:                 tag,
		layers:              []layer{{}},
		client:              s.client,
	}
}

// FromID looks up an Image from an ID
func (s *imageServiceImpl) FromID(ctx context.Context, imageID string) (*Image, error) {
	resp, err := s.client.cpClient.ImageFromId(
		ctx,
		pb.ImageFromIdRequest_builder{
			ImageId: imageID,
		}.Build(),
	)
	if status, ok := status.FromError(err); ok && status.Code() == codes.NotFound {
		return nil, NotFoundError{fmt.Sprintf("Image '%s' not found", imageID)}
	}
	if err != nil {
		return nil, err
	}

	return &Image{
		ImageID: resp.GetImageId(),
		layers:  []layer{{}},
		client:  s.client,
	}, nil
}

// DockerfileCommands extends an image with arbitrary Dockerfile-like commands.
//
// Each call creates a new Image layer that will be built sequentially.
// The provided options apply only to this layer.
func (image *Image) DockerfileCommands(commands []string, params *ImageDockerfileCommandsParams) *Image {
	if len(commands) == 0 {
		return image
	}

	if params == nil {
		params = &ImageDockerfileCommandsParams{}
	}

	newLayer := layer{
		commands:   append([]string{}, commands...),
		env:        params.Env,
		secrets:    params.Secrets,
		gpu:        params.GPU,
		forceBuild: params.ForceBuild,
	}

	newLayers := append([]layer{}, image.layers...)
	newLayers = append(newLayers, newLayer)

	return &Image{
		ImageID:             "",
		tag:                 image.tag,
		imageRegistryConfig: image.imageRegistryConfig,
		layers:              newLayers,
		client:              image.client,
	}
}

func validateDockerfileCommands(commands []string) error {
	for _, command := range commands {
		trimmed := strings.ToUpper(strings.TrimSpace(command))
		if strings.HasPrefix(trimmed, "COPY ") && !strings.HasPrefix(trimmed, "COPY --FROM=") {
			return InvalidError{"COPY commands that copy from local context are not yet supported."}
		}
	}
	return nil
}

// Build eagerly builds an Image on Modal.
func (image *Image) Build(ctx context.Context, app *App) (*Image, error) {
	// Image is already hyrdated
	if image.ImageID != "" {
		return image, nil
	}

	for _, currentLayer := range image.layers {
		if err := validateDockerfileCommands(currentLayer.commands); err != nil {
			return nil, err
		}
	}

	var currentImageID string

	for i, currentLayer := range image.layers {
		mergedSecrets, err := mergeEnvIntoSecrets(ctx, image.client, &currentLayer.env, &currentLayer.secrets)
		if err != nil {
			return nil, err
		}

		var secretIds []string
		for _, secret := range mergedSecrets {
			secretIds = append(secretIds, secret.SecretID)
		}

		var gpuConfig *pb.GPUConfig
		if currentLayer.gpu != "" {
			var err error
			gpuConfig, err = parseGPUConfig(currentLayer.gpu)
			if err != nil {
				return nil, err
			}
		}

		var dockerfileCommands []string
		var baseImages []*pb.BaseImage

		if i == 0 {
			dockerfileCommands = append([]string{fmt.Sprintf("FROM %s", image.tag)}, currentLayer.commands...)
			baseImages = []*pb.BaseImage{}
		} else {
			dockerfileCommands = append([]string{"FROM base"}, currentLayer.commands...)
			baseImages = []*pb.BaseImage{pb.BaseImage_builder{
				DockerTag: "base",
				ImageId:   currentImageID,
			}.Build()}
		}

		resp, err := image.client.cpClient.ImageGetOrCreate(
			ctx,
			pb.ImageGetOrCreateRequest_builder{
				AppId: app.AppID,
				Image: pb.Image_builder{
					DockerfileCommands:  dockerfileCommands,
					ImageRegistryConfig: image.imageRegistryConfig,
					SecretIds:           secretIds,
					GpuConfig:           gpuConfig,
					ContextFiles:        []*pb.ImageContextFile{},
					BaseImages:          baseImages,
				}.Build(),
				BuilderVersion: imageBuilderVersion("", image.client.profile),
				ForceBuild:     currentLayer.forceBuild,
			}.Build(),
		)
		if err != nil {
			return nil, err
		}

		result := resp.GetResult()

		if result == nil || result.GetStatus() == pb.GenericResult_GENERIC_STATUS_UNSPECIFIED {
			// Not built or in the process of building - wait for build
			lastEntryID := ""
			for result == nil {
				stream, err := image.client.cpClient.ImageJoinStreaming(ctx, pb.ImageJoinStreamingRequest_builder{
					ImageId:     resp.GetImageId(),
					Timeout:     55,
					LastEntryId: lastEntryID,
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
						lastEntryID = item.GetEntryId()
					}
					if item.GetResult() != nil && item.GetResult().GetStatus() != pb.GenericResult_GENERIC_STATUS_UNSPECIFIED {
						result = item.GetResult()
						break
					}
					// Ignore all log lines and progress updates.
				}
			}
		}

		switch result.GetStatus() {
		case pb.GenericResult_GENERIC_STATUS_FAILURE:
			return nil, RemoteError{fmt.Sprintf("Image build for %s failed with the exception:\n%s", resp.GetImageId(), result.GetException())}
		case pb.GenericResult_GENERIC_STATUS_TERMINATED:
			return nil, RemoteError{fmt.Sprintf("Image build for %s terminated due to external shut-down. Please try again.", resp.GetImageId())}
		case pb.GenericResult_GENERIC_STATUS_TIMEOUT:
			return nil, RemoteError{fmt.Sprintf("Image build for %s timed out. Please try again with a larger timeout parameter.", resp.GetImageId())}
		case pb.GenericResult_GENERIC_STATUS_SUCCESS:
			// Success, do nothing
		default:
			return nil, RemoteError{fmt.Sprintf("Image build for %s failed with unknown status: %s", resp.GetImageId(), result.GetStatus())}
		}

		// The new image becomes the base for the next layer
		currentImageID = resp.GetImageId()
	}

	image.ImageID = currentImageID
	return image, nil
}

// ImageDeleteParams are options for deleting an Image.
type ImageDeleteParams struct {
}

// Delete deletes an Image by ID. Warning: This removes an *entire Image*, and cannot be undone.
func (s *imageServiceImpl) Delete(ctx context.Context, imageID string, params *ImageDeleteParams) error {
	image, err := s.FromID(ctx, imageID)
	if err != nil {
		return err
	}

	_, err = s.client.cpClient.ImageDelete(ctx, pb.ImageDeleteRequest_builder{ImageId: image.ImageID}.Build())
	return err
}
