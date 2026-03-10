package com.modal.modalkt

import modal.client.*

private const val outputsTimeoutMs: Long = 55_000

internal interface Invocation {
    val functionCallId: String

    suspend fun awaitOutput(timeoutMs: Long? = null): Any?

    suspend fun retry(retryCount: Int)
}

internal class InputPlaneInvocation(
    private val client: ModalClient,
    private val ipClient: ControlPlaneClient,
    private val functionId: String,
    private val input: FunctionPutInputsItem,
    private var attemptToken: String,
) : Invocation {
    companion object {
        suspend fun create(
            client: ModalClient,
            inputPlaneUrl: String,
            functionId: String,
            input: FunctionInput,
        ): InputPlaneInvocation {
            val putInput = FunctionPutInputsItem.newBuilder()
                .setIdx(0)
                .setInput(input)
                .build()
            val ipClient = client.ipClient(inputPlaneUrl)
            val response = ipClient.attemptStart(
                AttemptStartRequest.newBuilder()
                    .setFunctionId(functionId)
                    .setInput(putInput)
                    .build(),
            )
            return InputPlaneInvocation(
                client = client,
                ipClient = ipClient,
                functionId = functionId,
                input = putInput,
                attemptToken = response.attemptToken,
            )
        }
    }

    override val functionCallId: String
        get() = attemptToken

    override suspend fun awaitOutput(timeoutMs: Long?): Any? {
        val start = System.currentTimeMillis()
        var pollTimeoutMs = outputsTimeoutMs
        if (timeoutMs != null) {
            pollTimeoutMs = minOf(pollTimeoutMs, timeoutMs)
        }

        while (true) {
            val output = ipClient.attemptAwait(
                AttemptAwaitRequest.newBuilder()
                    .setAttemptToken(attemptToken)
                    .setRequestedAt(System.currentTimeMillis() / 1000.0)
                    .setTimeoutSecs(pollTimeoutMs.toFloat() / 1000f)
                    .build(),
            ).output
            if (output != null) {
                return processResult(output.result ?: throw RemoteError("Missing function result"), output.dataFormat)
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
        val response = ipClient.attemptRetry(
            AttemptRetryRequest.newBuilder()
                .setFunctionId(functionId)
                .setInput(input)
                .setAttemptToken(attemptToken)
                .build(),
        )
        attemptToken = response.attemptToken
    }

    private suspend fun processResult(
        result: GenericResult,
        dataFormat: DataFormat,
    ): Any? {
        val data = when {
            result.hasData() -> result.data?.toByteArray()
            result.hasDataBlobId() -> blobDownload(client.cpClient, result.dataBlobId ?: "")
            else -> null
        }

        when (result.status) {
            GenericStatus.GENERIC_STATUS_TIMEOUT -> throw FunctionTimeoutError(result.exception)
            GenericStatus.GENERIC_STATUS_INTERNAL_FAILURE -> throw InternalFailure(result.exception)
            GenericStatus.GENERIC_STATUS_SUCCESS -> Unit
            else -> throw RemoteError(result.exception)
        }

        if (data == null || data.isEmpty()) {
            return null
        }
        return when (dataFormat) {
            DataFormat.DATA_FORMAT_CBOR -> cborDecode(data)
            DataFormat.DATA_FORMAT_GENERATOR_DONE -> null
            DataFormat.DATA_FORMAT_PICKLE -> throw InvalidError(
                "PICKLE output format is not supported - remote function must return CBOR format",
            )
            else -> throw InvalidError("Unsupported data format: ${dataFormat.number}")
        }
    }
}

internal class ControlPlaneInvocation(
    private val client: ModalClient,
    override val functionCallId: String,
    private val input: FunctionInput? = null,
    private val functionCallJwt: String? = null,
    private var inputJwt: String? = null,
) : Invocation {
    companion object {
        suspend fun create(
            client: ModalClient,
            functionId: String,
            input: FunctionInput,
            invocationType: FunctionCallInvocationType,
        ): ControlPlaneInvocation {
            val putInput = FunctionPutInputsItem.newBuilder()
                .setIdx(0)
                .setInput(input)
                .build()
            val response = client.cpClient.functionMap(
                FunctionMapRequest.newBuilder()
                    .setFunctionId(functionId)
                    .setFunctionCallType(FunctionCallType.FUNCTION_CALL_TYPE_UNARY)
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
                val result = item.result ?: throw RemoteError("Missing function result")
                return processResult(result, item.dataFormat)
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
            FunctionRetryInputsRequest.newBuilder()
                .setFunctionCallJwt(jwt)
                .addInputs(
                    FunctionRetryInputsItem.newBuilder()
                        .setInputJwt(inputJwt ?: "")
                        .setInput(originalInput)
                        .setRetryCount(retryCount)
                        .build(),
                )
                .build(),
        )
        inputJwt = response.inputJwtsList.firstOrNull()
    }

    private suspend fun getOutput(timeoutMs: Long): FunctionGetOutputsItem? {
        val response = client.cpClient.functionGetOutputs(
            FunctionGetOutputsRequest.newBuilder()
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
        result: GenericResult,
        dataFormat: DataFormat,
    ): Any? {
        val data = when {
            result.hasData() -> result.data?.toByteArray()
            result.hasDataBlobId() -> blobDownload(client.cpClient, result.dataBlobId ?: "")
            else -> null
        }

        when (result.status) {
            GenericStatus.GENERIC_STATUS_TIMEOUT -> {
                throw FunctionTimeoutError(result.exception)
            }
            GenericStatus.GENERIC_STATUS_INTERNAL_FAILURE -> {
                throw InternalFailure(result.exception)
            }
            GenericStatus.GENERIC_STATUS_SUCCESS -> Unit
            else -> {
                throw RemoteError(result.exception)
            }
        }

        if (data == null || data.isEmpty()) {
            return null
        }

        return when (dataFormat) {
            DataFormat.DATA_FORMAT_CBOR -> cborDecode(data)
            DataFormat.DATA_FORMAT_GENERATOR_DONE -> null
            DataFormat.DATA_FORMAT_PICKLE -> throw InvalidError(
                "PICKLE output format is not supported - remote function must return CBOR format",
            )
            else -> throw InvalidError("Unsupported data format: ${dataFormat.number}")
        }
    }
}
