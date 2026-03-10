package com.modal.modalkt

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlinx.coroutines.runBlocking
import modal.client.*
import modal.task_command_router.*

class SandboxSnapshotTest {
    @Test
    fun snapshotFilesystem() = runBlocking {
        val (client, control, _) = createMockModalClientsWithTaskRouter()
        control.handleUnary("/SandboxSnapshotFs") {
            SandboxSnapshotFsResponse.newBuilder()
                .setImageId("im-123")
                .setResult(
                    GenericResult.newBuilder()
                        .setStatus(GenericResult.GenericStatus.GENERIC_STATUS_SUCCESS)
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
            SandboxGetTaskIdResponse.newBuilder().setTaskId("ta-123").build()
        }
        router.handleUnary("/TaskSnapshotDirectory") { request ->
            request as TaskSnapshotDirectoryRequest
            TaskSnapshotDirectoryResponse.newBuilder()
                .setImageId("im-snap")
                .build()
        }
        router.handleUnary("/TaskMountDirectory") { Unit }

        val sandbox = SandboxHandle(client, "sb-123")
        val image = sandbox.snapshotDirectory("/mnt/data")
        assertEquals("im-snap", image.imageId)
        sandbox.mountImage("/mnt/data", image)
    }
}
