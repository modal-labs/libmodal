package modal

import (
	"context"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

// SecretService provides Secret related operations.
type SecretService interface {
	FromName(ctx context.Context, name string, params *SecretFromNameParams) (*Secret, error)
	FromMap(ctx context.Context, keyValuePairs map[string]string, params *SecretFromMapParams) (*Secret, error)
}

// secretServiceImpl is the real implementation of SecretService.
type secretServiceImpl struct{ client *Client }

// Secret represents a Modal Secret.
type Secret struct {
	SecretID string
	Name     string
}

// SecretFromNameParams are options for finding Modal Secrets.
type SecretFromNameParams struct {
	Environment  string
	RequiredKeys []string
}

// FromName references a Secret by its name.
func (s *secretServiceImpl) FromName(ctx context.Context, name string, params *SecretFromNameParams) (*Secret, error) {
	if params == nil {
		params = &SecretFromNameParams{}
	}

	resp, err := s.client.cpClient.SecretGetOrCreate(ctx, pb.SecretGetOrCreateRequest_builder{
		DeploymentName:  name,
		EnvironmentName: environmentName(params.Environment, s.client.profile),
		RequiredKeys:    params.RequiredKeys,
	}.Build())

	if err != nil {
		return nil, err
	}

	return &Secret{SecretID: resp.GetSecretId(), Name: name}, nil
}

// SecretFromMapParams are options for creating a Secret from a key/value map.
type SecretFromMapParams struct {
	Environment string
}

// FromMap creates a Secret from a map of key-value pairs.
func (s *secretServiceImpl) FromMap(ctx context.Context, keyValuePairs map[string]string, params *SecretFromMapParams) (*Secret, error) {
	if params == nil {
		params = &SecretFromMapParams{}
	}

	resp, err := s.client.cpClient.SecretGetOrCreate(ctx, pb.SecretGetOrCreateRequest_builder{
		ObjectCreationType: pb.ObjectCreationType_OBJECT_CREATION_TYPE_EPHEMERAL,
		EnvDict:            keyValuePairs,
		EnvironmentName:    environmentName(params.Environment, s.client.profile),
	}.Build())
	if err != nil {
		return nil, err
	}
	return &Secret{SecretID: resp.GetSecretId()}, nil
}
