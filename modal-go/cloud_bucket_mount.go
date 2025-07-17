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

	// Validate the endpoint URL if provided
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

func getBucketTypeFromEndpointURL(bucketEndpointURL *string) pb.CloudBucketMount_BucketType {
	if bucketEndpointURL == nil {
		return pb.CloudBucketMount_S3
	}

	parsedURL, err := url.Parse(*bucketEndpointURL)
	if err != nil {
		return pb.CloudBucketMount_S3
	}

	hostname := parsedURL.Hostname()
	if strings.HasSuffix(hostname, "r2.cloudflarestorage.com") {
		return pb.CloudBucketMount_R2
	} else if strings.HasSuffix(hostname, "storage.googleapis.com") {
		return pb.CloudBucketMount_GCP
	}
	return pb.CloudBucketMount_S3
}

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
		BucketType:          getBucketTypeFromEndpointURL(c.BucketEndpointUrl),
		RequesterPays:       c.RequesterPays,
		BucketEndpointUrl:   c.BucketEndpointUrl,
		KeyPrefix:           c.KeyPrefix,
		OidcAuthRoleArn:     c.OidcAuthRoleArn,
	}.Build()
}
