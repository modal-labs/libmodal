package com.modal.modalkt

import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.runBlocking
import modal.client.Api
import kotlin.test.Test
import kotlin.test.assertEquals

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
            request as Api.ImageGetOrCreateRequest
            assertEquals("ap-test", request.appId)
            assertEquals(listOf("FROM alpine:3.21"), request.image.dockerfileCommandsList)
            Api.ImageGetOrCreateResponse.newBuilder()
                .setImageId("im-base")
                .setResult(successResult())
                .build()
        }
        mock.handleUnary("/ImageGetOrCreate") { request ->
            request as Api.ImageGetOrCreateRequest
            assertEquals(listOf("FROM base", "RUN echo layer1"), request.image.dockerfileCommandsList)
            assertEquals(0, request.image.secretIdsCount)
            assertEquals(1, request.image.baseImagesCount)
            assertEquals("im-base", request.image.baseImagesList.first().imageId)
            Api.ImageGetOrCreateResponse.newBuilder()
                .setImageId("im-layer1")
                .setResult(successResult())
                .build()
        }
        mock.handleUnary("/ImageGetOrCreate") { request ->
            request as Api.ImageGetOrCreateRequest
            assertEquals(listOf("FROM base", "RUN echo layer2"), request.image.dockerfileCommandsList)
            assertEquals(listOf("sc-test"), request.image.secretIdsList)
            assertEquals("im-layer1", request.image.baseImagesList.first().imageId)
            assertEquals("A100", request.image.gpuConfig.gpuType)
            assertEquals(true, request.forceBuild)
            Api.ImageGetOrCreateResponse.newBuilder()
                .setImageId("im-layer2")
                .setResult(successResult())
                .build()
        }
        mock.handleUnary("/ImageGetOrCreate") { request ->
            request as Api.ImageGetOrCreateRequest
            assertEquals(listOf("FROM base", "RUN echo layer3"), request.image.dockerfileCommandsList)
            assertEquals("im-layer2", request.image.baseImagesList.first().imageId)
            assertEquals(true, request.forceBuild)
            Api.ImageGetOrCreateResponse.newBuilder()
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

    private fun successResult(): Api.GenericResult {
        return Api.GenericResult.newBuilder()
            .setStatus(Api.GenericResult.GenericStatus.GENERIC_STATUS_SUCCESS)
            .build()
    }
}
