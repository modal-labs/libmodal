package com.modal.modalkt

import io.grpc.Status
import modal.client.Api

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
                Api.FunctionGetRequest.newBuilder()
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
    private val handleMetadata: Api.FunctionHandleMetadata? = null,
) {
    suspend fun getCurrentStats(): FunctionStats {
        val response = client.cpClient.functionGetCurrentStats(
            Api.FunctionGetCurrentStatsRequest.newBuilder()
                .setFunctionId(functionId)
                .build(),
        )
        return FunctionStats(response.backlog, response.numTotalTasks)
    }

    suspend fun updateAutoscaler(params: FunctionUpdateAutoscalerParams) {
        val scaledownWindowMs = params.scaledownWindowMs
        if (scaledownWindowMs != null && scaledownWindowMs % 1000 != 0L) {
            throw InvalidError("scaledownWindowMs must be a multiple of 1000ms, got $scaledownWindowMs")
        }

        client.cpClient.functionUpdateSchedulingParams(
            Api.FunctionUpdateSchedulingParamsRequest.newBuilder()
                .setFunctionId(functionId)
                .setWarmPoolSizeOverride(0)
                .setSettings(
                    Api.AutoscalerSettings.newBuilder()
                        .apply {
                            if (params.minContainers != null) {
                                minContainers = params.minContainers
                            }
                            if (params.maxContainers != null) {
                                maxContainers = params.maxContainers
                            }
                            if (params.bufferContainers != null) {
                                bufferContainers = params.bufferContainers
                            }
                            if (scaledownWindowMs != null) {
                                scaledownWindow = (scaledownWindowMs / 1000).toInt()
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
        if (supportedFormats.isNotEmpty() && !supportedFormats.contains(Api.DataFormat.DATA_FORMAT_CBOR)) {
            throw InvalidError(
                "cannot call Modal Function from Kotlin SDK since it was deployed with an incompatible Python SDK version. Redeploy with Modal Python SDK >= 1.2",
            )
        }
        throw UnsupportedOperationException("Function.remote is not implemented yet")
    }

    suspend fun spawn(
        args: List<Any?> = emptyList(),
        kwargs: Map<String, Any?> = emptyMap(),
    ): Any? {
        checkNoWebUrl("spawn")
        throw UnsupportedOperationException("Function.spawn is not implemented yet")
    }

    private fun checkNoWebUrl(name: String) {
        val webUrl = handleMetadata?.webUrl ?: ""
        if (webUrl.isNotEmpty()) {
            throw InvalidError(
                "A webhook Function cannot be invoked for remote execution with '.$name'. Invoke this Function via its web url '$webUrl' instead.",
            )
        }
    }
}
