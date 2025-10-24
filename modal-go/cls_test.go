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

func TestBuildFunctionOptionsProto_WithCPUAndCPUMax(t *testing.T) {
	g := gomega.NewWithT(t)

	cpu := 2.0
	cpuMax := 4.0
	options, err := buildFunctionOptionsProto(&serviceOptions{
		cpu:    &cpu,
		cpuMax: &cpuMax,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(options).ShouldNot(gomega.BeNil())

	resources := options.GetResources()
	g.Expect(resources.GetMilliCpu()).To(gomega.Equal(uint32(2000)))
	g.Expect(resources.GetMilliCpuMax()).To(gomega.Equal(uint32(4000)))
}

func TestBuildFunctionOptionsProto_CPUMaxLowerThanCPU(t *testing.T) {
	g := gomega.NewWithT(t)

	cpu := 4.0
	cpuMax := 2.0
	_, err := buildFunctionOptionsProto(&serviceOptions{
		cpu:    &cpu,
		cpuMax: &cpuMax,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("the CPU request (4.000000) cannot be higher than CPUMax (2.000000)"))
}

func TestBuildFunctionOptionsProto_CPUMaxWithoutCPU(t *testing.T) {
	g := gomega.NewWithT(t)

	cpuMax := 4.0
	_, err := buildFunctionOptionsProto(&serviceOptions{
		cpuMax: &cpuMax,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("must also specify CPU request when CPUMax is specified"))
}

func TestBuildFunctionOptionsProto_WithMemoryAndMemoryMax(t *testing.T) {
	g := gomega.NewWithT(t)

	memory := 1024
	memoryMax := 2048
	options, err := buildFunctionOptionsProto(&serviceOptions{
		memory:    &memory,
		memoryMax: &memoryMax,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(options).ShouldNot(gomega.BeNil())

	resources := options.GetResources()
	g.Expect(resources.GetMemoryMb()).To(gomega.Equal(uint32(1024)))
	g.Expect(resources.GetMemoryMbMax()).To(gomega.Equal(uint32(2048)))
}

func TestBuildFunctionOptionsProto_MemoryMaxLowerThanMemory(t *testing.T) {
	g := gomega.NewWithT(t)

	memory := 2048
	memoryMax := 1024
	_, err := buildFunctionOptionsProto(&serviceOptions{
		memory:    &memory,
		memoryMax: &memoryMax,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("the Memory request (2048) cannot be higher than MemoryMax (1024)"))
}

func TestBuildFunctionOptionsProto_MemoryMaxWithoutMemory(t *testing.T) {
	g := gomega.NewWithT(t)

	memoryMax := 2048
	_, err := buildFunctionOptionsProto(&serviceOptions{
		memoryMax: &memoryMax,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("must also specify Memory request when MemoryMax is specified"))
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

	memory := -100
	_, err := buildFunctionOptionsProto(&serviceOptions{
		memory: &memory,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("must be a positive number"))
}

func TestBuildFunctionOptionsProto_ZeroMemory(t *testing.T) {
	g := gomega.NewWithT(t)

	memory := 0
	_, err := buildFunctionOptionsProto(&serviceOptions{
		memory: &memory,
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("must be a positive number"))
}
