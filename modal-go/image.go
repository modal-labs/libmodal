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

// ImageDockerfileCommandsOptions are options for Image.DockerfileCommands().
type ImageDockerfileCommandsOptions struct {
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
	ImageId string

	imageRegistryConfig *pb.ImageRegistryConfig
	tag                 string
	layers              []layer

	//lint:ignore U1000 may be used in future
	ctx context.Context
}

// NewImageFromRegistry builds a Modal Image from a public or private image registry without any changes.
func NewImageFromRegistry(tag string, options *ImageFromRegistryOptions) *Image {
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
		layers:              []layer{{}},
	}
}

// NewImageFromAwsEcr creates an Image from an AWS ECR tag.
func NewImageFromAwsEcr(tag string, secret *Secret) *Image {
	imageRegistryConfig := pb.ImageRegistryConfig_builder{
		RegistryAuthType: pb.RegistryAuthType_REGISTRY_AUTH_TYPE_AWS,
		SecretId:         secret.SecretId,
	}.Build()

	return &Image{
		ImageId:             "",
		imageRegistryConfig: imageRegistryConfig,
		tag:                 tag,
		layers:              []layer{{}},
	}
}

// NewImageFromGcpArtifactRegistry creates an Image from a GCP Artifact Registry tag.
func NewImageFromGcpArtifactRegistry(tag string, secret *Secret) *Image {
	imageRegistryConfig := pb.ImageRegistryConfig_builder{
		RegistryAuthType: pb.RegistryAuthType_REGISTRY_AUTH_TYPE_GCP,
		SecretId:         secret.SecretId,
	}.Build()
	return &Image{
		ImageId:             "",
		imageRegistryConfig: imageRegistryConfig,
		tag:                 tag,
		layers:              []layer{{}},
	}
}

// NewImageFromId looks up an Image from an ID
func NewImageFromId(ctx context.Context, imageId string) (*Image, error) {
	resp, err := client.ImageFromId(
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

	return &Image{
		ImageId: resp.GetImageId(),
		layers:  []layer{{}},
	}, nil
}

// DockerfileCommands extends an image with arbitrary Dockerfile-like commands.
//
// Each call creates a new Image layer that will be built sequentially.
// The provided options apply only to this layer.
func (image *Image) DockerfileCommands(commands []string, options *ImageDockerfileCommandsOptions) *Image {
	if len(commands) == 0 {
		return image
	}

	if options == nil {
		options = &ImageDockerfileCommandsOptions{}
	}

	newLayer := layer{
		commands:   append([]string{}, commands...),
		env:        options.Env,
		secrets:    options.Secrets,
		gpu:        options.GPU,
		forceBuild: options.ForceBuild,
	}

	newLayers := append([]layer{}, image.layers...)
	newLayers = append(newLayers, newLayer)

	return &Image{
		ImageId:             "",
		tag:                 image.tag,
		imageRegistryConfig: image.imageRegistryConfig,
		layers:              newLayers,
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
func (image *Image) Build(app *App) (*Image, error) {
	if image == nil {
		return nil, InvalidError{"image must be non-nil"}
	}

	// Image is already hyrdated
	if image.ImageId != "" {
		return image, nil
	}

	for _, currentLayer := range image.layers {
		if err := validateDockerfileCommands(currentLayer.commands); err != nil {
			return nil, err
		}
	}

	var currentImageId string

	for i, currentLayer := range image.layers {
		var secretIds []string
		for _, secret := range currentLayer.secrets {
			secretIds = append(secretIds, secret.SecretId)
		}
		if len(currentLayer.env) > 0 {
			envSecret, err := SecretFromMap(app.ctx, currentLayer.env, nil)
			if err != nil {
				return nil, err
			}
			secretIds = append(secretIds, envSecret.SecretId)
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
				ImageId:   currentImageId,
			}.Build()}
		}

		resp, err := client.ImageGetOrCreate(
			app.ctx,
			pb.ImageGetOrCreateRequest_builder{
				AppId: app.AppId,
				Image: pb.Image_builder{
					DockerfileCommands:  dockerfileCommands,
					ImageRegistryConfig: image.imageRegistryConfig,
					SecretIds:           secretIds,
					GpuConfig:           gpuConfig,
					ContextFiles:        []*pb.ImageContextFile{},
					BaseImages:          baseImages,
				}.Build(),
				BuilderVersion: imageBuilderVersion(""),
				ForceBuild:     currentLayer.forceBuild,
			}.Build(),
		)
		if err != nil {
			return nil, err
		}

		result := resp.GetResult()

		if result == nil || result.GetStatus() == pb.GenericResult_GENERIC_STATUS_UNSPECIFIED {
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
		currentImageId = resp.GetImageId()
	}

	image.ImageId = currentImageId
	image.ctx = app.ctx
	return image, nil
}

// ImageDeleteOptions are options for deleting an Image.
type ImageDeleteOptions struct {
}

// ImageDelete deletes an Image by ID. Warning: This removes an *entire Image*, and cannot be undone.
func ImageDelete(ctx context.Context, imageId string, options *ImageDeleteOptions) error {

	image, err := NewImageFromId(ctx, imageId)
	if err != nil {
		return err
	}

	_, err = client.ImageDelete(ctx, pb.ImageDeleteRequest_builder{ImageId: image.ImageId}.Build())
	return err
}
