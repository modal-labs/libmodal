package modal

import (
	"context"
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

	secretEnvVars := map[string]string{"A": "1"}
	secret, err := SecretFromMap(context.Background(), secretEnvVars, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	envVars := map[string]string{"B": "2"}
	envSecret, err := SecretFromMap(context.Background(), envVars, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	_, err = buildFunctionOptionsProto(&serviceOptions{env: &envVars}, nil)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("internal error: env and envSecret must both be provided or neither be provided"))

	mem := 1 // need to pass a non-empty serviceOptions to pass the !hasOptions(opts) check
	_, err = buildFunctionOptionsProto(&serviceOptions{memory: &mem}, envSecret)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("internal error: env and envSecret must both be provided or neither be provided"))

	functionOptions, err := buildFunctionOptionsProto(&serviceOptions{
		env:     &envVars,
		secrets: &[]*Secret{secret},
	}, envSecret)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(functionOptions.GetSecretIds()).To(gomega.HaveLen(2))
	g.Expect(functionOptions.GetSecretIds()).To(gomega.ContainElement(secret.SecretId))
	g.Expect(functionOptions.GetSecretIds()).To(gomega.ContainElement(envSecret.SecretId))
	g.Expect(functionOptions.GetReplaceSecretIds()).To(gomega.BeTrue())
}

func TestBuildFunctionOptionsProto_WithOnlyEnvParameter(t *testing.T) {
	g := gomega.NewWithT(t)

	envVars := map[string]string{"B": "2"}
	envSecret, err := SecretFromMap(context.Background(), envVars, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	functionOptions, err := buildFunctionOptionsProto(&serviceOptions{
		env: &envVars,
	}, envSecret)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(functionOptions.GetSecretIds()).To(gomega.HaveLen(1))
	g.Expect(functionOptions.GetSecretIds()).To(gomega.ContainElement(envSecret.SecretId))
	g.Expect(functionOptions.GetReplaceSecretIds()).To(gomega.BeTrue())
}

func TestBuildFunctionOptionsProto_EmptyEnv_WithSecrets(t *testing.T) {
	g := gomega.NewWithT(t)

	secret, err := SecretFromMap(context.Background(), map[string]string{"A": "1"}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	opts := &serviceOptions{env: &map[string]string{}, secrets: &[]*Secret{secret}}

	functionOptions, err := buildFunctionOptionsProto(opts, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(functionOptions.GetSecretIds()).To(gomega.HaveLen(1))
	g.Expect(functionOptions.GetSecretIds()).To(gomega.ContainElement(secret.SecretId))
	g.Expect(functionOptions.GetReplaceSecretIds()).To(gomega.BeTrue())
}

func TestBuildFunctionOptionsProto_EmptyEnv_NoSecrets(t *testing.T) {
	g := gomega.NewWithT(t)

	functionOptions, err := buildFunctionOptionsProto(&serviceOptions{env: &map[string]string{}}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(functionOptions.GetSecretIds()).To(gomega.HaveLen(0))
	g.Expect(functionOptions.GetReplaceSecretIds()).To(gomega.BeFalse())
}
