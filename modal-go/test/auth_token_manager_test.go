package test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/modal-labs/libmodal/modal-go"
	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"github.com/onsi/gomega"
	"google.golang.org/grpc"
)

type mockAuthClient struct {
	pb.ModalClientClient
	authToken string
}

func newMockAuthClient() *mockAuthClient {
	return &mockAuthClient{}
}

func (m *mockAuthClient) setAuthToken(token string) {
	m.authToken = token
}

func (m *mockAuthClient) AuthTokenGet(ctx context.Context, req *pb.AuthTokenGetRequest, opts ...grpc.CallOption) (*pb.AuthTokenGetResponse, error) {
	return pb.AuthTokenGetResponse_builder{
		Token: m.authToken,
	}.Build(), nil
}

// Creates a JWT token for testing
func createTestJWT(expiry int64) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": expiry,
		"iat": time.Now().Unix(),
	})

	tokenString, _ := token.SignedString([]byte("walter-test"))
	return tokenString
}

func TestAuthTokenManager_DecodeJWT(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := newMockAuthClient()
	manager := modal.NewAuthTokenManager(mockClient)

	validToken := createTestJWT(123456789)
	mockClient.setAuthToken(validToken)

	// FetchToken should decode and store the JWT
	_, err := manager.FetchToken(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(manager.GetCurrentToken()).Should(gomega.Equal(validToken))
}

// Setting the initial token and having it cached.
func TestAuthTokenManager_InitialFetch(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := newMockAuthClient()
	token := createTestJWT(time.Now().Unix() + 3600)
	mockClient.setAuthToken(token)

	manager := modal.NewAuthTokenManager(mockClient)
	err := manager.Start(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer manager.Stop()

	firstToken, firstErr := manager.GetToken(context.Background())
	g.Expect(firstErr).ShouldNot(gomega.HaveOccurred())
	g.Expect(firstToken).Should(gomega.Equal(token))

	secondToken, secondErr := manager.GetToken(context.Background())
	g.Expect(secondErr).ShouldNot(gomega.HaveOccurred())
	g.Expect(secondToken).Should(gomega.Equal(token))
}

func TestAuthTokenManager_IsExpired(t *testing.T) {
	g := gomega.NewWithT(t)

	manager := modal.NewAuthTokenManager(nil)

	// Test not expired
	manager.SetToken("token", time.Now().Unix()+3600)
	g.Expect(manager.IsExpired()).Should(gomega.BeFalse())

	// Test expired
	manager.SetToken("token", time.Now().Unix()-3600)
	g.Expect(manager.IsExpired()).Should(gomega.BeTrue())
}

// Refreshing an expired token. Unlikely to occur since we refresh in background before expiry.
func TestAuthTokenManager_RefreshExpiredToken(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := newMockAuthClient()
	now := time.Now().Unix()

	expiringToken := createTestJWT(now - 60)
	freshToken := createTestJWT(now + 3600)

	manager := modal.NewAuthTokenManager(mockClient)
	manager.SetToken(expiringToken, now-60)
	mockClient.setAuthToken(freshToken)

	// Start the background refresh goroutine
	err := manager.Start(context.Background())
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer manager.Stop()

	// Wait for background goroutine to refresh the token
	g.Eventually(func() string {
		return manager.GetCurrentToken()
	}, "1s", "10ms").Should(gomega.Equal(freshToken))
}

func TestAuthTokenManager_RefreshNearExpiryToken(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := newMockAuthClient()
	now := time.Now().Unix()

	expiringToken := createTestJWT(now + 60)
	freshToken := createTestJWT(now + 3600)

	manager := modal.NewAuthTokenManager(mockClient)
	manager.SetToken(expiringToken, now+60)
	mockClient.setAuthToken(freshToken)

	// Start the background refresh goroutine
	err := manager.Start(context.Background())
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer manager.Stop()

	// Wait for background goroutine to refresh the token
	g.Eventually(func() string {
		return manager.GetCurrentToken()
	}, "1s", "10ms").Should(gomega.Equal(freshToken))
}

// Should error out if no valid token is available.
func TestAuthTokenManager_GetToken_ExpiredToken(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := newMockAuthClient()
	manager := modal.NewAuthTokenManager(mockClient)

	_, err := manager.GetToken(context.Background())
	g.Expect(err).Should(gomega.HaveOccurred())
}

func TestClient_Close(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := newMockAuthClient()
	token := createTestJWT(time.Now().Unix() + 3600)
	mockClient.setAuthToken(token)

	client, err := modal.NewClientWithOptions(&modal.ClientParams{
		ControlPlaneClient: mockClient,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	client.Close()
}

func TestClient_MultipleInstancesSeparateManagers(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient1 := newMockAuthClient()
	token1 := createTestJWT(time.Now().Unix() + 3600)
	mockClient1.setAuthToken(token1)

	mockClient2 := newMockAuthClient()
	token2 := createTestJWT(time.Now().Unix() + 3600)
	mockClient2.setAuthToken(token2)

	client1, err1 := modal.NewClientWithOptions(&modal.ClientParams{
		ControlPlaneClient: mockClient1,
	})
	g.Expect(err1).ShouldNot(gomega.HaveOccurred())
	defer client1.Close()

	client2, err2 := modal.NewClientWithOptions(&modal.ClientParams{
		ControlPlaneClient: mockClient2,
	})
	g.Expect(err2).ShouldNot(gomega.HaveOccurred())
	defer client2.Close()
}
