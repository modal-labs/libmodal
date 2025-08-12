package modal

import (
	"context"
	"fmt"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Proxy represents a Modal proxy.
type Proxy struct {
	ProxyId string

	//lint:ignore U1000 may be used in future
	ctx context.Context
}

// ProxyFromNameOptions are options for looking up a Modal Proxy.
type ProxyFromNameOptions struct {
	Environment string
}

// ProxyFromName references a modal.Proxy by its name.
func ProxyFromName(ctx context.Context, name string, options *ProxyFromNameOptions) (*Proxy, error) {
	var err error
	ctx, err = clientContext(ctx)
	if err != nil {
		return nil, err
	}

	if options == nil {
		options = &ProxyFromNameOptions{}
	}

	resp, err := client.ProxyGet(ctx, pb.ProxyGetRequest_builder{
		Name:            name,
		EnvironmentName: environmentName(options.Environment),
	}.Build())

	if status, ok := status.FromError(err); ok && status.Code() == codes.NotFound {
		return nil, NotFoundError{fmt.Sprintf("Proxy '%s' not found", name)}
	}
	if err != nil {
		return nil, err
	}

	if resp.GetProxy() == nil || resp.GetProxy().GetProxyId() == "" {
		return nil, NotFoundError{fmt.Sprintf("Proxy '%s' not found", name)}
	}

	return &Proxy{ProxyId: resp.GetProxy().GetProxyId(), ctx: ctx}, nil
}
