package com.modal.modalkt

import io.grpc.Status
import modal.client.*
import okio.ByteString.Companion.toByteString

data class FunctionFromNameParams(
    val environment: String? = null,
    val createIfMissing: Boolean = false,
)

data class FunctionStats(
    val backlog: Int,
    val numTotalRunners: Int,
)

data class FunctionUpdateAutoscalerParams(
    val minContainers: Int? = null,
    val maxContainers: Int? = null,
    val bufferContainers: Int? = null,
    val scaledownWindowMs: Long? = null,
)

class FunctionService(
    private val client: ModalClient,
) {
    suspend fun fromName(
        appName: String,
        name: String,
        params: FunctionFromNameParams = FunctionFromNameParams(),
    ): Function {
        if (name.contains(".")) {
            val parts = name.split(".", limit = 2)
            throw InvalidError(
                "Cannot retrieve Cls methods using 'functions.fromName()'. Use:\n" +
                    "  const cls = await client.cls.fromName(\"$appName\", \"${parts[0]}\");\n" +
                    "  const instance = await cls.instance();\n" +
                    "  const m = instance.method(\"${parts[1]}\");",
            )
        }

        try {
            val response = client.cpClient.functionGet(
                FunctionGetRequest.newBuilder()
                    .setAppName(appName)
                    .setObjectTag(name)
                    .setEnvironmentName(client.environmentName(params.environment))
                    .build(),
            )
            client.logger.debug("Retrieved Function", "function_id", response.functionId, "app_name", appName, "function_name", name)
            return Function(client, response.functionId, null, response.handleMetadata)
        } catch (error: Throwable) {
            if (statusCode(error) == Status.Code.NOT_FOUND) {
                throw NotFoundError("Function '$appName/$name' not found")
            }
            throw error
        }
    }
}

class Function(
    private val client: ModalClient,
    val functionId: String,
    val methodName: String? = null,
    private val handleMetadata: FunctionHandleMetadata? = null,
) {
    suspend fun getCurrentStats(): FunctionStats {
        val response = client.cpClient.functionGetCurrentStats(
            FunctionGetCurrentStatsRequest.newBuilder()
                .setFunctionId(functionId)
                .build(),
        )
        return FunctionStats(response.backlog.toInt(), response.numTotalTasks.toInt())
    }

    suspend fun updateAutoscaler(params: FunctionUpdateAutoscalerParams) {
        val scaledownWindowMs = params.scaledownWindowMs
        if (scaledownWindowMs != null && scaledownWindowMs % 1000 != 0L) {
            throw InvalidError("scaledownWindowMs must be a multiple of 1000ms, got $scaledownWindowMs")
        }

        client.cpClient.functionUpdateSchedulingParams(
            FunctionUpdateSchedulingParamsRequest.newBuilder()
                .setFunctionId(functionId)
                .setWarmPoolSizeOverride(0)
                .setSettings(
                    AutoscalerSettings.newBuilder()
                        .apply {
                            if (params.minContainers != null) {
                                setMinContainers(params.minContainers)
                            }
                            if (params.maxContainers != null) {
                                setMaxContainers(params.maxContainers)
                            }
                            if (params.bufferContainers != null) {
                                setBufferContainers(params.bufferContainers)
                            }
                            if (scaledownWindowMs != null) {
                                setScaledownWindow((scaledownWindowMs / 1000).toInt())
                            }
                        }
                        .build(),
                )
                .build(),
        )
    }

    suspend fun getWebUrl(): String? {
        val url = handleMetadata?.webUrl ?: ""
        return url.ifEmpty { null }
    }

    suspend fun remote(
        args: List<Any?> = emptyList(),
        kwargs: Map<String, Any?> = emptyMap(),
    ): Any? {
        checkNoWebUrl("remote")
        val supportedFormats = handleMetadata?.supportedInputFormatsList ?: emptyList()
        if (supportedFormats.isNotEmpty() && !supportedFormats.contains(DataFormat.DATA_FORMAT_CBOR)) {
            throw InvalidError(
                "cannot call Modal Function from Kotlin SDK since it was deployed with an incompatible Python SDK version. Redeploy with Modal Python SDK >= 1.2",
            )
        }
        val input = createInput(args, kwargs)
        val inputPlaneUrl = handleMetadata?.inputPlaneUrl
        val invocation: Invocation = if (!inputPlaneUrl.isNullOrEmpty()) {
            InputPlaneInvocation.create(
                client,
                inputPlaneUrl,
                functionId,
                input,
            )
        } else {
            ControlPlaneInvocation.create(
                client,
                functionId,
                input,
                FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_SYNC,
            )
        }
        var retryCount = 0
        while (true) {
            try {
                return invocation.awaitOutput()
            } catch (error: Throwable) {
                if (error is InternalFailure && retryCount <= 8) {
                    invocation.retry(retryCount)
                    retryCount += 1
                    continue
                }
                throw error
            }
        }
    }

    suspend fun spawn(
        args: List<Any?> = emptyList(),
        kwargs: Map<String, Any?> = emptyMap(),
    ): FunctionCall {
        checkNoWebUrl("spawn")
        val input = createInput(args, kwargs)
        val invocation = ControlPlaneInvocation.create(
            client,
            functionId,
            input,
            FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_ASYNC,
        )
        return FunctionCall(client, invocation.functionCallId)
    }

    private fun checkNoWebUrl(name: String) {
        val webUrl = handleMetadata?.webUrl ?: ""
        if (webUrl.isNotEmpty()) {
            throw InvalidError(
                "A webhook Function cannot be invoked for remote execution with '.$name'. Invoke this Function via its web url '$webUrl' instead.",
            )
        }
    }

    private suspend fun createInput(
        args: List<Any?>,
        kwargs: Map<String, Any?>,
    ): FunctionInput {
        val payload = cborEncode(listOf(args, kwargs))
        val maxObjectSizeBytes = 2 * 1024 * 1024
        var blobId: String? = null
        if (payload.size > maxObjectSizeBytes) {
            blobId = blobUpload(client.cpClient, payload)
        }
        val methodName = handleMetadata?.useMethodName ?: ""
        val builder = FunctionInput.newBuilder()
            .setDataFormat(DataFormat.DATA_FORMAT_CBOR)
            .apply {
                if (blobId == null) {
                    setArgs(payload.toByteString())
                } else {
                    setArgsBlobId(blobId!!)
                }
                if (methodName.isNotEmpty()) {
                    setMethodName(methodName)
                }
            }
        return builder.build()
    }
}
