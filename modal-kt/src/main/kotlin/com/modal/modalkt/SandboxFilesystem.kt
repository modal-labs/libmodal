package com.modal.modalkt

import java.io.ByteArrayOutputStream
import kotlinx.coroutines.flow.collect
import modal.client.*
import okio.ByteString.Companion.toByteString

typealias SandboxFileMode = String

class SandboxFile(
    private val client: ModalClient,
    private val fileDescriptor: String,
    private val taskId: String,
) {
    suspend fun read(): ByteArray {
        val result = runFilesystemExec(
            client.cpClient,
            ContainerFilesystemExecRequest.newBuilder()
                .setTaskId(taskId)
                .setFileReadRequest(
                    ContainerFileReadRequest.newBuilder()
                        .setFileDescriptor(fileDescriptor)
                        .build(),
                )
                .build(),
        )
        return result.chunks
    }

    suspend fun write(data: ByteArray) {
        runFilesystemExec(
            client.cpClient,
            ContainerFilesystemExecRequest.newBuilder()
                .setTaskId(taskId)
                .setFileWriteRequest(
                    ContainerFileWriteRequest.newBuilder()
                        .setFileDescriptor(fileDescriptor)
                        .setData(data.toByteString())
                        .build(),
                )
                .build(),
        )
    }

    suspend fun flush() {
        runFilesystemExec(
            client.cpClient,
            ContainerFilesystemExecRequest.newBuilder()
                .setTaskId(taskId)
                .setFileFlushRequest(
                    ContainerFileFlushRequest.newBuilder()
                        .setFileDescriptor(fileDescriptor)
                        .build(),
                )
                .build(),
        )
    }

    suspend fun close() {
        runFilesystemExec(
            client.cpClient,
            ContainerFilesystemExecRequest.newBuilder()
                .setTaskId(taskId)
                .setFileCloseRequest(
                    ContainerFileCloseRequest.newBuilder()
                        .setFileDescriptor(fileDescriptor)
                        .build(),
                )
                .build(),
        )
    }
}

suspend fun runFilesystemExec(
    cpClient: ControlPlaneClient,
    request: ContainerFilesystemExecRequest,
): FilesystemExecResult {
    val response = cpClient.containerFilesystemExec(request)
    val output = ByteArrayOutputStream()
    var retries = 10
    var completed = false

    while (!completed) {
        try {
            cpClient.containerFilesystemExecGetOutput(
                ContainerFilesystemExecGetOutputRequest.newBuilder()
                    .setExecId(response.execId)
                    .setTimeout(55f)
                    .build(),
            ).collect { batch ->
                for (chunk in batch.outputList) {
                    output.write(chunk.toByteArray())
                }
                if (batch.hasError()) {
                    if (retries > 0) {
                        retries -= 1
                        return@collect
                    }
                    val errorMessage = batch.error?.errorMessage ?: "Unknown filesystem error"
                    throw SandboxFilesystemError(errorMessage)
                }
                if (batch.eof) {
                    completed = true
                }
            }
        } catch (error: Throwable) {
            if (isRetryableGrpc(error) && retries > 0) {
                retries -= 1
            } else {
                throw error
            }
        }
    }

    return FilesystemExecResult(output.toByteArray(), response)
}

data class FilesystemExecResult(
    val chunks: ByteArray,
    val response: ContainerFilesystemExecResponse,
)
