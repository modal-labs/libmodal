package com.modal.modalkt

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.runBlocking
import modal.client.*
import modal.task_command_router.*

class SandboxExecTest {
    @Test
    fun sandboxGetTaskIdPolling() = runBlocking {
        val (client, control, router) = createMockModalClientsWithTaskRouter()
        control.handleUnary("/SandboxGetTaskId") {
            SandboxGetTaskIdResponse.getDefaultInstance()
        }
        control.handleUnary("/SandboxGetTaskId") {
            SandboxGetTaskIdResponse.newBuilder().setTaskId("ta-123").build()
        }
        router.handleUnary("/TaskExecStart") { TaskExecStartResponse.getDefaultInstance() }

        val sandbox = SandboxHandle(client, "sb-123")
        val process = sandbox.exec(listOf("echo", "hello"))
        kotlin.test.assertNotNull(process)
    }

    @Test
    fun sandboxGetTaskIdTerminated() = runBlocking {
        val (client, control, _) = createMockModalClientsWithTaskRouter()
        control.handleUnary("/SandboxGetTaskId") {
            SandboxGetTaskIdResponse.newBuilder()
                .setTaskResult(
                    GenericResult.newBuilder()
                        .setStatus(GenericResult.GenericStatus.GENERIC_STATUS_TERMINATED)
                        .setException("boom")
                        .build(),
                )
                .build()
        }

        val sandbox = SandboxHandle(client, "sb-123")
        assertFailsWith<InvalidError> {
            sandbox.exec(listOf("echo", "hello"))
        }
    }

    @Test
    fun sandboxExecStdoutAndWait() = runBlocking {
        val (client, control, router) = createMockModalClientsWithTaskRouter()
        control.handleUnary("/SandboxGetTaskId") {
            SandboxGetTaskIdResponse.newBuilder().setTaskId("ta-123").build()
        }
        router.handleUnary("/TaskExecStart") { TaskExecStartResponse.getDefaultInstance() }
        router.handleStreaming("/TaskExecStdioRead") {
            flow {
                emit(
                    TaskExecStdioReadResponse.newBuilder()
                        .setData(com.google.protobuf.ByteString.copyFrom("hello\n".toByteArray()))
                        .build(),
                )
            }
        }
        router.handleStreaming("/TaskExecStdioRead") { flow { } }
        router.handleUnary("/TaskExecWait") {
            TaskExecWaitResponse.newBuilder()
                .setCode(0)
                .build()
        }

        val sandbox = SandboxHandle(client, "sb-123")
        val process = sandbox.exec(listOf("echo", "hello"))
        assertEquals("hello\n", process.stdout.readText())
        assertEquals("", process.stdout.readText())
        assertEquals(0, process.wait())
    }

    @Test
    fun sandboxDetachForbidsExec() = runBlocking {
        val (client, _, _) = createMockModalClientsWithTaskRouter()
        val sandbox = SandboxHandle(client, "sb-123")
        sandbox.detach()
        assertFailsWith<ClientClosedError> {
            sandbox.exec(listOf("echo", "hello"))
        }
    }

    @Test
    fun sandboxStdinWritesIncrementIndex() = runBlocking {
        val (client, control, _) = createMockModalClientsWithTaskRouter()
        val seen = mutableListOf<Int>()
        control.handleUnary("/SandboxStdinWrite") { request ->
            request as SandboxStdinWriteRequest
            seen += request.index
            Unit
        }
        control.handleUnary("/SandboxStdinWrite") { request ->
            request as SandboxStdinWriteRequest
            seen += request.index
            Unit
        }

        val sandbox = SandboxHandle(client, "sb-123")
        sandbox.stdin.writeText("hello")
        sandbox.stdin.close()
        assertEquals(listOf(1, 2), seen)
    }

    @Test
    fun sandboxExecMergesEnvIntoSecretsAndSupportsIgnore() = runBlocking {
        val (client, control, router) = createMockModalClientsWithTaskRouter()
        control.handleUnary("/SandboxGetTaskId") {
            SandboxGetTaskIdResponse.newBuilder().setTaskId("ta-123").build()
        }
        control.handleUnary("/SecretGetOrCreate") {
            SecretGetOrCreateResponse.newBuilder()
                .setSecretId("st-env")
                .build()
        }
        router.handleUnary("/TaskExecStart") { request ->
            request as TaskExecStartRequest
            assertEquals(listOf("st-env"), request.secretIdsList)
            TaskExecStartResponse.getDefaultInstance()
        }
        router.handleUnary("/TaskExecWait") {
            TaskExecWaitResponse.newBuilder()
                .setCode(0)
                .build()
        }

        val sandbox = SandboxHandle(client, "sb-123")
        val process = sandbox.exec(
            listOf("echo", "hello"),
            SandboxExecParams(
                env = mapOf("A" to "1"),
                stdout = StdioBehavior.IGNORE,
                stderr = StdioBehavior.IGNORE,
            ),
        )
        assertEquals("", process.stdout.readText())
        assertEquals("", process.stderr.readText())
        assertEquals(0, process.wait())
    }
}
