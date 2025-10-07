package modal

// Client construction, auth, timeout, and retry logic for Modal.

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

// timeoutCallOption carries a per-RPC absolute timeout.
type timeoutCallOption struct {
	grpc.EmptyCallOption
	timeout time.Duration
}

// retryCallOption carries per-RPC retry overrides.
type retryCallOption struct {
	grpc.EmptyCallOption
	retries         *int
	baseDelay       *time.Duration
	maxDelay        *time.Duration
	delayFactor     *float64
	additionalCodes []codes.Code
}

const (
	apiEndpoint            = "api.modal.com:443"
	maxMessageSize         = 100 * 1024 * 1024 // 100 MB
	defaultRetryAttempts   = 3
	defaultRetryBaseDelay  = 100 * time.Millisecond
	defaultRetryMaxDelay   = 1 * time.Second
	defaultRetryBackoffMul = 2.0
)

var retryableGrpcStatusCodes = map[codes.Code]struct{}{
	codes.DeadlineExceeded: {},
	codes.Unavailable:      {},
	codes.Canceled:         {},
	codes.Internal:         {},
	codes.Unknown:          {},
}

func isRetryableGrpc(err error) bool {
	if st, ok := status.FromError(err); ok {
		if _, ok := retryableGrpcStatusCodes[st.Code()]; ok {
			return true
		}
	}
	return false
}

// defaultConfig caches the parsed ~/.modal.toml contents (may be empty).
var defaultConfig config

// defaultProfile is resolved at package init from MODAL_PROFILE, ~/.modal.toml, etc.
var defaultProfile Profile

// clientProfile is the actual profile, from defaultProfile + InitializeClient().
var clientProfile Profile

// client is the default Modal client that talks to the control plane.
var client pb.ModalClientClient

// inputPlaneClients is a map of server URL to input-plane client.
var inputPlaneClients = map[string]pb.ModalClientClient{}

// authTokenManager manages auth tokens proactively in the background.
var authTokenManager *AuthTokenManager

func init() {
	defaultConfig, _ = readConfigFile()
	defaultProfile = getProfile(os.Getenv("MODAL_PROFILE"))
	clientProfile = defaultProfile
	var err error
	_, client, err = clientFactory(clientProfile)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize Modal client at startup: %v", err))
	}
	authTokenManager = NewAuthTokenManager(client)
}

// ClientOptions defines credentials and options for initializing the Modal client at runtime.
type ClientOptions struct {
	TokenId     string
	TokenSecret string
	Environment string // optional, defaults to the profile's environment
}

// InitializeClient updates the global Modal client configuration with the provided options.
//
// This function is useful when you want to set the client options programmatically. It
// should be called once at the start of your application.
func InitializeClient(options ClientOptions) error {
	mergedProfile := defaultProfile
	mergedProfile.TokenId = options.TokenId
	mergedProfile.TokenSecret = options.TokenSecret
	mergedProfile.Environment = firstNonEmpty(options.Environment, mergedProfile.Environment)
	clientProfile = mergedProfile
	var err error
	_, client, err = clientFactory(mergedProfile)
	if err != nil {
		return err
	}

	// Initialize new auth manager with client
	if authTokenManager == nil {
		authTokenManager = NewAuthTokenManager(client)
	}
	if err := authTokenManager.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start auth token manager: %w", err)
	}

	return nil
}

// Stops the auth token refresh.
func Close() {
	if authTokenManager != nil {
		authTokenManager.Stop()
	}
}

// getOrCreateInputPlaneClient returns a client for the given server URL, creating it if it doesn't exist.
func getOrCreateInputPlaneClient(serverURL string) (pb.ModalClientClient, error) {
	if client, ok := inputPlaneClients[serverURL]; ok {
		return client, nil
	}

	profile := clientProfile
	profile.ServerURL = serverURL
	_, client, err := clientFactory(profile)
	if err != nil {
		return nil, err
	}
	inputPlaneClients[serverURL] = client
	return client, nil
}

// clientFactory is the factory used to construct gRPC connections and stubs.
// Tests may override this variable to install a mock.
var clientFactory func(Profile) (grpc.ClientConnInterface, pb.ModalClientClient, error) = func(profile Profile) (grpc.ClientConnInterface, pb.ModalClientClient, error) {
	return newClient(profile)
}

// newClient dials the given server URL with auth/timeout/retry interceptors installed.
// It returns (conn, stub). Close the conn when done.
func newClient(profile Profile) (*grpc.ClientConn, pb.ModalClientClient, error) {
	var target string
	var creds credentials.TransportCredentials
	if after, ok := strings.CutPrefix(profile.ServerURL, "https://"); ok {
		target = after
		creds = credentials.NewTLS(&tls.Config{})
	} else if after, ok := strings.CutPrefix(profile.ServerURL, "http://"); ok {
		target = after
		creds = insecure.NewCredentials()
	} else {
		return nil, nil, status.Errorf(codes.InvalidArgument, "invalid server URL: %s", profile.ServerURL)
	}

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMessageSize),
			grpc.MaxCallSendMsgSize(maxMessageSize),
		),
		grpc.WithChainUnaryInterceptor(
			headerInjectorUnaryInterceptor(),
			authTokenInterceptor(),
			retryInterceptor(),
			timeoutInterceptor(),
		),
		grpc.WithChainStreamInterceptor(
			headerInjectorStreamInterceptor(),
		),
	)
	if err != nil {
		return nil, nil, err
	}
	return conn, pb.NewModalClientClient(conn), nil
}

