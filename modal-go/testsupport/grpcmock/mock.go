package grpcmock

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	modal "github.com/modal-labs/libmodal/modal-go"
	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

// unaryHandler handles a single unary RPC request and returns a response.
type unaryHandler func(proto.Message) (proto.Message, error)

type Mock struct {
	// mu guards access to internal state.
	mu sync.Mutex
	// methodHandlerQueues maps short RPC names to FIFO queues of handlers.
	methodHandlerQueues map[string][]unaryHandler
	// conn is the fake ClientConn used by the SDK client.
	conn *mockClientConn
}

// Install swaps the SDK client factory to use a mock gRPC connection.
// Register the returned cleanup function with t.Cleanup.
func Install() (*Mock, func()) {
	m := &Mock{methodHandlerQueues: map[string][]unaryHandler{}}
	m.conn = &mockClientConn{mock: m}

	restore := modal.SetClientFactoryForTesting(func(profile modal.Profile) (grpc.ClientConnInterface, pb.ModalClientClient, error) {
		return m.conn, pb.NewModalClientClient(m.conn), nil
	})
	cleanup := func() {
		if err := m.AssertExhausted(); err != nil {
			panic(err)
		}
		restore()
	}
	return m, cleanup
}

// HandleUnary registers a typed handler for a unary RPC, e.g. "/FunctionGetCurrentStats".
func HandleUnary[Req proto.Message, Resp proto.Message](m *Mock, rpc string, handler func(Req) (Resp, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	name := shortName(rpc)
	q := m.methodHandlerQueues[name]
	wrapped := unaryHandler(func(in proto.Message) (proto.Message, error) {
		req, ok := any(in).(Req)
		if !ok {
			return nil, fmt.Errorf("grpcmock: request type mismatch for %s: expected %T, got %T", name, *new(Req), in)
		}
		resp, err := handler(req)
		if err != nil {
			return nil, err
		}
		var out proto.Message = resp
		return out, nil
	})
	m.methodHandlerQueues[name] = append(q, wrapped)
}

// AssertExhausted errors unless all registered mock expectations have been consumed.
func (m *Mock) AssertExhausted() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var outstanding []string
	for k, q := range m.methodHandlerQueues {
		if len(q) > 0 {
			outstanding = append(outstanding, fmt.Sprintf("%s: %d remaining", k, len(q)))
		}
	}
	if len(outstanding) > 0 {
		return fmt.Errorf("not all expected gRPC calls were made:\n- %s", strings.Join(outstanding, "\n- "))
	}
	return nil
}

// mockClientConn implements grpc.ClientConnInterface for unary calls.
type mockClientConn struct{ mock *Mock }

// Invoke implements grpc.ClientConnInterface.Invoke for unary RPCs.
func (c *mockClientConn) Invoke(ctx context.Context, method string, in, out any, opts ...grpc.CallOption) error {
	name := shortName(method)
	handler, err := c.dequeueNextHandler(name)
	if err != nil {
		return err
	}
	resp, err := handler(in.(proto.Message))
	if err != nil {
		return err
	}
	if resp != nil {
		if outMsg, ok := out.(proto.Message); ok {
			proto.Merge(outMsg, resp)
		} else {
			return fmt.Errorf("grpcmock: response cannot be written into type %T", out)
		}
	}
	return nil
}

// NewStream returns an error because streaming RPCs are not supported yet.
func (c *mockClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("grpcmock: streaming not implemented for %s", shortName(method))
}

func (c *mockClientConn) dequeueNextHandler(method string) (unaryHandler, error) {
	c.mock.mu.Lock()
	defer c.mock.mu.Unlock()
	q := c.mock.methodHandlerQueues[method]
	if len(q) == 0 {
		return nil, fmt.Errorf("grpcmock: unexpected gRPC call to %s", method)
	}
	h := q[0]
	c.mock.methodHandlerQueues[method] = q[1:]
	return h, nil
}

func shortName(method string) string {
	if strings.HasPrefix(method, "/") {
		if idx := strings.LastIndex(method, "/"); idx >= 0 && idx+1 < len(method) {
			return method[idx+1:]
		}
	}
	return method
}
