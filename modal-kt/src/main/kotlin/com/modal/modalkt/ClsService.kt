package com.modal.modalkt

import io.grpc.Status
import modal.client.*
import okio.ByteString.Companion.toByteString

data class ClsFromNameParams(
    val environment: String? = null,
    val createIfMissing: Boolean = false,
)

data class ClsWithOptionsParams(
    val cpu: Double? = null,
    val cpuLimit: Double? = null,
    val memoryMiB: Int? = null,
    val memoryLimitMiB: Int? = null,
    val gpu: String? = null,
    val env: Map<String, String>? = null,
    val secrets: List<Secret>? = null,
    val volumes: Map<String, Volume>? = null,
    val retries: Any? = null,
    val maxContainers: Int? = null,
    val bufferContainers: Int? = null,
    val scaledownWindowMs: Long? = null,
    val timeoutMs: Long? = null,
    val memory: Int? = null,
    val memoryLimit: Int? = null,
    val scaledownWindow: Long? = null,
    val timeout: Long? = null,
)

data class ClsWithConcurrencyParams(
    val maxInputs: Int,
    val targetInputs: Int? = null,
)

data class ClsWithBatchingParams(
    val maxBatchSize: Int,
    val waitMs: Long,
)

internal data class ServiceOptions(
    val cpu: Double? = null,
    val cpuLimit: Double? = null,
    val memoryMiB: Int? = null,
    val memoryLimitMiB: Int? = null,
    val gpu: String? = null,
    val env: Map<String, String>? = null,
    val secrets: List<Secret>? = null,
    val volumes: Map<String, Volume>? = null,
    val retries: Any? = null,
    val maxContainers: Int? = null,
    val bufferContainers: Int? = null,
    val scaledownWindowMs: Long? = null,
    val timeoutMs: Long? = null,
    val maxConcurrentInputs: Int? = null,
    val targetConcurrentInputs: Int? = null,
    val batchMaxSize: Int? = null,
    val batchWaitMs: Long? = null,
)

class ClsService(
    private val client: ModalClient,
) {
    suspend fun fromName(
        appName: String,
        name: String,
        params: ClsFromNameParams = ClsFromNameParams(),
    ): Cls {
        try {
            val response = client.cpClient.functionGet(
                FunctionGetRequest.newBuilder()
                    .setAppName(appName)
                    .setObjectTag("$name.*")
                    .setEnvironmentName(client.environmentName(params.environment))
                    .build(),
            )
            val handleMetadata = response.handleMetadata
                ?: throw InvalidError("Class '$appName/$name' is missing handle metadata")
            return Cls(client, response.functionId, handleMetadata, null)
        } catch (error: Throwable) {
            if (statusCode(error) == Status.Code.NOT_FOUND) {
                throw NotFoundError("Class '$appName/$name' not found")
            }
            throw error
        }
    }
}

