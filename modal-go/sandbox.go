package modal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"
	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StdioBehavior defines how the standard input/output/error streams should behave.
type StdioBehavior string

const (
	// Pipe allows the sandbox to pipe the streams.
	Pipe StdioBehavior = "pipe"
	// Ignore ignores the streams, meaning they will not be available.
	Ignore StdioBehavior = "ignore"
)

// ExecOptions defines options for executing commands in a sandbox.
type ExecOptions struct {
	// Stdout defines whether to pipe or ignore standard output.
	Stdout StdioBehavior
	// Stderr defines whether to pipe or ignore standard error.
	Stderr StdioBehavior
	// Workdir is the working directory to run the command in.
	Workdir string
	// Timeout is the timeout for command execution. Defaults to 0 (no timeout).
	Timeout time.Duration
	// Secrets with environment variables for the command.
	Secrets []*Secret
}

// Tunnel represents a port forwarded from within a running Modal sandbox.
type Tunnel struct {
	Host            string // The public hostname for the tunnel
	Port            int    // The public port for the tunnel
	UnencryptedHost string // The unencrypted hostname (if applicable)
	UnencryptedPort int    // The unencrypted port (if applicable)
}

// Get the public HTTPS URL of the forwarded port.
func (t *Tunnel) URL() string {
	if t.Port == 443 {
		return fmt.Sprintf("https://%s", t.Host)
	}
	return fmt.Sprintf("https://%s:%d", t.Host, t.Port)
}

// Get the public TLS socket as a (host, port) tuple.
func (t *Tunnel) TLSSocket() (string, int) {
	return t.Host, t.Port
}

// Get the public TCP socket as a (host, port) tuple.
func (t *Tunnel) TCPSocket() (string, int, error) {
	if t.UnencryptedHost == "" || t.UnencryptedPort == 0 {
		return "", 0, InvalidError{Exception: "This tunnel is not configured for unencrypted TCP."}
	}
	return t.UnencryptedHost, t.UnencryptedPort, nil
}

// Sandbox represents a Modal sandbox, which can run commands and manage
// input/output streams for a remote process.
type Sandbox struct {
	SandboxId string
	Stdin     io.WriteCloser
	Stdout    io.ReadCloser
	Stderr    io.ReadCloser

	ctx     context.Context
	taskId  string
	tunnels map[int]*Tunnel
}

// newSandbox creates a new Sandbox object from ID.
func newSandbox(ctx context.Context, sandboxId string) *Sandbox {
	sb := &Sandbox{SandboxId: sandboxId, ctx: ctx}
	sb.Stdin = inputStreamSb(ctx, sandboxId)
	sb.Stdout = outputStreamSb(ctx, sandboxId, pb.FileDescriptor_FILE_DESCRIPTOR_STDOUT)
	sb.Stderr = outputStreamSb(ctx, sandboxId, pb.FileDescriptor_FILE_DESCRIPTOR_STDERR)
	return sb
}

// SandboxFromId returns a running Sandbox object from an ID.
func SandboxFromId(ctx context.Context, sandboxId string) (*Sandbox, error) {
	ctx, err := clientContext(ctx)
	if err != nil {
		return nil, err
	}

	_, err = client.SandboxWait(ctx, pb.SandboxWaitRequest_builder{
		SandboxId: sandboxId,
		Timeout:   0,
	}.Build())
	if status, ok := status.FromError(err); ok && status.Code() == codes.NotFound {
		return nil, NotFoundError{fmt.Sprintf("Sandbox with id: '%s' not found", sandboxId)}
	}
	if err != nil {
		return nil, err
	}
	return newSandbox(ctx, sandboxId), nil
}

// Exec runs a command in the sandbox and returns text streams.
func (sb *Sandbox) Exec(command []string, opts ExecOptions) (*ContainerProcess, error) {
	if err := sb.ensureTaskId(); err != nil {
		return nil, err
	}
	var workdir *string
	if opts.Workdir != "" {
		workdir = &opts.Workdir
	}
	secretIds := []string{}
	for _, secret := range opts.Secrets {
		if secret != nil {
			secretIds = append(secretIds, secret.SecretId)
		}
	}

	resp, err := client.ContainerExec(sb.ctx, pb.ContainerExecRequest_builder{
		TaskId:      sb.taskId,
		Command:     command,
		Workdir:     workdir,
		TimeoutSecs: uint32(opts.Timeout.Seconds()),
		SecretIds:   secretIds,
	}.Build())
	if err != nil {
		return nil, err
	}
	return newContainerProcess(sb.ctx, resp.GetExecId(), opts), nil
}

