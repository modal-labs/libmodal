package test

import (
	"context"
	"testing"

	modal "github.com/modal-labs/libmodal/modal-go"
	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"github.com/modal-labs/libmodal/modal-go/testsupport/grpcmock"
	"github.com/onsi/gomega"
)

func TestFunctionCall(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	function, err := modal.FunctionLookup(context.Background(), "libmodal-test-support", "echo_string", nil)
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

	function, err := modal.FunctionLookup(context.Background(), "libmodal-test-support", "bytelength", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	len := 3 * 1000 * 1000 // More than 2 MiB, offload to blob storage
	input := make([]byte, len)
	result, err := function.Remote([]any{input}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal(int64(len)))
}

func TestFunctionNotFound(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	_, err := modal.FunctionLookup(context.Background(), "libmodal-test-support", "not_a_real_function", nil)
	g.Expect(err).Should(gomega.BeAssignableToTypeOf(modal.NotFoundError{}))
}

func TestFunctionCallInputPlane(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	function, err := modal.FunctionLookup(context.Background(), "libmodal-test-support", "input_plane", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Try the same, but with args.
	result, err := function.Remote([]any{"hello"}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))
}

func TestFunctionGetCurrentStats(t *testing.T) {
	g := gomega.NewWithT(t)

	mock, cleanup := grpcmock.Install()
	t.Cleanup(cleanup)

	grpcmock.HandleUnary(
		mock, "/FunctionGetCurrentStats",
		func(req *pb.FunctionGetCurrentStatsRequest) (*pb.FunctionStats, error) {
			g.Expect(req.GetFunctionId()).To(gomega.Equal("fid-stats"))
			return pb.FunctionStats_builder{Backlog: 3, NumTotalTasks: 7}.Build(), nil
		},
	)

	f := &modal.Function{FunctionId: "fid-stats"}
	stats, err := f.GetCurrentStats()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(stats).To(gomega.Equal(&modal.FunctionStats{Backlog: 3, NumTotalRunners: 7}))
}

func TestFunctionUpdateAutoscaler(t *testing.T) {
	g := gomega.NewWithT(t)

	mock, cleanup := grpcmock.Install()
	t.Cleanup(cleanup)

	grpcmock.HandleUnary(
		mock, "/FunctionUpdateSchedulingParams",
		func(req *pb.FunctionUpdateSchedulingParamsRequest) (*pb.FunctionUpdateSchedulingParamsResponse, error) {
			g.Expect(req.GetFunctionId()).To(gomega.Equal("fid-auto"))
			s := req.GetSettings()
			g.Expect(s.GetMinContainers()).To(gomega.Equal(uint32(1)))
			g.Expect(s.GetMaxContainers()).To(gomega.Equal(uint32(10)))
			g.Expect(s.GetBufferContainers()).To(gomega.Equal(uint32(2)))
			g.Expect(s.GetScaledownWindow()).To(gomega.Equal(uint32(300)))
			return &pb.FunctionUpdateSchedulingParamsResponse{}, nil
		},
	)

	f := &modal.Function{FunctionId: "fid-auto"}

	err := f.UpdateAutoscaler(modal.UpdateAutoscalerOptions{
		MinContainers:    ptrU32(1),
		MaxContainers:    ptrU32(10),
		BufferContainers: ptrU32(2),
		ScaledownWindow:  ptrU32(300),
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	grpcmock.HandleUnary(
		mock, "/FunctionUpdateSchedulingParams",
		func(req *pb.FunctionUpdateSchedulingParamsRequest) (*pb.FunctionUpdateSchedulingParamsResponse, error) {
			g.Expect(req.GetFunctionId()).To(gomega.Equal("fid-auto"))
			g.Expect(req.GetSettings().GetMinContainers()).To(gomega.Equal(uint32(2)))
			return &pb.FunctionUpdateSchedulingParamsResponse{}, nil
		},
	)

	err = f.UpdateAutoscaler(modal.UpdateAutoscalerOptions{
		MinContainers: ptrU32(2),
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func ptrU32(v uint32) *uint32 { return &v }

func TestFunctionGetWebURL(t *testing.T) {
	g := gomega.NewWithT(t)

	f, err := modal.FunctionLookup(context.Background(), "libmodal-test-support", "echo_string", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(f.GetWebURL()).To(gomega.Equal(""))

	mock, cleanup := grpcmock.Install()
	t.Cleanup(cleanup)

	grpcmock.HandleUnary(
		mock, "FunctionGet",
		func(req *pb.FunctionGetRequest) (*pb.FunctionGetResponse, error) {
			g.Expect(req.GetAppName()).To(gomega.Equal("libmodal-test-support"))
			g.Expect(req.GetObjectTag()).To(gomega.Equal("web_endpoint"))
			return pb.FunctionGetResponse_builder{
				FunctionId:     "fid-web",
				HandleMetadata: pb.FunctionHandleMetadata_builder{WebUrl: "https://endpoint.internal"}.Build(),
			}.Build(), nil
		},
	)

	wef, err := modal.FunctionLookup(context.Background(), "libmodal-test-support", "web_endpoint", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(wef.GetWebURL()).To(gomega.Equal("https://endpoint.internal"))
}
