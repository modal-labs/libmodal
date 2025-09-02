package test

import (
	"context"
	"testing"

	"github.com/modal-labs/libmodal/modal-go"
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
