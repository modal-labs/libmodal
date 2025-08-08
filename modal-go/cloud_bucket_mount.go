package modal

import (
	"fmt"
	"net/url"
	"strings"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

// CloudBucketMount provides access to cloud storage buckets within Modal functions.
type CloudBucketMount struct {
	BucketName        string
	BucketType        pb.CloudBucketMount_BucketType
	Secret            *Secret
	ReadOnly          bool
	RequesterPays     bool
	BucketEndpointUrl *string
	KeyPrefix         *string
	OidcAuthRoleArn   *string
}

// CloudBucketMountOptions are options for creating a CloudBucketMount.
type CloudBucketMountOptions struct {
	Secret            *Secret
	ReadOnly          bool
	RequesterPays     bool
	BucketEndpointUrl *string
	KeyPrefix         *string
	OidcAuthRoleArn   *string
}

// NewCloudBucketMount creates a new CloudBucketMount.
func NewCloudBucketMount(bucketName string, options *CloudBucketMountOptions) (*CloudBucketMount, error) {
	if options == nil {
		options = &CloudBucketMountOptions{}
	}

	mount := &CloudBucketMount{
		BucketName:        bucketName,
		Secret:            options.Secret,
		ReadOnly:          options.ReadOnly,
		RequesterPays:     options.RequesterPays,
		BucketEndpointUrl: options.BucketEndpointUrl,
		KeyPrefix:         options.KeyPrefix,
		OidcAuthRoleArn:   options.OidcAuthRoleArn,
	}

	// Determine bucket type from endpoint URL
	if mount.BucketEndpointUrl != nil {
		parsedURL, err := url.Parse(*mount.BucketEndpointUrl)
		if err != nil {
			return nil, fmt.Errorf("invalid bucket endpoint URL: %w", err)
		}

		hostname := parsedURL.Hostname()
		if strings.HasSuffix(hostname, "r2.cloudflarestorage.com") {
			mount.BucketType = pb.CloudBucketMount_R2
		} else if strings.HasSuffix(hostname, "storage.googleapis.com") {
			mount.BucketType = pb.CloudBucketMount_GCP
		} else {
			mount.BucketType = pb.CloudBucketMount_S3
		}
	} else {
		// Just assume S3; this is backwards and forwards compatible
		mount.BucketType = pb.CloudBucketMount_S3
	}

	if mount.RequesterPays && mount.Secret == nil {
		return nil, fmt.Errorf("credentials required in order to use Requester Pays")
	}

	if mount.KeyPrefix != nil && !strings.HasSuffix(*mount.KeyPrefix, "/") {
		return nil, fmt.Errorf("keyPrefix will be prefixed to all object paths, so it must end in a '/'")
	}

	return mount, nil
}

// toProto converts the CloudBucketMount to a protobuf message.
func (c *CloudBucketMount) toProto(mountPath string) *pb.CloudBucketMount {
	credentialsSecretId := ""
	if c.Secret != nil {
		credentialsSecretId = c.Secret.SecretId
	}

	return pb.CloudBucketMount_builder{
		BucketName:          c.BucketName,
		MountPath:           mountPath,
		CredentialsSecretId: credentialsSecretId,
		ReadOnly:            c.ReadOnly,
		BucketType:          c.BucketType,
		RequesterPays:       c.RequesterPays,
		BucketEndpointUrl:   c.BucketEndpointUrl,
		KeyPrefix:           c.KeyPrefix,
		OidcAuthRoleArn:     c.OidcAuthRoleArn,
	}.Build()
}
