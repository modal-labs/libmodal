package modal

import (
	"context"
	"fmt"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProxyService provides Proxy related operations.
type ProxyService struct{ client *Client }

// Proxy represents a Modal Proxy.
type Proxy struct {
	ProxyId string
}

// ProxyFromNameOptions are options for looking up a Modal Proxy.
type ProxyFromNameOptions struct {
	Environment string
}

// FromName references a modal.Proxy by its name.
func (s *ProxyService) FromName(ctx context.Context, name string, options *ProxyFromNameOptions) (*Proxy, error) {
	if options == nil {
		options = &ProxyFromNameOptions{}
	}

	resp, err := s.client.cpClient.ProxyGet(ctx, pb.ProxyGetRequest_builder{
		Name:            name,
		EnvironmentName: environmentName(options.Environment, s.client.profile),
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

	return &Proxy{ProxyId: resp.GetProxy().GetProxyId()}, nil
}
