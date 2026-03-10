package com.modal.modalkt

import io.grpc.Status
import io.grpc.StatusException
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.toList
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.test.advanceTimeBy
import kotlinx.coroutines.test.runTest
import modal.client.*
import okio.ByteString.Companion.toByteString

class QueueServiceTest {
    @Test
    fun queueDeleteSuccess() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/QueueGetOrCreate") {
            QueueGetOrCreateResponse.newBuilder().setQueueId("qu-test-123").build()
        }
        mock.handleUnary("/QueueDelete") { request ->
            request as QueueDeleteRequest
            assertEquals("qu-test-123", request.queueId)
            Unit
        }

        client.queues.delete("test-queue")
        mock.assertExhausted()
    }

    @Test
    fun queueDeleteAllowMissing() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/QueueGetOrCreate") {
            throw StatusException(Status.NOT_FOUND.withDescription("Queue 'missing' not found"))
        }

        client.queues.delete("missing", QueueDeleteParams(allowMissing = true))
    }

    @OptIn(ExperimentalCoroutinesApi::class)
    @Test
    fun queueEphemeralHeartbeatStopsAfterClose() = runTest {
        val (client, mock) = createMockModalClients(
            backgroundScope = backgroundScope,
            ephemeralHeartbeatSleepMs = 10,
        )
        var heartbeatCount = 0
        mock.handleUnary("/QueueGetOrCreate") {
            QueueGetOrCreateResponse.newBuilder().setQueueId("test-queue-id").build()
        }
        mock.handleUnary("/QueueHeartbeat") {
            heartbeatCount += 1
            Unit
        }

        val queue = client.queues.ephemeral()
        delay(1)
        assertEquals(1, heartbeatCount)
        queue.closeEphemeral()
        advanceTimeBy(100)
        assertEquals(1, heartbeatCount)
    }

    @Test
    fun queuePutGetAndIterate() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/QueuePut") { Unit }
        mock.handleUnary("/QueueLen") {
            QueueLenResponse.newBuilder().setLen(1).build()
        }
        mock.handleUnary("/QueueGet") {
            QueueGetResponse.newBuilder()
                .addValues(PickleCodec.encode(123).toByteString())
                .build()
        }
        mock.handleUnary("/QueuePut") { Unit }
        mock.handleUnary("/QueueNextItems") {
            QueueNextItemsResponse.newBuilder()
                .addItems(
                    QueueItem.newBuilder()
                        .setEntryId("1-0")
                        .setValue(PickleCodec.encode(1).toByteString())
                        .build(),
                )
                .addItems(
                    QueueItem.newBuilder()
                        .setEntryId("1-1")
                        .setValue(PickleCodec.encode(2).toByteString())
                        .build(),
                )
                .build()
        }
        mock.handleUnary("/QueueNextItems") {
            QueueNextItemsResponse.getDefaultInstance()
        }
        mock.handleUnary("/QueueNextItems") {
            QueueNextItemsResponse.getDefaultInstance()
        }

        val queue = Queue(client, "qu-123", "test")
        queue.put(123)
        assertEquals(1, queue.len())
        assertEquals(123, queue.get())
        queue.putMany(listOf(1, 2))
        assertEquals(listOf(1, 2), queue.iterate().toList())
    }

    @Test
    fun queuePartitionValidation() = runBlocking {
        val (client, _) = createMockModalClients()
        val queue = Queue(client, "qu-123", "test")
        assertFailsWith<InvalidError> {
            queue.put(1, QueuePutParams(partition = "a".repeat(65)))
        }
    }
}
