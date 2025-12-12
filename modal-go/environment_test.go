package modal

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"github.com/onsi/gomega"
	"google.golang.org/grpc"
)

// mockAuthTokenResponse returns a mock AuthTokenGetResponse with a JWT that has a far-future
// expiration time. The JWT structure is: mock header + base64({"exp":9999999999}) + mock signature.
// This prevents warnings from AuthTokenManager.FetchToken() which expects an "exp" field.
func mockAuthTokenResponse() (*pb.AuthTokenGetResponse, error) {
	const mockJWT = "x.eyJleHAiOjk5OTk5OTk5OTl9.x"
	return pb.AuthTokenGetResponse_builder{Token: mockJWT}.Build(), nil
}

type mockEnvironmentClient struct {
	pb.ModalClientClient
	callCount atomic.Int64
	envs      sync.Map
}

func (m *mockEnvironmentClient) EnvironmentGetOrCreate(ctx context.Context, req *pb.EnvironmentGetOrCreateRequest, opts ...grpc.CallOption) (*pb.EnvironmentGetOrCreateResponse, error) {
	m.callCount.Add(1)
	if resp, ok := m.envs.Load(req.GetDeploymentName()); ok {
		return resp.(*pb.EnvironmentGetOrCreateResponse), nil
	}
	return pb.EnvironmentGetOrCreateResponse_builder{
		EnvironmentId: "en-default",
		Metadata: pb.EnvironmentMetadata_builder{
			Name: "",
			Settings: pb.EnvironmentSettings_builder{
				ImageBuilderVersion: "2024.10",
				WebhookSuffix:       "",
			}.Build(),
		}.Build(),
	}.Build(), nil
}

func (m *mockEnvironmentClient) AuthTokenGet(ctx context.Context, req *pb.AuthTokenGetRequest, opts ...grpc.CallOption) (*pb.AuthTokenGetResponse, error) {
	return mockAuthTokenResponse()
}

func (m *mockEnvironmentClient) getCallCount() int {
	return int(m.callCount.Load())
}

func TestGetEnvironmentCached(t *testing.T) {
	g := gomega.NewWithT(t)
	ctx := context.Background()

	t.Setenv("MODAL_IMAGE_BUILDER_VERSION", "")

	mockClient := &mockEnvironmentClient{}
	mockClient.envs.Store("", pb.EnvironmentGetOrCreateResponse_builder{
		EnvironmentId: "en-test123",
		Metadata: pb.EnvironmentMetadata_builder{
			Name: "main",
			Settings: pb.EnvironmentSettings_builder{
				ImageBuilderVersion: "2024.10",
				WebhookSuffix:       "modal.run",
			}.Build(),
		}.Build(),
	}.Build())
	mockClient.envs.Store("dev", pb.EnvironmentGetOrCreateResponse_builder{
		EnvironmentId: "en-dev",
		Metadata: pb.EnvironmentMetadata_builder{
			Name: "dev",
			Settings: pb.EnvironmentSettings_builder{
				ImageBuilderVersion: "2025.06",
			}.Build(),
		}.Build(),
	}.Build())

	client, err := NewClientWithOptions(&ClientParams{ControlPlaneClient: mockClient})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer client.Close()

	env, err := client.fetchEnvironment(ctx, "")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(env.Settings.ImageBuilderVersion).Should(gomega.Equal("2024.10"))
	g.Expect(env.Settings.WebhookSuffix).Should(gomega.Equal("modal.run"))
	g.Expect(mockClient.getCallCount()).Should(gomega.Equal(1))

	env2, err := client.fetchEnvironment(ctx, "")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(env2).Should(gomega.BeIdenticalTo(env))
	g.Expect(mockClient.getCallCount()).Should(gomega.Equal(1)) // got "" from cache

	envDev, err := client.fetchEnvironment(ctx, "dev")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(envDev.Settings.ImageBuilderVersion).Should(gomega.Equal("2025.06"))
	g.Expect(mockClient.getCallCount()).Should(gomega.Equal(2)) // got "dev" from server
}

func TestImageBuilderVersion_LocalConfigHasPrecedence(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	mockClient := &mockEnvironmentClient{}
	mockClient.envs.Store("", pb.EnvironmentGetOrCreateResponse_builder{
		EnvironmentId: "en-test",
		Metadata: pb.EnvironmentMetadata_builder{
			Settings: pb.EnvironmentSettings_builder{
				ImageBuilderVersion: "2024.10",
			}.Build(),
		}.Build(),
	}.Build())

	client, err := NewClientWithOptions(&ClientParams{
		ControlPlaneClient: mockClient,
		Config: &config{
			"default": rawProfile{
				ImageBuilderVersion: "2024.04",
				Active:              true,
			},
		},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer client.Close()

	version, err := client.imageBuilderVersion(ctx, "")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(version).Should(gomega.Equal("2024.04"))
	g.Expect(mockClient.getCallCount()).Should(gomega.Equal(0))
}
