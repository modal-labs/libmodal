package modal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

type AuthTokenManager struct {
	mutex  sync.RWMutex
	client pb.ModalClientClient
	token  string
	expiry int64
}

const (
	// Start refreshing this many seconds before the token expires
	refreshWindow = 5 * 60
	// If the token doesn't have an expiry field, default it to current time plus this value (not expected)
	DefaultExpiryOffset = 20 * 60
)

func NewAuthTokenManager(client pb.ModalClientClient) *AuthTokenManager {
	return &AuthTokenManager{
		client: client,
		token:  "",
		expiry: 0,
	}
}

/*
When called, the AuthTokenManager can be in one of three states:
1. Has a valid cached token. It is returned to the caller.
2. Has no cached token, or the token is expired. We fetch a new one and cache it. If `get_token` is called
concurrently by multiple coroutines, all requests will block until the token has been fetched. But only one
coroutine will actually make a request to the control plane to fetch the new token. This ensures we do not hit
the control plane with more requests than needed.
3. Has a valid cached token, but it is going to expire in the next 5 minutes. In this case we fetch a new token
and cache it. If `get_token` is called concurrently, all requests should receive the new token.
"""
*/
func (m *AuthTokenManager) GetToken(ctx context.Context) (string, error) {
	if m.token == "" || m.isExpired() {
		return m.refreshToken(ctx)
	} else if m.needsRefresh() {
		m.mutex.Lock()
		now := time.Now().Unix()
		if m.token != "" && now < (m.expiry-refreshWindow) {
			// Someone else refreshed the token while we were waiting for the lock
			token := m.token
			m.mutex.Unlock()
			return token, nil
		}
		m.mutex.Unlock()
		return m.refreshToken(ctx)
	}
	m.mutex.RLock()
	token := m.token
	m.mutex.RUnlock()
	return token, nil
}

/*
Fetch a new token from the control plane. If called concurrently, only one goroutine will make a request for a
new token. The others will block on a lock, until the first goroutine has fetched the new token.
*/
func (m *AuthTokenManager) refreshToken(ctx context.Context) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if someone else already refreshed the token while we were waiting for the lock
	now := time.Now().Unix()
	if m.token != "" && now < m.expiry && now < (m.expiry-refreshWindow) {
		return m.token, nil
	}

	resp, err := m.client.AuthTokenGet(ctx, &pb.AuthTokenGetRequest{})
	// Not expected
	if err != nil {
		return "", fmt.Errorf("failed to get auth token: %w", err)
	}

	token := resp.GetToken()
	// Not expected
	if token == "" {
		return "", fmt.Errorf("internal error: did not receive auth token from server, please contact Modal support")
	}

	m.token = token
	if exp := m.decodeJWT(token); exp > 0 {
		m.expiry = exp
	} else {
		fmt.Printf("Failed to decode x-modal-auth-token exp field")
		// We'll use the token, and set the expiry to the default expiry offset
		m.expiry = time.Now().Unix() + DefaultExpiryOffset
	}
	return m.token, nil

}

func (m *AuthTokenManager) decodeJWT(token string) int64 {
	parts := strings.Split(token, ".")
	// Invalid JWT
	if len(parts) != 3 {
		return 0
	}

	payload := parts[1]
	padding := len(payload) % 4
	if padding > 0 {
		payload += strings.Repeat("=", 4-padding)
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return 0
	}

	// Json parsing
	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return 0
	}

	if expiry, ok := claims["exp"]; ok {
		if expFloat, ok := expiry.(float64); ok {
			return int64(expFloat)
		}
	}

	return 0
}

func (m *AuthTokenManager) needsRefresh() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return time.Now().Unix() >= (m.expiry - refreshWindow)
}

// isExpired returns true if the token has expired.
func (m *AuthTokenManager) isExpired() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return time.Now().Unix() >= m.expiry
}

// SetTokenForTesting sets the token and expiry for testing purposes
func (m *AuthTokenManager) SetTokenForTesting(token string, expiry int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.token = token
	m.expiry = expiry
}

// GetExpiryForTesting returns the current expiry for testing purposes
func (m *AuthTokenManager) GetExpiryForTesting() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.expiry
}

// DecodeJWTForTesting exposes decodeJWT for testing purposes
func (m *AuthTokenManager) DecodeJWTForTesting(token string) int64 {
	return m.decodeJWT(token)
}

// NeedsRefreshForTesting exposes needsRefresh for testing purposes
func (m *AuthTokenManager) NeedsRefreshForTesting() bool {
	return m.needsRefresh()
}

// IsExpiredForTesting exposes isExpired for testing purposes
func (m *AuthTokenManager) IsExpiredForTesting() bool {
	return m.isExpired()
}
