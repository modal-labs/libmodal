package test

import (
	"context"
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
