package com.modal.modalkt

import kotlinx.coroutines.runBlocking
import modal.client.Api
import modal.task_command_router.TaskCommandRouterOuterClass
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class SandboxSnapshotTest {
    @Test
    fun snapshotFilesystem() = runBlocking {
        val (client, control, _) = createMockModalClientsWithTaskRouter()
        control.handleUnary("/SandboxSnapshotFs") {
            Api.SandboxSnapshotFsResponse.newBuilder()
                .setImageId("im-123")
                .setResult(
                    Api.GenericResult.newBuilder()
                        .setStatus(Api.GenericResult.GenericStatus.GENERIC_STATUS_SUCCESS)
                        .build(),
                )
                .build()
        }

        val sandbox = SandboxHandle(client, "sb-123")
        val image = sandbox.snapshotFilesystem()
        assertEquals("im-123", image.imageId)
    }

    @Test
    fun mountImageWithUnbuiltImageThrows() = runBlocking {
        val (client, _, _) = createMockModalClientsWithTaskRouter()
        val sandbox = SandboxHandle(client, "sb-123")
        val image = client.images.fromRegistry("alpine:3.21")
        assertFailsWith<InvalidError> {
            sandbox.mountImage("/mnt/data", image)
        }
    }

    @Test
    fun snapshotDirectoryAndMountImage() = runBlocking {
        val (client, control, router) = createMockModalClientsWithTaskRouter()
        control.handleUnary("/SandboxGetTaskId") {
            Api.SandboxGetTaskIdResponse.newBuilder().setTaskId("ta-123").build()
        }
        router.handleUnary("/TaskSnapshotDirectory") { request ->
            request as TaskCommandRouterOuterClass.TaskSnapshotDirectoryRequest
            TaskCommandRouterOuterClass.TaskSnapshotDirectoryResponse.newBuilder()
                .setImageId("im-snap")
                .build()
        }
        router.handleUnary("/TaskMountDirectory") { com.google.protobuf.Empty.getDefaultInstance() }

        val sandbox = SandboxHandle(client, "sb-123")
        val image = sandbox.snapshotDirectory("/mnt/data")
        assertEquals("im-snap", image.imageId)
        sandbox.mountImage("/mnt/data", image)
    }
}
