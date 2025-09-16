package modal

import (
	"context"
	"testing"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"github.com/onsi/gomega"
)

func TestContainerExecProto_WithoutPTY(t *testing.T) {
	g := gomega.NewWithT(t)
	req, err := buildContainerExecRequestProto("task-123", []string{"bash"}, ExecOptions{}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	ptyInfo := req.GetPtyInfo()
	g.Expect(ptyInfo).Should(gomega.BeNil())
}

func TestContainerExecProto_WithPTY(t *testing.T) {
	g := gomega.NewWithT(t)
	req, err := buildContainerExecRequestProto("task-123", []string{"bash"}, ExecOptions{
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

	secretEnvVars := map[string]string{"A": "1"}
	secret, err := SecretFromMap(context.Background(), secretEnvVars, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	envVars := map[string]string{"B": "2"}
	envSecret, err := SecretFromMap(context.Background(), envVars, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	_, err = buildContainerExecRequestProto("ta", []string{"echo", "hello"}, ExecOptions{
		Env: envVars,
	}, nil)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("internal error: Env and envSecret must both be provided or neither be provided"))

	_, err = buildContainerExecRequestProto("ta", []string{"echo", "hello"}, ExecOptions{}, envSecret)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("internal error: Env and envSecret must both be provided or neither be provided"))

	req, err := buildContainerExecRequestProto("ta", []string{"echo", "hello"}, ExecOptions{
		Env:     envVars,
		Secrets: []*Secret{secret},
	}, envSecret)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(req.GetSecretIds()).To(gomega.HaveLen(2))
	g.Expect(req.GetSecretIds()).To(gomega.ContainElement(secret.SecretId))
	g.Expect(req.GetSecretIds()).To(gomega.ContainElement(envSecret.SecretId))
}

func TestContainerExecRequestProto_WithOnlyEnvParameter(t *testing.T) {
	g := gomega.NewWithT(t)

	envVars := map[string]string{"B": "2"}
	envSecret, err := SecretFromMap(context.Background(), envVars, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	req, err := buildContainerExecRequestProto("ta", []string{"echo", "hello"}, ExecOptions{
		Env: envVars,
	}, envSecret)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(req.GetSecretIds()).To(gomega.HaveLen(1))
	g.Expect(req.GetSecretIds()).To(gomega.ContainElement(envSecret.SecretId))
}
