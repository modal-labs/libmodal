package com.modal.modalkt

import kotlinx.coroutines.flow.collect
import modal.client.Api
import java.io.ByteArrayOutputStream

typealias SandboxFileMode = String

class SandboxFile(
    private val client: ModalClient,
    private val fileDescriptor: String,
    private val taskId: String,
) {
    suspend fun read(): ByteArray {
        val result = runFilesystemExec(
            client.cpClient,
            Api.ContainerFilesystemExecRequest.newBuilder()
                .setTaskId(taskId)
                .setFileReadRequest(
                    Api.ContainerFileReadRequest.newBuilder()
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
            Api.ContainerFilesystemExecRequest.newBuilder()
                .setTaskId(taskId)
                .setFileWriteRequest(
                    Api.ContainerFileWriteRequest.newBuilder()
                        .setFileDescriptor(fileDescriptor)
                        .setData(com.google.protobuf.ByteString.copyFrom(data))
                        .build(),
                )
                .build(),
        )
    }

    suspend fun flush() {
        runFilesystemExec(
            client.cpClient,
            Api.ContainerFilesystemExecRequest.newBuilder()
                .setTaskId(taskId)
                .setFileFlushRequest(
                    Api.ContainerFileFlushRequest.newBuilder()
                        .setFileDescriptor(fileDescriptor)
                        .build(),
                )
                .build(),
        )
    }

    suspend fun close() {
        runFilesystemExec(
            client.cpClient,
            Api.ContainerFilesystemExecRequest.newBuilder()
                .setTaskId(taskId)
                .setFileCloseRequest(
                    Api.ContainerFileCloseRequest.newBuilder()
                        .setFileDescriptor(fileDescriptor)
                        .build(),
                )
                .build(),
        )
    }
}

suspend fun runFilesystemExec(
    cpClient: ControlPlaneClient,
    request: Api.ContainerFilesystemExecRequest,
): FilesystemExecResult {
    val response = cpClient.containerFilesystemExec(request)
    val output = ByteArrayOutputStream()
    var retries = 10
    var completed = false

    while (!completed) {
        try {
            cpClient.containerFilesystemExecGetOutput(
                Api.ContainerFilesystemExecGetOutputRequest.newBuilder()
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
                    throw SandboxFilesystemError(batch.error.errorMessage)
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
    val response: Api.ContainerFilesystemExecResponse,
)
