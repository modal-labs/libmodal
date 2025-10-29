package test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/modal-labs/libmodal/modal-go"
	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type slowModalServer struct {
	pb.UnimplementedModalClientServer
	sleepDuration time.Duration
}

// AppGetOrCreate is just chosen arbitrarily as a GRPC method to use for testing.
func (s *slowModalServer) AppGetOrCreate(ctx context.Context, req *pb.AppGetOrCreateRequest) (*pb.AppGetOrCreateResponse, error) {
	select {
	case <-time.After(s.sleepDuration):
		return pb.AppGetOrCreateResponse_builder{AppId: req.GetAppName()}.Build(), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *slowModalServer) AuthTokenGet(ctx context.Context, req *pb.AuthTokenGetRequest) (*pb.AuthTokenGetResponse, error) {
	return pb.AuthTokenGetResponse_builder{Token: "test-token"}.Build(), nil
}

func TestAppFromName_RespectsContextDeadline(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		serverSleep    time.Duration
		contextTimeout time.Duration
		expectTimeout  bool
	}{
		{
			name:           "deadline exceeded",
			serverSleep:    100 * time.Millisecond,
			contextTimeout: 10 * time.Millisecond,
			expectTimeout:  true,
		},
		{
			name:           "completes before deadline",
			serverSleep:    10 * time.Millisecond,
			contextTimeout: 100 * time.Millisecond,
			expectTimeout:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := gomega.NewWithT(t)

			lis := bufconn.Listen(1024 * 1024)

			grpcServer := grpc.NewServer()
			pb.RegisterModalClientServer(grpcServer, &slowModalServer{
				sleepDuration: tc.serverSleep,
			})

			go func() {
				if err := grpcServer.Serve(lis); err != nil {
					t.Logf("Server error: %v", err)
				}
			}()
			defer grpcServer.Stop()

			bufDialer := func(context.Context, string) (net.Conn, error) {
				return lis.Dial()
			}

			conn, err := grpc.NewClient("passthrough:///bufnet",
				grpc.WithContextDialer(bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
			defer conn.Close()

			client, err := modal.NewClientWithOptions(&modal.ClientParams{
				TokenID:            "test-token-id",
				TokenSecret:        "test-token-secret",
				Environment:        "test",
				ControlPlaneClient: pb.NewModalClientClient(conn),
			})
			g.Expect(err).ShouldNot(gomega.HaveOccurred())

			ctxWithTimeout, cancel := context.WithTimeout(context.Background(), tc.contextTimeout)
			defer cancel()

			app, err := client.Apps.FromName(ctxWithTimeout, "test-app", nil)

			if tc.expectTimeout {
				g.Expect(err).Should(gomega.HaveOccurred())
				st, ok := status.FromError(err)
				g.Expect(ok).To(gomega.BeTrue())
				g.Expect(st.Code()).To(gomega.Equal(codes.DeadlineExceeded))
			} else {
				g.Expect(err).ShouldNot(gomega.HaveOccurred())
				g.Expect(app.AppID).To(gomega.Equal("test-app"))
			}
		})
	}
}
