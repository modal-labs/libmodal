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

	function, err := modal.FunctionLookup(
		context.Background(),
		"libmodal-test-support", "echo_string", modal.LookupOptions{},
	)
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

	function, err := modal.FunctionLookup(
		context.Background(),
		"libmodal-test-support", "bytelength", modal.LookupOptions{},
	)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	len := 3 * 1000 * 1000 // More than 2 MiB, offload to blob storage
	input := make([]byte, len)
	result, err := function.Remote([]any{input}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal(int64(len)))
}
