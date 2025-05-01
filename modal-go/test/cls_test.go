package test

import (
	"context"
	"testing"

	"github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
)

func TestClsCall(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cls, err := modal.ClsLookup(
		context.Background(),
		"libmodal-test-support", "EchoCls", modal.LookupOptions{},
	)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Try accessing a non-existent method
	function, err := cls.Method("nonexistent")
	g.Expect(err).Should(gomega.HaveOccurred())

	// Call function
	function, err = cls.Method("echo_string")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	result, err := function.Remote(context.Background(), nil, map[string]any{"s": "hello"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))

	clsParametrized, err := modal.ClsLookup(
		context.Background(),
		"libmodal-test-support", "EchoClsParametrized", modal.LookupOptions{},
	)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// initialize
	instance, err := clsParametrized.Instance(map[string]any{"name": "hello-init"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	function, err = instance.Method("echo_parameter")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	result, err = function.Remote(context.Background(), nil, nil)
	g.Expect(result).Should(gomega.Equal("output: hello-init"))
}
