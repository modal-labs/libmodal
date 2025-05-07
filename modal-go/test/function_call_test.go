package test

import (
	"context"
	"testing"

	"github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
)

func TestFunctionSpawn(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	function, err := modal.FunctionLookup(
		context.Background(),
		"libmodal-test-support", "echo_string", modal.LookupOptions{},
	)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Call function using spawn
	functionCall, err := function.Spawn(nil, map[string]any{"s": "hello"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Get input later
	result, err := functionCall.Get()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))

	// Create FunctionCall instance and get output again
	functionCall, err = modal.FunctionCallLookup(context.Background(), functionCall.FunctionCallId)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	result, err = functionCall.Get()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))

}
