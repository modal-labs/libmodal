package com.modal.modalkt

import io.grpc.Status
import io.grpc.StatusException
import kotlinx.coroutines.runBlocking
import modal.client.Api
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class ImageServiceExtraTest {
    @Test
    fun imageFromIdSuccess() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/ImageFromId") { request ->
            request as Api.ImageFromIdRequest
            assertEquals("im-123", request.imageId)
            Api.ImageFromIdResponse.newBuilder()
                .setImageId("im-123")
                .build()
        }

        val image = client.images.fromId("im-123")
        assertEquals("im-123", image.imageId)
    }

    @Test
    fun imageDeleteNotFound() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/ImageDelete") {
            throw StatusException(Status.NOT_FOUND.withDescription("No Image with ID 'im-missing' found"))
        }

        assertFailsWith<NotFoundError> {
            client.images.delete("im-missing")
        }
    }
}