// injectRequiredHeaders adds required headers to the context.
func injectRequiredHeaders(ctx context.Context) (context.Context, error) {
	if clientProfile.TokenId == "" || clientProfile.TokenSecret == "" {
		return nil, fmt.Errorf("missing token_id or token_secret, please set in .modal.toml, environment variables, or via InitializeClient()")
	}

	clientType := strconv.Itoa(int(pb.ClientType_CLIENT_TYPE_LIBMODAL_GO))
	return metadata.AppendToOutgoingContext(
		ctx,
		"x-modal-client-type", clientType,
		"x-modal-client-version", "1.0.0", // CLIENT VERSION: Behaves like this Python SDK version
		"x-modal-token-id", clientProfile.TokenId,
		"x-modal-token-secret", clientProfile.TokenSecret,
	), nil
}

// headerInjectorUnaryInterceptor adds required headers to outgoing unary RPCs.
func headerInjectorUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		var err error
		ctx, err = injectRequiredHeaders(ctx)
		if err != nil {
			return err
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// headerInjectorStreamInterceptor adds required headers to outgoing streaming RPCs.
func headerInjectorStreamInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		var err error
		ctx, err = injectRequiredHeaders(ctx)
		if err != nil {
			return nil, err
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

// authTokenInterceptor adds auth tokens to outgoing requests using the background auth manager.
func authTokenInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		inv grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Skip auth token for AuthTokenGet requests to prevent it from getting stuck
		if method != "/modal.client.ModalClient/AuthTokenGet" {
			// InitializeClient() should have been called to create the authTokenManager.
			// Some of our tests don't call InitializeClient(), so we need to create the authTokenManager here for testing purposes.
			if authTokenManager == nil {
				authTokenManager = NewAuthTokenManager(client)
				if err := authTokenManager.Start(ctx); err != nil {
					return fmt.Errorf("failed to start auth token manager: %w", err)
				}
			} else if authTokenManager.GetCurrentToken() == "" {
				// Auth token manager exists but hasn't been started yet
				if err := authTokenManager.Start(ctx); err != nil {
					return fmt.Errorf("failed to start auth token manager: %w", err)
				}
			}
			token, err := authTokenManager.GetToken(ctx)
			if err != nil {
				return fmt.Errorf("failed to get auth token: %w", err)
			}
			if token != "" {
				ctx = metadata.AppendToOutgoingContext(ctx, "x-modal-auth-token", token)
			}
		}
		return inv(ctx, method, req, reply, cc, opts...)
	}
}

func timeoutInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		inv grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// pick the first TimeoutCallOption, if any
		for _, o := range opts {
			if to, ok := o.(timeoutCallOption); ok && to.timeout > 0 {
				// honour an existing, *earlier* deadline if present
				if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) <= to.timeout {
					break
				}
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, to.timeout)
				defer cancel()
				break
			}
		}
		return inv(ctx, method, req, reply, cc, opts...)
	}
}

func retryInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		inv grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// start with package defaults
		retries := defaultRetryAttempts
		baseDelay := defaultRetryBaseDelay
		maxDelay := defaultRetryMaxDelay
		factor := defaultRetryBackoffMul
		retryable := retryableGrpcStatusCodes

		// override from call-options (first one wins)
		for _, o := range opts {
			if rc, ok := o.(retryCallOption); ok {
				if rc.retries != nil {
					retries = *rc.retries
				}
				if rc.baseDelay != nil {
					baseDelay = *rc.baseDelay
				}
				if rc.maxDelay != nil {
					maxDelay = *rc.maxDelay
				}
				if rc.delayFactor != nil {
					factor = *rc.delayFactor
				}
				if len(rc.additionalCodes) > 0 {
					retryable = map[codes.Code]struct{}{}
					for k := range retryableGrpcStatusCodes {
						retryable[k] = struct{}{}
					}
					for _, c := range rc.additionalCodes {
						retryable[c] = struct{}{}
					}
				}
				break
			}
		}

		idempotency := uuid.NewString()
		start := time.Now()
		delay := baseDelay

		for attempt := 0; attempt <= retries; attempt++ {
			aCtx := metadata.AppendToOutgoingContext(
				ctx,
				"x-idempotency-key", idempotency,
				"x-retry-attempt", strconv.Itoa(attempt),
				"x-retry-delay", strconv.FormatFloat(time.Since(start).Seconds(), 'f', 3, 64),
			)

			err := inv(aCtx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}

			if st, ok := status.FromError(err); ok { // gRPC error
				if _, ok := retryable[st.Code()]; !ok || attempt == retries {
					return err
				}
			} else { // Unexpected, non-gRPC error
				return err
			}

			if sleepCtx(ctx, delay) != nil {
				return err // ctx cancelled or deadline exceeded
			}

			// exponential back-off
			delay = min(delay*time.Duration(factor), maxDelay)
		}
		return nil // unreachable
	}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
