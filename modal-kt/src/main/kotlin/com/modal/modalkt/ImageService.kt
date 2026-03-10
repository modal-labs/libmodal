package com.modal.modalkt

import io.grpc.Status
import kotlinx.coroutines.flow.collect
import modal.client.*

data class ImageDeleteParams(
    val unused: Unit? = null,
)

data class ImageDockerfileCommandsParams(
    val env: Map<String, String>? = null,
    val secrets: List<Secret>? = null,
    val gpu: String? = null,
    val forceBuild: Boolean = false,
)

internal data class ImageLayer(
    val commands: List<String>,
    val env: Map<String, String>? = null,
    val secrets: List<Secret>? = null,
    val gpuConfig: GPUConfig? = null,
    val forceBuild: Boolean = false,
)

class ImageService(
    private val client: ModalClient,
) {
    suspend fun fromId(imageId: String): Image {
        try {
            val response = client.cpClient.imageFromId(
                ImageFromIdRequest.newBuilder()
                    .setImageId(imageId)
                    .build(),
            )
            return Image(client, response.imageId, "")
        } catch (error: Throwable) {
            if (statusCode(error) == Status.Code.NOT_FOUND) {
                throw NotFoundError(statusMessage(error))
            }
            if (statusCode(error) == Status.Code.FAILED_PRECONDITION &&
                statusMessage(error).contains("Could not find image with ID")
            ) {
                throw NotFoundError(statusMessage(error))
            }
            throw error
        }
    }

    fun fromRegistry(tag: String, secret: Secret? = null): Image {
        val registryConfig = secret?.let {
            ImageRegistryConfig.newBuilder()
                .setRegistryAuthType(RegistryAuthType.REGISTRY_AUTH_TYPE_STATIC_CREDS)
                .setSecretId(it.secretId)
                .build()
        }
        return Image(client, "", tag, registryConfig)
    }

    fun fromAwsEcr(tag: String, secret: Secret): Image {
        val registryConfig = ImageRegistryConfig.newBuilder()
            .setRegistryAuthType(RegistryAuthType.REGISTRY_AUTH_TYPE_AWS)
            .setSecretId(secret.secretId)
            .build()
        return Image(client, "", tag, registryConfig)
    }

    fun fromGcpArtifactRegistry(tag: String, secret: Secret): Image {
        val registryConfig = ImageRegistryConfig.newBuilder()
            .setRegistryAuthType(RegistryAuthType.REGISTRY_AUTH_TYPE_GCP)
            .setSecretId(secret.secretId)
            .build()
        return Image(client, "", tag, registryConfig)
    }

    suspend fun delete(
        imageId: String,
        params: ImageDeleteParams = ImageDeleteParams(),
    ) {
        try {
            client.cpClient.imageDelete(
                ImageDeleteRequest.newBuilder()
                    .setImageId(imageId)
                    .build(),
            )
        } catch (error: Throwable) {
            if (statusCode(error) == Status.Code.NOT_FOUND) {
                throw NotFoundError(statusMessage(error))
            }
            if (statusCode(error) == Status.Code.FAILED_PRECONDITION &&
                statusMessage(error).contains("Could not find image with ID")
            ) {
                throw NotFoundError(statusMessage(error))
            }
            throw error
        }
    }
}

