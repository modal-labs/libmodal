package test

import (
	"context"
	"io"
	"testing"

	"github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
)

func TestSecretFromName(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	secret, err := tc.Secrets.FromName(ctx, "libmodal-test-secret", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(secret.SecretId).Should(gomega.HavePrefix("st-"))
	g.Expect(secret.Name).To(gomega.Equal("libmodal-test-secret"))

	_, err = tc.Secrets.FromName(ctx, "missing-secret", nil)
	g.Expect(err).Should(gomega.MatchError(gomega.ContainSubstring("Secret 'missing-secret' not found")))
}

func TestSecretFromNameWithRequiredKeys(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	secret, err := tc.Secrets.FromName(ctx, "libmodal-test-secret", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"a", "b", "c"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(secret.SecretId).Should(gomega.HavePrefix("st-"))

	_, err = tc.Secrets.FromName(ctx, "libmodal-test-secret", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"a", "b", "c", "missing-key"},
	})
	g.Expect(err).Should(gomega.MatchError(gomega.ContainSubstring("Secret is missing key(s): missing-key")))
}

func TestSecretFromMap(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	secret, err := tc.Secrets.FromMap(ctx, map[string]string{"key": "value"}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(secret.SecretId).Should(gomega.HavePrefix("st-"))

	sb, err := tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateOptions{Secrets: []*modal.Secret{secret}, Command: []string{"printenv", "key"}})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	output, err := io.ReadAll(sb.Stdout)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(string(output)).To(gomega.Equal("value\n"))
}
