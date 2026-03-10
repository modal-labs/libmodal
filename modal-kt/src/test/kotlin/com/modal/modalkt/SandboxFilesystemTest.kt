package com.modal.modalkt

import kotlin.test.Test
import kotlin.test.assertContentEquals
import kotlin.test.assertEquals
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.runBlocking
import modal.client.*
import okio.ByteString.Companion.toByteString

class SandboxFilesystemTest {
    @Test
    fun writeAndReadBinaryFile() = runBlocking {
        val (client, control, _) = createMockModalClientsWithTaskRouter()
        control.handleUnary("/SandboxGetTaskId") {
            SandboxGetTaskIdResponse.newBuilder().setTaskId("ta-123").build()
        }
        control.handleUnary("/ContainerFilesystemExec") {
            ContainerFilesystemExecResponse.newBuilder()
                .setExecId("exec-open-write")
                .setFileDescriptor("fd-write")
                .build()
        }
        control.handleStreaming("/ContainerFilesystemExecGetOutput") {
            flow {
                emit(FilesystemRuntimeOutputBatch.newBuilder().setEof(true).build())
            }
        }
        control.handleUnary("/ContainerFilesystemExec") {
            ContainerFilesystemExecResponse.newBuilder()
                .setExecId("exec-write")
                .build()
        }
        control.handleStreaming("/ContainerFilesystemExecGetOutput") {
            flow {
                emit(FilesystemRuntimeOutputBatch.newBuilder().setEof(true).build())
            }
        }
        control.handleUnary("/ContainerFilesystemExec") {
            ContainerFilesystemExecResponse.newBuilder()
                .setExecId("exec-close-write")
                .build()
        }
        control.handleStreaming("/ContainerFilesystemExecGetOutput") {
            flow {
                emit(FilesystemRuntimeOutputBatch.newBuilder().setEof(true).build())
            }
        }
        control.handleUnary("/ContainerFilesystemExec") {
            ContainerFilesystemExecResponse.newBuilder()
                .setExecId("exec-open-read")
                .setFileDescriptor("fd-read")
                .build()
        }
        control.handleStreaming("/ContainerFilesystemExecGetOutput") {
            flow {
                emit(FilesystemRuntimeOutputBatch.newBuilder().setEof(true).build())
            }
        }
        val testData = byteArrayOf(1, 2, 3, 4, 5)
        control.handleUnary("/ContainerFilesystemExec") {
            ContainerFilesystemExecResponse.newBuilder()
                .setExecId("exec-read")
                .build()
        }
        control.handleStreaming("/ContainerFilesystemExecGetOutput") {
            flow {
                emit(
                    FilesystemRuntimeOutputBatch.newBuilder()
                        .addOutput(testData.toByteString())
                        .setEof(true)
                        .build(),
                )
            }
        }

        val sandbox = SandboxHandle(client, "sb-123")
        val writeHandle = sandbox.open("/tmp/test.bin", "w")
        writeHandle.write(testData)
        writeHandle.close()
        val readHandle = sandbox.open("/tmp/test.bin", "r")
        val read = readHandle.read()
        assertContentEquals(testData, read)
    }

    @Test
    fun fileHandleFlush() = runBlocking {
        val (client, control, _) = createMockModalClientsWithTaskRouter()
        control.handleUnary("/SandboxGetTaskId") {
            SandboxGetTaskIdResponse.newBuilder().setTaskId("ta-123").build()
        }
        control.handleUnary("/ContainerFilesystemExec") {
            ContainerFilesystemExecResponse.newBuilder().setExecId("open").setFileDescriptor("fd").build()
        }
        control.handleStreaming("/ContainerFilesystemExecGetOutput") { flow { emit(FilesystemRuntimeOutputBatch.newBuilder().setEof(true).build()) } }
        control.handleUnary("/ContainerFilesystemExec") {
            ContainerFilesystemExecResponse.newBuilder().setExecId("flush").build()
        }
        control.handleStreaming("/ContainerFilesystemExecGetOutput") { flow { emit(FilesystemRuntimeOutputBatch.newBuilder().setEof(true).build()) } }

        val sandbox = SandboxHandle(client, "sb-123")
        val handle = sandbox.open("/tmp/flush.txt", "w")
        handle.flush()
    }
}