class Cls internal constructor(
    private val client: ModalClient,
    private val serviceFunctionId: String,
    private val serviceFunctionMetadata: FunctionHandleMetadata,
    private val serviceOptions: ServiceOptions?,
) {
    suspend fun instance(parameters: Map<String, Any?> = emptyMap()): ClsInstance {
        val functionId = if (schema().isEmpty() && serviceOptions == null) {
            serviceFunctionId
        } else {
            bindParameters(parameters)
        }

        val methods = serviceFunctionMetadata.methodHandleMetadataMap
            .mapNotNull { (name, metadata) ->
                metadata?.let { name to Function(client, functionId, name, it) }
            }
            .toMap()
        return ClsInstance(methods)
    }

    fun withOptions(options: ClsWithOptionsParams): Cls {
        checkForRenamedParams(
            mapOf(
                "memory" to options.memory,
                "memoryLimit" to options.memoryLimit,
                "scaledownWindow" to options.scaledownWindow,
                "timeout" to options.timeout,
            ).filterValues { it != null },
            mapOf(
                "memory" to "memoryMiB",
                "memoryLimit" to "memoryLimitMiB",
                "scaledownWindow" to "scaledownWindowMs",
                "timeout" to "timeoutMs",
            ),
        )

        return Cls(
            client,
            serviceFunctionId,
            serviceFunctionMetadata,
            mergeServiceOptions(
                serviceOptions,
                ServiceOptions(
                    cpu = options.cpu,
                    cpuLimit = options.cpuLimit,
                    memoryMiB = options.memoryMiB,
                    memoryLimitMiB = options.memoryLimitMiB,
                    gpu = options.gpu,
                    env = options.env,
                    secrets = options.secrets,
                    volumes = options.volumes,
                    retries = options.retries,
                    maxContainers = options.maxContainers,
                    bufferContainers = options.bufferContainers,
                    scaledownWindowMs = options.scaledownWindowMs,
                    timeoutMs = options.timeoutMs,
                ),
            ),
        )
    }

    fun withConcurrency(params: ClsWithConcurrencyParams): Cls {
        return Cls(
            client,
            serviceFunctionId,
            serviceFunctionMetadata,
            mergeServiceOptions(
                serviceOptions,
                ServiceOptions(
                    maxConcurrentInputs = params.maxInputs,
                    targetConcurrentInputs = params.targetInputs,
                ),
            ),
        )
    }

    fun withBatching(params: ClsWithBatchingParams): Cls {
        return Cls(
            client,
            serviceFunctionId,
            serviceFunctionMetadata,
            mergeServiceOptions(
                serviceOptions,
                ServiceOptions(
                    batchMaxSize = params.maxBatchSize,
                    batchWaitMs = params.waitMs,
                ),
            ),
        )
    }

    private fun schema(): List<ClassParameterSpec> {
        return serviceFunctionMetadata.classParameterInfo?.schemaList ?: emptyList()
    }

    private suspend fun bindParameters(parameters: Map<String, Any?>): String {
        val mergedSecrets = mergeEnvIntoSecrets(client, serviceOptions?.env, serviceOptions?.secrets)
        val options = if (serviceOptions == null) {
            null
        } else {
            serviceOptions.copy(secrets = mergedSecrets, env = null)
        }

        val request = FunctionBindParamsRequest.newBuilder()
            .setFunctionId(serviceFunctionId)
            .setSerializedParams(encodeParameterSet(schema(), parameters).toByteString())
            .setEnvironmentName(client.environmentName())
            .apply {
                val functionOptions = buildFunctionOptionsProto(options)
                if (functionOptions != null) {
                    setFunctionOptions(functionOptions)
                }
            }
            .build()
        val response = client.cpClient.functionBindParams(request)
        return response.boundFunctionId
    }
}

class ClsInstance(
    private val methods: Map<String, Function>,
) {
    fun method(name: String): Function {
        return methods[name] ?: throw NotFoundError("Method '$name' not found on class")
    }
}

private fun mergeServiceOptions(base: ServiceOptions?, diff: ServiceOptions): ServiceOptions {
    return ServiceOptions(
        cpu = diff.cpu ?: base?.cpu,
        cpuLimit = diff.cpuLimit ?: base?.cpuLimit,
        memoryMiB = diff.memoryMiB ?: base?.memoryMiB,
        memoryLimitMiB = diff.memoryLimitMiB ?: base?.memoryLimitMiB,
        gpu = diff.gpu ?: base?.gpu,
        env = diff.env ?: base?.env,
        secrets = diff.secrets ?: base?.secrets,
        volumes = diff.volumes ?: base?.volumes,
        retries = diff.retries ?: base?.retries,
        maxContainers = diff.maxContainers ?: base?.maxContainers,
        bufferContainers = diff.bufferContainers ?: base?.bufferContainers,
        scaledownWindowMs = diff.scaledownWindowMs ?: base?.scaledownWindowMs,
        timeoutMs = diff.timeoutMs ?: base?.timeoutMs,
        maxConcurrentInputs = diff.maxConcurrentInputs ?: base?.maxConcurrentInputs,
        targetConcurrentInputs = diff.targetConcurrentInputs ?: base?.targetConcurrentInputs,
        batchMaxSize = diff.batchMaxSize ?: base?.batchMaxSize,
        batchWaitMs = diff.batchWaitMs ?: base?.batchWaitMs,
    )
}

