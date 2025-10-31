package modal

// Client construction, auth, timeout, and retry logic for Modal.

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
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

var sdkVersion = sync.OnceValue(func() string {
	const mod = "github.com/modal-labs/libmodal/modal-go"

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == mod {
				return dep.Version
			}
		}
	}
	return "v0.0.0"
})

// Client exposes services for interacting with Modal resources.
// You should not instantiate it directly, and instead use [NewClient]/[NewClientWithOptions].
type Client struct {
	Apps              AppService
	CloudBucketMounts CloudBucketMountService
	Cls               ClsService
	Functions         FunctionService
	FunctionCalls     FunctionCallService
	Images            ImageService
	Proxies           ProxyService
	Queues            QueueService
	Sandboxes         SandboxService
	Secrets           SecretService
	Volumes           VolumeService

	config           config
	profile          Profile
	sdkVersion       string
	logger           *slog.Logger
	cpClient         pb.ModalClientClient            // control plane client
	ipClients        map[string]pb.ModalClientClient // input plane clients
	authTokenManager *AuthTokenManager
	mu               sync.RWMutex
}

// NewClient generates a new client with the default profile configuration read from environment variables and ~/.modal.toml.
func NewClient() (*Client, error) {
	return NewClientWithOptions(nil)
}

// ClientParams defines credentials and options for initializing the Modal client.
type ClientParams struct {
	TokenID            string
	TokenSecret        string
	Environment        string
	Config             *config
	Logger             *slog.Logger
	ControlPlaneClient pb.ModalClientClient
}

// NewClientWithOptions generates a new client and allows overriding options in the default profile configuration.
func NewClientWithOptions(params *ClientParams) (*Client, error) {
	if params == nil {
		params = &ClientParams{}
	}

	var cfg config
	if params.Config != nil {
		cfg = *params.Config
	} else {
		var err error
		cfg, err = readConfigFile()
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	profile := getProfile(os.Getenv("MODAL_PROFILE"), cfg)

	if params.TokenID != "" {
		profile.TokenID = params.TokenID
	}
	if params.TokenSecret != "" {
		profile.TokenSecret = params.TokenSecret
	}
	if params.Environment != "" {
		profile.Environment = params.Environment
	}

	var logger *slog.Logger
	var err error

	if params.Logger != nil {
		logger = params.Logger
	} else {
		logger, err = newLogger(profile)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize logger: %w", err)
		}
	}

	c := &Client{
		config:     cfg,
		profile:    profile,
		sdkVersion: sdkVersion(),
		logger:     logger,
		ipClients:  make(map[string]pb.ModalClientClient),
	}

	logger.Debug("Initializing Modal client", "version", sdkVersion(), "server_url", profile.ServerURL)

	if params.ControlPlaneClient != nil {
		c.cpClient = params.ControlPlaneClient
	} else {
		_, c.cpClient, err = newClient(profile, c)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create control plane client: %w", err)
	}

	c.authTokenManager = NewAuthTokenManager(c.cpClient, c.logger)
	if err := c.authTokenManager.Start(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to start auth token manager: %w", err)
	}

	logger.Debug("Modal client initialized successfully")

	c.Apps = &appServiceImpl{client: c}
	c.CloudBucketMounts = &cloudBucketMountServiceImpl{client: c}
	c.Cls = &clsServiceImpl{client: c}
	c.Functions = &functionServiceImpl{client: c}
	c.FunctionCalls = &functionCallServiceImpl{client: c}
	c.Images = &imageServiceImpl{client: c}
	c.Proxies = &proxyServiceImpl{client: c}
	c.Queues = &queueServiceImpl{client: c}
	c.Sandboxes = &sandboxServiceImpl{client: c}
	c.Secrets = &secretServiceImpl{client: c}
	c.Volumes = &volumeServiceImpl{client: c}

	return c, nil
}

