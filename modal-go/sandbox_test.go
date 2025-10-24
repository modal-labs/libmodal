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

func TestSandboxCreateRequestProto_WithCPUAndCPUMax(t *testing.T) {
	g := gomega.NewWithT(t)

	req, err := buildSandboxCreateRequestProto("app-123", "img-456", SandboxCreateParams{
		CPU:    2.0,
		CPUMax: 4.0,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	resources := req.GetDefinition().GetResources()
	g.Expect(resources.GetMilliCpu()).To(gomega.Equal(uint32(2000)))
	g.Expect(resources.GetMilliCpuMax()).To(gomega.Equal(uint32(4000)))
}

func TestSandboxCreateRequestProto_CPUMaxLowerThanCPU(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := buildSandboxCreateRequestProto("app-123", "img-456", SandboxCreateParams{
		CPU:    4.0,
		CPUMax: 2.0,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("the CPU request (4.000000) cannot be higher than CPUMax (2.000000)"))
}

func TestSandboxCreateRequestProto_CPUMaxWithoutCPU(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := buildSandboxCreateRequestProto("app-123", "img-456", SandboxCreateParams{
		CPUMax: 4.0,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("must also specify CPU request when CPUMax is specified"))
}

func TestSandboxCreateRequestProto_WithMemoryAndMemoryMax(t *testing.T) {
	g := gomega.NewWithT(t)

	req, err := buildSandboxCreateRequestProto("app-123", "img-456", SandboxCreateParams{
		Memory:    1024,
		MemoryMax: 2048,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	resources := req.GetDefinition().GetResources()
	g.Expect(resources.GetMemoryMb()).To(gomega.Equal(uint32(1024)))
	g.Expect(resources.GetMemoryMbMax()).To(gomega.Equal(uint32(2048)))
}

func TestSandboxCreateRequestProto_MemoryMaxLowerThanMemory(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := buildSandboxCreateRequestProto("app-123", "img-456", SandboxCreateParams{
		Memory:    2048,
		MemoryMax: 1024,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("the Memory request (2048) cannot be higher than MemoryMax (1024)"))
}

func TestSandboxCreateRequestProto_MemoryMaxWithoutMemory(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := buildSandboxCreateRequestProto("app-123", "img-456", SandboxCreateParams{
		MemoryMax: 2048,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("must also specify Memory request when MemoryMax is specified"))
}

func TestSandboxCreateRequestProto_NegativeCPU(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := buildSandboxCreateRequestProto("app-123", "img-456", SandboxCreateParams{
		CPU: -1.0,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("must be a positive number"))
}

func TestSandboxCreateRequestProto_NegativeMemory(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := buildSandboxCreateRequestProto("app-123", "img-456", SandboxCreateParams{
		Memory: -100,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("must be a positive number"))
}
