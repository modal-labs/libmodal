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

	app, err := tc.Apps.Lookup(ctx, "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := tc.Images.FromRegistry("alpine:3.21", nil).Build(ctx, app)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	imageFromId, err := tc.Images.FromId(ctx, image.ImageId)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(imageFromId.ImageId).Should(gomega.Equal(image.ImageId))

	_, err = tc.Images.FromId(ctx, "im-nonexistent")
	g.Expect(err).Should(gomega.HaveOccurred())
}

//nolint:staticcheck
func TestImageFromRegistry(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.Lookup(ctx, "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := tc.Images.FromRegistry("alpine:3.21", nil).Build(ctx, app)
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
	ctx := context.Background()

	app, err := tc.Apps.Lookup(ctx, "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	secret, err := tc.Secrets.FromName(ctx, "libmodal-gcp-artifact-registry-test", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"REGISTRY_USERNAME", "REGISTRY_PASSWORD"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := tc.Images.FromRegistry("us-east1-docker.pkg.dev/modal-prod-367916/private-repo-test/my-image", &modal.ImageFromRegistryOptions{
		Secret: secret,
	}).Build(ctx, app)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))
}

//nolint:staticcheck
func TestImageFromAwsEcr(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.Lookup(ctx, "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	secret, err := tc.Secrets.FromName(ctx, "libmodal-aws-ecr-test", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := tc.Images.FromAwsEcr("459781239556.dkr.ecr.us-east-1.amazonaws.com/ecr-private-registry-test-7522615:python", secret).Build(ctx, app)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))
}

//nolint:staticcheck
func TestImageFromGcpArtifactRegistry(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.Lookup(ctx, "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	secret, err := tc.Secrets.FromName(ctx, "libmodal-gcp-artifact-registry-test", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"SERVICE_ACCOUNT_JSON"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := tc.Images.FromGcpArtifactRegistry("us-east1-docker.pkg.dev/modal-prod-367916/private-repo-test/my-image", secret).Build(ctx, app)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))
}

func TestCreateOneSandboxTopLevelImageAPI(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.Lookup(ctx, "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)
	g.Expect(image.ImageId).Should(gomega.BeEmpty())

	sb, err := tc.Sandboxes.Create(ctx, app, image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer sb.Terminate(ctx)

	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))
}

func TestCreateOneSandboxTopLevelImageAPISecret(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.Lookup(ctx, "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	secret, err := tc.Secrets.FromName(ctx, "libmodal-gcp-artifact-registry-test", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"REGISTRY_USERNAME", "REGISTRY_PASSWORD"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("us-east1-docker.pkg.dev/modal-prod-367916/private-repo-test/my-image", &modal.ImageFromRegistryOptions{
		Secret: secret,
	})
	g.Expect(image.ImageId).Should(gomega.BeEmpty())

	sb, err := tc.Sandboxes.Create(ctx, app, image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer sb.Terminate(ctx)

	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))
}

func TestImageFromAwsEcrTopLevel(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.Lookup(ctx, "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	secret, err := tc.Secrets.FromName(ctx, "libmodal-aws-ecr-test", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromAwsEcr("459781239556.dkr.ecr.us-east-1.amazonaws.com/ecr-private-registry-test-7522615:python", secret)
	g.Expect(image.ImageId).Should(gomega.Equal(""))

	sb, err := tc.Sandboxes.Create(ctx, app, image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer sb.Terminate(ctx)

	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))
}

func TestImageFromGcpEcrTopLevel(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.Lookup(ctx, "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	secret, err := tc.Secrets.FromName(ctx, "libmodal-gcp-artifact-registry-test", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"SERVICE_ACCOUNT_JSON"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromGcpArtifactRegistry("us-east1-docker.pkg.dev/modal-prod-367916/private-repo-test/my-image", secret)
	g.Expect(image.ImageId).Should(gomega.Equal(""))

	sb, err := tc.Sandboxes.Create(ctx, app, image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer sb.Terminate(ctx)

	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))
}

func TestImageDelete(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.Lookup(ctx, "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := tc.Images.FromRegistry("alpine:3.13", nil).Build(ctx, app)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(image.ImageId).Should(gomega.HavePrefix("im-"))

	imageFromId, err := tc.Images.FromId(ctx, image.ImageId)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(imageFromId.ImageId).Should(gomega.Equal(image.ImageId))

	err = tc.Images.Delete(ctx, image.ImageId, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	_, err = tc.Images.FromId(ctx, image.ImageId)
	g.Expect(err).Should(gomega.MatchError(gomega.MatchRegexp("Image .+ not found")))

	newImage, err := tc.Images.FromRegistry("alpine:3.13", nil).Build(ctx, app)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(newImage.ImageId).ShouldNot(gomega.Equal(image.ImageId))

	_, err = tc.Images.FromId(ctx, "im-nonexistent")
	g.Expect(err).Should(gomega.MatchError(gomega.MatchRegexp("Image .+ not found")))
}
