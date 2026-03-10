package com.modal.modalkt

import io.grpc.Status
import modal.client.*

data class AppFromNameParams(
    val environment: String? = null,
    val createIfMissing: Boolean = false,
)

class AppService(
    private val client: ModalClient,
) {
    suspend fun fromName(
        name: String,
        params: AppFromNameParams = AppFromNameParams(),
    ): App {
        try {
            val response = client.cpClient.appGetOrCreate(
                AppGetOrCreateRequest.newBuilder()
                    .setAppName(name)
                    .setEnvironmentName(client.environmentName(params.environment))
                    .setObjectCreationType(
                        if (params.createIfMissing) {
                            ObjectCreationType.OBJECT_CREATION_TYPE_CREATE_IF_MISSING
                        } else {
                            ObjectCreationType.OBJECT_CREATION_TYPE_UNSPECIFIED
                        },
                    )
                    .build(),
            )
            client.logger.debug("Retrieved App", "app_id", response.appId, "app_name", name)
            return App(response.appId, name)
        } catch (error: Throwable) {
            if (statusCode(error) == Status.Code.NOT_FOUND) {
                throw NotFoundError("App '$name' not found")
            }
            throw error
        }
    }
}

fun parseGpuConfig(gpu: String?): GPUConfig {
    if (gpu.isNullOrBlank()) {
        return GPUConfig.getDefaultInstance()
    }

    var gpuType = gpu
    var count = 1

    if (gpu.contains(":")) {
        val parts = gpu.split(":", limit = 2)
        gpuType = parts[0]
        count = parts[1].toIntOrNull()
            ?: throw InvalidError(
                "Invalid GPU count: ${parts[1]}. Value must be a positive integer.",
            )
        if (count < 1) {
            throw InvalidError(
                "Invalid GPU count: ${parts[1]}. Value must be a positive integer.",
            )
        }
    }

    return GPUConfig.newBuilder()
        .setCount(count)
        .setGpuType(gpuType.uppercase())
        .build()
}
