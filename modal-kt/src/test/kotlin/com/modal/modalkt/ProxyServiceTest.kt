package com.modal.modalkt

import io.grpc.Status
import io.grpc.StatusException
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlinx.coroutines.runBlocking
import modal.client.*

class ProxyServiceTest {
    @Test
    fun proxyFromNameSuccess() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/ProxyGet") {
            ProxyGetResponse.newBuilder()
                .setProxy(modal.client.Proxy.newBuilder().setProxyId("pr-123").build())
                .build()
        }

        val proxy = client.proxies.fromName("libmodal-test-proxy")
        assertEquals("pr-123", proxy.proxyId)
    }

    @Test
    fun proxyNotFound() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/ProxyGet") {
            throw StatusException(Status.NOT_FOUND.withDescription("missing"))
        }

        assertFailsWith<NotFoundError> {
            client.proxies.fromName("non-existent-proxy-name")
        }
    }
}