// Open opens a file in the sandbox filesystem.
// The mode parameter follows the same conventions as os.OpenFile:
// "r" for read-only, "w" for write-only (truncates), "a" for append, etc.
func (sb *Sandbox) Open(path, mode string) (*SandboxFile, error) {
	if err := sb.ensureTaskId(); err != nil {
		return nil, err
	}

	_, resp, err := runFilesystemExec(sb.ctx, pb.ContainerFilesystemExecRequest_builder{
		FileOpenRequest: pb.ContainerFileOpenRequest_builder{
			Path: path,
			Mode: mode,
		}.Build(),
		TaskId: sb.taskId,
	}.Build(), nil)

	if err != nil {
		return nil, err
	}

	return &SandboxFile{
		fileDescriptor: resp.GetFileDescriptor(),
		taskId:         sb.taskId,
		ctx:            sb.ctx,
	}, nil
}

func (sb *Sandbox) ensureTaskId() error {
	if sb.taskId == "" {
		resp, err := client.SandboxGetTaskId(sb.ctx, pb.SandboxGetTaskIdRequest_builder{
			SandboxId: sb.SandboxId,
		}.Build())
		if err != nil {
			return err
		}
		if resp.GetTaskId() == "" {
			return fmt.Errorf("Sandbox %s does not have a task ID, it may not be running", sb.SandboxId)
		}
		if resp.GetTaskResult() != nil {
			return fmt.Errorf("Sandbox %s has already completed with result: %v", sb.SandboxId, resp.GetTaskResult())
		}
		sb.taskId = resp.GetTaskId()
	}
	return nil
}

// Terminate stops the sandbox.
func (sb *Sandbox) Terminate() error {
	_, err := client.SandboxTerminate(sb.ctx, pb.SandboxTerminateRequest_builder{
		SandboxId: sb.SandboxId,
	}.Build())
	if err != nil {
		return err
	}
	sb.taskId = ""
	return nil
}

// Wait blocks until the sandbox exits.
func (sb *Sandbox) Wait() (int, error) {
	for {
		resp, err := client.SandboxWait(sb.ctx, pb.SandboxWaitRequest_builder{
			SandboxId: sb.SandboxId,
			Timeout:   55,
		}.Build())
		if err != nil {
			return 0, err
		}
		if resp.GetResult() != nil {
			returnCode := getReturnCode(resp.GetResult())
			if returnCode != nil {
				return *returnCode, nil
			}
			return 0, nil
		}
	}
}

// Tunnels gets Tunnel metadata for the sandbox.
// Returns SandboxTimeoutError if the tunnels are not available after the timeout.
// Returns a map of Tunnel objects keyed by the container port.
func (sb *Sandbox) Tunnels(timeout time.Duration) (map[int]*Tunnel, error) {
	if sb.tunnels != nil {
		return sb.tunnels, nil
	}

	resp, err := client.SandboxGetTunnels(sb.ctx, pb.SandboxGetTunnelsRequest_builder{
		SandboxId: sb.SandboxId,
		Timeout:   float32(timeout.Seconds()),
	}.Build())
	if err != nil {
		return nil, err
	}

	if resp.GetResult() != nil && resp.GetResult().GetStatus() == pb.GenericResult_GENERIC_STATUS_TIMEOUT {
		return nil, SandboxTimeoutError{Exception: "Sandbox operation timed out"}
	}

	sb.tunnels = make(map[int]*Tunnel)
	for _, t := range resp.GetTunnels() {
		sb.tunnels[int(t.GetContainerPort())] = &Tunnel{
			Host:            t.GetHost(),
			Port:            int(t.GetPort()),
			UnencryptedHost: t.GetUnencryptedHost(),
			UnencryptedPort: int(t.GetUnencryptedPort()),
		}
	}

	return sb.tunnels, nil
}

