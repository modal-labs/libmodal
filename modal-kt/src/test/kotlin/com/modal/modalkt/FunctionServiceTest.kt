package com.modal.modalkt

import io.grpc.Status
import io.grpc.StatusException
import kotlinx.coroutines.runBlocking
import modal.client.Api
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class FunctionServiceTest {
    @Test
    fun functionRemoteSuccess() = runBlocking {
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

        val function = Function(
            client = client,
            functionId = "fid-123",
            handleMetadata = Api.FunctionHandleMetadata.newBuilder()
                .addSupportedInputFormats(Api.DataFormat.DATA_FORMAT_CBOR)
                .build(),
        )
        assertEquals("output: hello", function.remote(kwargs = mapOf("s" to "hello")))
    }

    @Test
    fun functionGetCurrentStats() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/FunctionGetCurrentStats") { request ->
            request as Api.FunctionGetCurrentStatsRequest
            assertEquals("fid-stats", request.functionId)
            Api.FunctionStats.newBuilder()
                .setBacklog(3)
                .setNumTotalTasks(7)
                .build()
        }

        val function = Function(client, "fid-stats")
        val stats = function.getCurrentStats()
        assertEquals(FunctionStats(3, 7), stats)
    }

    @Test
    fun functionUpdateAutoscaler() = runBlocking {
        val (client, mock) = createMockModalClients()

        mock.handleUnary("/FunctionUpdateSchedulingParams") { request ->
            request as Api.FunctionUpdateSchedulingParamsRequest
            assertEquals("fid-auto", request.functionId)
            assertEquals(1, request.settings.minContainers)
            assertEquals(10, request.settings.maxContainers)
            assertEquals(2, request.settings.bufferContainers)
            assertEquals(300, request.settings.scaledownWindow)
            com.google.protobuf.Empty.getDefaultInstance()
        }

        val function = Function(client, "fid-auto")
        function.updateAutoscaler(
            FunctionUpdateAutoscalerParams(
                minContainers = 1,
                maxContainers = 10,
                bufferContainers = 2,
                scaledownWindowMs = 300_000,
            ),
        )
    }

    @Test
    fun functionGetWebUrl() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/FunctionGet") { request ->
            request as Api.FunctionGetRequest
            assertEquals("libmodal-test-support", request.appName)
            assertEquals("web_endpoint", request.objectTag)
            Api.FunctionGetResponse.newBuilder()
                .setFunctionId("fid-web")
                .setHandleMetadata(
                    Api.FunctionHandleMetadata.newBuilder()
                        .setWebUrl("https://endpoint.internal")
                        .build(),
                )
                .build()
        }

        val function = client.functions.fromName("libmodal-test-support", "web_endpoint")
        assertEquals("https://endpoint.internal", function.getWebUrl())
    }

    @Test
    fun functionFromNameWithDotNotation() = runBlocking {
        val (client, _) = createMockModalClients()
        assertFailsWith<InvalidError> {
            client.functions.fromName("libmodal-test-support", "MyClass.myMethod")
        }
    }

    @Test
    fun functionCallPreCborVersionError() = runBlocking {
        val (client, _) = createMockModalClients()
        val function = Function(
            client = client,
            functionId = "fid-123",
            handleMetadata = Api.FunctionHandleMetadata.getDefaultInstance(),
        )
        assertFailsWith<InvalidError> {
            function.remote(kwargs = mapOf("s" to "hello"))
        }
    }

    @Test
    fun webEndpointRemoteCallError() = runBlocking {
        val function = Function(
            ModalClient(ModalClientParams(controlPlaneClient = MockControlPlaneClient(), authTokenProvider = MockControlPlaneClient())),
            "fid",
            null,
            Api.FunctionHandleMetadata.newBuilder()
                .setWebUrl("https://endpoint.internal")
                .build(),
        )

        assertFailsWith<InvalidError> {
            function.remote(listOf("hello"))
        }
        assertFailsWith<InvalidError> {
            function.spawn(listOf("hello"))
        }
    }

    @Test
    fun functionNotFound() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/FunctionGet") {
            throw StatusException(Status.NOT_FOUND.withDescription("missing"))
        }
        assertFailsWith<NotFoundError> {
            client.functions.fromName("libmodal-test-support", "not_a_real_function")
        }
    }
}
