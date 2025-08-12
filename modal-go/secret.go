package modal

import (
	"context"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

// Secret represents a Modal secret.
type Secret struct {
	SecretId string
	Name     string

	//lint:ignore U1000 may be used in future
	ctx context.Context
}

// SecretFromNameOptions are options for finding Modal secrets.
type SecretFromNameOptions struct {
	Environment  string
	RequiredKeys []string
}

// SecretFromName references a modal.Secret by its name.
func SecretFromName(ctx context.Context, name string, options *SecretFromNameOptions) (*Secret, error) {
	var err error
	ctx, err = clientContext(ctx)
	if err != nil {
		return nil, err
	}

	if options == nil {
		options = &SecretFromNameOptions{}
	}

	resp, err := client.SecretGetOrCreate(ctx, pb.SecretGetOrCreateRequest_builder{
		DeploymentName:  name,
		EnvironmentName: environmentName(options.Environment),
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

// SecretFromMap creates a Secret from a map of key-value pairs.
func SecretFromMap(ctx context.Context, keyValuePairs map[string]string, options *SecretFromMapOptions) (*Secret, error) {
	var err error
	ctx, err = clientContext(ctx)
	if err != nil {
		return nil, err
	}

	if options == nil {
		options = &SecretFromMapOptions{}
	}

	resp, err := client.SecretGetOrCreate(ctx, pb.SecretGetOrCreateRequest_builder{
		ObjectCreationType: pb.ObjectCreationType_OBJECT_CREATION_TYPE_EPHEMERAL,
		EnvDict:            keyValuePairs,
		EnvironmentName:    environmentName(options.Environment),
	}.Build())
	if err != nil {
		return nil, err
	}
	return &Secret{SecretId: resp.GetSecretId()}, nil
}
