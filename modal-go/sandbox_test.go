package modal

import (
	"testing"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"github.com/onsi/gomega"
)

func TestSandboxCreateRequestProto_WithoutPTY(t *testing.T) {
	g := gomega.NewWithT(t)

	req, err := buildSandboxCreateRequestProto("app-123", "img-456", SandboxCreateParams{})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	definition := req.GetDefinition()
	ptyInfo := definition.GetPtyInfo()
	g.Expect(ptyInfo).Should(gomega.BeNil())
}

func TestSandboxCreateRequestProto_WithPTY(t *testing.T) {
	g := gomega.NewWithT(t)

	req, err := buildSandboxCreateRequestProto("app-123", "img-456", SandboxCreateParams{
		PTY: true,
	})
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

func TestContainerExecProto_WithoutPTY(t *testing.T) {
	g := gomega.NewWithT(t)
	req, err := buildContainerExecRequestProto("task-123", []string{"bash"}, SandboxExecParams{})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	ptyInfo := req.GetPtyInfo()
	g.Expect(ptyInfo).Should(gomega.BeNil())
}

func TestContainerExecProto_WithPTY(t *testing.T) {
	g := gomega.NewWithT(t)
	req, err := buildContainerExecRequestProto("task-123", []string{"bash"}, SandboxExecParams{
		PTY: true,
	})
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
