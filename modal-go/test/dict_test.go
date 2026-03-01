package test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/modal-labs/libmodal/modal-go"
	"github.com/modal-labs/libmodal/modal-go/internal/grpcmock"
	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestDictEphemeral(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()
	tc := newTestClient(t)

	dict, err := tc.Dicts.Ephemeral(ctx, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer dict.CloseEphemeral()
	g.Expect(dict.Name).To(gomega.BeEmpty())

	created, err := dict.Put(ctx, "key1", "value1", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(created).To(gomega.BeTrue())

	n, err := dict.Len(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(n).To(gomega.Equal(1))

	val, found, err := dict.Get(ctx, "key1")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(val).To(gomega.Equal("value1"))
}

func TestDictPutAndGet(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()
	tc := newTestClient(t)

	dict, err := tc.Dicts.Ephemeral(ctx, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer dict.CloseEphemeral()

	_, err = dict.Put(ctx, "hello", "world", nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	val, found, err := dict.Get(ctx, "hello")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(val).To(gomega.Equal("world"))

	// missing key
	_, found, err = dict.Get(ctx, "missing")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(found).To(gomega.BeFalse())
}

func TestDictContains(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()
	tc := newTestClient(t)

	dict, err := tc.Dicts.Ephemeral(ctx, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer dict.CloseEphemeral()

	_, err = dict.Put(ctx, "exists", 123, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	found, err := dict.Contains(ctx, "exists")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(found).To(gomega.BeTrue())

	found, err = dict.Contains(ctx, "nope")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(found).To(gomega.BeFalse())
}

func TestDictLen(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()
	tc := newTestClient(t)

	dict, err := tc.Dicts.Ephemeral(ctx, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer dict.CloseEphemeral()

	n, err := dict.Len(ctx)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(n).To(gomega.Equal(0))

	_, err = dict.Put(ctx, "a", 1, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = dict.Put(ctx, "b", 2, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = dict.Put(ctx, "c", 3, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	n, err = dict.Len(ctx)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(n).To(gomega.Equal(3))
}

func TestDictPop(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()
	tc := newTestClient(t)

	dict, err := tc.Dicts.Ephemeral(ctx, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer dict.CloseEphemeral()

	_, err = dict.Put(ctx, "key", "value", nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	val, found, err := dict.Pop(ctx, "key")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(val).To(gomega.Equal("value"))

	// key should be gone
	_, found, err = dict.Get(ctx, "key")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(found).To(gomega.BeFalse())

	// pop missing key
	_, found, err = dict.Pop(ctx, "key")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(found).To(gomega.BeFalse())
}

func TestDictClear(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()
	tc := newTestClient(t)

	dict, err := tc.Dicts.Ephemeral(ctx, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer dict.CloseEphemeral()

	_, err = dict.Put(ctx, "a", 1, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = dict.Put(ctx, "b", 2, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = dict.Clear(ctx)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	n, err := dict.Len(ctx)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(n).To(gomega.Equal(0))
}

func TestDictUpdate(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()
	tc := newTestClient(t)

	dict, err := tc.Dicts.Ephemeral(ctx, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer dict.CloseEphemeral()

	err = dict.Update(ctx, map[any]any{
		"x": 10,
		"y": 20,
		"z": 30,
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	n, err := dict.Len(ctx)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(n).To(gomega.Equal(3))

	val, found, err := dict.Get(ctx, "x")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(val).To(gomega.Equal(int64(10)))
}

func TestDictPutSkipIfExists(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()
	tc := newTestClient(t)

	dict, err := tc.Dicts.Ephemeral(ctx, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer dict.CloseEphemeral()

	created, err := dict.Put(ctx, "key", "first", nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(created).To(gomega.BeTrue())

	created, err = dict.Put(ctx, "key", "second", &modal.DictPutParams{SkipIfExists: true})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(created).To(gomega.BeFalse())

	// original value should be preserved
	val, found, err := dict.Get(ctx, "key")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(val).To(gomega.Equal("first"))
}

func TestDictNonEphemeral(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()
	tc := newTestClient(t)

	dictName := "test-dict-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	dict1, err := tc.Dicts.FromName(ctx, dictName, &modal.DictFromNameParams{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(dict1.Name).To(gomega.Equal(dictName))

	defer func() {
		err := tc.Dicts.Delete(ctx, dictName, nil)
		g.Expect(err).ShouldNot(gomega.HaveOccurred())

		_, err = tc.Dicts.FromName(ctx, dictName, nil)
		g.Expect(err).Should(gomega.HaveOccurred())
	}()

	_, err = dict1.Put(ctx, "data-key", "data-value", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	dict2, err := tc.Dicts.FromName(ctx, dictName, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	val, found, err := dict2.Get(ctx, "data-key")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(val).To(gomega.Equal("data-value"))
}

func TestDictKeysValuesItems(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()
	tc := newTestClient(t)

	dict, err := tc.Dicts.Ephemeral(ctx, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer dict.CloseEphemeral()

	err = dict.Update(ctx, map[any]any{
		"a": int64(1),
		"b": int64(2),
		"c": int64(3),
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Keys
	keys := make(map[any]bool)
	for k, err := range dict.Keys(ctx) {
		g.Expect(err).ToNot(gomega.HaveOccurred())
		keys[k] = true
	}
	g.Expect(keys).To(gomega.HaveLen(3))
	g.Expect(keys).To(gomega.HaveKey("a"))
	g.Expect(keys).To(gomega.HaveKey("b"))
	g.Expect(keys).To(gomega.HaveKey("c"))

	// Values
	values := make([]any, 0, 3)
	for v, err := range dict.Values(ctx) {
		g.Expect(err).ToNot(gomega.HaveOccurred())
		values = append(values, v)
	}
	g.Expect(values).To(gomega.HaveLen(3))
	g.Expect(values).To(gomega.ContainElements(int64(1), int64(2), int64(3)))

	// Items
	items := make(map[any]any)
	for item, err := range dict.Items(ctx) {
		g.Expect(err).ToNot(gomega.HaveOccurred())
		items[item.Key] = item.Value
	}
	g.Expect(items).To(gomega.HaveLen(3))
	g.Expect(items["a"]).To(gomega.Equal(int64(1)))
	g.Expect(items["b"]).To(gomega.Equal(int64(2)))
	g.Expect(items["c"]).To(gomega.Equal(int64(3)))
}

func TestDictDeleteSuccess(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	mock := newGRPCMockClient(t)

	grpcmock.HandleUnary(
		mock, "/DictGetOrCreate",
		func(req *pb.DictGetOrCreateRequest) (*pb.DictGetOrCreateResponse, error) {
			return pb.DictGetOrCreateResponse_builder{
				DictId: "di-test-123",
			}.Build(), nil
		},
	)

	grpcmock.HandleUnary(
		mock, "/DictDelete",
		func(req *pb.DictDeleteRequest) (*emptypb.Empty, error) {
			g.Expect(req.GetDictId()).To(gomega.Equal("di-test-123"))
			return &emptypb.Empty{}, nil
		},
	)

	err := mock.Dicts.Delete(ctx, "test-dict", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(mock.AssertExhausted()).ShouldNot(gomega.HaveOccurred())
}

func TestDictDeleteWithAllowMissing(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	mock := newGRPCMockClient(t)

	grpcmock.HandleUnary(
		mock, "/DictGetOrCreate",
		func(req *pb.DictGetOrCreateRequest) (*pb.DictGetOrCreateResponse, error) {
			return nil, modal.NotFoundError{Exception: "Dict 'missing' not found"}
		},
	)

	err := mock.Dicts.Delete(ctx, "missing", &modal.DictDeleteParams{
		AllowMissing: true,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(mock.AssertExhausted()).ShouldNot(gomega.HaveOccurred())
}

func TestDictDeleteWithAllowMissingDeleteRPCNotFound(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	mock := newGRPCMockClient(t)

	grpcmock.HandleUnary(mock, "/DictGetOrCreate",
		func(req *pb.DictGetOrCreateRequest) (*pb.DictGetOrCreateResponse, error) {
			return pb.DictGetOrCreateResponse_builder{DictId: "di-test-123"}.Build(), nil
		},
	)

	grpcmock.HandleUnary(mock, "/DictDelete",
		func(req *pb.DictDeleteRequest) (*emptypb.Empty, error) {
			return nil, status.Errorf(codes.NotFound, "Dict not found")
		},
	)

	err := mock.Dicts.Delete(ctx, "test-dict", &modal.DictDeleteParams{AllowMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(mock.AssertExhausted()).ShouldNot(gomega.HaveOccurred())
}

func TestDictDeleteWithAllowMissingFalseThrows(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	mock := newGRPCMockClient(t)

	grpcmock.HandleUnary(
		mock, "/DictGetOrCreate",
		func(req *pb.DictGetOrCreateRequest) (*pb.DictGetOrCreateResponse, error) {
			return nil, modal.NotFoundError{Exception: "Dict 'missing' not found"}
		},
	)

	err := mock.Dicts.Delete(ctx, "missing", &modal.DictDeleteParams{
		AllowMissing: false,
	})
	g.Expect(err).Should(gomega.HaveOccurred())

	g.Expect(mock.AssertExhausted()).ShouldNot(gomega.HaveOccurred())
}
