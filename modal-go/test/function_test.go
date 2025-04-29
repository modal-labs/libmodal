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

	function, err := modal.FunctionLookup(context.Background(), "libmodal-test", "libmodal_function_test", modal.LookupOptions{Environment: "luis-dev"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Represent Python args and kwargs.
	args := []interface{}{}
	kwargs := map[string]interface{}{
		"s": "hello",
	}
	result, err := function.Remote(context.Background(), args, kwargs)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Parse the result as a string using type assertion.
	resultString, ok := result.(string)
	g.Expect(ok).Should(gomega.BeTrueBecause("output should be a string"))
	g.Expect(resultString).Should(gomega.BeEquivalentTo("output: hello"))

}
