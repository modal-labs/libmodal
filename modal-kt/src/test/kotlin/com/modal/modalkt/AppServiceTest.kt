package com.modal.modalkt

import io.grpc.Status
import io.grpc.StatusException
import kotlinx.coroutines.runBlocking
import modal.client.Api
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class AppServiceTest {
    @Test
    fun appFromNameSuccess() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/AppGetOrCreate") { request ->
            request as Api.AppGetOrCreateRequest
            assertEquals("my-app", request.appName)
            Api.AppGetOrCreateResponse.newBuilder()
                .setAppId("ap-123")
                .build()
        }

        val app = client.apps.fromName("my-app")
        assertEquals("ap-123", app.appId)
        assertEquals("my-app", app.name)
    }

    @Test
    fun appFromNameNotFound() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/AppGetOrCreate") {
            throw StatusException(Status.NOT_FOUND.withDescription("missing"))
        }
        assertFailsWith<NotFoundError> {
            client.apps.fromName("missing")
        }
    }
}
