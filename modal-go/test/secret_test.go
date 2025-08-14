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
	secret, err := modal.SecretFromName(context.Background(), "libmodal-test-secret", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(secret.SecretId).Should(gomega.HavePrefix("st-"))

	_, err = modal.SecretFromName(context.Background(), "missing-secret", nil)
	g.Expect(err).Should(gomega.MatchError(gomega.ContainSubstring("Secret 'missing-secret' not found")))
}

func TestSecretFromNameWithRequiredKeys(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	secret, err := modal.SecretFromName(context.Background(), "libmodal-test-secret", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"a", "b", "c"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(secret.SecretId).Should(gomega.HavePrefix("st-"))

	_, err = modal.SecretFromName(context.Background(), "libmodal-test-secret", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"a", "b", "c", "missing-key"},
	})
	g.Expect(err).Should(gomega.MatchError(gomega.ContainSubstring("Secret is missing key(s): missing-key")))
}

func TestSecretFromMap(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := modal.NewImageFromRegistry("alpine:3.21", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	secret, err := modal.SecretFromMap(context.Background(), map[string]string{"key": "value"}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(secret.SecretId).Should(gomega.HavePrefix("st-"))

	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{Secrets: []*modal.Secret{secret}, Command: []string{"printenv", "key"}})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	output, err := io.ReadAll(sb.Stdout)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(string(output)).To(gomega.Equal("value\n"))
}