// ipClient returns the input plane client for the given server URL.
// It creates a new client if one doesn't exist for that specific server URL, otherwise it returns the existing client.
func (c *Client) ipClient(serverURL string) (pb.ModalClientClient, error) {
	c.mu.RLock()
	if client, ok := c.ipClients[serverURL]; ok {
		c.mu.RUnlock()
		return client, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if client, ok := c.ipClients[serverURL]; ok {
		return client, nil
	}

	c.logger.Debug("Creating input plane client", "server_url", serverURL)
	prof := c.profile
	prof.ServerURL = serverURL
	_, client, err := newClient(prof, c)
	if err != nil {
		return nil, err
	}
	c.ipClients[serverURL] = client
	return client, nil
}

// Close stops the background auth token refresh.
func (c *Client) Close() {
	c.logger.Debug("Closing Modal client")
	c.authTokenManager.Stop()
	c.logger.Debug("Modal client closed")
}

// Version returns the SDK version.
func (c *Client) Version() string {
	return c.sdkVersion
}

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

// newClient dials the given server URL with auth/timeout/retry interceptors installed.
// It returns (conn, stub). Close the conn when done.
func newClient(profile Profile, c *Client) (*grpc.ClientConn, pb.ModalClientClient, error) {
	var target string
	var creds credentials.TransportCredentials
	var scheme string
	if after, ok := strings.CutPrefix(profile.ServerURL, "https://"); ok {
		target = after
		creds = credentials.NewTLS(&tls.Config{})
		scheme = "https"
	} else if after, ok := strings.CutPrefix(profile.ServerURL, "http://"); ok {
		target = after
		creds = insecure.NewCredentials()
		scheme = "http"
	} else {
		return nil, nil, status.Errorf(codes.InvalidArgument, "invalid server URL: %s", profile.ServerURL)
	}

	c.logger.Debug("Connecting to Modal server", "target", target, "scheme", scheme)

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMessageSize),
			grpc.MaxCallSendMsgSize(maxMessageSize),
		),
		grpc.WithChainUnaryInterceptor(
			headerInjectorUnaryInterceptor(profile, c.sdkVersion),
			authTokenInterceptor(c),
			retryInterceptor(c),
			timeoutInterceptor(),
		),
		grpc.WithChainStreamInterceptor(
			headerInjectorStreamInterceptor(profile, c.sdkVersion),
		),
	)
	if err != nil {
		return nil, nil, err
	}
	return conn, pb.NewModalClientClient(conn), nil
}

// injectRequiredHeaders adds required headers to the context.
func injectRequiredHeaders(ctx context.Context, profile Profile, sdkVersion string) (context.Context, error) {
	if profile.TokenID == "" || profile.TokenSecret == "" {
		return nil, fmt.Errorf("missing token_id or token_secret, please set in .modal.toml, environment variables, or via NewClientWithOptions()")
	}

	clientType := strconv.Itoa(int(pb.ClientType_CLIENT_TYPE_LIBMODAL_GO))
	return metadata.AppendToOutgoingContext(
		ctx,
		"x-modal-client-type", clientType,
		"x-modal-client-version", "1.0.0", // CLIENT VERSION: Behaves like this Python SDK version
		"x-modal-libmodal-version", "modal-go/"+sdkVersion,
		"x-modal-token-id", profile.TokenID,
		"x-modal-token-secret", profile.TokenSecret,
	), nil
}

// headerInjectorUnaryInterceptor adds required headers to outgoing unary RPCs.
func headerInjectorUnaryInterceptor(profile Profile, sdkVersion string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		var err error
		ctx, err = injectRequiredHeaders(ctx, profile, sdkVersion)
		if err != nil {
			return err
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// headerInjectorStreamInterceptor adds required headers to outgoing streaming RPCs.
func headerInjectorStreamInterceptor(profile Profile, sdkVersion string) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		var err error
		ctx, err = injectRequiredHeaders(ctx, profile, sdkVersion)
		if err != nil {
			return nil, err
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

// authTokenInterceptor handles proactive auth token management.
// Injects auth tokens into outgoing requests.
func authTokenInterceptor(c *Client) grpc.UnaryClientInterceptor {
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
			token, err := c.authTokenManager.GetToken(ctx)
			if err != nil || token == "" {
				return fmt.Errorf("failed to get auth token: %w", err)
			}
			ctx = metadata.AppendToOutgoingContext(ctx, "x-modal-auth-token", token)
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

func retryInterceptor(c *Client) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		inv grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		c.logger.DebugContext(ctx, "Sending gRPC request", "method", method)
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
					if attempt == retries && attempt > 0 {
						c.logger.DebugContext(ctx, "Final retry attempt failed",
							"error", err,
							"retries", attempt,
							"delay", delay,
							"method", method,
							"idempotency_key", idempotency[:8])
					}
					return err
				}
			} else { // Unexpected, non-gRPC error
				return err
			}

			if attempt > 0 {
				c.logger.DebugContext(ctx, "Retryable failure",
					"error", err,
					"retries", attempt,
					"delay", delay,
					"method", method,
					"idempotency_key", idempotency[:8])
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
