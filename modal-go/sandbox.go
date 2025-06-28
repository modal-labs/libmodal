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
}

// Sandbox represents a Modal sandbox, which can run commands and manage
// input/output streams for a remote process.
type Sandbox struct {
	SandboxId string
	Stdin     io.WriteCloser
	Stdout    io.ReadCloser
	Stderr    io.ReadCloser

	ctx    context.Context
	taskId string
}

// newSandbox creates a new Sandbox object from ID.
func newSandbox(ctx context.Context, sandboxId string) *Sandbox {
	sb := &Sandbox{SandboxId: sandboxId, ctx: ctx}
	sb.Stdin = inputStreamSb(ctx, sandboxId)
	sb.Stdout = outputStreamSb(ctx, sandboxId, pb.FileDescriptor_FILE_DESCRIPTOR_STDOUT)
	sb.Stderr = outputStreamSb(ctx, sandboxId, pb.FileDescriptor_FILE_DESCRIPTOR_STDERR)
	return sb
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
	resp, err := client.ContainerExec(sb.ctx, pb.ContainerExecRequest_builder{
		TaskId:      sb.taskId,
		Command:     command,
		Workdir:     workdir,
		TimeoutSecs: uint32(opts.Timeout.Seconds()),
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
func (sb *Sandbox) Wait() (int32, error) {
	for {
		resp, err := client.SandboxWait(sb.ctx, pb.SandboxWaitRequest_builder{
			SandboxId: sb.SandboxId,
			Timeout:   55,
		}.Build())
		if err != nil {
			return 0, err
		}
		if resp.GetResult() != nil {
			return resp.GetResult().GetExitcode(), nil
		}
	}
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
func (cp *ContainerProcess) Wait() (int32, error) {
	for {
		resp, err := client.ContainerExecWait(cp.ctx, pb.ContainerExecWaitRequest_builder{
			ExecId:  cp.execId,
			Timeout: 55,
		}.Build())
		if err != nil {
			return 0, err
		}
		if resp.GetCompleted() {
			return resp.GetExitCode(), nil
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
