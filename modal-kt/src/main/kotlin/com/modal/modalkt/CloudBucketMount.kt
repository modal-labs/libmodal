package com.modal.modalkt

import modal.client.*

data class CloudBucketMountCreateParams(
    val secret: Secret? = null,
    val readOnly: Boolean = false,
    val requesterPays: Boolean = false,
    val bucketEndpointUrl: String? = null,
    val keyPrefix: String? = null,
    val oidcAuthRoleArn: String? = null,
)

class CloudBucketMountService(
    private val client: ModalClient,
) {
    fun create(
        bucketName: String,
        params: CloudBucketMountCreateParams = CloudBucketMountCreateParams(),
    ): CloudBucketMount {
        var bucketType = CloudBucketMountBucketType.S3
        val bucketEndpointUrl = params.bucketEndpointUrl
        if (!bucketEndpointUrl.isNullOrEmpty()) {
            val url = try {
                java.net.URI(bucketEndpointUrl).toURL()
            } catch (error: java.net.URISyntaxException) {
                throw java.net.MalformedURLException(error.message)
            }
            val host = url.host
            bucketType = when {
                host.endsWith("r2.cloudflarestorage.com") -> CloudBucketMountBucketType.R2
                host.endsWith("storage.googleapis.com") -> CloudBucketMountBucketType.GCP
                else -> {
                    client.logger.debug(
                        "CloudBucketMount received unrecognized bucket endpoint URL. Assuming AWS S3 configuration as fallback.",
                        "bucketEndpointUrl",
                        bucketEndpointUrl,
                    )
                    CloudBucketMountBucketType.S3
                }
            }
        }

        if (params.requesterPays && params.secret == null) {
            throw InvalidError("Credentials required in order to use Requester Pays.")
        }
        val keyPrefix = params.keyPrefix
        if (keyPrefix != null && !keyPrefix.endsWith("/")) {
            throw InvalidError(
                "keyPrefix will be prefixed to all object paths, so it must end in a '/'",
            )
        }

        return CloudBucketMount(
            bucketName = bucketName,
            secret = params.secret,
            readOnly = params.readOnly,
            requesterPays = params.requesterPays,
            bucketEndpointUrl = bucketEndpointUrl,
            keyPrefix = keyPrefix,
            oidcAuthRoleArn = params.oidcAuthRoleArn,
            bucketType = bucketType,
        )
    }
}

data class CloudBucketMount(
    val bucketName: String,
    val secret: Secret? = null,
    val readOnly: Boolean = false,
    val requesterPays: Boolean = false,
    val bucketEndpointUrl: String? = null,
    val keyPrefix: String? = null,
    val oidcAuthRoleArn: String? = null,
    private val bucketType: CloudBucketMountBucketType = CloudBucketMountBucketType.S3,
) {
    fun toProto(mountPath: String): CloudBucketMountProto {
        return CloudBucketMountProto.newBuilder()
            .setBucketName(bucketName)
            .setMountPath(mountPath)
            .setCredentialsSecretId(secret?.secretId ?: "")
            .setReadOnly(readOnly)
            .setBucketType(bucketType)
            .setRequesterPays(requesterPays)
            .apply {
                if (this@CloudBucketMount.bucketEndpointUrl != null) {
                    setBucketEndpointUrl(this@CloudBucketMount.bucketEndpointUrl!!)
                }
                if (this@CloudBucketMount.keyPrefix != null) {
                    setKeyPrefix(this@CloudBucketMount.keyPrefix!!)
                }
                if (this@CloudBucketMount.oidcAuthRoleArn != null) {
                    setOidcAuthRoleArn(this@CloudBucketMount.oidcAuthRoleArn!!)
                }
            }
            .build()
    }
}
