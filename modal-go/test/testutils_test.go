package test

import (
	"context"
	"net/http"
	"testing"

	modal "github.com/modal-labs/libmodal/modal-go"
	"github.com/modal-labs/libmodal/modal-go/internal/grpcmock"
	"github.com/onsi/gomega"
)

func newModalClient(t *testing.T) *modal.Client {
	t.Helper()

	c, err := modal.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		c.Close()

		// Close idle http connections to silence goleak.
		http.DefaultClient.CloseIdleConnections()
	})

	return c
}

func newGRPCMockClient(t *testing.T) *grpcmock.MockClient {
	t.Helper()

	mock := grpcmock.NewMockClient()

	t.Cleanup(func() {
		mock.Close()
	})

	return mock
}

func terminateSandbox(g *gomega.WithT, sb *modal.Sandbox) {
	err := sb.Terminate(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}
