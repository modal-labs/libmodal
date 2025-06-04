package modal

import (
	"context"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

// Secret represents a Modal secret.
type Secret struct {
	SecretId string

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
	ctx = clientContext(ctx)

	if options == nil {
		options = &SecretFromNameOptions{}
	}

	resp, err := client.SecretGetOrCreate(ctx, pb.SecretGetOrCreateRequest_builder{
		DeploymentName:  name,
		Namespace:       pb.DeploymentNamespace_DEPLOYMENT_NAMESPACE_WORKSPACE,
		EnvironmentName: environmentName(options.Environment),
		RequiredKeys:    options.RequiredKeys,
	}.Build())

	if err != nil {
		return nil, err
	}

	return &Secret{SecretId: resp.GetSecretId()}, nil
}
