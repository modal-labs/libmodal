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
	ctx := context.Background()

	cls, err := tc.Cls.Lookup(ctx, "libmodal-test-support", "EchoCls", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	instance, err := cls.Instance(ctx, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Try accessing a non-existent method
	_, err = instance.Method("nonexistent")
	g.Expect(err).Should(gomega.BeAssignableToTypeOf(modal.NotFoundError{}))

	function, err := instance.Method("echo_string")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	result, err := function.Remote(ctx, nil, map[string]any{"s": "hello"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))

	cls, err = tc.Cls.Lookup(ctx, "libmodal-test-support", "EchoClsParametrized", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	instance, err = cls.Instance(ctx, map[string]any{"name": "hello-init"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	function, err = instance.Method("echo_parameter")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	result, _ = function.Remote(ctx, nil, nil)
	g.Expect(result).Should(gomega.Equal("output: hello-init"))
}

func TestClsNotFound(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	_, err := tc.Cls.Lookup(context.Background(), "libmodal-test-support", "NotRealClassName", nil)
	g.Expect(err).Should(gomega.BeAssignableToTypeOf(modal.NotFoundError{}))
}

func TestClsCallInputPlane(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	cls, err := tc.Cls.Lookup(ctx, "libmodal-test-support", "EchoClsInputPlane", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	instance, err := cls.Instance(ctx, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	function, err := instance.Method("echo_string")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	result, err := function.Remote(ctx, nil, map[string]any{"s": "hello"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))
}
