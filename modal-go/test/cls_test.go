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

	cls, err := modal.ClsLookup(context.Background(), "libmodal-test-support", "EchoCls", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	instance, err := cls.Instance(nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Try accessing a non-existent method
	_, err = instance.Method("nonexistent")
	g.Expect(err).Should(gomega.BeAssignableToTypeOf(modal.NotFoundError{}))

	function, err := instance.Method("echo_string")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	result, err := function.Remote(nil, map[string]any{"s": "hello"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))

	cls, err = modal.ClsLookup(context.Background(), "libmodal-test-support", "EchoClsParametrized", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	instance, err = cls.Instance(map[string]any{"name": "hello-init"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	function, err = instance.Method("echo_parameter")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	result, _ = function.Remote(nil, nil)
	g.Expect(result).Should(gomega.Equal("output: hello-init"))
}

func TestClsNotFound(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	_, err := modal.ClsLookup(context.Background(), "libmodal-test-support", "NotRealClassName", nil)
	g.Expect(err).Should(gomega.BeAssignableToTypeOf(modal.NotFoundError{}))
}