// Snapshot the filesystem of the Sandbox.
// Returns an Image object which can be used to spawn a new Sandbox with the same filesystem.
func (sb *Sandbox) SnapshotFilesystem(timeout time.Duration) (*Image, error) {
	resp, err := client.SandboxSnapshotFs(sb.ctx, pb.SandboxSnapshotFsRequest_builder{
		SandboxId: sb.SandboxId,
		Timeout:   float32(timeout.Seconds()),
	}.Build())
	if err != nil {
		return nil, err
	}

	if resp.GetResult() != nil && resp.GetResult().GetStatus() != pb.GenericResult_GENERIC_STATUS_SUCCESS {
		return nil, ExecutionError{Exception: fmt.Sprintf("Sandbox snapshot failed: %s", resp.GetResult().GetException())}
	}

	if resp.GetImageId() == "" {
		return nil, ExecutionError{Exception: "Sandbox snapshot response missing image ID"}
	}

	return &Image{ImageId: resp.GetImageId(), ctx: sb.ctx}, nil
}

// Poll checks if the Sandbox has finished running.
// Returns nil if the Sandbox is still running, else returns the exit code.
func (sb *Sandbox) Poll() (*int, error) {
	resp, err := client.SandboxWait(sb.ctx, pb.SandboxWaitRequest_builder{
		SandboxId: sb.SandboxId,
		Timeout:   0,
	}.Build())
	if err != nil {
		return nil, err
	}

	return getReturnCode(resp.GetResult()), nil
}

func getReturnCode(result *pb.GenericResult) *int {
	if result == nil || result.GetStatus() == pb.GenericResult_GENERIC_STATUS_UNSPECIFIED {
		return nil
	}

	// Statuses are converted to exitcodes so we can conform to subprocess API.
	var exitCode int
	switch result.GetStatus() {
	case pb.GenericResult_GENERIC_STATUS_TIMEOUT:
		exitCode = 124
	case pb.GenericResult_GENERIC_STATUS_TERMINATED:
		exitCode = 137
	default:
		exitCode = int(result.GetExitcode())
	}

	return &exitCode
}

// ContainerProcess represents a process running in a Modal container, allowing
// interaction with its standard input/output/error streams.
//
// It is created by executing a command in a sandbox.
type ContainerProcess struct {
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser

	ctx    context.Context
	execId string
}

func newContainerProcess(ctx context.Context, execId string, opts ExecOptions) *ContainerProcess {
	stdoutBehavior := Pipe
	stderrBehavior := Pipe
	if opts.Stdout != "" {
		stdoutBehavior = opts.Stdout
	}
	if opts.Stderr != "" {
		stderrBehavior = opts.Stderr
	}

	cp := &ContainerProcess{execId: execId, ctx: ctx}
	cp.Stdin = inputStreamCp(ctx, execId)

	cp.Stdout = outputStreamCp(ctx, execId, pb.FileDescriptor_FILE_DESCRIPTOR_STDOUT)
	if stdoutBehavior == Ignore {
		cp.Stdout.Close()
		cp.Stdout = io.NopCloser(bytes.NewReader(nil))
	}
	cp.Stderr = outputStreamCp(ctx, execId, pb.FileDescriptor_FILE_DESCRIPTOR_STDERR)
	if stderrBehavior == Ignore {
		cp.Stderr.Close()
		cp.Stderr = io.NopCloser(bytes.NewReader(nil))
	}

	return cp
}

// Wait blocks until the container process exits and returns its exit code.
func (cp *ContainerProcess) Wait() (int, error) {
	for {
		resp, err := client.ContainerExecWait(cp.ctx, pb.ContainerExecWaitRequest_builder{
			ExecId:  cp.execId,
			Timeout: 55,
		}.Build())
		if err != nil {
			return 0, err
		}
		if resp.GetCompleted() {
			return int(resp.GetExitCode()), nil
		}
	}
}

func inputStreamSb(ctx context.Context, sandboxId string) io.WriteCloser {
	return &sbStdin{sandboxId: sandboxId, ctx: ctx, index: 1}
}

type sbStdin struct {
	sandboxId string
	ctx       context.Context // context for the sandbox operations

	mu    sync.Mutex // protects index
	index uint32
}

