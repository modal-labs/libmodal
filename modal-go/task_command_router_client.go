package modal

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// tlsCredsNoALPN is a TLS credential that skips ALPN enforcement, implementing
// the google.golang.org/grpc/credentials#TransportCredentials interface.
//
// Starting in grpc-go v1.67, ALPN is enforced by default for TLS connections.
// However, the task command router server doesn't negotiate ALPN.
// This performs the TLS handshake without that check.
// See: https://github.com/grpc/grpc-go/issues/434
type tlsCredsNoALPN struct{}

func (c *tlsCredsNoALPN) ClientHandshake(ctx context.Context, authority string, rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	serverName, _, err := net.SplitHostPort(authority)
	if err != nil {
		serverName = authority
	}
	cfg := &tls.Config{
		ServerName: serverName,
		NextProtos: []string{"h2"},
	}

	conn := tls.Client(rawConn, cfg)
	if err := conn.HandshakeContext(ctx); err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	return conn, credentials.TLSInfo{
		State: conn.ConnectionState(),
		CommonAuthInfo: credentials.CommonAuthInfo{
			SecurityLevel: credentials.PrivacyAndIntegrity,
		},
	}, nil
}

func (c *tlsCredsNoALPN) ServerHandshake(net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return nil, nil, errors.New("tlsCredsNoALPN: server-side not supported")
}

func (c *tlsCredsNoALPN) Info() credentials.ProtocolInfo {
	return credentials.ProtocolInfo{SecurityProtocol: "tls", SecurityVersion: "1.2"}
}

func (c *tlsCredsNoALPN) Clone() credentials.TransportCredentials {
	return &tlsCredsNoALPN{}
}

func (c *tlsCredsNoALPN) OverrideServerName(string) error {
	return errors.New("tlsCredsNoALPN: OverrideServerName not supported")
}

// retryOptions configures retry behavior for callWithRetriesOnTransientErrors.
type retryOptions struct {
	BaseDelay   time.Duration
	DelayFactor float64
	MaxRetries  *int // nil means retry forever
	Deadline    *time.Time
}

// defaultRetryOptions returns the default retry options.
func defaultRetryOptions() retryOptions {
	maxRetries := 10
	return retryOptions{
		BaseDelay:   10 * time.Millisecond,
		DelayFactor: 2.0,
		MaxRetries:  &maxRetries,
		Deadline:    nil,
	}
}

var commandRouterRetryableCodes = map[codes.Code]struct{}{
	codes.DeadlineExceeded: {},
	codes.Unavailable:      {},
	codes.Canceled:         {},
	codes.Internal:         {},
	codes.Unknown:          {},
}

// parseJwtExpiration extracts the expiration time from a JWT token.
// Returns nil if the token is malformed or doesn't have an exp claim.
func parseJwtExpiration(ctx context.Context, jwt string, logger *slog.Logger) *int64 {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil
	}

	payloadB64 := parts[1]
	switch len(payloadB64) % 4 {
	case 2:
		payloadB64 += "=="
	case 3:
		payloadB64 += "="
	}

	payloadJSON, err := base64.StdEncoding.DecodeString(payloadB64)
	if err != nil {
		payloadJSON, err = base64.URLEncoding.DecodeString(payloadB64)
		if err != nil {
			logger.WarnContext(ctx, "Failed to decode JWT payload", "error", err)
			return nil
		}
	}

	var payload struct {
		Exp json.Number `json:"exp"`
	}
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		logger.WarnContext(ctx, "Failed to parse JWT payload", "error", err)
		return nil
	}

	if payload.Exp == "" {
		return nil
	}

	exp, err := payload.Exp.Int64()
	if err != nil {
		return nil
	}

	return &exp
}

var errDeadlineExceeded = errors.New("deadline exceeded")

