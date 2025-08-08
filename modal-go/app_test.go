package modal

import (
	"testing"

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
