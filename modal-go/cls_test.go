package modal

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestBuildFunctionOptionsProto_NilOptions(t *testing.T) {
	g := gomega.NewWithT(t)

	options, err := buildFunctionOptionsProto(nil, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(options).Should(gomega.BeNil())
}

func TestBuildFunctionOptionsProto_MergesEnvAndSecrets(t *testing.T) {
	g := gomega.NewWithT(t)

	secret := &Secret{SecretID: "test-secret-1"}

	envVars := map[string]string{"B": "2"}
	envSecret := &Secret{SecretID: "test-env-secret"}

	_, err := buildFunctionOptionsProto(&serviceOptions{env: &envVars}, nil)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("internal error: env and envSecret must both be provided or neither be provided"))

	mem := 1 // need to pass a non-empty serviceOptions to pass the !hasParams() check
	_, err = buildFunctionOptionsProto(&serviceOptions{memory: &mem}, envSecret)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("internal error: env and envSecret must both be provided or neither be provided"))

	functionOptions, err := buildFunctionOptionsProto(&serviceOptions{
		env:     &envVars,
		secrets: &[]*Secret{secret},
	}, envSecret)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(functionOptions.GetSecretIds()).To(gomega.HaveLen(2))
	g.Expect(functionOptions.GetSecretIds()).To(gomega.ContainElement(secret.SecretID))
	g.Expect(functionOptions.GetSecretIds()).To(gomega.ContainElement(envSecret.SecretID))
	g.Expect(functionOptions.GetReplaceSecretIds()).To(gomega.BeTrue())
}

func TestBuildFunctionOptionsProto_WithOnlyEnvParameter(t *testing.T) {
	g := gomega.NewWithT(t)

	envVars := map[string]string{"B": "2"}
	envSecret := &Secret{SecretID: "test-env-secret"}

	functionOptions, err := buildFunctionOptionsProto(&serviceOptions{
		env: &envVars,
	}, envSecret)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(functionOptions.GetSecretIds()).To(gomega.HaveLen(1))
	g.Expect(functionOptions.GetSecretIds()).To(gomega.ContainElement(envSecret.SecretID))
	g.Expect(functionOptions.GetReplaceSecretIds()).To(gomega.BeTrue())
}

func TestBuildFunctionOptionsProto_EmptyEnv_WithSecrets(t *testing.T) {
	g := gomega.NewWithT(t)

	secret := &Secret{SecretID: "test-secret-1"}

	params := &serviceOptions{env: &map[string]string{}, secrets: &[]*Secret{secret}}

	functionOptions, err := buildFunctionOptionsProto(params, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(functionOptions.GetSecretIds()).To(gomega.HaveLen(1))
	g.Expect(functionOptions.GetSecretIds()).To(gomega.ContainElement(secret.SecretID))
	g.Expect(functionOptions.GetReplaceSecretIds()).To(gomega.BeTrue())
}

func TestBuildFunctionOptionsProto_EmptyEnv_NoSecrets(t *testing.T) {
	g := gomega.NewWithT(t)

	functionOptions, err := buildFunctionOptionsProto(&serviceOptions{env: &map[string]string{}}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(functionOptions.GetSecretIds()).To(gomega.HaveLen(0))
	g.Expect(functionOptions.GetReplaceSecretIds()).To(gomega.BeFalse())
}
