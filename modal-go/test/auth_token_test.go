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

func TestAuthToken_DecodeJWT(t *testing.T) {
	g := gomega.NewWithT(t)

	manager := modal.NewAuthTokenManager(nil)

	// Decoding valid JWT
	validToken := createTestJWT(123456789)
	exp := manager.DecodeJWT(validToken)
	g.Expect(exp).Should(gomega.Equal(int64(123456789)))

	// Decoding invalid JWT
	invalidExp := manager.DecodeJWT("invalid.jwt.token")
	g.Expect(invalidExp).Should(gomega.Equal(int64(0)))
}

// Setting the initial token and having it cached.
func TestAuthToken_InitialFetch(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := newMockAuthClient()
	token := createTestJWT(time.Now().Unix() + 3600)
	mockClient.setAuthToken(token)

	manager := modal.NewAuthTokenManager(mockClient)

	first_token, first_err := manager.GetToken(context.Background())
	g.Expect(first_err).ShouldNot(gomega.HaveOccurred())
	g.Expect(first_token).Should(gomega.Equal(token))

	second_token, second_err := manager.GetToken(context.Background())
	g.Expect(second_err).ShouldNot(gomega.HaveOccurred())
	g.Expect(second_token).Should(gomega.Equal(token))
}

func TestAuthToken_IsExpired(t *testing.T) {
	g := gomega.NewWithT(t)

	manager := modal.NewAuthTokenManager(nil)

	// Test not expired
	manager.SetToken("token", time.Now().Unix()+3600)
	g.Expect(manager.IsExpired()).Should(gomega.BeFalse())

	// Test expired
	manager.SetToken("token", time.Now().Unix()-3600)
	g.Expect(manager.IsExpired()).Should(gomega.BeTrue())
}

func TestAuthToken_NeedsRefresh(t *testing.T) {
	g := gomega.NewWithT(t)

	manager := modal.NewAuthTokenManager(nil)

	// Doesn't need refresh
	manager.SetToken("token", time.Now().Unix()+600)
	g.Expect(manager.NeedsRefresh()).Should(gomega.BeFalse())

	// Needs refresh
	manager.SetToken("token", time.Now().Unix()+240)
	g.Expect(manager.NeedsRefresh()).Should(gomega.BeTrue())
}

// Refreshing an expired token. Unlikely to occur since we refresh in background before expiry.
func TestAuthToken_RefreshExpiredToken(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := newMockAuthClient()
	now := time.Now().Unix()

	expiringToken := createTestJWT(now - 60)
	freshToken := createTestJWT(now + 3600)

	manager := modal.NewAuthTokenManager(mockClient)
	manager.SetToken(expiringToken, now-60)
	mockClient.setAuthToken(freshToken)

	// Start the background refresh goroutine
	manager.Start(context.Background())

	// Brief sleep for background goroutine to complete. TODO(walter): Can adjust if flaky.
	time.Sleep(10 * time.Millisecond)

	// Should have the new token cached
	g.Expect(manager.GetCurrentToken()).Should(gomega.Equal(freshToken))
	g.Expect(manager.NeedsRefresh()).Should(gomega.BeFalse())
}

func TestAuthToken_RefreshNearExpiryToken(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := newMockAuthClient()
	now := time.Now().Unix()

	expiringToken := createTestJWT(now + 60)
	freshToken := createTestJWT(now + 3600)

	manager := modal.NewAuthTokenManager(mockClient)
	manager.SetToken(expiringToken, now+60)
	mockClient.setAuthToken(freshToken)

	// Start the background refresh goroutine
	manager.Start(context.Background())

	// Brief sleep for background goroutine to complete. TODO(walter): Can adjust if flaky.
	time.Sleep(10 * time.Millisecond)

	// Should have the new token cached
	g.Expect(manager.GetCurrentToken()).Should(gomega.Equal(freshToken))
	g.Expect(manager.NeedsRefresh()).Should(gomega.BeFalse())
}

// Calling GetToken() with an expired token should trigger a refresh. Unlikely to occur since we refresh in background.
func TestAuthToken_GetToken_ExpiredToken(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := newMockAuthClient()
	expiredToken := createTestJWT(time.Now().Unix() - 60)
	freshToken := createTestJWT(time.Now().Unix() + 3600)

	manager := modal.NewAuthTokenManager(mockClient)

	manager.SetToken(expiredToken, time.Now().Unix()-60)

	mockClient.setAuthToken(freshToken)

	// GetToken() should fetch new token since cached one is expired
	result, err := manager.GetToken(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).Should(gomega.Equal(freshToken))
}
