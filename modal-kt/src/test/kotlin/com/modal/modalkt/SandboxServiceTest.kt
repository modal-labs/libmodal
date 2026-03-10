package com.modal.modalkt

import io.grpc.Status
import io.grpc.StatusException
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlinx.coroutines.runBlocking
import modal.client.*

class SandboxServiceTest {
    @Test
    fun sandboxCreateMergesEnvIntoSecrets() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/SecretGetOrCreate") {
            SecretGetOrCreateResponse.newBuilder()
                .setSecretId("st-env")
                .build()
        }
        mock.handleUnary("/SandboxCreate") { request ->
            request as SandboxCreateRequest
            assertEquals(listOf("st-env"), requireNotNull(request.definition).secretIdsList)
            SandboxCreateResponse.newBuilder()
                .setSandboxId("sb-123")
                .build()
        }

        val app = App("ap-123", "app")
        val image = Image(client, "im-123", "")
        val sandbox = client.sandboxes.create(
            app,
            image,
            SandboxCreateParams(env = mapOf("A" to "1")),
        )
        assertEquals("sb-123", sandbox.sandboxId)
    }

    @Test
    fun createConnectToken() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/SandboxCreateConnectToken") {
            SandboxCreateConnectTokenResponse.newBuilder()
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
            SandboxTerminateResponse.getDefaultInstance()
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
        mock.handleUnary("/SandboxTagsSet") { Unit }
        mock.handleUnary("/SandboxTagsGet") {
            SandboxTagsGetResponse.newBuilder()
                .addTags(SandboxTag.newBuilder().setTagName("key-a").setTagValue("A").build())
                .addTags(SandboxTag.newBuilder().setTagName("key-b").setTagValue("B").build())
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
            SandboxGetTunnelsResponse.newBuilder()
                .setResult(
                    GenericResult.newBuilder()
                        .setStatus(GenericResult.GenericStatus.GENERIC_STATUS_SUCCESS)
                        .build(),
                )
                .addTunnels(
                    TunnelData.newBuilder()
                        .setContainerPort(8443)
                        .setHost("example.modal.host")
                        .setPort(443)
                        .build(),
                )
                .addTunnels(
                    TunnelData.newBuilder()
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
