package modal

import (
	"context"
	"fmt"
	"io"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

// SandboxFile represents an open file in the sandbox filesystem.
// It implements io.Reader, io.Writer, io.Seeker, and io.Closer interfaces.
type SandboxFile struct {
	fileDescriptor string
	taskId         string
	ctx            context.Context
}

// newFile creates a new File object.
func newSandboxFile(ctx context.Context, fileDescriptor, taskId string) *SandboxFile {
	return &SandboxFile{
		fileDescriptor: fileDescriptor,
		taskId:         taskId,
		ctx:            ctx,
	}
}

// Read reads up to len(p) bytes from the file into p.
// It returns the number of bytes read and any error encountered.
func (f *SandboxFile) Read(p []byte) (n int, err error) {
	nBytes := uint32(len(p))
	resp, err := runFilesystemExec(f.ctx, pb.ContainerFilesystemExecRequest_builder{
		FileReadRequest: pb.ContainerFileReadRequest_builder{
			FileDescriptor: f.fileDescriptor,
			N:              &nBytes,
		}.Build(),
		TaskId: f.taskId,
	}.Build())
	if err != nil {
		return 0, err
	}

	// Get the output stream to read the actual data
	outputIterator, err := client.ContainerFilesystemExecGetOutput(f.ctx, pb.ContainerFilesystemExecGetOutputRequest_builder{
		ExecId:  resp.GetExecId(),
		Timeout: 55,
	}.Build())
	if err != nil {
		return 0, err
	}

	var totalRead int
	for {
		batch, err := outputIterator.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return totalRead, err
		}

		for _, chunk := range batch.GetOutput() {
			copyLen := copy(p[totalRead:], chunk)
			totalRead += copyLen
			if totalRead >= len(p) {
				return totalRead, nil
			}
		}

		if batch.GetEof() {
			break
		}
	}

	if totalRead == 0 {
		return 0, io.EOF
	}
	return totalRead, nil
}

// Write writes len(p) bytes from p to the file.
// It returns the number of bytes written and any error encountered.
func (f *SandboxFile) Write(p []byte) (n int, err error) {
	_, err = runFilesystemExec(f.ctx, pb.ContainerFilesystemExecRequest_builder{
		FileWriteRequest: pb.ContainerFileWriteRequest_builder{
			FileDescriptor: f.fileDescriptor,
			Data:           p,
		}.Build(),
		TaskId: f.taskId,
	}.Build())
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Flush flushes any buffered data to the file.
func (f *SandboxFile) Flush() error {
	_, err := runFilesystemExec(f.ctx, pb.ContainerFilesystemExecRequest_builder{
		FileFlushRequest: pb.ContainerFileFlushRequest_builder{
			FileDescriptor: f.fileDescriptor,
		}.Build(),
		TaskId: f.taskId,
	}.Build())
	return err
}

// Close closes the file, rendering it unusable for I/O.
func (f *SandboxFile) Close() error {
	_, err := runFilesystemExec(f.ctx, pb.ContainerFilesystemExecRequest_builder{
		FileCloseRequest: pb.ContainerFileCloseRequest_builder{
			FileDescriptor: f.fileDescriptor,
		}.Build(),
		TaskId: f.taskId,
	}.Build())
	return err
}

func runFilesystemExec(ctx context.Context, req *pb.ContainerFilesystemExecRequest) (*pb.ContainerFilesystemExecResponse, error) {
	resp, err := client.ContainerFilesystemExec(ctx, req)
	if err != nil {
		return nil, err
	}

	retries := 10
	completed := false

	for !completed {
		outputIterator, err := client.ContainerFilesystemExecGetOutput(ctx, pb.ContainerFilesystemExecGetOutputRequest_builder{
			ExecId:  resp.GetExecId(),
			Timeout: 55,
		}.Build())
		if err != nil {
			if retries > 0 {
				retries--
				continue
			}
			return nil, err
		}
		for {
			batch, err := outputIterator.Recv()
			if err == io.EOF {
				completed = true
				break
			}
			// Invalid error
			if err != nil {
				if retries > 0 {
					retries--
					break
				}
				return nil, err
			}

			if batch.GetEof() {
				completed = true
				break
			}

			if batch.GetError() != nil {
				if retries > 0 {
					retries--
					break
				}
				return nil, fmt.Errorf("filesystem exec error: %s", batch.GetError().GetErrorMessage())
			}
		}
	}
	return resp, nil
}