// callWithRetriesOnTransientErrors retries the given function on transient gRPC errors.
func callWithRetriesOnTransientErrors[T any](
	ctx context.Context,
	fn func() (T, error),
	opts retryOptions,
) (T, error) {
	var zero T
	delay := opts.BaseDelay
	numRetries := 0

	for {
		if opts.Deadline != nil && time.Now().After(*opts.Deadline) {
			return zero, errDeadlineExceeded
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		st, ok := status.FromError(err)
		if !ok {
			return zero, err
		}

		if _, retryable := commandRouterRetryableCodes[st.Code()]; !retryable {
			return zero, err
		}

		if opts.MaxRetries != nil && numRetries >= *opts.MaxRetries {
			return zero, err
		}

		if opts.Deadline != nil && time.Now().Add(delay).After(*opts.Deadline) {
			return zero, errDeadlineExceeded
		}

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(delay):
		}

		delay = time.Duration(float64(delay) * opts.DelayFactor)
		numRetries++
	}
}

// TaskCommandRouterClient provides a client for the TaskCommandRouter gRPC service.
type TaskCommandRouterClient struct {
	stub         pb.TaskCommandRouterClient
	conn         *grpc.ClientConn
	serverClient pb.ModalClientClient
	taskID       string
	serverURL    string
	jwt          atomic.Pointer[string]
	jwtExp       atomic.Pointer[int64]
	logger       *slog.Logger
	closed       atomic.Bool

	refreshJwtGroup singleflight.Group
}

// TryInitTaskCommandRouterClient attempts to initialize a TaskCommandRouterClient.
// Returns nil if command router access is not available for this task.
func TryInitTaskCommandRouterClient(
	ctx context.Context,
	serverClient pb.ModalClientClient,
	taskID string,
	logger *slog.Logger,
	profile Profile,
) (*TaskCommandRouterClient, error) {
	resp, err := serverClient.TaskGetCommandRouterAccess(ctx, pb.TaskGetCommandRouterAccessRequest_builder{
		TaskId: taskID,
	}.Build())
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.FailedPrecondition {
			logger.DebugContext(ctx, "Command router access is not enabled for task", "task_id", taskID)
			return nil, nil
		}
		return nil, err
	}

	logger.DebugContext(ctx, "Using command router access for task", "task_id", taskID, "url", resp.GetUrl())

	url, err := url.Parse(resp.GetUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to parse task router URL: %w", err)
	}

	if url.Scheme != "https" {
		return nil, fmt.Errorf("task router URL must be https, got: %s", resp.GetUrl())
	}

	host := url.Hostname()
	port := url.Port()
	if port == "" {
		port = "443"
	}
	target := fmt.Sprintf("%s:%s", host, port)

	var creds credentials.TransportCredentials
	if profile.TaskCommandRouterInsecure {
		logger.WarnContext(ctx, "Using insecure TLS for task command router due to MODAL_TASK_COMMAND_ROUTER_INSECURE")
		creds = insecure.NewCredentials()
	} else {
		// Use custom TLS credentials that skip ALPN enforcement.
		// The command router server may not negotiate ALPN, which causes
		// grpc-go v1.67+ to fail the handshake.
		creds = &tlsCredsNoALPN{}
	}

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMessageSize),
			grpc.MaxCallSendMsgSize(maxMessageSize),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create task command router connection: %w", err)
	}

	client := &TaskCommandRouterClient{
		stub:         pb.NewTaskCommandRouterClient(conn),
		conn:         conn,
		serverClient: serverClient,
		taskID:       taskID,
		serverURL:    resp.GetUrl(),
		logger:       logger,
	}
	jwt := resp.GetJwt()
	client.jwt.Store(&jwt)
	jwtExp := parseJwtExpiration(ctx, jwt, logger)
	client.jwtExp.Store(jwtExp)

	logger.DebugContext(ctx, "Successfully initialized command router client", "task_id", taskID)
	return client, nil
}

// Close closes the gRPC connection.
func (c *TaskCommandRouterClient) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *TaskCommandRouterClient) authContext(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+*c.jwt.Load())
}

