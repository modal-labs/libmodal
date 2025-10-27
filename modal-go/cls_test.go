package modal

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestBuildFunctionOptionsProto_NilOptions(t *testing.T) {
	g := gomega.NewWithT(t)

	options, err := buildFunctionOptionsProto(nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(options).Should(gomega.BeNil())
}

func TestBuildFunctionOptionsProto_WithCPUAndCPULimit(t *testing.T) {
	g := gomega.NewWithT(t)

	cpu := 2.0
	cpuLimit := 4.5
	options, err := buildFunctionOptionsProto(&serviceOptions{
		cpu:      &cpu,
		cpuLimit: &cpuLimit,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(options).ShouldNot(gomega.BeNil())

	resources := options.GetResources()
	g.Expect(resources.GetMilliCpu()).To(gomega.Equal(uint32(2000)))
	g.Expect(resources.GetMilliCpuMax()).To(gomega.Equal(uint32(4500)))
}

func TestBuildFunctionOptionsProto_CPULimitLowerThanCPU(t *testing.T) {
	g := gomega.NewWithT(t)

	cpu := 4.0
	cpuLimit := 2.0
	_, err := buildFunctionOptionsProto(&serviceOptions{
		cpu:      &cpu,
		cpuLimit: &cpuLimit,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("the CPU request (4.000000) cannot be higher than CPULimit (2.000000)"))
}

func TestBuildFunctionOptionsProto_CPULimitWithoutCPU(t *testing.T) {
	g := gomega.NewWithT(t)

	cpuLimit := 4.0
	_, err := buildFunctionOptionsProto(&serviceOptions{
		cpuLimit: &cpuLimit,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("must also specify non-zero CPU request when CPULimit is specified"))
}

func TestBuildFunctionOptionsProto_WithMemoryAndMemoryLimit(t *testing.T) {
	g := gomega.NewWithT(t)

	memoryMib := 1024
	memoryLimitMib := 2048
	options, err := buildFunctionOptionsProto(&serviceOptions{
		memoryMib:      &memoryMib,
		memoryLimitMib: &memoryLimitMib,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(options).ShouldNot(gomega.BeNil())

	resources := options.GetResources()
	g.Expect(resources.GetMemoryMb()).To(gomega.Equal(uint32(1024)))
	g.Expect(resources.GetMemoryMbMax()).To(gomega.Equal(uint32(2048)))
}

func TestBuildFunctionOptionsProto_MemoryLimitLowerThanMemory(t *testing.T) {
	g := gomega.NewWithT(t)

	memoryMib := 2048
	memoryLimitMib := 1024
	_, err := buildFunctionOptionsProto(&serviceOptions{
		memoryMib:      &memoryMib,
		memoryLimitMib: &memoryLimitMib,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("the MemoryMib request (2048) cannot be higher than MemoryLimitMib (1024)"))
}

func TestBuildFunctionOptionsProto_MemoryLimitWithoutMemory(t *testing.T) {
	g := gomega.NewWithT(t)

	memoryLimitMib := 2048
	_, err := buildFunctionOptionsProto(&serviceOptions{
		memoryLimitMib: &memoryLimitMib,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("must also specify non-zero MemoryMib request when MemoryLimitMib is specified"))
}

func TestBuildFunctionOptionsProto_NegativeCPU(t *testing.T) {
	g := gomega.NewWithT(t)

	cpu := -1.0
	_, err := buildFunctionOptionsProto(&serviceOptions{
		cpu: &cpu,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("must be a positive number"))
}

func TestBuildFunctionOptionsProto_ZeroCPU(t *testing.T) {
	g := gomega.NewWithT(t)

	cpu := 0.0
	_, err := buildFunctionOptionsProto(&serviceOptions{
		cpu: &cpu,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("must be a positive number"))
}

func TestBuildFunctionOptionsProto_NegativeMemory(t *testing.T) {
	g := gomega.NewWithT(t)

	memoryMib := -100
	_, err := buildFunctionOptionsProto(&serviceOptions{
		memoryMib: &memoryMib,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("must be a positive number"))
}

func TestBuildFunctionOptionsProto_ZeroMemory(t *testing.T) {
	g := gomega.NewWithT(t)

	memoryMib := 0
	_, err := buildFunctionOptionsProto(&serviceOptions{
		memoryMib: &memoryMib,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("must be a positive number"))
}