func (s *sbStdin) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	index := s.index
	s.index++
	_, err = client.SandboxStdinWrite(s.ctx, pb.SandboxStdinWriteRequest_builder{
		SandboxId: s.sandboxId,
		Input:     p,
		Index:     index,
	}.Build())
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (s *sbStdin) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := client.SandboxStdinWrite(s.ctx, pb.SandboxStdinWriteRequest_builder{
		SandboxId: s.sandboxId,
		Index:     s.index,
		Eof:       true,
	}.Build())
	return err
}

func inputStreamCp(ctx context.Context, execId string) io.WriteCloser {
	return &cpStdin{execId: execId, messageIndex: 1, ctx: ctx}
}

type cpStdin struct {
	execId       string
	messageIndex uint64
	ctx          context.Context // context for the exec operations
}

func (c *cpStdin) Write(p []byte) (n int, err error) {
	_, err = client.ContainerExecPutInput(c.ctx, pb.ContainerExecPutInputRequest_builder{
		ExecId: c.execId,
		Input: pb.RuntimeInputMessage_builder{
			Message:      p,
			MessageIndex: c.messageIndex,
		}.Build(),
	}.Build())
	if err != nil {
		return 0, err
	}
	c.messageIndex++
	return len(p), nil
}

func (c *cpStdin) Close() error {
	_, err := client.ContainerExecPutInput(c.ctx, pb.ContainerExecPutInputRequest_builder{
		ExecId: c.execId,
		Input: pb.RuntimeInputMessage_builder{
			MessageIndex: c.messageIndex,
			Eof:          true,
		}.Build(),
	}.Build())
	return err
}

func outputStreamSb(ctx context.Context, sandboxId string, fd pb.FileDescriptor) io.ReadCloser {
	pr, pw := nio.Pipe(buffer.New(64 * 1024))
	go func() {
		defer pw.Close()
		lastIndex := "0-0"
		completed := false
		retries := 10
		for !completed {
			stream, err := client.SandboxGetLogs(ctx, pb.SandboxGetLogsRequest_builder{
				SandboxId:      sandboxId,
				FileDescriptor: fd,
				Timeout:        55,
				LastEntryId:    lastIndex,
			}.Build())
			if err != nil {
				if isRetryableGrpc(err) && retries > 0 {
					retries--
					continue
				}
				pw.CloseWithError(fmt.Errorf("error getting output stream: %w", err))
				return
			}
			for {
				batch, err := stream.Recv()
				if err != nil {
					if err != io.EOF {
						if isRetryableGrpc(err) && retries > 0 {
							retries--
						} else {
							pw.CloseWithError(fmt.Errorf("error getting output stream: %w", err))
							return
						}
					}
					break // we need to retry, either from an EOF or gRPC error
				}
				lastIndex = batch.GetEntryId()
				for _, item := range batch.GetItems() {
					// On error, writer has been closed. Still consume the rest of the channel.
					pw.Write([]byte(item.GetData()))
				}
				if batch.GetEof() {
					completed = true
					break
				}
			}
		}
	}()
	return pr
}

func outputStreamCp(ctx context.Context, execId string, fd pb.FileDescriptor) io.ReadCloser {
	pr, pw := nio.Pipe(buffer.New(64 * 1024))
	go func() {
		defer pw.Close()
		var lastIndex uint64
		completed := false
		retries := 10
		for !completed {
			stream, err := client.ContainerExecGetOutput(ctx, pb.ContainerExecGetOutputRequest_builder{
				ExecId:         execId,
				FileDescriptor: fd,
				Timeout:        55,
				GetRawBytes:    true,
				LastBatchIndex: lastIndex,
			}.Build())
			if err != nil {
				if isRetryableGrpc(err) && retries > 0 {
					retries--
					continue
				}
				pw.CloseWithError(fmt.Errorf("error getting output stream: %w", err))
				return
			}
			for {
				batch, err := stream.Recv()
				if err != nil {
					if err != io.EOF {
						if isRetryableGrpc(err) && retries > 0 {
							retries--
						} else {
							pw.CloseWithError(fmt.Errorf("error getting output stream: %w", err))
							return
						}
					}
					break // we need to retry, either from an EOF or gRPC error
				}
				lastIndex = batch.GetBatchIndex()
				for _, item := range batch.GetItems() {
					// On error, writer has been closed. Still consume the rest of the channel.
					pw.Write(item.GetMessageBytes())
				}
				if batch.HasExitCode() {
					completed = true
					break
				}
			}
		}
	}()
	return pr
}
