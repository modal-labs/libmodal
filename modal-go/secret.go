package modal

import (
	"context"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

// SecretService provides Secret related operations.
type SecretService struct{ client *Client }

// Secret represents a Modal Secret.
type Secret struct {
	SecretId string
	Name     string
}

// SecretFromNameOptions are options for finding Modal Secrets.
type SecretFromNameOptions struct {
	Environment  string
	RequiredKeys []string
}

// FromName references a Secret by its name.
func (s *SecretService) FromName(ctx context.Context, name string, options *SecretFromNameOptions) (*Secret, error) {
	if options == nil {
		options = &SecretFromNameOptions{}
	}

	resp, err := s.client.cpClient.SecretGetOrCreate(ctx, pb.SecretGetOrCreateRequest_builder{
		DeploymentName:  name,
		EnvironmentName: environmentName(options.Environment, s.client.profile),
		RequiredKeys:    options.RequiredKeys,
	}.Build())

	if err != nil {
		return nil, err
	}

	return &Secret{SecretId: resp.GetSecretId(), Name: name}, nil
}

// SecretFromMapOptions are options for creating a Secret from a key/value map.
type SecretFromMapOptions struct {
	Environment string
}

// FromMap creates a Secret from a map of key-value pairs.
func (s *SecretService) FromMap(ctx context.Context, keyValuePairs map[string]string, options *SecretFromMapOptions) (*Secret, error) {
	if options == nil {
		options = &SecretFromMapOptions{}
	}

	resp, err := s.client.cpClient.SecretGetOrCreate(ctx, pb.SecretGetOrCreateRequest_builder{
		ObjectCreationType: pb.ObjectCreationType_OBJECT_CREATION_TYPE_EPHEMERAL,
		EnvDict:            keyValuePairs,
		EnvironmentName:    environmentName(options.Environment, s.client.profile),
	}.Build())
	if err != nil {
		return nil, err
	}
	return &Secret{SecretId: resp.GetSecretId()}, nil
}
