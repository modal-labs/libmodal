package com.modal.modalkt

import modal.client.Api

data class AppFromNameParams(
    val environment: String? = null,
    val createIfMissing: Boolean = false,
)

fun parseGpuConfig(gpu: String?): Api.GPUConfig {
    if (gpu.isNullOrBlank()) {
        return Api.GPUConfig.getDefaultInstance()
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

    return Api.GPUConfig.newBuilder()
        .setCount(count)
        .setGpuType(gpuType.uppercase())
        .build()
}
