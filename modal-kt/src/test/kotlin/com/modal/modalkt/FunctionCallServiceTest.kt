package com.modal.modalkt

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlinx.coroutines.runBlocking
import modal.client.*

class FunctionCallServiceTest {
    @Test
    fun functionCallGetTimeoutAndNull() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/FunctionGetOutputs") {
            FunctionGetOutputsResponse.getDefaultInstance()
        }
        val functionCall = FunctionCall(client, "fc-123")
        assertFailsWith<FunctionTimeoutError> {
            functionCall.get(FunctionCallGetParams(timeoutMs = 0))
        }
    }

    @Test
    fun functionCallGetSuccess() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/FunctionGetOutputs") {
            FunctionGetOutputsResponse.newBuilder()
                .addOutputs(
                    FunctionGetOutputsItem.newBuilder()
                        .setResult(
                            GenericResult.newBuilder()
                                .setStatus(GenericResult.GenericStatus.GENERIC_STATUS_SUCCESS)
                                .setData(com.google.protobuf.ByteString.copyFrom(cborEncode("output: hello")))
                                .build(),
                        )
                        .setDataFormat(DataFormat.DATA_FORMAT_CBOR)
                        .build(),
                )
                .build()
        }

        val functionCall = FunctionCall(client, "fc-123")
        assertEquals("output: hello", functionCall.get())
    }

    @Test
    fun functionSpawnCreatesFunctionCall() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/FunctionMap") {
            FunctionMapResponse.newBuilder()
                .setFunctionCallId("fc-123")
                .setFunctionCallJwt("jwt")
                .addPipelinedInputs(
                    FunctionPutInputsResponseItem.newBuilder()
                        .setInputJwt("input-jwt")
                        .build(),
                )
                .build()
        }

        val function = Function(
            client = client,
            functionId = "fid-123",
            handleMetadata = FunctionHandleMetadata.newBuilder()
                .addSupportedInputFormats(DataFormat.DATA_FORMAT_CBOR)
                .build(),
        )
        val call = function.spawn(kwargs = mapOf("s" to "hello"))
        assertEquals("fc-123", call.functionCallId)
    }
}
