package test

import (
	"context"
	"io"
	"testing"

	"github.com/modal-labs/libmodal/modal-go"
	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"github.com/modal-labs/libmodal/modal-go/testsupport/grpcmock"
	"github.com/onsi/gomega"
)

func TestImageFromId(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	ctx := context.Background()

	app, err := modal.AppLookup(ctx, "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := modal.NewImageFromRegistry("alpine:3.21", nil).Build(app)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	imageFromId, err := modal.NewImageFromId(ctx, image.ImageId)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(imageFromId.ImageId).Should(gomega.Equal(image.ImageId))

	_, err = modal.NewImageFromId(ctx, "im-nonexistent")
	g.Expect(err).Should(gomega.HaveOccurred())
}

//nolint:staticcheck
func TestImageFromRegistry(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromRegistry("alpine:3.21", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))
}

//nolint:staticcheck
func TestImageFromRegistryWithSecret(t *testing.T) {
	// GCP Artifact Registry also supports auth using username and password, if the username is "_json_key"
	// and the password is the service account JSON blob. See:
	// https://cloud.google.com/artifact-registry/docs/docker/authentication#json-key
	// So we use GCP Artifact Registry to test this too.

	t.Parallel()
	g := gomega.NewWithT(t)

	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	secret, err := modal.SecretFromName(context.Background(), "libmodal-gcp-artifact-registry-test", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"REGISTRY_USERNAME", "REGISTRY_PASSWORD"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromRegistry("us-east1-docker.pkg.dev/modal-prod-367916/private-repo-test/my-image", &modal.ImageFromRegistryOptions{
		Secret: secret,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))
}

//nolint:staticcheck
func TestImageFromAwsEcr(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	secret, err := modal.SecretFromName(context.Background(), "libmodal-aws-ecr-test", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromAwsEcr("459781239556.dkr.ecr.us-east-1.amazonaws.com/ecr-private-registry-test-7522615:python", secret)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))
}

//nolint:staticcheck
func TestImageFromGcpArtifactRegistry(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	secret, err := modal.SecretFromName(context.Background(), "libmodal-gcp-artifact-registry-test", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"SERVICE_ACCOUNT_JSON"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromGcpArtifactRegistry("us-east1-docker.pkg.dev/modal-prod-367916/private-repo-test/my-image", secret)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))
}

func TestCreateOneSandboxTopLevelImageAPI(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := modal.NewImageFromRegistry("alpine:3.21", nil)
	g.Expect(image.ImageId).Should(gomega.BeEmpty())

	sb, err := app.CreateSandbox(image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer sb.Terminate()

	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))
}

func TestCreateOneSandboxTopLevelImageAPISecret(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	secret, err := modal.SecretFromName(context.Background(), "libmodal-gcp-artifact-registry-test", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"REGISTRY_USERNAME", "REGISTRY_PASSWORD"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := modal.NewImageFromRegistry("us-east1-docker.pkg.dev/modal-prod-367916/private-repo-test/my-image", &modal.ImageFromRegistryOptions{
		Secret: secret,
	})
	g.Expect(image.ImageId).Should(gomega.BeEmpty())

	sb, err := app.CreateSandbox(image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer sb.Terminate()

	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))
}

func TestImageFromAwsEcrTopLevel(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	secret, err := modal.SecretFromName(context.Background(), "libmodal-aws-ecr-test", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := modal.NewImageFromAwsEcr("459781239556.dkr.ecr.us-east-1.amazonaws.com/ecr-private-registry-test-7522615:python", secret)
	g.Expect(image.ImageId).Should(gomega.Equal(""))

	sb, err := app.CreateSandbox(image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer sb.Terminate()

	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))
}

func TestImageFromGcpEcrTopLevel(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	secret, err := modal.SecretFromName(context.Background(), "libmodal-gcp-artifact-registry-test", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"SERVICE_ACCOUNT_JSON"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := modal.NewImageFromGcpArtifactRegistry("us-east1-docker.pkg.dev/modal-prod-367916/private-repo-test/my-image", secret)
	g.Expect(image.ImageId).Should(gomega.Equal(""))

	sb, err := app.CreateSandbox(image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer sb.Terminate()

	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))
}

