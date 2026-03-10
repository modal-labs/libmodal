package com.modal.modalkt

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import modal.client.*

class CloudBucketMountTest {
    private val client = ModalClient()

    @Test
    fun createWithMinimalOptions() {
        val mount = client.cloudBucketMounts.create("my-bucket")

        assertEquals("my-bucket", mount.bucketName)
        assertEquals(false, mount.readOnly)
        assertEquals(false, mount.requesterPays)
        assertEquals(null, mount.secret)
        assertEquals(CloudBucketMountBucketType.S3, mount.toProto("/").bucketType)
    }

    @Test
    fun createWithAllOptions() {
        val secret = Secret("sec-123")

        val mount = client.cloudBucketMounts.create(
            "my-bucket",
            CloudBucketMountCreateParams(
                secret = secret,
                readOnly = true,
                requesterPays = true,
                bucketEndpointUrl = "https://my-bucket.r2.cloudflarestorage.com",
                keyPrefix = "prefix/",
                oidcAuthRoleArn = "arn:aws:iam::123456789:role/MyRole",
            ),
        )

        assertEquals(secret, mount.secret)
        assertEquals(CloudBucketMountBucketType.R2, mount.toProto("/").bucketType)
    }

    @Test
    fun detectsBucketTypesFromEndpointUrl() {
        assertEquals(
            CloudBucketMountBucketType.S3,
            client.cloudBucketMounts.create(
                "my-bucket",
                CloudBucketMountCreateParams(bucketEndpointUrl = ""),
            ).toProto("/").bucketType,
        )
        assertEquals(
            CloudBucketMountBucketType.R2,
            client.cloudBucketMounts.create(
                "my-bucket",
                CloudBucketMountCreateParams(bucketEndpointUrl = "https://my-bucket.r2.cloudflarestorage.com"),
            ).toProto("/").bucketType,
        )
        assertEquals(
            CloudBucketMountBucketType.GCP,
            client.cloudBucketMounts.create(
                "my-bucket",
                CloudBucketMountCreateParams(bucketEndpointUrl = "https://storage.googleapis.com/my-bucket"),
            ).toProto("/").bucketType,
        )
        assertFailsWith<java.net.MalformedURLException> {
            client.cloudBucketMounts.create(
                "my-bucket",
                CloudBucketMountCreateParams(bucketEndpointUrl = "://invalid-url"),
            )
        }
    }

    @Test
    fun requesterPaysWithoutSecretFails() {
        assertFailsWith<InvalidError> {
            client.cloudBucketMounts.create(
                "my-bucket",
                CloudBucketMountCreateParams(requesterPays = true),
            )
        }
    }

    @Test
    fun keyPrefixWithoutTrailingSlashFails() {
        assertFailsWith<InvalidError> {
            client.cloudBucketMounts.create(
                "my-bucket",
                CloudBucketMountCreateParams(keyPrefix = "prefix"),
            )
        }
    }

    @Test
    fun toProtoWithMinimalOptions() {
        val proto = client.cloudBucketMounts.create("my-bucket").toProto("/mnt/bucket")

        assertEquals("my-bucket", proto.bucketName)
        assertEquals("/mnt/bucket", proto.mountPath)
        assertEquals("", proto.credentialsSecretId)
        assertEquals(false, proto.readOnly)
        assertEquals(CloudBucketMountBucketType.S3, proto.bucketType)
    }
}