class Image internal constructor(
    private val client: ModalClient,
    private var internalImageId: String,
    private val tag: String,
    private val imageRegistryConfig: ImageRegistryConfig? = null,
    private val layers: List<ImageLayer> = listOf(ImageLayer(emptyList())),
) {
    val imageId: String
        get() = internalImageId

    fun dockerfileCommands(
        commands: List<String>,
        params: ImageDockerfileCommandsParams = ImageDockerfileCommandsParams(),
    ): Image {
        if (commands.isEmpty()) {
            return this
        }

        validateDockerfileCommands(commands)

        val layer = ImageLayer(
            commands = commands.toList(),
            env = params.env,
            secrets = params.secrets,
            gpuConfig = params.gpu?.let(::parseGpuConfig),
            forceBuild = params.forceBuild,
        )
        return Image(client, "", tag, imageRegistryConfig, layers + layer)
    }

    suspend fun build(app: App): Image {
        if (internalImageId.isNotEmpty()) {
            return this
        }

        var baseImageId: String? = null
        for ((index, layer) in layers.withIndex()) {
            val mergedSecrets = mergeEnvIntoSecrets(client, layer.env, layer.secrets)
            val dockerfileCommands: List<String>
            val baseImages: List<BaseImage>

            if (index == 0) {
                dockerfileCommands = listOf("FROM $tag") + layer.commands
                baseImages = emptyList()
            } else {
                dockerfileCommands = listOf("FROM base") + layer.commands
                baseImages = listOf(
                    BaseImage.newBuilder()
                        .setDockerTag("base")
                        .setImageId(baseImageId ?: "")
                        .build(),
                )
            }

            val response = client.cpClient.imageGetOrCreate(
                ImageGetOrCreateRequest.newBuilder()
                    .setAppId(app.appId)
                    .setImage(
                        ImageProto.newBuilder()
                            .addAllDockerfileCommands(dockerfileCommands)
                            .apply {
                                if (this@Image.imageRegistryConfig != null) {
                                    setImageRegistryConfig(this@Image.imageRegistryConfig)
                                }
                                addAllSecretIds(mergedSecrets.map { it.secretId })
                                addAllBaseImages(baseImages)
                                if (layer.gpuConfig != null) {
                                    setGpuConfig(layer.gpuConfig)
                                }
                            }
                            .build(),
                    )
                    .setBuilderVersion(client.imageBuilderVersion())
                    .setForceBuild(layer.forceBuild)
                    .build(),
            )

            val result = if (response.hasResult()) {
                response.result ?: waitForImageBuild(response.imageId)
            } else {
                waitForImageBuild(response.imageId)
            }

            when (result.status) {
                GenericStatus.GENERIC_STATUS_SUCCESS -> Unit
                GenericStatus.GENERIC_STATUS_FAILURE -> {
                    throw InvalidError(
                        "Image build for ${response.imageId} failed with the exception:\n${result.exception}",
                    )
                }
                GenericStatus.GENERIC_STATUS_TERMINATED -> {
                    throw InvalidError(
                        "Image build for ${response.imageId} terminated due to external shut-down. Please try again.",
                    )
                }
                GenericStatus.GENERIC_STATUS_TIMEOUT -> {
                    throw InvalidError(
                        "Image build for ${response.imageId} timed out. Please try again with a larger timeout parameter.",
                    )
                }
                else -> {
                    throw InvalidError(
                        "Image build for ${response.imageId} failed with unknown status: ${result.status.number}",
                    )
                }
            }

            baseImageId = response.imageId
        }

        internalImageId = baseImageId ?: internalImageId
        return this
    }

    private suspend fun waitForImageBuild(imageId: String): GenericResult {
        var result: GenericResult? = null
        var lastEntryId = ""
        client.cpClient.imageJoinStreaming(
            ImageJoinStreamingRequest.newBuilder()
                .setImageId(imageId)
                .setTimeout(55f)
                .setLastEntryId(lastEntryId)
                .build(),
        ).collect { item ->
            if (item.entryId.isNotEmpty()) {
                lastEntryId = item.entryId
            }
            val itemResult = item.result
            if (item.hasResult() && itemResult != null && itemResult.status != GenericStatus.GENERIC_STATUS_UNSPECIFIED) {
                result = itemResult
            }
        }
        return result ?: throw InvalidError("Image build for $imageId failed with unknown status: missing result")
    }

    private fun validateDockerfileCommands(commands: List<String>) {
        for (command in commands) {
            val trimmed = command.trim().uppercase()
            if (trimmed.startsWith("COPY ") && !trimmed.startsWith("COPY --FROM=")) {
                throw InvalidError("COPY commands that copy from local context are not yet supported.")
            }
        }
    }
}
