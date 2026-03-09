package com.modal.modalkt

import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.runBlocking
import modal.client.Api
import modal.task_command_router.TaskCommandRouterOuterClass
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class SandboxExecTest {
    @Test
    fun sandboxGetTaskIdPolling() = runBlocking {
        val (client, control, router) = createMockModalClientsWithTaskRouter()
        control.handleUnary("/SandboxGetTaskId") {
            Api.SandboxGetTaskIdResponse.getDefaultInstance()
        }
        control.handleUnary("/SandboxGetTaskId") {
            Api.SandboxGetTaskIdResponse.newBuilder().setTaskId("ta-123").build()
        }
        router.handleUnary("/TaskExecStart") { TaskCommandRouterOuterClass.TaskExecStartResponse.getDefaultInstance() }

        val sandbox = SandboxHandle(client, "sb-123")
        val process = sandbox.exec(listOf("echo", "hello"))
        kotlin.test.assertNotNull(process)
    }

    @Test
    fun sandboxGetTaskIdTerminated() = runBlocking {
        val (client, control, _) = createMockModalClientsWithTaskRouter()
        control.handleUnary("/SandboxGetTaskId") {
            Api.SandboxGetTaskIdResponse.newBuilder()
                .setTaskResult(
                    Api.GenericResult.newBuilder()
                        .setStatus(Api.GenericResult.GenericStatus.GENERIC_STATUS_TERMINATED)
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
            Api.SandboxGetTaskIdResponse.newBuilder().setTaskId("ta-123").build()
        }
        router.handleUnary("/TaskExecStart") { TaskCommandRouterOuterClass.TaskExecStartResponse.getDefaultInstance() }
        router.handleStreaming("/TaskExecStdioRead") {
            flow {
                emit(
                    TaskCommandRouterOuterClass.TaskExecStdioReadResponse.newBuilder()
                        .setData(com.google.protobuf.ByteString.copyFrom("hello\n".toByteArray()))
                        .build(),
                )
            }
        }
        router.handleStreaming("/TaskExecStdioRead") { flow { } }
        router.handleUnary("/TaskExecWait") {
            TaskCommandRouterOuterClass.TaskExecWaitResponse.newBuilder()
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
}
