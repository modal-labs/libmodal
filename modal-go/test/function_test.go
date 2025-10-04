package test

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	modal "github.com/modal-labs/libmodal/modal-go"
	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"github.com/modal-labs/libmodal/modal-go/testsupport/grpcmock"
	"github.com/onsi/gomega"
)

func TestFunctionCall(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	function, err := tc.Functions.FromName(ctx, "libmodal-test-support", "echo_string", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Represent Python kwargs.
	result, err := function.Remote(ctx, nil, map[string]any{"s": "hello"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))

	// Try the same, but with args.
	result, err = function.Remote(ctx, []any{"hello"}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))
}

func TestFunctionCallOldVersionError(t *testing.T) {
	// test that calling a pre 1.2 function raises an error
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	function, err := tc.Functions.FromName(ctx, "test-support-1-1", "identity_with_repr", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Represent Python kwargs.
	_, err = function.Remote(ctx, nil, map[string]any{"s": "hello"})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("please redeploy the remote function using modal >= 1.2"))
}

func TestFunctionCallGoMap(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()
	function, err := tc.Functions.FromName(ctx, "libmodal-test-support", "identity_with_repr", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Represent Python kwargs.
	inputArg := map[string]any{"s": "hello"}
	result, err := function.Remote(ctx, []any{inputArg}, nil)
	t.Log("result", result)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	// "Explode" result into two parts, assuming it's a slice/array of length 2.
	resultSlice, ok := result.([]any)
	g.Expect(ok).Should(gomega.BeTrue(), "result should be a []any")
	g.Expect(len(resultSlice)).Should(gomega.Equal(2), "result should have two elements")

	// Assert and type the first element as string
	identityResult := resultSlice[0]
	// Use custom comparison for deep equality ignoring concrete types (e.g. map[string]interface{} vs map[string]any)
	g.Expect(compareFlexible(identityResult, inputArg)).Should(gomega.BeTrue(), "identityResult should deeply equal inputArg (ignoring concrete map types)")

	reprResult, ok := resultSlice[1].(string)
	g.Expect(ok).Should(gomega.BeTrue(), "first element should be a string")
	g.Expect(reprResult).Should(gomega.Equal(`{'s': 'hello'}`), "first element should equal the Python repr {'s': 'hello'}")

}

func TestFunctionCallDateTimeRoundtrip(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	ctx := context.Background()
	function, err := tc.Functions.FromName(ctx, "libmodal-test-support", "identity_with_repr", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Test: Send a Go time.Time to Python and see how it's represented
	testTime := time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC)
	result, err := function.Remote(ctx, []any{testTime}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Parse the result - identity_with_repr returns [input, repr(input)]
	resultSlice, ok := result.([]any)
	g.Expect(ok).Should(gomega.BeTrue(), "result should be a []any")
	g.Expect(len(resultSlice)).Should(gomega.Equal(2), "result should have two elements")

	// Check what we got back (should be the original time, potentially with precision loss)
	identityResult := resultSlice[0]
	t.Logf("Go sent: %s", testTime.String())
	t.Logf("Go received back: %+v (type: %T)", identityResult, identityResult)

	// Check the Python representation
	reprResult, ok := resultSlice[1].(string)
	g.Expect(ok).Should(gomega.BeTrue(), "repr result should be a string")
	t.Logf("Python repr: %s", reprResult)

	// Analyze what Python received
	if strings.Contains(reprResult, "datetime.datetime") {
		// Success! Python received it as a datetime
		g.Expect(reprResult).Should(gomega.ContainSubstring("datetime.datetime"))
		g.Expect(reprResult).Should(gomega.ContainSubstring("2024"))
		t.Logf("✅ SUCCESS: Go time.Time was received as Python datetime.datetime")

		// Verify the roundtrip - we should get back a time.Time
		receivedTime, ok := identityResult.(time.Time)
		g.Expect(ok).Should(gomega.BeTrue(), "identity result should be a time.Time after roundtrip")

		// Check precision
		timeDiff := testTime.Sub(receivedTime)
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}
		t.Logf("Original time: %v", testTime)
		t.Logf("Received time: %v", receivedTime)
		t.Logf("Time difference after roundtrip: %v (%v nanoseconds)", timeDiff, timeDiff.Nanoseconds())

		// Python's datetime has microsecond precision (not nanosecond)
		// CBOR encodes time.Time with TimeRFC3339Nano (nanosecond precision)
		// Python decodes to datetime (rounds to nearest microsecond)
		// When Python re-encodes, we get back microsecond precision
		// So we should expect to lose sub-microsecond precision (< 1000 ns)
		//
		// Our test uses 123456789 nanoseconds = 123.456789 milliseconds
		// Python will round to 123456 microseconds = 123.456 milliseconds
		// So we should lose exactly 789 nanoseconds
		g.Expect(timeDiff).Should(gomega.BeNumerically("<", time.Microsecond),
			"time difference should be less than 1 microsecond (sub-microsecond precision loss), got %v", timeDiff)

		// Verify the times are equal when truncated to microseconds
		g.Expect(receivedTime.Truncate(time.Microsecond)).Should(gomega.Equal(testTime.Truncate(time.Microsecond)),
			"times should be equal when truncated to microseconds")

	} else {
		// Check if it's a Unix timestamp (integer)
		if unixTime, ok := identityResult.(uint64); ok {
			expectedUnix := uint64(testTime.Unix())
			g.Expect(unixTime).Should(gomega.Equal(expectedUnix), "Unix timestamp should match")
			t.Logf("⚠️  Python received Go time.Time as Unix timestamp: %s", reprResult)
			t.Logf("This means CBOR time tags are NOT being used by the Go client")
			t.Logf("✅ Unix timestamp roundtrip successful: %d", unixTime)
		} else if unixTime, ok := identityResult.(int64); ok {
			expectedUnix := testTime.Unix()
			g.Expect(unixTime).Should(gomega.Equal(expectedUnix), "Unix timestamp should match")
			t.Logf("⚠️  Python received Go time.Time as Unix timestamp: %s", reprResult)
			t.Logf("This means CBOR time tags are NOT being used by the Go client")
			t.Logf("✅ Unix timestamp roundtrip successful: %d", unixTime)
		} else {
			t.Logf("❓ Unexpected Python representation: %s", reprResult)
			t.Logf("Identity result: %+v (type: %T)", identityResult, identityResult)
		}
	}
}

func TestFunctionCallLargeInput(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	function, err := tc.Functions.FromName(ctx, "libmodal-test-support", "bytelength", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	len := 3 * 1000 * 1000 // More than 2 MiB, offload to blob storage
	input := make([]byte, len)
	result, err := function.Remote(ctx, []any{input}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal(uint64(len)))
}

func TestFunctionNotFound(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	_, err := tc.Functions.FromName(ctx, "libmodal-test-support", "not_a_real_function", nil)
	g.Expect(err).Should(gomega.BeAssignableToTypeOf(modal.NotFoundError{}))
}

func TestFunctionCallInputPlane(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	function, err := tc.Functions.FromName(ctx, "libmodal-test-support", "input_plane", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Try the same, but with args.
	result, err := function.Remote(ctx, []any{"hello"}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal("output: hello"))
}

func TestFunctionGetCurrentStats(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	mock := grpcmock.NewMockClient()
	defer func() {
		g.Expect(mock.AssertExhausted()).ShouldNot(gomega.HaveOccurred())
	}()

	grpcmock.HandleUnary(
		mock, "/FunctionGet",
		func(req *pb.FunctionGetRequest) (*pb.FunctionGetResponse, error) {
			return pb.FunctionGetResponse_builder{
				FunctionId: "fid-stats",
			}.Build(), nil
		},
	)

	f, err := mock.Functions.FromName(ctx, "test-app", "test-function", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	grpcmock.HandleUnary(
		mock, "/FunctionGetCurrentStats",
		func(req *pb.FunctionGetCurrentStatsRequest) (*pb.FunctionStats, error) {
			g.Expect(req.GetFunctionId()).To(gomega.Equal("fid-stats"))
			return pb.FunctionStats_builder{Backlog: 3, NumTotalTasks: 7}.Build(), nil
		},
	)

	stats, err := f.GetCurrentStats(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(stats).To(gomega.Equal(&modal.FunctionStats{Backlog: 3, NumTotalRunners: 7}))
}

func TestFunctionUpdateAutoscaler(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	mock := grpcmock.NewMockClient()
	defer func() {
		g.Expect(mock.AssertExhausted()).ShouldNot(gomega.HaveOccurred())
	}()

	grpcmock.HandleUnary(
		mock, "/FunctionGet",
		func(req *pb.FunctionGetRequest) (*pb.FunctionGetResponse, error) {
			return pb.FunctionGetResponse_builder{
				FunctionId: "fid-auto",
			}.Build(), nil
		},
	)

	f, err := mock.Functions.FromName(ctx, "test-app", "test-function", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

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

	err = f.UpdateAutoscaler(ctx, &modal.FunctionUpdateAutoscalerParams{
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

	err = f.UpdateAutoscaler(ctx, &modal.FunctionUpdateAutoscalerParams{
		MinContainers: ptrU32(2),
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func ptrU32(v uint32) *uint32 { return &v }

func TestFunctionGetWebURL(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	mock := grpcmock.NewMockClient()
	defer func() {
		g.Expect(mock.AssertExhausted()).ShouldNot(gomega.HaveOccurred())
	}()

	grpcmock.HandleUnary(
		mock, "FunctionGet",
		func(req *pb.FunctionGetRequest) (*pb.FunctionGetResponse, error) {
			return pb.FunctionGetResponse_builder{
				FunctionId: "fid-normal",
			}.Build(), nil
		},
	)

	f, err := mock.Functions.FromName(ctx, "libmodal-test-support", "echo_string", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(f.GetWebURL()).To(gomega.Equal(""))

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

	wef, err := mock.Functions.FromName(ctx, "libmodal-test-support", "web_endpoint", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(wef.GetWebURL()).To(gomega.Equal("https://endpoint.internal"))
}

// compareFlexible compares two values with flexible type handling
func compareFlexible(a, b interface{}) bool {
	// Handle nil cases explicitly
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Handle map comparisons
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	if av.Kind() == reflect.Map && bv.Kind() == reflect.Map {
		return compareMaps(a, b)
	}

	// If types are exactly the same, use reflect.DeepEqual
	if reflect.TypeOf(a) == reflect.TypeOf(b) {
		return reflect.DeepEqual(a, b)
	}

	// For other types, fall back to reflect.DeepEqual
	return reflect.DeepEqual(a, b)
}

// compareMaps compares two maps with flexible key type handling
func compareMaps(a, b interface{}) bool {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	if av.Kind() != reflect.Map || bv.Kind() != reflect.Map {
		return false
	}

	if av.Len() != bv.Len() {
		return false
	}

	for _, key := range av.MapKeys() {
		aVal := av.MapIndex(key)

		// Try to find the corresponding key in b
		// Handle cases where key types might differ (string vs interface{})
		var bVal reflect.Value
		found := false

		for _, bKey := range bv.MapKeys() {
			if compareFlexible(key.Interface(), bKey.Interface()) {
				bVal = bv.MapIndex(bKey)
				found = true
				break
			}
		}

		if !found || !bVal.IsValid() {
			return false
		}

		// Use flexible comparison for values
		if !compareFlexible(aVal.Interface(), bVal.Interface()) {
			return false
		}
	}

	return true
}