func (c *TaskCommandRouterClient) refreshJwt(ctx context.Context) error {
	const jwtRefreshBufferSeconds = 30

	if c.closed.Load() {
		return errors.New("client is closed")
	}

	// If the current JWT expiration is already far enough in the future, don't refresh.
	if exp := c.jwtExp.Load(); exp != nil && *exp-time.Now().Unix() > jwtRefreshBufferSeconds {
		c.logger.DebugContext(ctx, "Skipping JWT refresh because expiration is far enough in the future", "task_id", c.taskID)
		return nil
	}

	_, err, _ := c.refreshJwtGroup.Do("refresh", func() (any, error) {
		if exp := c.jwtExp.Load(); exp != nil && *exp-time.Now().Unix() > jwtRefreshBufferSeconds {
			return nil, nil
		}

		resp, err := c.serverClient.TaskGetCommandRouterAccess(ctx, pb.TaskGetCommandRouterAccessRequest_builder{
			TaskId: c.taskID,
		}.Build())
		if err != nil {
			return nil, fmt.Errorf("failed to refresh JWT: %w", err)
		}

		if resp.GetUrl() != c.serverURL {
			return nil, errors.New("task router URL changed during session")
		}

		jwt := resp.GetJwt()
		c.jwt.Store(&jwt)
		jwtExp := parseJwtExpiration(ctx, jwt, c.logger)
		c.jwtExp.Store(jwtExp)
		return nil, nil
	})
	return err
}

func (c *TaskCommandRouterClient) callWithAuthRetry(ctx context.Context, fn func(context.Context) error) error {
	err := fn(c.authContext(ctx))
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated {
			if refreshErr := c.refreshJwt(ctx); refreshErr != nil {
				return refreshErr
			}
			return fn(c.authContext(ctx))
		}
	}
	return err
}

// ExecStart starts a command execution.
func (c *TaskCommandRouterClient) ExecStart(ctx context.Context, request *pb.TaskExecStartRequest) (*pb.TaskExecStartResponse, error) {
	var resp *pb.TaskExecStartResponse
	_, err := callWithRetriesOnTransientErrors(ctx, func() (struct{}, error) {
		callErr := c.callWithAuthRetry(ctx, func(authCtx context.Context) error {
			var err error
			resp, err = c.stub.TaskExecStart(authCtx, request)
			return err
		})
		return struct{}{}, callErr
	}, defaultRetryOptions())
	return resp, err
}

// ExecStdinWrite writes data to stdin of an exec.
func (c *TaskCommandRouterClient) ExecStdinWrite(ctx context.Context, taskID, execID string, offset uint64, data []byte, eof bool) error {
	request := pb.TaskExecStdinWriteRequest_builder{
		TaskId: taskID,
		ExecId: execID,
		Offset: offset,
		Data:   data,
		Eof:    eof,
	}.Build()

	_, err := callWithRetriesOnTransientErrors(ctx, func() (struct{}, error) {
		callErr := c.callWithAuthRetry(ctx, func(authCtx context.Context) error {
			_, err := c.stub.TaskExecStdinWrite(authCtx, request)
			return err
		})
		return struct{}{}, callErr
	}, defaultRetryOptions())
	return err
}

// ExecPoll polls for the exit status of an exec.
func (c *TaskCommandRouterClient) ExecPoll(ctx context.Context, taskID, execID string, deadline *time.Time) (*pb.TaskExecPollResponse, error) {
	request := pb.TaskExecPollRequest_builder{
		TaskId: taskID,
		ExecId: execID,
	}.Build()

	if deadline != nil && time.Now().After(*deadline) {
		return nil, fmt.Errorf("deadline exceeded while polling for exec %s", execID)
	}

	opts := defaultRetryOptions()
	opts.Deadline = deadline

	resp, err := callWithRetriesOnTransientErrors(ctx, func() (*pb.TaskExecPollResponse, error) {
		var resp *pb.TaskExecPollResponse
		callErr := c.callWithAuthRetry(ctx, func(authCtx context.Context) error {
			var err error
			resp, err = c.stub.TaskExecPoll(authCtx, request)
			return err
		})
		return resp, callErr
	}, opts)

	if err != nil {
		st, ok := status.FromError(err)
		if (ok && st.Code() == codes.DeadlineExceeded) || errors.Is(err, errDeadlineExceeded) {
			return nil, fmt.Errorf("deadline exceeded while polling for exec %s", execID)
		}
	}
	return resp, err
}

