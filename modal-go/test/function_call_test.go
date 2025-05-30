package test

import (
	"context"
	"testing"
	"time"

	pickle "github.com/kisielk/og-rek"
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
	sleep, err := modal.FunctionLookup(
		context.Background(),
		"libmodal-test-support", "sleep", modal.LookupOptions{},
	)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	functionCall, err = sleep.Spawn(nil, map[string]any{"t": 5})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Cancel function call.
	err = functionCall.Cancel(modal.FunctionCallCancelOptions{})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Attempting to get outputs for a cancelled function call
	// is expected to return an error.
	_, err = functionCall.Get(modal.FunctionCallGetOptions{})
	g.Expect(err).Should(gomega.HaveOccurred())

	// Spawn function with long running input.
	functionCall, err = sleep.Spawn(nil, map[string]any{"t": 5})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Get is now expected to timeout.
	timeout := 1 * time.Second
	_, err = functionCall.Get(modal.FunctionCallGetOptions{Timeout: &timeout})
	g.Expect(err).Should(gomega.HaveOccurred())
}

func TestFunctionCallGet0(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	sleep, _ := modal.FunctionLookup(
		context.Background(),
		"libmodal-test-support", "sleep", modal.LookupOptions{},
	)

	functionCall, err := sleep.Spawn([]any{0.5}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Polling for output with timeout 0 should raise an error, since the
	// function call has not finished yet.
	timeout := 0 * time.Second
	_, err = functionCall.Get(modal.FunctionCallGetOptions{Timeout: &timeout})
	g.Expect(err).Should(gomega.HaveOccurred())

	// Wait for the function call to finish.
	result, err := functionCall.Get(modal.FunctionCallGetOptions{})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal(pickle.None{}))

	// Now we can get the result.
	result, err = functionCall.Get(modal.FunctionCallGetOptions{Timeout: &timeout})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal(pickle.None{}))
}
