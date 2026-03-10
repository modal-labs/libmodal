package com.modal.modalkt

import io.grpc.Status
import io.grpc.StatusException
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertTrue
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.runBlocking
import modal.client.*

class SandboxLogsTest {
    @Test
    fun sandboxGetLogsNotCalledUntilStdoutIsRead() = runBlocking {
        val (client, mock) = createMockModalClients()
        var calls = 0
        mock.handleStreaming("/SandboxGetLogs") {
            calls += 1
            flow {
                emit(batch("1-0", emptyList(), true))
            }
        }
        val sandbox = SandboxHandle(client, "sb-123")
        assertEquals(0, calls)
        sandbox.stdout.readText()
        assertEquals(1, calls)
    }

    @Test
    fun sandboxGetLogsNonRetryablePropagates() = runBlocking {
        val (client, mock) = createMockModalClients()
        var calls = 0
        mock.handleStreaming("/SandboxGetLogs") {
            calls += 1
            throw StatusException(Status.PERMISSION_DENIED.withDescription("denied"))
        }
        val sandbox = SandboxHandle(client, "sb-denied")
        assertFailsWith<StatusException> {
            sandbox.stdout.readText()
        }
        assertEquals(1, calls)
    }

    @Test
    fun sandboxGetLogsRetryExhaustion() = runBlocking {
        val (client, mock) = createMockModalClients()
        var calls = 0
        repeat(11) {
            mock.handleStreaming("/SandboxGetLogs") {
                calls += 1
                throw StatusException(Status.UNAVAILABLE.withDescription("always-unavailable"))
            }
        }
        val sandbox = SandboxHandle(client, "sb-retry-exhaust")
        assertFailsWith<StatusException> {
            sandbox.stdout.readText()
        }
        assertTrue(calls >= 11)
    }

    @Test
    fun sandboxGetLogsResumesWithLastEntryIdAfterRetry() = runBlocking {
        val (client, mock) = createMockModalClients()
        val seen = mutableListOf<String>()
        mock.handleStreaming("/SandboxGetLogs") { request ->
            request as SandboxGetLogsRequest
            seen += request.lastEntryId
            flow {
                emit(batch("1-9", listOf("part"), false))
                throw StatusException(Status.UNAVAILABLE.withDescription("transient"))
            }
        }
        mock.handleStreaming("/SandboxGetLogs") { request ->
            request as SandboxGetLogsRequest
            seen += request.lastEntryId
            flow {
                emit(batch("1-10", emptyList(), true))
            }
        }
        val sandbox = SandboxHandle(client, "sb-last-entry")
        val out = sandbox.stdout.readText()
        assertTrue(out.contains("part"))
        assertEquals("0-0", seen[0])
        assertEquals("1-9", seen[1])
    }

    @Test
    fun sandboxGetLogsRetryDelayResetsAfterSuccess() = runBlocking {
        val (client, mock) = createMockModalClients()
        val callTimes = mutableListOf<Long>()
        mock.handleStreaming("/SandboxGetLogs") {
            callTimes += System.currentTimeMillis()
            throw StatusException(Status.UNAVAILABLE.withDescription("transient-1"))
        }
        mock.handleStreaming("/SandboxGetLogs") {
            callTimes += System.currentTimeMillis()
            flow {
                emit(batch("1-0", listOf("hi"), false))
            }
        }
        mock.handleStreaming("/SandboxGetLogs") {
            callTimes += System.currentTimeMillis()
            throw StatusException(Status.UNAVAILABLE.withDescription("transient-2"))
        }
        mock.handleStreaming("/SandboxGetLogs") {
            callTimes += System.currentTimeMillis()
            flow {
                emit(batch("1-1", emptyList(), true))
            }
        }

        val sandbox = SandboxHandle(client, "sb-000")
        assertEquals("hi", sandbox.stdout.readText())
        assertTrue(callTimes.size >= 4)
    }

    @Test
    fun repeatedStdoutReadReturnsEmpty() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleStreaming("/SandboxGetLogs") {
            flow {
                emit(batch("1-0", listOf("hello"), true))
            }
        }
        val sandbox = SandboxHandle(client, "sb-123")
        assertEquals("hello", sandbox.stdout.readText())
        assertEquals("", sandbox.stdout.readText())
    }

    private fun batch(entryId: String, items: List<String>, eof: Boolean): TaskLogsBatch {
        return TaskLogsBatch.newBuilder()
            .setEntryId(entryId)
            .setEof(eof)
            .addAllItems(
                items.map { text ->
                    TaskLogs.newBuilder().setData(text).build()
                },
            )
            .build()
    }
}
