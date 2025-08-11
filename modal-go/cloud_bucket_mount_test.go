package modal

import (
	"testing"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"github.com/onsi/gomega"
)

func TestNewCloudBucketMount_MinimalOptions(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	mount, err := NewCloudBucketMount("my-bucket", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(mount.BucketName).Should(gomega.Equal("my-bucket"))
	g.Expect(mount.ReadOnly).Should(gomega.BeFalse())
	g.Expect(mount.RequesterPays).Should(gomega.BeFalse())
	g.Expect(mount.Secret).Should(gomega.BeNil())
	g.Expect(mount.BucketEndpointUrl).Should(gomega.BeNil())
	g.Expect(mount.KeyPrefix).Should(gomega.BeNil())
	g.Expect(mount.OidcAuthRoleArn).Should(gomega.BeNil())

	bucketType, err := getBucketTypeFromEndpointURL(mount.BucketEndpointUrl)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(bucketType).Should(gomega.Equal(pb.CloudBucketMount_S3))
}

func TestNewCloudBucketMount_AllOptions(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	mockSecret := &Secret{SecretId: "sec-123"}
	endpointURL := "https://my-bucket.r2.cloudflarestorage.com"
	keyPrefix := "prefix/"
	oidcRole := "arn:aws:iam::123456789:role/MyRole"

	mount, err := NewCloudBucketMount("my-bucket", &CloudBucketMountOptions{
		Secret:            mockSecret,
		ReadOnly:          true,
		RequesterPays:     true,
		BucketEndpointUrl: &endpointURL,
		KeyPrefix:         &keyPrefix,
		OidcAuthRoleArn:   &oidcRole,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(mount.BucketName).Should(gomega.Equal("my-bucket"))
	g.Expect(mount.ReadOnly).Should(gomega.BeTrue())
	g.Expect(mount.RequesterPays).Should(gomega.BeTrue())
	g.Expect(mount.Secret).Should(gomega.Equal(mockSecret))
	g.Expect(mount.BucketEndpointUrl).ShouldNot(gomega.BeNil())
	g.Expect(*mount.BucketEndpointUrl).Should(gomega.Equal(endpointURL))
	g.Expect(mount.KeyPrefix).ShouldNot(gomega.BeNil())
	g.Expect(*mount.KeyPrefix).Should(gomega.Equal(keyPrefix))
	g.Expect(mount.OidcAuthRoleArn).ShouldNot(gomega.BeNil())
	g.Expect(*mount.OidcAuthRoleArn).Should(gomega.Equal(oidcRole))

	bucketType, err := getBucketTypeFromEndpointURL(mount.BucketEndpointUrl)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(bucketType).Should(gomega.Equal(pb.CloudBucketMount_R2))
}

func TestGetBucketTypeFromEndpointURL(t *testing.T) {
	t.Parallel()
	test_cases := []struct {
		name         string
		endpointURL  string
		expectedType pb.CloudBucketMount_BucketType
	}{
		{
			name:         "Empty defaults to S3",
			endpointURL:  "",
			expectedType: pb.CloudBucketMount_S3,
		},
		{
			name:         "R2",
			endpointURL:  "https://my-bucket.r2.cloudflarestorage.com",
			expectedType: pb.CloudBucketMount_R2,
		},
		{
			name:         "GCP",
			endpointURL:  "https://storage.googleapis.com/my-bucket",
			expectedType: pb.CloudBucketMount_GCP,
		},
		{
			name:         "Unknown defaults to S3",
			endpointURL:  "https://unknown-endpoint.com/my-bucket",
			expectedType: pb.CloudBucketMount_S3,
		},
	}

	for _, tc := range test_cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := gomega.NewWithT(t)

			options := &CloudBucketMountOptions{}
			if tc.endpointURL != "" {
				options.BucketEndpointUrl = &tc.endpointURL
			}

			mount, err := NewCloudBucketMount("my-bucket", options)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())

			bucketType, err := getBucketTypeFromEndpointURL(mount.BucketEndpointUrl)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
			g.Expect(bucketType).Should(gomega.Equal(tc.expectedType))
		})
	}
}

