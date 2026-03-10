package com.modal.modalkt

import modal.client.*

data class FunctionCallGetParams(
    val timeoutMs: Long? = null,
)

data class FunctionCallCancelParams(
    val terminateContainers: Boolean? = null,
)

class FunctionCallService(
    private val client: ModalClient,
) {
    suspend fun fromId(functionCallId: String): FunctionCall {
        return FunctionCall(client, functionCallId)
    }
}

class FunctionCall(
    private val client: ModalClient,
    val functionCallId: String,
) {
    suspend fun get(params: FunctionCallGetParams = FunctionCallGetParams()): Any? {
        checkForRenamedParams(
            emptyMap(),
            mapOf("timeout" to "timeoutMs"),
        )
        val invocation = ControlPlaneInvocation.fromFunctionCallId(client, functionCallId)
        return invocation.awaitOutput(params.timeoutMs)
    }

    suspend fun cancel(params: FunctionCallCancelParams = FunctionCallCancelParams()) {
        client.cpClient.functionCallCancel(
            FunctionCallCancelRequest.newBuilder()
                .setFunctionCallId(functionCallId)
                .apply {
                    if (params.terminateContainers != null) {
                        setTerminateContainers(params.terminateContainers)
                    }
                }
                .build(),
        )
    }
}