// ExecWait waits for an exec to complete and returns the exit code.
func (c *TaskCommandRouterClient) ExecWait(ctx context.Context, taskID, execID string, deadline *time.Time) (*pb.TaskExecWaitResponse, error) {
	request := pb.TaskExecWaitRequest_builder{
		TaskId: taskID,
		ExecId: execID,
	}.Build()

	if deadline != nil && time.Now().After(*deadline) {
		return nil, fmt.Errorf("deadline exceeded while waiting for exec %s", execID)
	}

	opts := retryOptions{
		BaseDelay:   1 * time.Second, // Retry after 1s since total time is expected to be long.
		DelayFactor: 1,               // Fixed delay.
		MaxRetries:  nil,             // Retry forever.
		Deadline:    deadline,
	}

	resp, err := callWithRetriesOnTransientErrors(ctx, func() (*pb.TaskExecWaitResponse, error) {
		var resp *pb.TaskExecWaitResponse
		callErr := c.callWithAuthRetry(ctx, func(authCtx context.Context) error {
			// Set a per-call timeout of 60 seconds
			callCtx, cancel := context.WithTimeout(authCtx, 60*time.Second)
			defer cancel()
			var err error
			resp, err = c.stub.TaskExecWait(callCtx, request)
			return err
		})
		return resp, callErr
	}, opts)

	if err != nil {
		st, ok := status.FromError(err)
		if (ok && st.Code() == codes.DeadlineExceeded) || errors.Is(err, errDeadlineExceeded) {
			return nil, fmt.Errorf("deadline exceeded while waiting for exec %s", execID)
		}
	}
	return resp, err
}

// stdioReadResult represents a result from the stdio read stream.
type stdioReadResult struct {
	Response *pb.TaskExecStdioReadResponse
	Err      error
}

// ExecStdioRead reads stdout or stderr from an exec.
// The returned channel will be closed when the stream ends or an error occurs.
func (c *TaskCommandRouterClient) ExecStdioRead(
	ctx context.Context,
	taskID, execID string,
	fd pb.FileDescriptor,
	deadline *time.Time,
) <-chan stdioReadResult {
	resultCh := make(chan stdioReadResult)

	go func() {
		defer close(resultCh)

		var srFd pb.TaskExecStdioFileDescriptor
		switch fd {
		case pb.FileDescriptor_FILE_DESCRIPTOR_STDOUT:
			srFd = pb.TaskExecStdioFileDescriptor_TASK_EXEC_STDIO_FILE_DESCRIPTOR_STDOUT
		case pb.FileDescriptor_FILE_DESCRIPTOR_STDERR:
			srFd = pb.TaskExecStdioFileDescriptor_TASK_EXEC_STDIO_FILE_DESCRIPTOR_STDERR
		case pb.FileDescriptor_FILE_DESCRIPTOR_INFO, pb.FileDescriptor_FILE_DESCRIPTOR_UNSPECIFIED:
			resultCh <- stdioReadResult{Err: fmt.Errorf("unsupported file descriptor: %v", fd)}
			return
		default:
			resultCh <- stdioReadResult{Err: fmt.Errorf("invalid file descriptor: %v", fd)}
			return
		}

		c.streamStdio(ctx, resultCh, taskID, execID, srFd, deadline)
	}()

	return resultCh
}

// MountDirectory mounts an image at a directory in the container.
func (c *TaskCommandRouterClient) MountDirectory(ctx context.Context, request *pb.TaskMountDirectoryRequest) error {
	_, err := callWithRetriesOnTransientErrors(ctx, func() (struct{}, error) {
		callErr := c.callWithAuthRetry(ctx, func(authCtx context.Context) error {
			_, err := c.stub.TaskMountDirectory(authCtx, request)
			return err
		})
		return struct{}{}, callErr
	}, defaultRetryOptions())
	return err
}

