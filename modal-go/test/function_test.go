package test

import (
	"context"
	"testing"

	"github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
)

func TestFunctionCall(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	function, err := modal.FunctionLookup(context.Background(), "libmodal-test-support", "echo_string", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Represent Python kwargs.
	result, err := function.Remote(nil, map[string]any{"s": "hello"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))

	// Try the same, but with args.
	result, err = function.Remote([]any{"hello"}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))
}

func TestFunctionCallLargeInput(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	function, err := modal.FunctionLookup(context.Background(), "libmodal-test-support", "bytelength", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	len := 3 * 1000 * 1000 // More than 2 MiB, offload to blob storage
	input := make([]byte, len)
	result, err := function.Remote([]any{input}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal(int64(len)))
}

func TestFunctionNotFound(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	_, err := modal.FunctionLookup(context.Background(), "libmodal-test-support", "not_a_real_function", nil)
	g.Expect(err).Should(gomega.BeAssignableToTypeOf(modal.NotFoundError{}))
}

func TestFunctionCallInputPlane(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	function, err := modal.FunctionLookup(context.Background(), "libmodal-test-support", "input_plane", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Try the same, but with args.
	result, err := function.Remote([]any{"hello"}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))
}

func TestFunctionGetCurrentStats(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	function, err := modal.FunctionLookup(context.Background(), "libmodal-test-support", "echo_string", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	stats, err := function.GetCurrentStats()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(stats.Backlog).Should(gomega.BeNumerically(">=", 0))
	g.Expect(stats.NumTotalRunners).Should(gomega.BeNumerically(">=", 0))
}

func TestFunctionUpdateAutoscaler(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	function, err := modal.FunctionLookup(context.Background(), "libmodal-test-support", "echo_string", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Test updating various autoscaler settings - should not error
	minContainers := uint32(1)
	maxContainers := uint32(10)
	bufferContainers := uint32(2)
	scaledownWindow := uint32(300)

	err = function.UpdateAutoscaler(modal.UpdateAutoscalerOptions{
		MinContainers:    &minContainers,
		MaxContainers:    &maxContainers,
		BufferContainers: &bufferContainers,
		ScaledownWindow:  &scaledownWindow,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Test partial updates
	minContainers2 := uint32(2)
	err = function.UpdateAutoscaler(modal.UpdateAutoscalerOptions{
		MinContainers: &minContainers2,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	scaledownWindow2 := uint32(600)
	err = function.UpdateAutoscaler(modal.UpdateAutoscalerOptions{
		ScaledownWindow: &scaledownWindow2,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}
