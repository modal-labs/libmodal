package com.modal.modalkt

import io.grpc.Status
import io.grpc.StatusException
import kotlinx.coroutines.runBlocking
import modal.client.Api
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class SandboxServiceTest {
    @Test
    fun createConnectToken() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/SandboxCreateConnectToken") {
            Api.SandboxCreateConnectTokenResponse.newBuilder()
                .setUrl("https://sandbox.modal.run")
                .setToken("token-123")
                .build()
        }

        val sandbox = SandboxHandle(client, "sb-123")
        val creds = sandbox.createConnectToken(SandboxCreateConnectTokenParams("abc"))
        assertEquals("https://sandbox.modal.run", creds.url)
        assertEquals("token-123", creds.token)
    }

    @Test
    fun sandboxDetachForbidsOperations() = runBlocking {
        val (client, _) = createMockModalClients()
        val sandbox = SandboxHandle(client, "sb-123")
        sandbox.detach()

        assertFailsWith<ClientClosedError> { sandbox.createConnectToken() }
        assertFailsWith<ClientClosedError> { sandbox.poll() }
        assertFailsWith<ClientClosedError> { sandbox.wait() }
        assertFailsWith<ClientClosedError> { sandbox.getTags() }
    }

    @Test
    fun sandboxTerminateThenDetach() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/SandboxTerminate") {
            Api.SandboxTerminateResponse.getDefaultInstance()
        }
        val sandbox = SandboxHandle(client, "sb-123")
        sandbox.terminate()
        sandbox.detach()
    }

    @Test
    fun namedSandboxNotFound() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/SandboxGetFromName") {
            throw StatusException(Status.NOT_FOUND.withDescription("missing"))
        }

        assertFailsWith<NotFoundError> {
            client.sandboxes.fromName("libmodal-test", "non-existent-sandbox")
        }
    }

    @Test
    fun sandboxSetAndGetTags() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/SandboxTagsSet") { com.google.protobuf.Empty.getDefaultInstance() }
        mock.handleUnary("/SandboxTagsGet") {
            Api.SandboxTagsGetResponse.newBuilder()
                .addTags(Api.SandboxTag.newBuilder().setTagName("key-a").setTagValue("A").build())
                .addTags(Api.SandboxTag.newBuilder().setTagName("key-b").setTagValue("B").build())
                .build()
        }

        val sandbox = SandboxHandle(client, "sb-123")
        sandbox.setTags(mapOf("key-a" to "A", "key-b" to "B"))
        val tags = sandbox.getTags()
        assertEquals(mapOf("key-a" to "A", "key-b" to "B"), tags)
    }

    @Test
    fun sandboxTunnels() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/SandboxGetTunnels") {
            Api.SandboxGetTunnelsResponse.newBuilder()
                .setResult(
                    Api.GenericResult.newBuilder()
                        .setStatus(Api.GenericResult.GenericStatus.GENERIC_STATUS_SUCCESS)
                        .build(),
                )
                .addTunnels(
                    Api.TunnelData.newBuilder()
                        .setContainerPort(8443)
                        .setHost("example.modal.host")
                        .setPort(443)
                        .build(),
                )
                .addTunnels(
                    Api.TunnelData.newBuilder()
                        .setContainerPort(8080)
                        .setHost("example2.modal.host")
                        .setPort(443)
                        .setUnencryptedHost("tcp.modal.host")
                        .setUnencryptedPort(12345)
                        .build(),
                )
                .build()
        }

        val sandbox = SandboxHandle(client, "sb-123")
        val tunnels = sandbox.tunnels()
        assertEquals("https://example.modal.host", tunnels[8443]?.url)
        assertEquals("tcp.modal.host", tunnels[8080]?.unencryptedHost)
    }
}