// SnapshotDirectory snapshots a directory into a new image.
func (c *TaskCommandRouterClient) SnapshotDirectory(ctx context.Context, request *pb.TaskSnapshotDirectoryRequest) (*pb.TaskSnapshotDirectoryResponse, error) {
	resp, err := callWithRetriesOnTransientErrors(ctx, func() (*pb.TaskSnapshotDirectoryResponse, error) {
		var resp *pb.TaskSnapshotDirectoryResponse
		callErr := c.callWithAuthRetry(ctx, func(authCtx context.Context) error {
			var err error
			resp, err = c.stub.TaskSnapshotDirectory(authCtx, request)
			return err
		})
		return resp, callErr
	}, defaultRetryOptions())
	return resp, err
}

func (c *TaskCommandRouterClient) streamStdio(
	ctx context.Context,
	resultCh chan<- stdioReadResult,
	taskID, execID string,
	fd pb.TaskExecStdioFileDescriptor,
	deadline *time.Time,
) {
	if deadline != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, *deadline)
		defer cancel()
	}

	var offset int64
	delay := 10 * time.Millisecond
	delayFactor := 2.0
	numRetriesRemaining := 10
	didAuthRetry := false

	for {
		if ctx.Err() != nil {
			if deadline != nil && ctx.Err() == context.DeadlineExceeded {
				resultCh <- stdioReadResult{Err: fmt.Errorf("deadline exceeded while streaming stdio for exec %s", execID)}
			} else {
				resultCh <- stdioReadResult{Err: ctx.Err()}
			}
			return
		}

		callCtx := c.authContext(ctx)

		request := pb.TaskExecStdioReadRequest_builder{
			TaskId:         taskID,
			ExecId:         execID,
			Offset:         uint64(offset),
			FileDescriptor: fd,
		}.Build()

		stream, err := c.stub.TaskExecStdioRead(callCtx, request)
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated && !didAuthRetry {
				if refreshErr := c.refreshJwt(ctx); refreshErr != nil {
					resultCh <- stdioReadResult{Err: refreshErr}
					return
				}
				didAuthRetry = true
				continue
			}
			if _, retryable := commandRouterRetryableCodes[status.Code(err)]; retryable && numRetriesRemaining > 0 {
				if deadline != nil && time.Until(*deadline) <= delay {
					resultCh <- stdioReadResult{Err: fmt.Errorf("deadline exceeded while streaming stdio for exec %s", execID)}
					return
				}
				c.logger.DebugContext(ctx, "Retrying stdio read with delay", "delay", delay, "error", err)
				time.Sleep(delay)
				delay = time.Duration(float64(delay) * delayFactor)
				numRetriesRemaining--
				continue
			}
			resultCh <- stdioReadResult{Err: err}
			return
		}

		for {
			item, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated && !didAuthRetry {
					if refreshErr := c.refreshJwt(ctx); refreshErr != nil {
						resultCh <- stdioReadResult{Err: refreshErr}
						return
					}
					didAuthRetry = true
					break
				}
				if _, retryable := commandRouterRetryableCodes[status.Code(err)]; retryable && numRetriesRemaining > 0 {
					if deadline != nil && time.Until(*deadline) <= delay {
						resultCh <- stdioReadResult{Err: fmt.Errorf("deadline exceeded while streaming stdio for exec %s", execID)}
						return
					}
					c.logger.DebugContext(ctx, "Retrying stdio read with delay", "delay", delay, "error", err)
					time.Sleep(delay)
					delay = time.Duration(float64(delay) * delayFactor)
					numRetriesRemaining--
					break
				}
				resultCh <- stdioReadResult{Err: err}
				return
			}

			if didAuthRetry {
				didAuthRetry = false
			}
			delay = 10 * time.Millisecond
			offset += int64(len(item.GetData()))

			resultCh <- stdioReadResult{Response: item}
		}
	}
}
