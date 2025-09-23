package modal

import (
	"testing"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"github.com/onsi/gomega"
)

func TestSandboxCreateRequestProto_WithoutPTY(t *testing.T) {
	g := gomega.NewWithT(t)

	req, err := buildSandboxCreateRequestProto("app-123", "img-456", SandboxCreateParams{}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	definition := req.GetDefinition()
	ptyInfo := definition.GetPtyInfo()
	g.Expect(ptyInfo).Should(gomega.BeNil())
}

func TestSandboxCreateRequestProto_WithPTY(t *testing.T) {
	g := gomega.NewWithT(t)

	req, err := buildSandboxCreateRequestProto("app-123", "img-456", SandboxCreateParams{
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

	secret := &Secret{SecretID: "test-secret-1"}

	envVars := map[string]string{"B": "2"}
	envSecret := &Secret{SecretID: "test-env-secret"}

	_, err := buildSandboxCreateRequestProto("ap", "im", SandboxCreateParams{
		Env: envVars,
	}, nil)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("internal error: Env and envSecret must both be provided or neither be provided"))

	_, err = buildSandboxCreateRequestProto("ap", "im", SandboxCreateParams{}, envSecret)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("internal error: Env and envSecret must both be provided or neither be provided"))

	req, err := buildSandboxCreateRequestProto("ap", "im", SandboxCreateParams{
		Env:     envVars,
		Secrets: []*Secret{secret},
	}, envSecret)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	definition := req.GetDefinition()
	g.Expect(definition.GetSecretIds()).To(gomega.HaveLen(2))
	g.Expect(definition.GetSecretIds()).To(gomega.ContainElement(secret.SecretID))
	g.Expect(definition.GetSecretIds()).To(gomega.ContainElement(envSecret.SecretID))
}

func TestSandboxCreateRequestProto_WithOnlyEnvParameter(t *testing.T) {
	g := gomega.NewWithT(t)

	envVars := map[string]string{"B": "2", "C": "3"}
	envSecret := &Secret{SecretID: "test-env-secret"}

	req, err := buildSandboxCreateRequestProto("ap", "im", SandboxCreateParams{
		Env: envVars,
	}, envSecret)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	definition := req.GetDefinition()
	g.Expect(definition.GetSecretIds()).To(gomega.HaveLen(1))
	g.Expect(definition.GetSecretIds()).To(gomega.ContainElement(envSecret.SecretID))
}

func TestContainerExecProto_WithoutPTY(t *testing.T) {
	g := gomega.NewWithT(t)
	req, err := buildContainerExecRequestProto("task-123", []string{"bash"}, SandboxExecParams{}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	ptyInfo := req.GetPtyInfo()
	g.Expect(ptyInfo).Should(gomega.BeNil())
}

func TestContainerExecProto_WithPTY(t *testing.T) {
	g := gomega.NewWithT(t)
	req, err := buildContainerExecRequestProto("task-123", []string{"bash"}, SandboxExecParams{
		PTY: true,
	}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	ptyInfo := req.GetPtyInfo()
	g.Expect(ptyInfo).ShouldNot(gomega.BeNil())
	g.Expect(ptyInfo.GetEnabled()).To(gomega.BeTrue())
	g.Expect(ptyInfo.GetWinszRows()).To(gomega.Equal(uint32(24)))
	g.Expect(ptyInfo.GetWinszCols()).To(gomega.Equal(uint32(80)))
	g.Expect(ptyInfo.GetEnvTerm()).To(gomega.Equal("xterm-256color"))
	g.Expect(ptyInfo.GetEnvColorterm()).To(gomega.Equal("truecolor"))
	g.Expect(ptyInfo.GetPtyType()).To(gomega.Equal(pb.PTYInfo_PTY_TYPE_SHELL))
	g.Expect(ptyInfo.GetNoTerminateOnIdleStdin()).To(gomega.BeTrue())
}

func TestContainerExecRequestProto_MergesEnvAndSecrets(t *testing.T) {
	g := gomega.NewWithT(t)

	secret := &Secret{SecretID: "test-secret-1"}

	envVars := map[string]string{"B": "2"}
	envSecret := &Secret{SecretID: "test-env-secret"}

	_, err := buildContainerExecRequestProto("ta", []string{"echo", "hello"}, SandboxExecParams{
		Env: envVars,
	}, nil)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("internal error: Env and envSecret must both be provided or neither be provided"))

	_, err = buildContainerExecRequestProto("ta", []string{"echo", "hello"}, SandboxExecParams{}, envSecret)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("internal error: Env and envSecret must both be provided or neither be provided"))

	req, err := buildContainerExecRequestProto("ta", []string{"echo", "hello"}, SandboxExecParams{
		Env:     envVars,
		Secrets: []*Secret{secret},
	}, envSecret)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(req.GetSecretIds()).To(gomega.HaveLen(2))
	g.Expect(req.GetSecretIds()).To(gomega.ContainElement(secret.SecretID))
	g.Expect(req.GetSecretIds()).To(gomega.ContainElement(envSecret.SecretID))
}

func TestContainerExecRequestProto_WithOnlyEnvParameter(t *testing.T) {
	g := gomega.NewWithT(t)

	envVars := map[string]string{"B": "2"}
	envSecret := &Secret{SecretID: "test-env-secret"}

	req, err := buildContainerExecRequestProto("ta", []string{"echo", "hello"}, SandboxExecParams{
		Env: envVars,
	}, envSecret)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(req.GetSecretIds()).To(gomega.HaveLen(1))
	g.Expect(req.GetSecretIds()).To(gomega.ContainElement(envSecret.SecretID))
}
