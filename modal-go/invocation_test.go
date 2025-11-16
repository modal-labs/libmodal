package modal

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc"
)

type fakeBlobGetter struct {
	urls map[string]string
}

func (f *fakeBlobGetter) BlobGet(ctx context.Context, req *pb.BlobGetRequest, opts ...grpc.CallOption) (*pb.BlobGetResponse, error) {
	url, ok := f.urls[req.GetBlobId()]
	if !ok {
		return nil, fmt.Errorf("unknown blob id: %s", req.GetBlobId())
	}
	return pb.BlobGetResponse_builder{
		DownloadUrl: url,
	}.Build(), nil
}

func TestBlobDownloadSuccess(t *testing.T) {
	t.Parallel()

	want := []byte("hello world")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(want); err != nil {
			t.Fatalf("failed to write response body: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := &fakeBlobGetter{
		urls: map[string]string{"blob-id": server.URL},
	}

	got, err := blobDownload(context.Background(), client, "blob-id")
	if err != nil {
		t.Fatalf("blobDownload returned error: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("blobDownload = %q, want %q", got, want)
	}
}

func TestBlobDownloadErrorStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "permission denied", http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	client := &fakeBlobGetter{
		urls: map[string]string{"blob-id": server.URL},
	}

	_, err := blobDownload(context.Background(), client, "blob-id")
	if err == nil {
		t.Fatal("blobDownload expected error, got nil")
	}
	if !strings.Contains(err.Error(), "status=403") {
		t.Fatalf("error %q missing status detail", err)
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("error %q missing body preview", err)
	}
}
