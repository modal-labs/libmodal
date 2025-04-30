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
		"libmodal-test-support", "EchoCls", modal.LookupOptions{
			Environment: "libmodal",
		},
	)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Call function
	function, ok := cls.Methods["echo_string"]
	g.Expect(ok).Should(gomega.BeTrue())

	result, err := function.Remote(context.Background(), nil, map[string]any{"s": "hello"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))

}
