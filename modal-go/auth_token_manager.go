package modal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

const (
	// Start refreshing this many seconds before the token expires
	RefreshWindow = 5 * 60
	// If the token doesn't have an expiry field, default to current time plus this value (not expected).
	DefaultExpiryOffset = 20 * 60
)

type TokenAndExpiry struct {
	token  string
	expiry int64
}

// AuthTokenManager manages authentication tokens, refreshing them lazily
// when GetToken is called. Tokens are refreshed when expired or within
// RefreshWindow seconds of expiry.
type AuthTokenManager struct {
	client pb.ModalClientClient
	logger *slog.Logger

	tokenAndExpiry atomic.Value
	refreshMu      sync.Mutex
}

func NewAuthTokenManager(client pb.ModalClientClient, logger *slog.Logger) *AuthTokenManager {
	manager := &AuthTokenManager{
		client: client,
		logger: logger,
	}

	manager.tokenAndExpiry.Store(TokenAndExpiry{
		token:  "",
		expiry: 0,
	})

	return manager
}

// GetToken returns a valid auth token, fetching or refreshing as needed.
//
// Three states:
//  1. Valid token (not near expiry): returned immediately, no locking.
//  2. No token or expired: all callers block until a fresh token is fetched.
//     Only one goroutine makes the RPC; others wait on the mutex then see the
//     new token via a double-check.
//  3. Valid but within RefreshWindow of expiry: one goroutine refreshes
//     (blocking only itself); concurrent callers get the old, still-valid token.
func (m *AuthTokenManager) GetToken(ctx context.Context) (string, error) {
	data := m.tokenAndExpiry.Load().(TokenAndExpiry)

	if data.token == "" || isExpired(data) {
		if err := m.lockedRefresh(ctx); err != nil {
			return "", err
		}
		return m.tokenAndExpiry.Load().(TokenAndExpiry).token, nil
	}

	if needsRefresh(data) {
		if m.refreshMu.TryLock() {
			data = m.tokenAndExpiry.Load().(TokenAndExpiry)
			if needsRefresh(data) {
				if _, err := m.FetchToken(ctx); err != nil {
					m.logger.ErrorContext(ctx, "refreshing auth token", "error", err)
				}
			}
			m.refreshMu.Unlock()
		}
	}

	return m.tokenAndExpiry.Load().(TokenAndExpiry).token, nil
}

// lockedRefresh blocks until the mutex is acquired, then refreshes if still needed.
func (m *AuthTokenManager) lockedRefresh(ctx context.Context) error {
	m.refreshMu.Lock()
	defer m.refreshMu.Unlock()

	data := m.tokenAndExpiry.Load().(TokenAndExpiry)
	if data.token != "" && !needsRefresh(data) {
		return nil
	}
	_, err := m.FetchToken(ctx)
	return err
}

// FetchToken fetches a new token using AuthTokenGet() and stores it.
func (m *AuthTokenManager) FetchToken(ctx context.Context) (string, error) {
	resp, err := m.client.AuthTokenGet(ctx, &pb.AuthTokenGetRequest{})
	if err != nil {
		return "", fmt.Errorf("AuthTokenGet: %w", err)
	}

	token := resp.GetToken()
	if token == "" {
		return "", fmt.Errorf("internal error: did not receive auth token from server, please contact Modal support")
	}

	var expiry int64
	if exp := m.decodeJWT(token); exp > 0 {
		expiry = exp
	} else {
		m.logger.WarnContext(ctx, "x-modal-auth-token does not contain exp field")
		// We'll use the token, and set the expiry to 20 min from now.
		expiry = time.Now().Unix() + DefaultExpiryOffset
	}

	m.tokenAndExpiry.Store(TokenAndExpiry{
		token:  token,
		expiry: expiry,
	})

	timeUntilRefresh := time.Duration(expiry-time.Now().Unix()-RefreshWindow) * time.Second
	m.logger.DebugContext(ctx, "Fetched auth token",
		"expires_in", time.Until(time.Unix(expiry, 0)),
		"refresh_in", timeUntilRefresh)

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

// IsExpired checks if the current token is expired.
func (m *AuthTokenManager) IsExpired() bool {
	data := m.tokenAndExpiry.Load().(TokenAndExpiry)
	return isExpired(data)
}

func isExpired(data TokenAndExpiry) bool {
	return time.Now().Unix() >= data.expiry
}

func needsRefresh(data TokenAndExpiry) bool {
	return time.Now().Unix() >= data.expiry-RefreshWindow
}

// SetToken sets the token and expiry (for testing).
func (m *AuthTokenManager) SetToken(token string, expiry int64) {
	m.tokenAndExpiry.Store(TokenAndExpiry{
		token:  token,
		expiry: expiry,
	})
}
