package com.modal.modalkt

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.runBlocking
import modal.client.*

class ImageServiceTest {
    @Test
    fun dockerfileCommandsEmptyArrayNoOp() {
        val client = ModalClient(
            ModalClientParams(
                controlPlaneClient = MockControlPlaneClient(),
                authTokenProvider = MockControlPlaneClient(),
            ),
        )
        val image1 = client.images.fromRegistry("alpine:3.21")
        val image2 = image1.dockerfileCommands(emptyList())
        assertEquals(image1, image2)
    }

    @Test
    fun dockerfileCommandsCopyCommandValidation() {
        val client = ModalClient(
            ModalClientParams(
                controlPlaneClient = MockControlPlaneClient(),
                authTokenProvider = MockControlPlaneClient(),
            ),
        )
        client.images.fromRegistry("alpine:3.21")
            .dockerfileCommands(listOf("COPY --from=alpine:latest /etc/os-release /tmp/os-release"))

        kotlin.test.assertFailsWith<InvalidError> {
            client.images.fromRegistry("alpine:3.21")
                .dockerfileCommands(listOf("COPY ./file.txt /root/"))
        }
    }

    @Test
    fun dockerfileCommandsWithOptions() = runBlocking {
        val (client, mock) = createMockModalClients()

        mock.handleUnary("/ImageGetOrCreate") { request ->
            request as ImageGetOrCreateRequest
            assertEquals("ap-test", request.appId)
            val image = requireNotNull(request.image)
            assertEquals(listOf("FROM alpine:3.21"), image.dockerfileCommandsList)
            ImageGetOrCreateResponse.newBuilder()
                .setImageId("im-base")
                .setResult(successResult())
                .build()
        }
        mock.handleUnary("/ImageGetOrCreate") { request ->
            request as ImageGetOrCreateRequest
            val image = requireNotNull(request.image)
            assertEquals(listOf("FROM base", "RUN echo layer1"), image.dockerfileCommandsList)
            assertEquals(0, image.secretIdsCount)
            assertEquals(1, image.baseImagesCount)
            assertEquals("im-base", image.baseImagesList.first().imageId)
            ImageGetOrCreateResponse.newBuilder()
                .setImageId("im-layer1")
                .setResult(successResult())
                .build()
        }
        mock.handleUnary("/ImageGetOrCreate") { request ->
            request as ImageGetOrCreateRequest
            val image = requireNotNull(request.image)
            assertEquals(listOf("FROM base", "RUN echo layer2"), image.dockerfileCommandsList)
            assertEquals(listOf("sc-test"), image.secretIdsList)
            assertEquals("im-layer1", image.baseImagesList.first().imageId)
            assertEquals("A100", requireNotNull(image.gpuConfig).gpuType)
            assertEquals(true, request.forceBuild)
            ImageGetOrCreateResponse.newBuilder()
                .setImageId("im-layer2")
                .setResult(successResult())
                .build()
        }
        mock.handleUnary("/ImageGetOrCreate") { request ->
            request as ImageGetOrCreateRequest
            val image = requireNotNull(request.image)
            assertEquals(listOf("FROM base", "RUN echo layer3"), image.dockerfileCommandsList)
            assertEquals("im-layer2", image.baseImagesList.first().imageId)
            assertEquals(true, request.forceBuild)
            ImageGetOrCreateResponse.newBuilder()
                .setImageId("im-layer3")
                .setResult(successResult())
                .build()
        }

        val image = client.images
            .fromRegistry("alpine:3.21")
            .dockerfileCommands(listOf("RUN echo layer1"))
            .dockerfileCommands(
                listOf("RUN echo layer2"),
                ImageDockerfileCommandsParams(
                    secrets = listOf(Secret("sc-test", "test_secret")),
                    gpu = "A100",
                    forceBuild = true,
                ),
            )
            .dockerfileCommands(
                listOf("RUN echo layer3"),
                ImageDockerfileCommandsParams(forceBuild = true),
            )
            .build(App("ap-test", "libmodal-test"))

        assertEquals("im-layer3", image.imageId)
        mock.assertExhausted()
    }

    private fun successResult(): GenericResult {
        return GenericResult.newBuilder()
            .setStatus(GenericResult.GenericStatus.GENERIC_STATUS_SUCCESS)
            .build()
    }
}