func TestImageDelete(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	ctx := context.Background()

	app, err := modal.AppLookup(ctx, "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := modal.NewImageFromRegistry("alpine:3.13", nil).Build(app)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))

	imageFromId, err := modal.NewImageFromId(ctx, image.ImageId)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(imageFromId.ImageId).Should(gomega.Equal(image.ImageId))

	err = modal.ImageDelete(ctx, image.ImageId, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	_, err = modal.NewImageFromId(ctx, image.ImageId)
	g.Expect(err).Should(gomega.MatchError(gomega.MatchRegexp("Image .+ not found")))

	newImage, err := modal.NewImageFromRegistry("alpine:3.13", nil).Build(app)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(newImage.ImageId).ShouldNot(gomega.Equal(image.ImageId))

	_, err = modal.NewImageFromId(ctx, "im-nonexistent")
	g.Expect(err).Should(gomega.MatchError(gomega.MatchRegexp("Image .+ not found")))
}

func TestDockerfileCommands(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := modal.NewImageFromRegistry("alpine:3.21", nil).DockerfileCommands(
		[]string{"RUN echo hey > /root/hello.txt"},
		nil,
	)

	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command: []string{"cat", "/root/hello.txt"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	stdout, err := io.ReadAll(sb.Stdout)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(string(stdout)).Should(gomega.Equal("hey\n"))

	err = sb.Terminate()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestDockerfileCommandsEmptyArrayNoOp(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	image1 := modal.NewImageFromRegistry("alpine:3.21", nil)
	image2 := image1.DockerfileCommands([]string{}, nil)
	g.Expect(image2).Should(gomega.BeIdenticalTo(image1))
}

func TestDockerfileCommandsChaining(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	secret, err := modal.SecretFromMap(context.Background(), map[string]string{"SECRET": "hello"}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := modal.NewImageFromRegistry("alpine:3.21", nil).
		DockerfileCommands([]string{"RUN echo ${SECRET:-unset} > /root/layer1.txt"}, nil).
		DockerfileCommands([]string{"RUN echo ${SECRET:-unset} > /root/layer2.txt"}, &modal.ImageDockerfileCommandsOptions{
			Secrets: []*modal.Secret{secret},
		}).
		DockerfileCommands([]string{"RUN echo ${SECRET:-unset} > /root/layer3.txt"}, nil)

	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command: []string{
			"cat",
			"/root/layer1.txt",
			"/root/layer2.txt",
			"/root/layer3.txt",
		},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	stdout, err := io.ReadAll(sb.Stdout)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(string(stdout)).Should(gomega.Equal("unset\nhello\nunset\n"))

	err = sb.Terminate()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestDockerfileCommandsCopyCommandValidation(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := modal.NewImageFromRegistry("alpine:3.21", nil)

	validImage := image.DockerfileCommands(
		[]string{"COPY --from=alpine:latest /etc/os-release /tmp/os-release"},
		nil,
	)
	_, err = validImage.Build(app)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	invalidImage := image.DockerfileCommands(
		[]string{"COPY ./file.txt /root/"},
		nil,
	)
	_, err = invalidImage.Build(app)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("COPY commands that copy from local context are not yet supported"))

	runImage := image.DockerfileCommands(
		[]string{"RUN echo 'COPY ./file.txt /root/'"},
		nil,
	)
	_, err = runImage.Build(app)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	multiInvalidImage := image.DockerfileCommands(
		[]string{
			"RUN echo hey",
			"copy ./file.txt /root/",
			"RUN echo hey",
		},
		nil,
	)
	_, err = multiInvalidImage.Build(app)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("COPY commands that copy from local context are not yet supported"))
}

func TestDockerfileCommandsWithOptions(t *testing.T) {
	g := gomega.NewWithT(t)

	mock, cleanup := grpcmock.Install()
	t.Cleanup(cleanup)

	grpcmock.HandleUnary(
		mock, "ImageGetOrCreate",
		func(req *pb.ImageGetOrCreateRequest) (*pb.ImageGetOrCreateResponse, error) {
			g.Expect(req.GetAppId()).To(gomega.Equal("ap-test"))
			image := req.GetImage()
			g.Expect(image.GetDockerfileCommands()).To(gomega.Equal([]string{"FROM alpine:3.21"}))
			g.Expect(image.GetSecretIds()).To(gomega.BeEmpty())
			g.Expect(image.GetBaseImages()).To(gomega.BeEmpty())
			g.Expect(image.GetGpuConfig()).To(gomega.BeNil())
			g.Expect(req.GetForceBuild()).To(gomega.BeFalse())

			return pb.ImageGetOrCreateResponse_builder{
				ImageId: "im-base",
				Result:  pb.GenericResult_builder{Status: pb.GenericResult_GENERIC_STATUS_SUCCESS}.Build(),
			}.Build(), nil
		},
	)

	grpcmock.HandleUnary(
		mock, "ImageGetOrCreate",
		func(req *pb.ImageGetOrCreateRequest) (*pb.ImageGetOrCreateResponse, error) {
			g.Expect(req.GetAppId()).To(gomega.Equal("ap-test"))
			image := req.GetImage()
			g.Expect(image.GetDockerfileCommands()).To(gomega.Equal([]string{"FROM base", "RUN echo layer1"}))
			g.Expect(image.GetSecretIds()).To(gomega.BeEmpty())
			g.Expect(image.GetBaseImages()).To(gomega.HaveLen(1))
			g.Expect(image.GetBaseImages()[0].GetDockerTag()).To(gomega.Equal("base"))
			g.Expect(image.GetBaseImages()[0].GetImageId()).To(gomega.Equal("im-base"))
			g.Expect(image.GetGpuConfig()).To(gomega.BeNil())
			g.Expect(req.GetForceBuild()).To(gomega.BeFalse())

			return pb.ImageGetOrCreateResponse_builder{
				ImageId: "im-layer1",
				Result:  pb.GenericResult_builder{Status: pb.GenericResult_GENERIC_STATUS_SUCCESS}.Build(),
			}.Build(), nil
		},
	)

	grpcmock.HandleUnary(
		mock, "ImageGetOrCreate",
		func(req *pb.ImageGetOrCreateRequest) (*pb.ImageGetOrCreateResponse, error) {
			g.Expect(req.GetAppId()).To(gomega.Equal("ap-test"))
			image := req.GetImage()
			g.Expect(image.GetDockerfileCommands()).To(gomega.Equal([]string{"FROM base", "RUN echo layer2"}))
			g.Expect(image.GetSecretIds()).To(gomega.Equal([]string{"sc-test"}))
			g.Expect(image.GetBaseImages()).To(gomega.HaveLen(1))
			g.Expect(image.GetBaseImages()[0].GetDockerTag()).To(gomega.Equal("base"))
			g.Expect(image.GetBaseImages()[0].GetImageId()).To(gomega.Equal("im-layer1"))
			g.Expect(image.GetGpuConfig()).ToNot(gomega.BeNil())
			g.Expect(image.GetGpuConfig().GetType()).To(gomega.Equal(pb.GPUType(0)))
			g.Expect(image.GetGpuConfig().GetCount()).To(gomega.Equal(uint32(1)))
			g.Expect(image.GetGpuConfig().GetGpuType()).To(gomega.Equal("A100"))
			g.Expect(req.GetForceBuild()).To(gomega.BeTrue())

			return pb.ImageGetOrCreateResponse_builder{
				ImageId: "im-layer2",
				Result:  pb.GenericResult_builder{Status: pb.GenericResult_GENERIC_STATUS_SUCCESS}.Build(),
			}.Build(), nil
		},
	)

	grpcmock.HandleUnary(
		mock, "ImageGetOrCreate",
		func(req *pb.ImageGetOrCreateRequest) (*pb.ImageGetOrCreateResponse, error) {
			g.Expect(req.GetAppId()).To(gomega.Equal("ap-test"))
			image := req.GetImage()
			g.Expect(image.GetDockerfileCommands()).To(gomega.Equal([]string{"FROM base", "RUN echo layer3"}))
			g.Expect(image.GetSecretIds()).To(gomega.BeEmpty())
			g.Expect(image.GetBaseImages()).To(gomega.HaveLen(1))
			g.Expect(image.GetBaseImages()[0].GetDockerTag()).To(gomega.Equal("base"))
			g.Expect(image.GetBaseImages()[0].GetImageId()).To(gomega.Equal("im-layer2"))
			g.Expect(image.GetGpuConfig()).To(gomega.BeNil())
			g.Expect(req.GetForceBuild()).To(gomega.BeTrue())

			return pb.ImageGetOrCreateResponse_builder{
				ImageId: "im-layer3",
				Result:  pb.GenericResult_builder{Status: pb.GenericResult_GENERIC_STATUS_SUCCESS}.Build(),
			}.Build(), nil
		},
	)

	app := &modal.App{AppId: "ap-test"}
	secret := &modal.Secret{SecretId: "sc-test"}

	builtImage, err := modal.NewImageFromRegistry("alpine:3.21", nil).
		DockerfileCommands([]string{"RUN echo layer1"}, nil).
		DockerfileCommands([]string{"RUN echo layer2"}, &modal.ImageDockerfileCommandsOptions{
			Secrets:    []*modal.Secret{secret},
			GPU:        "A100",
			ForceBuild: true,
		}).
		DockerfileCommands([]string{"RUN echo layer3"}, &modal.ImageDockerfileCommandsOptions{
			ForceBuild: true,
		}).
		Build(app)

	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(builtImage.ImageId).To(gomega.Equal("im-layer3"))
}
