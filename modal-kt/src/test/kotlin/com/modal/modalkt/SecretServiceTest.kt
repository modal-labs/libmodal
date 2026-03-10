package com.modal.modalkt

import io.grpc.Status
import io.grpc.StatusException
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlinx.coroutines.runBlocking
import modal.client.*

class SecretServiceTest {
    @Test
    fun secretDeleteSuccess() = runBlocking {
        val (client, mock) = createMockModalClients()

        mock.handleUnary("/SecretGetOrCreate") {
            SecretGetOrCreateResponse.newBuilder()
                .setSecretId("st-test-123")
                .build()
        }
        mock.handleUnary("/SecretDelete") { request ->
            request as SecretDeleteRequest
            assertEquals("st-test-123", request.secretId)
            Unit
        }

        client.secrets.delete("test-secret")
        mock.assertExhausted()
    }

    @Test
    fun secretDeleteAllowMissing() = runBlocking {
        val (client, mock) = createMockModalClients()

        mock.handleUnary("/SecretGetOrCreate") {
            throw NotFoundError("Secret 'missing' not found")
        }

        client.secrets.delete("missing", SecretDeleteParams(allowMissing = true))
        mock.assertExhausted()
    }

    @Test
    fun secretDeleteAllowMissingOnDeleteRpcNotFound() = runBlocking {
        val (client, mock) = createMockModalClients()

        mock.handleUnary("/SecretGetOrCreate") {
            SecretGetOrCreateResponse.newBuilder()
                .setSecretId("st-test-123")
                .build()
        }
        mock.handleUnary("/SecretDelete") {
            throw StatusException(Status.NOT_FOUND.withDescription("No Secret with ID 'st-test-123' found"))
        }

        client.secrets.delete("test-secret", SecretDeleteParams(allowMissing = true))
        mock.assertExhausted()
    }

    @Test
    fun secretDeleteWithAllowMissingFalseThrows() = runBlocking {
        val (client, mock) = createMockModalClients()

        mock.handleUnary("/SecretGetOrCreate") {
            throw NotFoundError("Secret 'missing' not found")
        }

        assertFailsWith<NotFoundError> {
            client.secrets.delete("missing", SecretDeleteParams(allowMissing = false))
        }
    }

    @Test
    fun secretFromObjectInvalid() = runBlocking {
        val (client, _) = createMockModalClients()

        assertFailsWith<InvalidError> {
            client.secrets.fromObject(mapOf("key" to 123))
        }
    }

    @Test
    fun mergeEnvIntoSecretsMergesEnvWithExistingSecrets() = runBlocking {
        val (client, mock) = createMockModalClients()
        val existingSecret = Secret("st-existing")
        mock.handleUnary("/SecretGetOrCreate") {
            SecretGetOrCreateResponse.newBuilder()
                .setSecretId("st-generated")
                .build()
        }

        val result = mergeEnvIntoSecrets(client, mapOf("B" to "2", "C" to "3"), listOf(existingSecret))

        assertEquals(2, result.size)
        assertEquals(existingSecret, result[0])
        assertEquals("st-generated", result[1].secretId)
    }
}
