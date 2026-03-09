package com.modal.modalkt

import kotlinx.coroutines.runBlocking
import modal.client.Api
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class FunctionCallServiceTest {
    @Test
    fun functionCallGetTimeoutAndNull() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/FunctionGetOutputs") {
            Api.FunctionGetOutputsResponse.getDefaultInstance()
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
            Api.FunctionGetOutputsResponse.newBuilder()
                .addOutputs(
                    Api.FunctionGetOutputsItem.newBuilder()
                        .setResult(
                            Api.GenericResult.newBuilder()
                                .setStatus(Api.GenericResult.GenericStatus.GENERIC_STATUS_SUCCESS)
                                .setData(com.google.protobuf.ByteString.copyFrom(cborEncode("output: hello")))
                                .build(),
                        )
                        .setDataFormat(Api.DataFormat.DATA_FORMAT_CBOR)
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
            Api.FunctionMapResponse.newBuilder()
                .setFunctionCallId("fc-123")
                .setFunctionCallJwt("jwt")
                .addPipelinedInputs(
                    Api.FunctionPutInputsResponseItem.newBuilder()
                        .setInputJwt("input-jwt")
                        .build(),
                )
                .build()
        }

        val function = Function(
            client = client,
            functionId = "fid-123",
            handleMetadata = Api.FunctionHandleMetadata.newBuilder()
                .addSupportedInputFormats(Api.DataFormat.DATA_FORMAT_CBOR)
                .build(),
        )
        val call = function.spawn(kwargs = mapOf("s" to "hello"))
        assertEquals("fc-123", call.functionCallId)
    }
}
