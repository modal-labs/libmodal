package test

import (
	"context"
	"testing"
	"time"

	modal "github.com/modal-labs/libmodal/modal-go"
	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"github.com/modal-labs/libmodal/modal-go/testsupport/grpcmock"
	"github.com/onsi/gomega"
)

var mockFunctionProto = pb.FunctionGetResponse_builder{
	FunctionId: "fid",
	HandleMetadata: pb.FunctionHandleMetadata_builder{
		MethodHandleMetadata: map[string]*pb.FunctionHandleMetadata{"echo_string": {}},
		ClassParameterInfo:   pb.ClassParameterInfo_builder{Schema: []*pb.ClassParameterSpec{}}.Build(),
	}.Build(),
}.Build()

func TestClsWithOptionsStacking(t *testing.T) {
	g := gomega.NewWithT(t)

	mock, cleanup := grpcmock.Install()
	t.Cleanup(cleanup)

	grpcmock.HandleUnary(
		mock, "FunctionGet",
		func(req *pb.FunctionGetRequest) (*pb.FunctionGetResponse, error) {
			return mockFunctionProto, nil
		},
	)

	cls, err := modal.ClsLookup(context.Background(), "libmodal-test-support", "EchoCls", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	grpcmock.HandleUnary(
		mock, "FunctionBindParams",
		func(req *pb.FunctionBindParamsRequest) (*pb.FunctionBindParamsResponse, error) {
			g.Expect(req.GetFunctionId()).To(gomega.Equal("fid"))
			fo := req.GetFunctionOptions()
			g.Expect(fo).ToNot(gomega.BeNil())
			g.Expect(fo.GetTimeoutSecs()).To(gomega.Equal(uint32(60)))
			g.Expect(fo.GetResources()).ToNot(gomega.BeNil())
			g.Expect(fo.GetResources().GetMilliCpu()).To(gomega.Equal(uint32(250)))
			g.Expect(fo.GetResources().GetMemoryMb()).To(gomega.Equal(uint32(256)))
			g.Expect(fo.GetResources().GetGpuConfig()).ToNot(gomega.BeNil())
			g.Expect(fo.GetSecretIds()).To(gomega.Equal([]string{"sec-1"}))
			g.Expect(fo.GetReplaceSecretIds()).To(gomega.BeTrue())
			g.Expect(fo.GetReplaceVolumeMounts()).To(gomega.BeTrue())
			g.Expect(fo.GetVolumeMounts()).To(gomega.HaveLen(1))
			g.Expect(fo.GetVolumeMounts()[0].GetMountPath()).To(gomega.Equal("/mnt/test"))
			g.Expect(fo.GetVolumeMounts()[0].GetVolumeId()).To(gomega.Equal("vol-1"))
			g.Expect(fo.GetVolumeMounts()[0].GetAllowBackgroundCommits()).To(gomega.BeTrue())
			g.Expect(fo.GetVolumeMounts()[0].GetReadOnly()).To(gomega.BeFalse())
			return pb.FunctionBindParamsResponse_builder{BoundFunctionId: "fid-1", HandleMetadata: &pb.FunctionHandleMetadata{}}.Build(), nil
		},
	)

	secret := &modal.Secret{SecretId: "sec-1"}
	volume := &modal.Volume{VolumeId: "vol-1"}
	cpu := 0.25
	memory := 256
	gpu := "T4"
	timeout := 45 * time.Second
	newTimeout := 60 * time.Second

	optioned := cls.
		WithOptions(modal.ClsOptions{Timeout: &timeout, CPU: &cpu}).
		WithOptions(modal.ClsOptions{Timeout: &newTimeout, Memory: &memory, GPU: &gpu}).
		WithOptions(modal.ClsOptions{Secrets: []*modal.Secret{secret}, Volumes: map[string]*modal.Volume{"/mnt/test": volume}})

	instance, err := optioned.Instance(nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(instance).ToNot(gomega.BeNil())
}

func TestClsWithConcurrencyWithBatchingChaining(t *testing.T) {
	g := gomega.NewWithT(t)

	mock, cleanup := grpcmock.Install()
	t.Cleanup(cleanup)

	grpcmock.HandleUnary(
		mock, "FunctionGet",
		func(req *pb.FunctionGetRequest) (*pb.FunctionGetResponse, error) {
			return mockFunctionProto, nil
		},
	)

	cls, err := modal.ClsLookup(context.Background(), "libmodal-test-support", "EchoCls", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	grpcmock.HandleUnary(
		mock, "FunctionBindParams",
		func(req *pb.FunctionBindParamsRequest) (*pb.FunctionBindParamsResponse, error) {
			g.Expect(req.GetFunctionId()).To(gomega.Equal("fid"))
			fo := req.GetFunctionOptions()
			g.Expect(fo).ToNot(gomega.BeNil())
			g.Expect(fo.GetTimeoutSecs()).To(gomega.Equal(uint32(60)))
			g.Expect(fo.GetMaxConcurrentInputs()).To(gomega.Equal(uint32(10)))
			g.Expect(fo.GetBatchMaxSize()).To(gomega.Equal(uint32(11)))
			g.Expect(fo.GetBatchLingerMs()).To(gomega.Equal(uint64(12)))
			return pb.FunctionBindParamsResponse_builder{BoundFunctionId: "fid-1", HandleMetadata: &pb.FunctionHandleMetadata{}}.Build(), nil
		},
	)

	timeout := 60 * time.Second
	chained := cls.
		WithOptions(modal.ClsOptions{Timeout: &timeout}).
		WithConcurrency(modal.ClsConcurrencyOptions{MaxInputs: 10}).
		WithBatching(modal.ClsBatchingOptions{MaxBatchSize: 11, Wait: 12 * time.Millisecond})

	instance, err := chained.Instance(nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(instance).ToNot(gomega.BeNil())
}

func TestClsWithOptionsRetries(t *testing.T) {
	g := gomega.NewWithT(t)

	mock, cleanup := grpcmock.Install()
	t.Cleanup(cleanup)

	grpcmock.HandleUnary(
		mock, "FunctionGet",
		func(req *pb.FunctionGetRequest) (*pb.FunctionGetResponse, error) {
			return mockFunctionProto, nil
		},
	)

	cls, err := modal.ClsLookup(context.Background(), "libmodal-test-support", "EchoCls", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	grpcmock.HandleUnary(
		mock, "FunctionBindParams",
		func(req *pb.FunctionBindParamsRequest) (*pb.FunctionBindParamsResponse, error) {
			fo := req.GetFunctionOptions()
			g.Expect(fo).ToNot(gomega.BeNil())
			g.Expect(fo.GetRetryPolicy()).ToNot(gomega.BeNil())
			g.Expect(fo.GetRetryPolicy().GetRetries()).To(gomega.Equal(uint32(2)))
			g.Expect(fo.GetRetryPolicy().GetBackoffCoefficient()).To(gomega.Equal(float32(2.0)))
			g.Expect(fo.GetRetryPolicy().GetInitialDelayMs()).To(gomega.Equal(uint32(2000)))
			g.Expect(fo.GetRetryPolicy().GetMaxDelayMs()).To(gomega.Equal(uint32(5000)))
			return pb.FunctionBindParamsResponse_builder{BoundFunctionId: "fid-2", HandleMetadata: &pb.FunctionHandleMetadata{}}.Build(), nil
		},
	)

	backoff := float32(2.0)
	initial := 2 * time.Second
	max := 5 * time.Second
	retries, err := modal.NewRetries(2, &modal.RetriesOptions{
		BackoffCoefficient: &backoff,
		InitialDelay:       &initial,
		MaxDelay:           &max,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	_, err = cls.WithOptions(modal.ClsOptions{Retries: retries}).Instance(nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestClsWithOptionsInvalidValues(t *testing.T) {
	g := gomega.NewWithT(t)

	mock, cleanup := grpcmock.Install()
	t.Cleanup(cleanup)

	grpcmock.HandleUnary(
		mock, "FunctionGet",
		func(req *pb.FunctionGetRequest) (*pb.FunctionGetResponse, error) {
			return mockFunctionProto, nil
		},
	)

	cls, err := modal.ClsLookup(context.Background(), "libmodal-test-support", "EchoCls", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	timeout := 500 * time.Millisecond
	_, err = cls.WithOptions(modal.ClsOptions{Timeout: &timeout}).Instance(nil)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("timeout must be at least 1 second"))

	scaledownWindow := 100 * time.Millisecond
	_, err = cls.WithOptions(modal.ClsOptions{ScaledownWindow: &scaledownWindow}).Instance(nil)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("scaledownWindow must be at least 1 second"))

	fractionalTimeout := 1500 * time.Millisecond
	_, err = cls.WithOptions(modal.ClsOptions{Timeout: &fractionalTimeout}).Instance(nil)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("whole number of seconds"))

	fractionalScaledown := 1500 * time.Millisecond
	_, err = cls.WithOptions(modal.ClsOptions{ScaledownWindow: &fractionalScaledown}).Instance(nil)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("whole number of seconds"))
}

func TestWithOptionsEmptySecretsDoesNotReplace(t *testing.T) {
	g := gomega.NewWithT(t)

	mock, cleanup := grpcmock.Install()
	t.Cleanup(cleanup)

	grpcmock.HandleUnary(
		mock, "FunctionGet",
		func(req *pb.FunctionGetRequest) (*pb.FunctionGetResponse, error) {
			return mockFunctionProto, nil
		},
	)

	cls, err := modal.ClsLookup(context.Background(), "libmodal-test-support", "EchoCls", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	grpcmock.HandleUnary(
		mock, "FunctionBindParams",
		func(req *pb.FunctionBindParamsRequest) (*pb.FunctionBindParamsResponse, error) {
			g.Expect(req.GetFunctionId()).To(gomega.Equal("fid"))
			fo := req.GetFunctionOptions()
			g.Expect(fo.GetSecretIds()).To(gomega.HaveLen(0))
			g.Expect(fo.GetReplaceSecretIds()).To(gomega.BeFalse())

			return pb.FunctionBindParamsResponse_builder{BoundFunctionId: "fid-1", HandleMetadata: &pb.FunctionHandleMetadata{}}.Build(), nil
		},
	)

	_, err = cls.WithOptions(modal.ClsOptions{Secrets: []*modal.Secret{}}).Instance(nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestWithOptionsEmptyVolumesDoesNotReplace(t *testing.T) {
	g := gomega.NewWithT(t)

	mock, cleanup := grpcmock.Install()
	t.Cleanup(cleanup)

	grpcmock.HandleUnary(
		mock, "FunctionGet",
		func(req *pb.FunctionGetRequest) (*pb.FunctionGetResponse, error) {
			return mockFunctionProto, nil
		},
	)

	cls, err := modal.ClsLookup(context.Background(), "libmodal-test-support", "EchoCls", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	grpcmock.HandleUnary(
		mock, "FunctionBindParams",
		func(req *pb.FunctionBindParamsRequest) (*pb.FunctionBindParamsResponse, error) {
			g.Expect(req.GetFunctionId()).To(gomega.Equal("fid"))
			fo := req.GetFunctionOptions()
			g.Expect(fo.GetVolumeMounts()).To(gomega.HaveLen(0))
			g.Expect(fo.GetReplaceVolumeMounts()).To(gomega.BeFalse())

			return pb.FunctionBindParamsResponse_builder{BoundFunctionId: "fid-1", HandleMetadata: &pb.FunctionHandleMetadata{}}.Build(), nil
		},
	)

	_, err = cls.WithOptions(modal.ClsOptions{Volumes: map[string]*modal.Volume{}}).Instance(nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}