func TestGetBucketTypeFromEndpointURL_InvalidURL(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	invalidURL := "://invalid-url"
	_, err := getBucketTypeFromEndpointURL(&invalidURL)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("failed to parse bucketEndpointURL"))
	g.Expect(err.Error()).Should(gomega.ContainSubstring(invalidURL))
}

func TestNewCloudBucketMount_ValidationRequesterPaysWithoutSecret(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	_, err := NewCloudBucketMount("my-bucket", &CloudBucketMountOptions{
		RequesterPays: true,
	})

	g.Expect(err).Should(gomega.MatchError("credentials required in order to use Requester Pays"))
}

func TestNewCloudBucketMount_ValidationKeyPrefixWithoutTrailingSlash(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	keyPrefix := "prefix"
	_, err := NewCloudBucketMount("my-bucket", &CloudBucketMountOptions{
		KeyPrefix: &keyPrefix,
	})

	g.Expect(err).Should(gomega.MatchError("keyPrefix will be prefixed to all object paths, so it must end in a '/'"))
}

func TestCloudBucketMount_ToProtoMinimalOptions(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	mount, err := NewCloudBucketMount("my-bucket", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	proto, err := mount.toProto("/mnt/bucket")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(proto.GetBucketName()).Should(gomega.Equal("my-bucket"))
	g.Expect(proto.GetMountPath()).Should(gomega.Equal("/mnt/bucket"))
	g.Expect(proto.GetCredentialsSecretId()).Should(gomega.BeEmpty())
	g.Expect(proto.GetReadOnly()).Should(gomega.BeFalse())
	g.Expect(proto.GetBucketType()).Should(gomega.Equal(pb.CloudBucketMount_S3))
	g.Expect(proto.GetRequesterPays()).Should(gomega.BeFalse())
	g.Expect(proto.GetBucketEndpointUrl()).Should(gomega.BeEmpty())
	g.Expect(proto.GetKeyPrefix()).Should(gomega.BeEmpty())
	g.Expect(proto.GetOidcAuthRoleArn()).Should(gomega.BeEmpty())
}

func TestCloudBucketMount_ToProtoAllOptions(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	mockSecret := &Secret{SecretId: "sec-123"}
	endpointURL := "https://my-bucket.r2.cloudflarestorage.com"
	keyPrefix := "prefix/"
	oidcRole := "arn:aws:iam::123456789:role/MyRole"

	mount, err := NewCloudBucketMount("my-bucket", &CloudBucketMountOptions{
		Secret:            mockSecret,
		ReadOnly:          true,
		RequesterPays:     true,
		BucketEndpointUrl: &endpointURL,
		KeyPrefix:         &keyPrefix,
		OidcAuthRoleArn:   &oidcRole,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	proto, err := mount.toProto("/mnt/bucket")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(proto.GetBucketName()).Should(gomega.Equal("my-bucket"))
	g.Expect(proto.GetMountPath()).Should(gomega.Equal("/mnt/bucket"))
	g.Expect(proto.GetCredentialsSecretId()).Should(gomega.Equal("sec-123"))
	g.Expect(proto.GetReadOnly()).Should(gomega.BeTrue())
	g.Expect(proto.GetBucketType()).Should(gomega.Equal(pb.CloudBucketMount_R2))
	g.Expect(proto.GetRequesterPays()).Should(gomega.BeTrue())
	g.Expect(proto.GetBucketEndpointUrl()).Should(gomega.Equal(endpointURL))
	g.Expect(proto.GetKeyPrefix()).Should(gomega.Equal(keyPrefix))
	g.Expect(proto.GetOidcAuthRoleArn()).Should(gomega.Equal(oidcRole))
}
