package com.modal.modalkt

import io.grpc.Status
import io.grpc.StatusException
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlinx.coroutines.runBlocking
import modal.client.*

class FunctionServiceTest {
    @Test
    fun functionRemoteSuccess() = runBlocking {
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

        val function = Function(
            client = client,
            functionId = "fid-123",
            handleMetadata = FunctionHandleMetadata.newBuilder()
                .addSupportedInputFormats(DataFormat.DATA_FORMAT_CBOR)
                .build(),
        )
        assertEquals("output: hello", function.remote(kwargs = mapOf("s" to "hello")))
    }

    @Test
    fun functionGetCurrentStats() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/FunctionGetCurrentStats") { request ->
            request as FunctionGetCurrentStatsRequest
            assertEquals("fid-stats", request.functionId)
            ProtoFunctionStats.newBuilder()
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
            request as FunctionUpdateSchedulingParamsRequest
            val settings = requireNotNull(request.settings)
            assertEquals("fid-auto", request.functionId)
            assertEquals(1, settings.minContainers)
            assertEquals(10, settings.maxContainers)
            assertEquals(2, settings.bufferContainers)
            assertEquals(300, settings.scaledownWindow)
            Unit
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
            request as FunctionGetRequest
            assertEquals("libmodal-test-support", request.appName)
            assertEquals("web_endpoint", request.objectTag)
            FunctionGetResponse.newBuilder()
                .setFunctionId("fid-web")
                .setHandleMetadata(
                    FunctionHandleMetadata.newBuilder()
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
            handleMetadata = FunctionHandleMetadata.getDefaultInstance(),
        )
        assertFailsWith<InvalidError> {
            function.remote(kwargs = mapOf("s" to "hello"))
        }
    }

    @Test
    fun functionCallInputPlane() = runBlocking {
        val control = MockControlPlaneClient()
        val inputPlane = MockControlPlaneClient()
        val modal = ModalClient(
            ModalClientParams(
                controlPlaneClient = control,
                authTokenProvider = control,
                inputPlaneClientFactory = { inputPlane },
            ),
        )
        inputPlane.handleUnary("/AttemptStart") {
            AttemptStartResponse.newBuilder()
                .setAttemptToken("attempt-1")
                .build()
        }
        inputPlane.handleUnary("/AttemptAwait") {
            AttemptAwaitResponse.newBuilder()
                .setOutput(
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

        val function = Function(
            client = modal,
            functionId = "fid-input",
            handleMetadata = FunctionHandleMetadata.newBuilder()
                .setInputPlaneUrl("https://input-plane.example")
                .addSupportedInputFormats(DataFormat.DATA_FORMAT_CBOR)
                .build(),
        )
        assertEquals("output: hello", function.remote(listOf("hello")))
    }

    @Test
    fun webEndpointRemoteCallError() = runBlocking {
        val function = Function(
            ModalClient(ModalClientParams(controlPlaneClient = MockControlPlaneClient(), authTokenProvider = MockControlPlaneClient())),
            "fid",
            null,
            FunctionHandleMetadata.newBuilder()
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
