package modal

import (
	"fmt"
	"net/url"
	"strings"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

// CloudBucketMountService provides CloudBucketMount related operations.
type CloudBucketMountService struct{ client *Client }

// CloudBucketMount provides access to cloud storage buckets within Modal Functions.
type CloudBucketMount struct {
	BucketName        string
	Secret            *Secret
	ReadOnly          bool
	RequesterPays     bool
	BucketEndpointUrl *string
	KeyPrefix         *string
	OidcAuthRoleArn   *string
}

// CloudBucketMountParams are options for creating a CloudBucketMount.
type CloudBucketMountParams struct {
	Secret            *Secret
	ReadOnly          bool
	RequesterPays     bool
	BucketEndpointUrl *string
	KeyPrefix         *string
	OidcAuthRoleArn   *string
}

// New creates a new CloudBucketMount.
func (s *CloudBucketMountService) New(bucketName string, params *CloudBucketMountParams) (*CloudBucketMount, error) {
	if params == nil {
		params = &CloudBucketMountParams{}
	}

	mount := &CloudBucketMount{
		BucketName:        bucketName,
		Secret:            params.Secret,
		ReadOnly:          params.ReadOnly,
		RequesterPays:     params.RequesterPays,
		BucketEndpointUrl: params.BucketEndpointUrl,
		KeyPrefix:         params.KeyPrefix,
		OidcAuthRoleArn:   params.OidcAuthRoleArn,
	}

	if mount.BucketEndpointUrl != nil {
		_, err := url.Parse(*mount.BucketEndpointUrl)
		if err != nil {
			return nil, fmt.Errorf("invalid bucket endpoint URL: %w", err)
		}
	}

	if mount.RequesterPays && mount.Secret == nil {
		return nil, fmt.Errorf("credentials required in order to use Requester Pays")
	}

	if mount.KeyPrefix != nil && !strings.HasSuffix(*mount.KeyPrefix, "/") {
		return nil, fmt.Errorf("keyPrefix will be prefixed to all object paths, so it must end in a '/'")
	}

	return mount, nil
}

func getBucketTypeFromEndpointURL(bucketEndpointURL *string) (pb.CloudBucketMount_BucketType, error) {
	if bucketEndpointURL == nil {
		return pb.CloudBucketMount_S3, nil
	}

	parsedURL, err := url.Parse(*bucketEndpointURL)
	if err != nil {
		return pb.CloudBucketMount_S3, fmt.Errorf("failed to parse bucketEndpointURL '%s': %w", *bucketEndpointURL, err)
	}

	hostname := parsedURL.Hostname()
	if strings.HasSuffix(hostname, "r2.cloudflarestorage.com") {
		return pb.CloudBucketMount_R2, nil
	} else if strings.HasSuffix(hostname, "storage.googleapis.com") {
		return pb.CloudBucketMount_GCP, nil
	}
	return pb.CloudBucketMount_S3, nil
}

func (c *CloudBucketMount) toProto(mountPath string) (*pb.CloudBucketMount, error) {
	credentialsSecretId := ""
	if c.Secret != nil {
		credentialsSecretId = c.Secret.SecretId
	}

	bucketType, err := getBucketTypeFromEndpointURL(c.BucketEndpointUrl)
	if err != nil {
		return nil, err
	}

	return pb.CloudBucketMount_builder{
		BucketName:          c.BucketName,
		MountPath:           mountPath,
		CredentialsSecretId: credentialsSecretId,
		ReadOnly:            c.ReadOnly,
		BucketType:          bucketType,
		RequesterPays:       c.RequesterPays,
		BucketEndpointUrl:   c.BucketEndpointUrl,
		KeyPrefix:           c.KeyPrefix,
		OidcAuthRoleArn:     c.OidcAuthRoleArn,
	}.Build(), nil
}
