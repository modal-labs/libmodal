package modal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

const (
	// Start refreshing this many seconds before the token expires
	REFRESH_WINDOW = 5 * 60
	// If the token doesn't have an expiry field, default to current time plus this value (not expected).
	DEFAULT_EXPIRY_OFFSET = 20 * 60
)

type TokenAndExpiry struct {
	token  string
	expiry int64
}

// Manages authentication tokens using a goroutine, and refreshes the token REFRESH_WINDOW seconds before it expires.
type AuthTokenManager struct {
	client   pb.ModalClientClient
	cancelFn context.CancelFunc

	tokenAndExpiry atomic.Value
}

func NewAuthTokenManager(client pb.ModalClientClient) *AuthTokenManager {
	manager := &AuthTokenManager{
		client: client,
	}

	manager.tokenAndExpiry.Store(TokenAndExpiry{
		token:  "",
		expiry: 0,
	})

	return manager
}

// Start the token refresh goroutine.
func (m *AuthTokenManager) Start(ctx context.Context) {
	refreshCtx, cancel := context.WithCancel(ctx)
	m.cancelFn = cancel
	if _, err := m.FetchToken(refreshCtx); err != nil {
		fmt.Printf("Failed to fetch initial auth token: %v\n", err)
	}
	go m.backgroundRefresh(refreshCtx)
}

// Stop the refresh goroutine.
func (m *AuthTokenManager) Stop() {
	if m.cancelFn != nil {
		m.cancelFn()
		m.cancelFn = nil
	}
}

// GetToken returns the current cached token.
// If no token is available or the token is expired, it will fetch a new one.
// Ideally GetToken() should not refresh tokens, since we have a background goroutine that refreshes tokens near expiry.
func (m *AuthTokenManager) GetToken(ctx context.Context) (string, error) {
	token := m.GetCurrentToken()

	// Return an existing, non-expired token.
	if token != "" && !m.IsExpired() {
		return token, nil
	}

	// Should almost never happen.
	return m.FetchToken(ctx)
}

// backgroundRefresh runs in a goroutine and refreshes tokens REFRESH_WINDOW seconds before they expire.
func (m *AuthTokenManager) backgroundRefresh(ctx context.Context) {
	for {
		data := m.tokenAndExpiry.Load().(TokenAndExpiry)
		now := time.Now().Unix()
		refreshTime := data.expiry - REFRESH_WINDOW

		var delay time.Duration
		if refreshTime > now {
			// Token does not need refreshing yet. Set a delay to wait until the refresh time.
			delay = time.Duration(refreshTime-now) * time.Second
		} else {
			// Token needs refresh now or is expired, refresh immediately
			// This should almost never happen.
			delay = 0
		}

		if delay > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
		}

		// Refresh the token
		if _, err := m.FetchToken(ctx); err != nil {
			fmt.Printf("Failed to refresh auth token: %v\n", err)
		}
	}
}

// FetchToken fetches a new token using AuthTokenGet() and stores it.
func (m *AuthTokenManager) FetchToken(ctx context.Context) (string, error) {
	resp, err := m.client.AuthTokenGet(ctx, &pb.AuthTokenGetRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to get auth token: %w", err)
	}

	token := resp.GetToken()
	if token == "" {
		return "", fmt.Errorf("internal error: did not receive auth token from server, please contact Modal support")
	}

	var expiry int64
	if exp := m.decodeJWT(token); exp > 0 {
		expiry = exp
	} else {
		fmt.Printf("Failed to decode x-modal-auth-token exp field\n")
		// We'll use the token, and set the expiry to 20 min from now.
		expiry = time.Now().Unix() + DEFAULT_EXPIRY_OFFSET
	}

	m.tokenAndExpiry.Store(TokenAndExpiry{
		token:  token,
		expiry: expiry,
	})

	return token, nil
}

// Extracts the exp claim from a JWT token.
func (m *AuthTokenManager) decodeJWT(token string) int64 {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0
	}

	payload := parts[1]
	for len(payload)%4 != 0 {
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return 0
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return 0
	}

	if exp, ok := claims["exp"].(float64); ok {
		return int64(exp)
	}

	return 0
}

// GetCurrentToken returns the current cached token.
func (m *AuthTokenManager) GetCurrentToken() string {
	data := m.tokenAndExpiry.Load().(TokenAndExpiry)
	return data.token
}

// IsExpired checks if token is expired.
func (m *AuthTokenManager) IsExpired() bool {
	data := m.tokenAndExpiry.Load().(TokenAndExpiry)
	now := time.Now().Unix()
	return now >= data.expiry
}

// SetToken sets the token and expiry (for testing).
func (m *AuthTokenManager) SetToken(token string, expiry int64) {
	m.tokenAndExpiry.Store(TokenAndExpiry{
		token:  token,
		expiry: expiry,
	})
}
