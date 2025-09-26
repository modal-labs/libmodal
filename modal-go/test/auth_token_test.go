package test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	modal "github.com/modal-labs/libmodal/modal-go"
	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"github.com/onsi/gomega"
	"google.golang.org/grpc"
)

// mockModalClient implements pb.ModalClientClient for testing
type mockModalClient struct {
	pb.ModalClientClient
	mutex     sync.Mutex
	authToken string
	callCount int
}

func (m *mockModalClient) AuthTokenGet(ctx context.Context, req *pb.AuthTokenGetRequest, opts ...grpc.CallOption) (*pb.AuthTokenGetResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.callCount++
	if m.authToken == "" {
		return nil, fmt.Errorf("no auth token configured")
	}

	builder := pb.AuthTokenGetResponse_builder{
		Token: m.authToken,
	}
	return builder.Build(), nil
}

func (m *mockModalClient) getCallCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.callCount
}

func (m *mockModalClient) setAuthToken(token string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.authToken = token
}

func createTestJWT(exp int64) string {
	claims := jwt.MapClaims{}
	if exp > 0 {
		claims["exp"] = exp
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("walter-secret-key"))
	return tokenString
}

// captureStdout captures stdout during the execution of a function
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestGetToken(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := &mockModalClient{}
	authTokenManager := modal.NewAuthTokenManager(mockClient)

	validToken := createTestJWT(time.Now().Unix() + 3600)
	mockClient.setAuthToken(validToken)

	token, err := authTokenManager.GetToken(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(token).Should(gomega.Equal(validToken))
	g.Expect(mockClient.getCallCount()).Should(gomega.Equal(1))
}

// Test that cached token is returned without making new request
func TestGetTokenCached(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := &mockModalClient{}
	authTokenManager := modal.NewAuthTokenManager(mockClient)

	firstToken := createTestJWT(time.Now().Unix() + 3600)
	mockClient.setAuthToken(firstToken)

	// Set up initial token
	token1, err := authTokenManager.GetToken(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(token1).Should(gomega.Equal(firstToken))

	// Set a bogus token in the servier, and verify we get the cached valid token
	mockClient.setAuthToken("bogus")
	token2, err := authTokenManager.GetToken(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(token2).Should(gomega.Equal(firstToken))
	g.Expect(mockClient.getCallCount()).Should(gomega.Equal(1))
}

// Test that expired token triggers refresh
func TestGetExpiredToken(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := &mockModalClient{}
	authTokenManager := modal.NewAuthTokenManager(mockClient)

	expiredToken := createTestJWT(time.Now().Unix() - 3600)
	authTokenManager.SetTokenForTesting(expiredToken, time.Now().Unix()-3600)

	// Set up new valid token
	refreshedToken := createTestJWT(time.Now().Unix() + 3600)
	mockClient.setAuthToken(refreshedToken)

	token, err := authTokenManager.GetToken(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(token).Should(gomega.Equal(refreshedToken))
	g.Expect(mockClient.getCallCount()).Should(gomega.Equal(1))
}

// Test that token is refreshed when it is close to expiry
func TestGetTokenNearExpiry(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := &mockModalClient{}
	authTokenManager := modal.NewAuthTokenManager(mockClient)

	// Set up expiring token that expires within the refresh window
	nearExpiryTime := time.Now().Unix() + 180
	nearExpiryToken := createTestJWT(nearExpiryTime)
	authTokenManager.SetTokenForTesting(nearExpiryToken, nearExpiryTime)

	// Set up new valid token for refresh
	refreshedToken := createTestJWT(time.Now().Unix() + 3600)
	mockClient.setAuthToken(refreshedToken)

	token, err := authTokenManager.GetToken(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(token).Should(gomega.Equal(refreshedToken))
	g.Expect(mockClient.getCallCount()).Should(gomega.Equal(1))
}

// Test handling of token without exp claim
func TestGetTokenWithoutExpClaim(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := &mockModalClient{}
	authTokenManager := modal.NewAuthTokenManager(mockClient)

	tokenWithoutExp := createTestJWT(0) // 0 means no exp claim
	mockClient.setAuthToken(tokenWithoutExp)

	var token string
	var err error
	output := captureStdout(func() {
		token, err = authTokenManager.GetToken(context.Background())
	})

	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(token).Should(gomega.Equal(tokenWithoutExp))

	// Verify the warning message was printed for missing exp claim
	g.Expect(output).Should(gomega.ContainSubstring("Failed to decode x-modal-auth-token exp field"))

	// Verify that the manager set the default expiry
	expiry := authTokenManager.GetExpiryForTesting()
	g.Expect(expiry).Should(gomega.BeNumerically(">", time.Now().Unix()))
	g.Expect(expiry).Should(gomega.BeNumerically("<=", time.Now().Unix()+modal.DefaultExpiryOffset))

}

func TestGetTokenEmptyResponse(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := &mockModalClient{}
	authTokenManager := modal.NewAuthTokenManager(mockClient)

	mockClient.setAuthToken("")

	_, err := authTokenManager.GetToken(context.Background())
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("failed to get auth token: no auth token configured"))
}

func TestConcurrentGetToken(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := &mockModalClient{}
	authTokenManager := modal.NewAuthTokenManager(mockClient)

	validToken := createTestJWT(time.Now().Unix() + 3600)
	mockClient.setAuthToken(validToken)

	// Make concurrent calls
	const numCalls = 10
	results := make([]string, numCalls)
	errors := make([]error, numCalls)
	var wg sync.WaitGroup

	for i := 0; i < numCalls; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			token, err := authTokenManager.GetToken(context.Background())
			results[idx] = token
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	// All should succeed and return the same token
	for i := 0; i < numCalls; i++ {
		g.Expect(errors[i]).ShouldNot(gomega.HaveOccurred())
		g.Expect(results[i]).Should(gomega.Equal(validToken))
	}

	// Should have made only one call to the server
	g.Expect(mockClient.getCallCount()).Should(gomega.Equal(1))
}

// Test that when getToken is called concurrently
func TestConcurrentRefresh(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := &mockModalClient{}
	authTokenManager := modal.NewAuthTokenManager(mockClient)

	// Set up expiring token that needs refresh
	nearExpiryTime := time.Now().Unix() + 180
	authTokenManager.SetTokenForTesting("old.but.valid.token", nearExpiryTime)

	newToken := createTestJWT(time.Now().Unix() + 3600)
	mockClient.setAuthToken(newToken)

	// Make concurrent calls
	const numCalls = 10
	results := make([]string, numCalls)
	errors := make([]error, numCalls)
	var wg sync.WaitGroup

	for i := 0; i < numCalls; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			token, err := authTokenManager.GetToken(context.Background())
			results[idx] = token
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	// All should succeed
	for i := 0; i < numCalls; i++ {
		g.Expect(errors[i]).ShouldNot(gomega.HaveOccurred())
	}

	// At least one call should have returned the new token
	// Typically, all calls should have returned the new token, but due to timing, some may have returned the old token
	g.Expect(results).Should(gomega.ContainElement(newToken))

	// Should have made only one call to AuthTokenGet
	g.Expect(mockClient.getCallCount()).Should(gomega.Equal(1))

	// The new token should be cached now
	finalToken, err := authTokenManager.GetToken(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(finalToken).Should(gomega.Equal(newToken))
}

func TestDecodeJWTValid(t *testing.T) {
	g := gomega.NewWithT(t)

	authTokenManager := modal.NewAuthTokenManager(nil)
	validToken := createTestJWT(time.Now().Unix() + 3600)

	exp := authTokenManager.DecodeJWTForTesting(validToken)
	g.Expect(exp).Should(gomega.BeNumerically(">", time.Now().Unix()))
}

func TestDecodeJWTWithoutExpClaim(t *testing.T) {
	g := gomega.NewWithT(t)

	authTokenManager := modal.NewAuthTokenManager(nil)
	tokenWithoutExp := createTestJWT(0) // No exp claim

	exp := authTokenManager.DecodeJWTForTesting(tokenWithoutExp)
	g.Expect(exp).Should(gomega.Equal(int64(0)))
}

// Test JWT decoding with invalid format
func TestDecodeJWTInvalidFormat(t *testing.T) {
	g := gomega.NewWithT(t)

	authTokenManager := modal.NewAuthTokenManager(nil)

	exp := authTokenManager.DecodeJWTForTesting("invalid.token")
	g.Expect(exp).Should(gomega.Equal(int64(0))) // Should return 0 for invalid format
}

// Test needsRefresh helper returns true when token expires soon
func TestNeedsRefreshTrue(t *testing.T) {
	g := gomega.NewWithT(t)

	authTokenManager := modal.NewAuthTokenManager(nil)
	authTokenManager.SetTokenForTesting("walter-test-token", time.Now().Unix()+180)

	g.Expect(authTokenManager.NeedsRefreshForTesting()).Should(gomega.BeTrue())
}

// Test needsRefresh helper returns false when token is not close to expiry
func TestNeedsRefreshFalse(t *testing.T) {
	g := gomega.NewWithT(t)

	authTokenManager := modal.NewAuthTokenManager(nil)
	authTokenManager.SetTokenForTesting("walter-test-token", time.Now().Unix()+600)

	g.Expect(authTokenManager.NeedsRefreshForTesting()).Should(gomega.BeFalse())
}

// Test isExpired helper returns true for expired token
func TestIsExpiredTrue(t *testing.T) {
	g := gomega.NewWithT(t)

	authTokenManager := modal.NewAuthTokenManager(nil)
	authTokenManager.SetTokenForTesting("walter-test-token", time.Now().Unix()-60)

	g.Expect(authTokenManager.IsExpiredForTesting()).Should(gomega.BeTrue())
}

// Test isExpired returns false for valid token
func TestIsExpiredFalse(t *testing.T) {
	g := gomega.NewWithT(t)

	authTokenManager := modal.NewAuthTokenManager(nil)
	authTokenManager.SetTokenForTesting("test-token", time.Now().Unix()+60)

	g.Expect(authTokenManager.IsExpiredForTesting()).Should(gomega.BeFalse())
}

// Test multiple refresh cycles work correctly
func TestMultipleRefreshCycles(t *testing.T) {
	g := gomega.NewWithT(t)

	mockClient := &mockModalClient{}
	manager := modal.NewAuthTokenManager(mockClient)

	exp := time.Now().Unix() + 3600
	tokens := []string{
		createTestJWT(exp),
		createTestJWT(exp),
		createTestJWT(exp),
	}

	// First call
	mockClient.setAuthToken(tokens[0])
	token0, err := manager.GetToken(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(token0).Should(gomega.Equal(tokens[0]))

	// Expire the token
	manager.SetTokenForTesting(token0, time.Now().Unix()-100)

	// Second call
	mockClient.setAuthToken(tokens[1])
	token1, err := manager.GetToken(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(token1).Should(gomega.Equal(tokens[1]))

	// Expire again
	manager.SetTokenForTesting(token1, time.Now().Unix()-100)

	// Third call
	mockClient.setAuthToken(tokens[2])
	token2, err := manager.GetToken(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(token2).Should(gomega.Equal(tokens[2]))
}
