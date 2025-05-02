package test

import (
	"context"
	"testing"

	"github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

func TestClsCall(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cls, err := modal.ClsLookup(
		context.Background(),
		"libmodal-test-support", "EchoCls", modal.LookupOptions{},
	)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	instance, err := cls.Instance(nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Try accessing a non-existent method
	function, err := cls.Method("nonexistent")
	g.Expect(err).Should(gomega.HaveOccurred())

	function, err = cls.Method("echo_string")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	result, err := function.Remote(context.Background(), nil, map[string]any{"s": "hello"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))

	cls, err = modal.ClsLookup(
		context.Background(),
		"libmodal-test-support", "EchoClsParametrized", modal.LookupOptions{},
	)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	instance, err = cls.Instance(map[string]any{"name": "hello-init"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	function, err = instance.Method("echo_parameter")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	result, err = function.Remote(context.Background(), nil, nil)
	g.Expect(result).Should(gomega.Equal("output: hello-init"))
}

type Parameter struct {
	Name  string
	Value interface{}
	Type  pb.ParameterType
}

func serializeParameterSet(params []Parameter) ([]byte, error) {
	paramSet := pb.ClassParameterSet_builder{
		Parameters: []*pb.ClassParameterValue{},
	}.Build()

	for _, param := range params {
		serializedParam, err := modal.EncodeParameterValue(param.Name, param.Value, param.Type)
		if err != nil {
			return nil, err
		}

		newParameters := append(paramSet.GetParameters(), serializedParam)
		paramSet.SetParameters(newParameters)
	}

	serializedParams, err := proto.Marshal(paramSet)
	if err != nil {
		return nil, err
	}

	return serializedParams, nil
}

// Reproduce serialization test from the Python SDK.
// https://github.com/modal-labs/modal-client/blob/4c62d67ee2816146a2a5d42581f6fe7349fa1bf6/test/serialization_test.py
func TestParameterSerialization(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	serializedParams, err := serializeParameterSet(
		[]Parameter{
			{Name: "foo", Value: "bar", Type: pb.ParameterType_PARAM_TYPE_STRING},
			{Name: "i", Value: 5, Type: pb.ParameterType_PARAM_TYPE_INT},
		},
	)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	byteData := []byte("\n\x0c\n\x03foo\x10\x01\x1a\x03bar\n\x07\n\x01i\x10\x02 \x05")
	g.Expect(serializedParams).Should(gomega.Equal(byteData))

	serializedParams, err = serializeParameterSet(
		[]Parameter{
			{Name: "x", Value: []byte("\x00"), Type: pb.ParameterType_PARAM_TYPE_BYTES},
		},
	)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	byteData = []byte("\n\x08\n\x01x\x10\x042\x01\x00")
	g.Expect(serializedParams).Should(gomega.Equal(byteData))
}
