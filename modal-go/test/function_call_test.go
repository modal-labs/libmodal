package test

import (
	"context"
	"testing"
	"time"

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

	// Call function using spawn.
	functionCall, err := function.Spawn(nil, map[string]any{"s": "hello"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Get outputs.
	result, err := functionCall.Get(modal.FunctionCallGetOptions{})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))

	// Create FunctionCall instance and get output again.
	functionCall, err = modal.FunctionCallFromId(context.Background(), functionCall.FunctionCallId)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	result, err = functionCall.Get(modal.FunctionCallGetOptions{})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))

	// Looking function that takes a long time to complete.
	functionSleep, err := modal.FunctionLookup(
		context.Background(),
		"libmodal-test-support", "sleep", modal.LookupOptions{},
	)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	functionCall, err = functionSleep.Spawn(nil, map[string]any{"t": 5})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Cancel function call.
	err = functionCall.Cancel(modal.FunctionCallCancelOptions{})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Attempting to get outputs for a cancelled function call
	// is expected to return an error.
	_, err = functionCall.Get(modal.FunctionCallGetOptions{})
	g.Expect(err).Should(gomega.HaveOccurred())

	// Spawn function with long running input.
	functionCall, err = functionSleep.Spawn(nil, map[string]any{"t": 5})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Get is now expected to timeout.
	_, err = functionCall.Get(modal.FunctionCallGetOptions{Timeout: 1 * time.Second})
	g.Expect(err).Should(gomega.HaveOccurred())
}
