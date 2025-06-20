package modal

import (
	"context"
	"fmt"
	"io"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

// File represents an open file in the sandbox filesystem.
// It implements io.Reader, io.Writer, io.Seeker, and io.Closer interfaces.
type File struct {
	fileDescriptor string
	taskId         string
	ctx            context.Context
}

// newFile creates a new File object.
func newFile(ctx context.Context, fileDescriptor, taskId string) *File {
	return &File{
		fileDescriptor: fileDescriptor,
		taskId:         taskId,
		ctx:            ctx,
	}
}

// Read reads up to len(p) bytes from the file into p.
// It returns the number of bytes read and any error encountered.
func (f *File) Read(p []byte) (n int, err error) {
	nBytes := uint32(len(p))
	resp, err := client.ContainerFilesystemExec(f.ctx, pb.ContainerFilesystemExecRequest_builder{
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
func (f *File) Write(p []byte) (n int, err error) {
	_, err = client.ContainerFilesystemExec(f.ctx, pb.ContainerFilesystemExecRequest_builder{
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

// Seek sets the offset for the next Read or Write on file to offset,
// interpreted according to whence: 0 means relative to the origin of the file,
// 1 means relative to the current offset, and 2 means relative to the end.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	var seekWhence pb.SeekWhence
	switch whence {
	case 0:
		seekWhence = pb.SeekWhence_SEEK_SET
	case 1:
		seekWhence = pb.SeekWhence_SEEK_CUR
	case 2:
		seekWhence = pb.SeekWhence_SEEK_END
	default:
		return 0, fmt.Errorf("invalid whence value: %d", whence)
	}

	offset32 := int32(offset)
	_, err := client.ContainerFilesystemExec(f.ctx, pb.ContainerFilesystemExecRequest_builder{
		FileSeekRequest: pb.ContainerFileSeekRequest_builder{
			FileDescriptor: f.fileDescriptor,
			Offset:         offset32,
			Whence:         seekWhence,
		}.Build(),
		TaskId: f.taskId,
	}.Build())
	if err != nil {
		return 0, err
	}
	return offset, nil
}

// Flush flushes any buffered data to the file.
func (f *File) Flush() error {
	_, err := client.ContainerFilesystemExec(f.ctx, pb.ContainerFilesystemExecRequest_builder{
		FileFlushRequest: pb.ContainerFileFlushRequest_builder{
			FileDescriptor: f.fileDescriptor,
		}.Build(),
		TaskId: f.taskId,
	}.Build())
	return err
}

// Close closes the file, rendering it unusable for I/O.
func (f *File) Close() error {
	_, err := client.ContainerFilesystemExec(f.ctx, pb.ContainerFilesystemExecRequest_builder{
		FileCloseRequest: pb.ContainerFileCloseRequest_builder{
			FileDescriptor: f.fileDescriptor,
		}.Build(),
		TaskId: f.taskId,
	}.Build())
	return err
}

// ReadAll reads from the file until an error or EOF and returns the data it read.
func (f *File) ReadAll() ([]byte, error) {
	return io.ReadAll(f)
}

// WriteString writes the contents of the string s to the file.
func (f *File) WriteString(s string) (n int, err error) {
	return f.Write([]byte(s))
}
