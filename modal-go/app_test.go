package modal

import (
	"context"
	"testing"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"github.com/onsi/gomega"
)

func TestParseGPUConfig(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Test empty string returns nil
	config, err := parseGPUConfig("")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(config).Should(gomega.BeNil())

	// Test single GPU type
	config, err = parseGPUConfig("T4")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(config.GetCount()).To(gomega.Equal(uint32(1)))
	g.Expect(config.GetGpuType()).To(gomega.Equal("T4"))

	config, err = parseGPUConfig("A10G")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(config.GetCount()).To(gomega.Equal(uint32(1)))
	g.Expect(config.GetGpuType()).To(gomega.Equal("A10G"))

	config, err = parseGPUConfig("A100-80GB")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(config.GetCount()).To(gomega.Equal(uint32(1)))
	g.Expect(config.GetGpuType()).To(gomega.Equal("A100-80GB"))

	// Test GPU type with count
	config, err = parseGPUConfig("A100-80GB:3")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(config.GetCount()).To(gomega.Equal(uint32(3)))
	g.Expect(config.GetGpuType()).To(gomega.Equal("A100-80GB"))

	config, err = parseGPUConfig("T4:2")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(config.GetCount()).To(gomega.Equal(uint32(2)))
	g.Expect(config.GetGpuType()).To(gomega.Equal("T4"))

	// Test lowercase conversion
	config, err = parseGPUConfig("a100:4")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(config.GetCount()).To(gomega.Equal(uint32(4)))
	g.Expect(config.GetGpuType()).To(gomega.Equal("A100"))

	// Test invalid count formats
	_, err = parseGPUConfig("T4:invalid")
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("invalid GPU count: invalid"))

	_, err = parseGPUConfig("T4:")
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("invalid GPU count: "))

	_, err = parseGPUConfig("T4:0")
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("invalid GPU count: 0"))

	_, err = parseGPUConfig("T4:-1")
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("invalid GPU count: -1"))
}

func TestSandboxCreateRequestProto_WithoutPTY(t *testing.T) {
	g := gomega.NewWithT(t)

	req, err := buildSandboxCreateRequestProto("app-123", "img-456", SandboxOptions{}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	definition := req.GetDefinition()
	ptyInfo := definition.GetPtyInfo()
	g.Expect(ptyInfo).Should(gomega.BeNil())
}

func TestSandboxCreateRequestProto_WithPTY(t *testing.T) {
	g := gomega.NewWithT(t)

	req, err := buildSandboxCreateRequestProto("app-123", "img-456", SandboxOptions{
		PTY: true,
	}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	definition := req.GetDefinition()
	ptyInfo := definition.GetPtyInfo()
	g.Expect(ptyInfo.GetEnabled()).To(gomega.BeTrue())
	g.Expect(ptyInfo.GetWinszRows()).To(gomega.Equal(uint32(24)))
	g.Expect(ptyInfo.GetWinszCols()).To(gomega.Equal(uint32(80)))
	g.Expect(ptyInfo.GetEnvTerm()).To(gomega.Equal("xterm-256color"))
	g.Expect(ptyInfo.GetEnvColorterm()).To(gomega.Equal("truecolor"))
	g.Expect(ptyInfo.GetPtyType()).To(gomega.Equal(pb.PTYInfo_PTY_TYPE_SHELL))
}

func TestSandboxCreateRequestProto_MergesEnvAndSecrets(t *testing.T) {
	g := gomega.NewWithT(t)

	secretEnvVars := map[string]string{"A": "1"}
	secret, err := SecretFromMap(context.Background(), secretEnvVars, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	envVars := map[string]string{"B": "2"}
	envSecret, err := SecretFromMap(context.Background(), envVars, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	_, err = buildSandboxCreateRequestProto("ap", "im", SandboxOptions{
		Env: envVars,
	}, nil)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("internal error: Env and envSecret must both be provided or neither be provided"))

	_, err = buildSandboxCreateRequestProto("ap", "im", SandboxOptions{}, envSecret)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("internal error: Env and envSecret must both be provided or neither be provided"))

	req, err := buildSandboxCreateRequestProto("ap", "im", SandboxOptions{
		Env:     envVars,
		Secrets: []*Secret{secret},
	}, envSecret)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	definition := req.GetDefinition()
	g.Expect(definition.GetSecretIds()).To(gomega.HaveLen(2))
	g.Expect(definition.GetSecretIds()).To(gomega.ContainElement(secret.SecretId))
	g.Expect(definition.GetSecretIds()).To(gomega.ContainElement(envSecret.SecretId))
}

func TestSandboxCreateRequestProto_WithOnlyEnvParameter(t *testing.T) {
	g := gomega.NewWithT(t)

	envVars := map[string]string{"B": "2", "C": "3"}
	envSecret, err := SecretFromMap(context.Background(), envVars, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	req, err := buildSandboxCreateRequestProto("ap", "im", SandboxOptions{
		Env: envVars,
	}, envSecret)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	definition := req.GetDefinition()
	g.Expect(definition.GetSecretIds()).To(gomega.HaveLen(1))
	g.Expect(definition.GetSecretIds()).To(gomega.ContainElement(envSecret.SecretId))
}
