package com.modal.modalkt

import modal.client.Api

private const val outputsTimeoutMs: Long = 55_000

internal interface Invocation {
    val functionCallId: String

    suspend fun awaitOutput(timeoutMs: Long? = null): Any?

    suspend fun retry(retryCount: Int)
}

internal class ControlPlaneInvocation(
    private val client: ModalClient,
    override val functionCallId: String,
    private val input: Api.FunctionInput? = null,
    private val functionCallJwt: String? = null,
    private var inputJwt: String? = null,
) : Invocation {
    companion object {
        suspend fun create(
            client: ModalClient,
            functionId: String,
            input: Api.FunctionInput,
            invocationType: Api.FunctionCallInvocationType,
        ): ControlPlaneInvocation {
            val putInput = Api.FunctionPutInputsItem.newBuilder()
                .setIdx(0)
                .setInput(input)
                .build()
            val response = client.cpClient.functionMap(
                Api.FunctionMapRequest.newBuilder()
                    .setFunctionId(functionId)
                    .setFunctionCallType(Api.FunctionCallType.FUNCTION_CALL_TYPE_UNARY)
                    .setFunctionCallInvocationType(invocationType)
                    .addPipelinedInputs(putInput)
                    .build(),
            )
            return ControlPlaneInvocation(
                client = client,
                functionCallId = response.functionCallId,
                input = input,
                functionCallJwt = response.functionCallJwt,
                inputJwt = response.pipelinedInputsList.firstOrNull()?.inputJwt,
            )
        }

        fun fromFunctionCallId(
            client: ModalClient,
            functionCallId: String,
        ): ControlPlaneInvocation {
            return ControlPlaneInvocation(client, functionCallId)
        }
    }

    override suspend fun awaitOutput(timeoutMs: Long?): Any? {
        val start = System.currentTimeMillis()
        var pollTimeoutMs = outputsTimeoutMs
        if (timeoutMs != null) {
            pollTimeoutMs = minOf(pollTimeoutMs, timeoutMs)
        }

        while (true) {
            val item = getOutput(pollTimeoutMs)
            if (item != null) {
                return processResult(item.result, item.dataFormat)
            }

            if (timeoutMs != null) {
                val remaining = timeoutMs - (System.currentTimeMillis() - start)
                if (remaining <= 0) {
                    throw FunctionTimeoutError("Timeout exceeded: ${timeoutMs}ms")
                }
                pollTimeoutMs = minOf(outputsTimeoutMs, remaining)
            }
        }
    }

    override suspend fun retry(retryCount: Int) {
        val originalInput = input ?: throw InvalidError("Cannot retry Function invocation - input missing")
        val jwt = functionCallJwt ?: throw InvalidError("Cannot retry Function invocation - input jwt missing")
        val response = client.cpClient.functionRetryInputs(
            Api.FunctionRetryInputsRequest.newBuilder()
                .setFunctionCallJwt(jwt)
                .addInputs(
                    Api.FunctionRetryInputsItem.newBuilder()
                        .setInputJwt(inputJwt ?: "")
                        .setInput(originalInput)
                        .setRetryCount(retryCount)
                        .build(),
                )
                .build(),
        )
        inputJwt = response.inputJwtsList.firstOrNull()
    }

    private suspend fun getOutput(timeoutMs: Long): Api.FunctionGetOutputsItem? {
        val response = client.cpClient.functionGetOutputs(
            Api.FunctionGetOutputsRequest.newBuilder()
                .setFunctionCallId(functionCallId)
                .setMaxValues(1)
                .setTimeout(timeoutMs.toFloat() / 1000f)
                .setLastEntryId("0-0")
                .setClearOnSuccess(true)
                .setRequestedAt(System.currentTimeMillis() / 1000.0)
                .build(),
        )
        return response.outputsList.firstOrNull()
    }

    private suspend fun processResult(
        result: Api.GenericResult,
        dataFormat: Api.DataFormat,
    ): Any? {
        val data = when {
            result.hasData() -> result.data.toByteArray()
            else -> null
        }

        when (result.status) {
            Api.GenericResult.GenericStatus.GENERIC_STATUS_TIMEOUT -> {
                throw FunctionTimeoutError(result.exception)
            }
            Api.GenericResult.GenericStatus.GENERIC_STATUS_INTERNAL_FAILURE -> {
                throw InternalFailure(result.exception)
            }
            Api.GenericResult.GenericStatus.GENERIC_STATUS_SUCCESS -> Unit
            else -> {
                throw RemoteError(result.exception)
            }
        }

        if (data == null || data.isEmpty()) {
            return null
        }

        return when (dataFormat) {
            Api.DataFormat.DATA_FORMAT_CBOR -> cborDecode(data)
            Api.DataFormat.DATA_FORMAT_GENERATOR_DONE -> null
            Api.DataFormat.DATA_FORMAT_PICKLE -> throw InvalidError(
                "PICKLE output format is not supported - remote function must return CBOR format",
            )
            else -> throw InvalidError("Unsupported data format: ${dataFormat.number}")
        }
    }
}
