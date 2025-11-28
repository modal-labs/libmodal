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
	"golang.org/x/sync/singleflight"
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

// AuthTokenManager manages authentication tokens using a goroutine, and refreshes the token REFRESH_WINDOW seconds before it expires.
type AuthTokenManager struct {
	client   pb.ModalClientClient
	logger   *slog.Logger
	cancelFn context.CancelFunc

	tokenAndExpiry atomic.Value

	mu         sync.Mutex
	running    bool
	fetchGroup singleflight.Group
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

// Start the token refresh goroutine.
// Returns an error if the initial token fetch fails.
func (m *AuthTokenManager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = true
	refreshCtx, cancel := context.WithCancel(ctx)
	m.cancelFn = cancel
	m.mu.Unlock()

	if err := m.runFetch(refreshCtx); err != nil {
		m.Stop()
		return fmt.Errorf("failed to fetch initial auth token: %w", err)
	}

	go m.backgroundRefresh(refreshCtx)
	return nil
}

// Stop the refresh goroutine.
func (m *AuthTokenManager) Stop() {
	m.mu.Lock()
	m.running = false
	cancelFn := m.cancelFn
	m.cancelFn = nil
	m.mu.Unlock()

	if cancelFn != nil {
		cancelFn()
	}
}

// runFetch fetches a token, ensuring only one fetch is in progress at a time.
func (m *AuthTokenManager) runFetch(ctx context.Context) error {
	_, err, _ := m.fetchGroup.Do("fetch", func() (interface{}, error) {
		return m.FetchToken(ctx)
	})
	return err
}

// GetToken returns a valid auth token.
// If the current token is expired and the manager is running, triggers an on-demand refresh.
func (m *AuthTokenManager) GetToken(ctx context.Context) (string, error) {
	if token := m.GetCurrentToken(); token != "" && !m.IsExpired() {
		return token, nil
	}

	m.mu.Lock()
	running := m.running
	m.mu.Unlock()

	if running {
		if err := m.runFetch(ctx); err == nil {
			if token := m.GetCurrentToken(); token != "" && !m.IsExpired() {
				return token, nil
			}
		}
	}

	return "", fmt.Errorf("no valid auth token available")
}

// backgroundRefresh runs in a goroutine and refreshes tokens REFRESH_WINDOW seconds before they expire.
func (m *AuthTokenManager) backgroundRefresh(ctx context.Context) {
	for {
		data := m.tokenAndExpiry.Load().(TokenAndExpiry)
		now := time.Now().Unix()
		refreshTime := data.expiry - RefreshWindow

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
		if err := m.runFetch(ctx); err != nil {
			m.logger.ErrorContext(ctx, "Failed to refresh auth token", "error", err)
			// Sleep for 5 seconds before trying again on failure
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
		}
	}
}

// FetchToken fetches a new token using AuthTokenGet() and stores it.
func (m *AuthTokenManager) FetchToken(ctx context.Context) (string, error) {
	resp, err := m.client.AuthTokenGet(ctx, &pb.AuthTokenGetRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to get new auth token: %w", err)
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