private fun buildFunctionOptionsProto(
    options: ServiceOptions?,
): FunctionOptions? {
    if (options == null) {
        return null
    }

    val gpuConfig = parseGpuConfig(options.gpu)

    val resourcesBuilder = Resources.newBuilder()
    var hasResources = false

    val cpu = options.cpu
    val cpuLimit = options.cpuLimit
    if (cpu == null && cpuLimit != null) {
        throw InvalidError("must also specify cpu when cpuLimit is specified")
    }
    if (cpu != null) {
        if (cpu <= 0.0) {
            throw InvalidError("cpu ($cpu) must be a positive number")
        }
        resourcesBuilder.setMilliCpu((cpu * 1000).toInt())
        hasResources = true
        if (cpuLimit != null) {
            if (cpuLimit < cpu) {
                throw InvalidError("cpu ($cpu) cannot be higher than cpuLimit ($cpuLimit)")
            }
            resourcesBuilder.setMilliCpuMax((cpuLimit * 1000).toInt())
        }
    }

    val memoryMiB = options.memoryMiB
    val memoryLimitMiB = options.memoryLimitMiB
    if (memoryMiB == null && memoryLimitMiB != null) {
        throw InvalidError("must also specify memoryMiB when memoryLimitMiB is specified")
    }
    if (memoryMiB != null) {
        if (memoryMiB <= 0) {
            throw InvalidError("memoryMiB ($memoryMiB) must be a positive number")
        }
        resourcesBuilder.setMemoryMb(memoryMiB)
        hasResources = true
        if (memoryLimitMiB != null) {
            if (memoryLimitMiB < memoryMiB) {
                throw InvalidError("memoryMiB ($memoryMiB) cannot be higher than memoryLimitMiB ($memoryLimitMiB)")
            }
            resourcesBuilder.setMemoryMbMax(memoryLimitMiB)
        }
    }
    if (gpuConfig != GPUConfig.getDefaultInstance()) {
        resourcesBuilder.setGpuConfig(gpuConfig)
        hasResources = true
    }

    val retries = parseRetries(options.retries)

    val scaledownWindowMs = options.scaledownWindowMs
    if (scaledownWindowMs != null && scaledownWindowMs % 1000 != 0L) {
        throw InvalidError("scaledownWindowMs must be a multiple of 1000ms, got $scaledownWindowMs")
    }
    val timeoutMs = options.timeoutMs
    if (timeoutMs != null && timeoutMs % 1000 != 0L) {
        throw InvalidError("timeoutMs must be a multiple of 1000ms, got $timeoutMs")
    }

    val secretIds = options.secrets?.map { it.secretId } ?: emptyList()
    val volumeMounts = options.volumes?.map { (mountPath, volume) ->
        VolumeMount.newBuilder()
            .setVolumeId(volume.volumeId)
            .setMountPath(mountPath)
            .setAllowBackgroundCommits(true)
            .setReadOnly(volume.isReadOnly)
            .build()
    } ?: emptyList()

    return FunctionOptions.newBuilder()
        .addAllSecretIds(secretIds)
        .setReplaceSecretIds(secretIds.isNotEmpty())
        .setReplaceVolumeMounts(volumeMounts.isNotEmpty())
        .addAllVolumeMounts(volumeMounts)
        .apply {
            if (hasResources) {
                setResources(resourcesBuilder.build())
            }
            if (retries != null) {
                setRetryPolicy(
                    FunctionRetryPolicy.newBuilder()
                    .setRetries(retries.maxRetries)
                    .setBackoffCoefficient(retries.backoffCoefficient.toFloat())
                    .setInitialDelayMs(retries.initialDelayMs.toInt())
                    .setMaxDelayMs(retries.maxDelayMs.toInt())
                    .build(),
                )
            }
            if (options.maxContainers != null) {
                setConcurrencyLimit(options.maxContainers)
            }
            if (options.bufferContainers != null) {
                setBufferContainers(options.bufferContainers)
            }
            if (scaledownWindowMs != null) {
                setTaskIdleTimeoutSecs((scaledownWindowMs / 1000).toInt())
            }
            if (timeoutMs != null) {
                setTimeoutSecs((timeoutMs / 1000).toInt())
            }
            if (options.maxConcurrentInputs != null) {
                setMaxConcurrentInputs(options.maxConcurrentInputs)
            }
            if (options.targetConcurrentInputs != null) {
                setTargetConcurrentInputs(options.targetConcurrentInputs)
            }
            if (options.batchMaxSize != null) {
                setBatchMaxSize(options.batchMaxSize)
            }
            if (options.batchWaitMs != null) {
                setBatchLingerMs(options.batchWaitMs)
            }
        }
        .build()
}
