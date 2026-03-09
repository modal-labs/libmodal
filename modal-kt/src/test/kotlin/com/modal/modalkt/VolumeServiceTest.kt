package com.modal.modalkt

import io.grpc.Status
import io.grpc.StatusException
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.delay
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.test.advanceTimeBy
import kotlinx.coroutines.test.runTest
import modal.client.Api
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class VolumeServiceTest {
    @Test
    fun volumeDeleteSuccess() = runBlocking {
        val (client, mock) = createMockModalClients()

        mock.handleUnary("/VolumeGetOrCreate") {
            Api.VolumeGetOrCreateResponse.newBuilder()
                .setVolumeId("vo-test-123")
                .setMetadata(
                    Api.VolumeMetadata.newBuilder()
                        .setName("test-volume")
                        .build(),
                )
                .build()
        }
        mock.handleUnary("/VolumeDelete") { request ->
            request as Api.VolumeDeleteRequest
            assertEquals("vo-test-123", request.volumeId)
            com.google.protobuf.Empty.getDefaultInstance()
        }

        client.volumes.delete("test-volume")
        mock.assertExhausted()
    }

    @Test
    fun volumeDeleteAllowMissing() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/VolumeGetOrCreate") {
            throw NotFoundError("Volume 'missing' not found")
        }

        client.volumes.delete("missing", VolumeDeleteParams(allowMissing = true))
        mock.assertExhausted()
    }

    @Test
    fun volumeDeleteAllowMissingDeleteRpcNotFound() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/VolumeGetOrCreate") {
            Api.VolumeGetOrCreateResponse.newBuilder()
                .setVolumeId("vo-test-123")
                .setMetadata(Api.VolumeMetadata.newBuilder().setName("test-volume").build())
                .build()
        }
        mock.handleUnary("/VolumeDelete") {
            throw StatusException(Status.NOT_FOUND.withDescription("No Volume with ID 'vo-test-123' found"))
        }

        client.volumes.delete("test-volume", VolumeDeleteParams(allowMissing = true))
        mock.assertExhausted()
    }

    @Test
    fun volumeDeleteAllowMissingFalseThrows() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/VolumeGetOrCreate") {
            throw NotFoundError("Volume 'missing' not found")
        }

        assertFailsWith<NotFoundError> {
            client.volumes.delete("missing", VolumeDeleteParams(allowMissing = false))
        }
    }

    @OptIn(ExperimentalCoroutinesApi::class)
    @Test
    fun volumeEphemeralHeartbeatStopsAfterClose() = runTest {
        val (client, mock) = createMockModalClients(
            backgroundScope = backgroundScope,
            ephemeralHeartbeatSleepMs = 10,
        )
        var heartbeatCount = 0

        mock.handleUnary("/VolumeGetOrCreate") {
            Api.VolumeGetOrCreateResponse.newBuilder()
                .setVolumeId("vo-test-123")
                .build()
        }
        mock.handleUnary("/VolumeHeartbeat") {
            heartbeatCount += 1
            com.google.protobuf.Empty.getDefaultInstance()
        }

        val volume = client.volumes.ephemeral()
        delay(1)
        assertEquals(1, heartbeatCount)

        volume.closeEphemeral()
        advanceTimeBy(100)
        assertEquals(1, heartbeatCount)
    }
}
