package com.modal.modalkt

import io.grpc.Status
import io.grpc.StatusException
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlinx.coroutines.runBlocking
import modal.client.*

class AppServiceTest {
    @Test
    fun appFromNameSuccess() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/AppGetOrCreate") { request ->
            request as AppGetOrCreateRequest
            assertEquals("my-app", request.appName)
            AppGetOrCreateResponse.newBuilder()
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
