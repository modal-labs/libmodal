package com.modal.modalkt

import io.grpc.Status
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.flow
import modal.client.Api
import modal.task_command_router.TaskCommandRouterOuterClass

data class SandboxListParams(
    val appId: String? = null,
    val tags: Map<String, String>? = null,
    val environment: String? = null,
)

data class SandboxFromNameParams(
    val environment: String? = null,
)

data class SandboxTerminateParams(
    val wait: Boolean = false,
)

data class SandboxCreateConnectTokenParams(
    val userMetadata: String? = null,
)

data class SandboxCreateConnectCredentials(
    val url: String,
    val token: String,
)

class SandboxService(
    private val client: ModalClient,
) {
    suspend fun create(
        app: App,
        image: Image,
        params: SandboxCreateParams = SandboxCreateParams(),
    ): SandboxHandle {
        image.build(app)
        val request = buildSandboxCreateRequestProto(app.appId, image.imageId, params)
        try {
            val response = client.cpClient.sandboxCreate(request)
            return SandboxHandle(client, response.sandboxId)
        } catch (error: Throwable) {
            if (statusCode(error) == Status.Code.ALREADY_EXISTS) {
                throw AlreadyExistsError(statusMessage(error))
            }
            throw error
        }
    }

    suspend fun fromId(sandboxId: String): SandboxHandle {
        try {
            client.cpClient.sandboxWait(
                Api.SandboxWaitRequest.newBuilder()
                    .setSandboxId(sandboxId)
                    .setTimeout(0f)
                    .build(),
            )
        } catch (error: Throwable) {
            if (statusCode(error) == Status.Code.NOT_FOUND) {
                throw NotFoundError("Sandbox with id: '$sandboxId' not found")
            }
            throw error
        }
        return SandboxHandle(client, sandboxId)
    }

    suspend fun fromName(
        appName: String,
        name: String,
        params: SandboxFromNameParams = SandboxFromNameParams(),
    ): SandboxHandle {
        try {
            val response = client.cpClient.sandboxGetFromName(
                Api.SandboxGetFromNameRequest.newBuilder()
                    .setSandboxName(name)
                    .setAppName(appName)
                    .setEnvironmentName(client.environmentName(params.environment))
                    .build(),
            )
            return SandboxHandle(client, response.sandboxId)
        } catch (error: Throwable) {
            if (statusCode(error) == Status.Code.NOT_FOUND) {
                throw NotFoundError("Sandbox with name '$name' not found in App '$appName'")
            }
            throw error
        }
    }

    suspend fun list(
        params: SandboxListParams = SandboxListParams(),
    ): List<SandboxHandle> {
        val response = client.cpClient.sandboxList(
            Api.SandboxListRequest.newBuilder()
                .apply {
                    if (params.appId != null) {
                        appId = params.appId
                    }
                    setEnvironmentName(client.environmentName(params.environment))
                    setIncludeFinished(false)
                    if (params.tags != null) {
                        addAllTags(
                            params.tags.map { (name, value) ->
                                Api.SandboxTag.newBuilder()
                                    .setTagName(name)
                                    .setTagValue(value)
                                    .build()
                            },
                        )
                    }
                }
                .build(),
        )
        return response.sandboxesList.map { SandboxHandle(client, it.id) }
    }
}

class SandboxHandle(
    private val client: ModalClient,
    val sandboxId: String,
) {
    private var attached = true
    private var taskId: String? = null
    private var commandRouter: TaskCommandRouter? = null

    val stdout: ModalReadStream by lazy {
        ModalReadStream {
            outputStream(Api.FileDescriptor.FILE_DESCRIPTOR_STDOUT)
        }
    }

    val stderr: ModalReadStream by lazy {
        ModalReadStream {
            outputStream(Api.FileDescriptor.FILE_DESCRIPTOR_STDERR)
        }
    }

    suspend fun createConnectToken(
        params: SandboxCreateConnectTokenParams = SandboxCreateConnectTokenParams(),
    ): SandboxCreateConnectCredentials {
        ensureAttached()
        val response = client.cpClient.sandboxCreateConnectToken(
            Api.SandboxCreateConnectTokenRequest.newBuilder()
                .setSandboxId(sandboxId)
                .apply {
                    if (params.userMetadata != null) {
                        userMetadata = params.userMetadata
                    }
                }
                .build(),
        )
        return SandboxCreateConnectCredentials(response.url, response.token)
    }

    suspend fun exec(
        command: List<String>,
        params: SandboxExecParams = SandboxExecParams(),
    ): ContainerProcess {
        ensureAttached()
        validateExecArgs(command)
        val currentTaskId = getTaskId()
        val router = getOrCreateCommandRouter(currentTaskId)
        val execId = java.util.UUID.randomUUID().toString()
        router.execStart(buildTaskExecStartRequestProto(currentTaskId, execId, command, params))
        return ContainerProcess(currentTaskId, execId, router)
    }

    suspend fun open(path: String, mode: SandboxFileMode = "r"): SandboxFile {
        ensureAttached()
        val currentTaskId = getTaskId()
        val result = runFilesystemExec(
            client.cpClient,
            Api.ContainerFilesystemExecRequest.newBuilder()
                .setTaskId(currentTaskId)
                .setFileOpenRequest(
                    Api.ContainerFileOpenRequest.newBuilder()
                        .setPath(path)
                        .setMode(mode)
                        .build(),
                )
                .build(),
        )
        val descriptor = result.response.fileDescriptor
        return SandboxFile(client, descriptor, currentTaskId)
    }

    suspend fun terminate(params: SandboxTerminateParams = SandboxTerminateParams()): Int? {
        ensureAttached()
        client.cpClient.sandboxTerminate(
            Api.SandboxTerminateRequest.newBuilder()
                .setSandboxId(sandboxId)
                .build(),
        )
        val exitCode = if (params.wait) wait() else null
        detach()
        return exitCode
    }

    suspend fun wait(): Int {
        ensureAttached()
        while (true) {
            val response = client.cpClient.sandboxWait(
                Api.SandboxWaitRequest.newBuilder()
                    .setSandboxId(sandboxId)
                    .setTimeout(10f)
                    .build(),
            )
            if (response.hasResult()) {
                return getReturnCode(response.result) ?: 0
            }
        }
    }

    suspend fun poll(): Int? {
        ensureAttached()
        val response = client.cpClient.sandboxWait(
            Api.SandboxWaitRequest.newBuilder()
                .setSandboxId(sandboxId)
                .setTimeout(0f)
                .build(),
        )
        return if (response.hasResult()) getReturnCode(response.result) else null
    }

    suspend fun tunnels(timeoutMs: Long = 50_000): Map<Int, Tunnel> {
        ensureAttached()
        val response = client.cpClient.sandboxGetTunnels(
            Api.SandboxGetTunnelsRequest.newBuilder()
                .setSandboxId(sandboxId)
                .setTimeout(timeoutMs.toFloat() / 1000f)
                .build(),
        )
        if (response.result.status == Api.GenericResult.GenericStatus.GENERIC_STATUS_TIMEOUT) {
            throw SandboxTimeoutError()
        }
        return response.tunnelsList.associate { tunnel ->
            tunnel.containerPort to Tunnel(
                tunnel.host,
                tunnel.port,
                tunnel.unencryptedHost,
                if (tunnel.hasUnencryptedPort()) tunnel.unencryptedPort else null,
            )
        }
    }

    suspend fun snapshotFilesystem(timeoutMs: Long = 55_000): Image {
        ensureAttached()
        val response = client.cpClient.sandboxSnapshotFs(
            Api.SandboxSnapshotFsRequest.newBuilder()
                .setSandboxId(sandboxId)
                .setTimeout(timeoutMs.toFloat() / 1000f)
                .build(),
        )
        if (response.result.status != Api.GenericResult.GenericStatus.GENERIC_STATUS_SUCCESS) {
            throw InvalidError(
                "Sandbox snapshot failed: ${response.result.exception.ifEmpty { "Unknown error" }}",
            )
        }
        if (response.imageId.isEmpty()) {
            throw InvalidError("Sandbox snapshot response missing `imageId`")
        }
        return Image(client, response.imageId, "")
    }

    suspend fun mountImage(path: String, image: Image? = null) {
        ensureAttached()
        if (image != null && image.imageId.isEmpty()) {
            throw InvalidError("Image must be built before mounting. Call `image.build(app)` first.")
        }
        val currentTaskId = getTaskId()
        val router = getOrCreateCommandRouter(currentTaskId)
        router.mountDirectory(
            TaskCommandRouterOuterClass.TaskMountDirectoryRequest.newBuilder()
                .setTaskId(currentTaskId)
                .setPath(com.google.protobuf.ByteString.copyFrom(path.toByteArray()))
                .setImageId(image?.imageId ?: "")
                .build(),
        )
    }

    suspend fun snapshotDirectory(path: String): Image {
        ensureAttached()
        val currentTaskId = getTaskId()
        val router = getOrCreateCommandRouter(currentTaskId)
        val response = router.snapshotDirectory(
            TaskCommandRouterOuterClass.TaskSnapshotDirectoryRequest.newBuilder()
                .setTaskId(currentTaskId)
                .setPath(com.google.protobuf.ByteString.copyFrom(path.toByteArray()))
                .build(),
        )
        if (response.imageId.isEmpty()) {
            throw InvalidError("Sandbox snapshot directory response missing `imageId`")
        }
        return Image(client, response.imageId, "")
    }

    suspend fun setTags(tags: Map<String, String>) {
        ensureAttached()
        client.cpClient.sandboxTagsSet(
            Api.SandboxTagsSetRequest.newBuilder()
                .setEnvironmentName(client.environmentName())
                .setSandboxId(sandboxId)
                .addAllTags(
                    tags.map { (name, value) ->
                        Api.SandboxTag.newBuilder()
                            .setTagName(name)
                            .setTagValue(value)
                            .build()
                    },
                )
                .build(),
        )
    }

    suspend fun getTags(): Map<String, String> {
        ensureAttached()
        val response = client.cpClient.sandboxTagsGet(
            Api.SandboxTagsGetRequest.newBuilder()
                .setSandboxId(sandboxId)
                .build(),
        )
        return response.tagsList.associate { it.tagName to it.tagValue }
    }

    fun detach() {
        attached = false
        commandRouter?.close()
    }

    private fun outputStream(fileDescriptor: Api.FileDescriptor) = flow {
        var lastEntryId = "0-0"
        var retriesRemaining = 10
        var delayMs = 10L
        var completed = false

        while (!completed) {
            try {
                client.cpClient.sandboxGetLogs(
                    Api.SandboxGetLogsRequest.newBuilder()
                        .setSandboxId(sandboxId)
                        .setFileDescriptor(fileDescriptor)
                        .setTimeout(55f)
                        .setLastEntryId(lastEntryId)
                        .build(),
                ).collect { batch ->
                    delayMs = 10
                    retriesRemaining = 10
                    lastEntryId = batch.entryId
                    for (item in batch.itemsList) {
                        emit(item.data.toByteArray())
                    }
                    if (batch.eof) {
                        completed = true
                    }
                }
            } catch (error: Throwable) {
                if (isRetryableGrpc(error) && retriesRemaining > 0) {
                    delay(delayMs)
                    delayMs *= 2
                    retriesRemaining -= 1
                } else {
                    throw error
                }
            }
        }
    }

    private suspend fun getTaskId(): String {
        taskId?.let { return it }
        repeat(600) {
            val response = client.cpClient.sandboxGetTaskId(
                Api.SandboxGetTaskIdRequest.newBuilder()
                    .setSandboxId(sandboxId)
                    .build(),
            )
            if (response.hasTaskResult()) {
                if (response.taskResult.status == Api.GenericResult.GenericStatus.GENERIC_STATUS_SUCCESS ||
                    response.taskResult.exception.isEmpty()
                ) {
                    throw InvalidError("Sandbox $sandboxId has already completed")
                }
                throw InvalidError("Sandbox $sandboxId has already completed with result: exception:\"${response.taskResult.exception}\"")
            }
            if (response.hasTaskId()) {
                taskId = response.taskId
                return response.taskId
            }
            delay(500)
        }
        throw InvalidError("Timed out waiting for task ID for Sandbox $sandboxId")
    }

    private suspend fun getOrCreateCommandRouter(taskId: String): TaskCommandRouter {
        commandRouter?.let { return it }
        val created = client.taskCommandRouterFactory(client, taskId)
        commandRouter = created
        return created
    }

    private fun ensureAttached() {
        if (!attached) {
            throw ClientClosedError()
        }
    }

    private fun getReturnCode(result: Api.GenericResult): Int? {
        return when (result.status) {
            Api.GenericResult.GenericStatus.GENERIC_STATUS_UNSPECIFIED -> null
            Api.GenericResult.GenericStatus.GENERIC_STATUS_TIMEOUT -> 124
            Api.GenericResult.GenericStatus.GENERIC_STATUS_TERMINATED -> 137
            else -> result.exitcode
        }
    }
}

class ContainerProcess(
    private val taskId: String,
    private val execId: String,
    private val commandRouter: TaskCommandRouter,
) {
    private var stdinOffset = 0L

    val stdin = ModalWriteStream(
        writeBlock = { bytes ->
            commandRouter.execStdinWrite(
                TaskCommandRouterOuterClass.TaskExecStdinWriteRequest.newBuilder()
                    .setTaskId(taskId)
                    .setExecId(execId)
                    .setOffset(stdinOffset)
                    .setData(com.google.protobuf.ByteString.copyFrom(bytes))
                    .setEof(false)
                    .build(),
            )
            stdinOffset += bytes.size
        },
        closeBlock = {
            commandRouter.execStdinWrite(
                TaskCommandRouterOuterClass.TaskExecStdinWriteRequest.newBuilder()
                    .setTaskId(taskId)
                    .setExecId(execId)
                    .setOffset(stdinOffset)
                    .setData(com.google.protobuf.ByteString.EMPTY)
                    .setEof(true)
                    .build(),
            )
        },
    )

    val stdout = ModalReadStream {
        stdio(TaskCommandRouterOuterClass.TaskExecStdioFileDescriptor.TASK_EXEC_STDIO_FILE_DESCRIPTOR_STDOUT)
    }

    val stderr = ModalReadStream {
        stdio(TaskCommandRouterOuterClass.TaskExecStdioFileDescriptor.TASK_EXEC_STDIO_FILE_DESCRIPTOR_STDERR)
    }

    suspend fun wait(): Int {
        val response = commandRouter.execWait(
            TaskCommandRouterOuterClass.TaskExecWaitRequest.newBuilder()
                .setTaskId(taskId)
                .setExecId(execId)
                .build(),
        )
        return when {
            response.hasCode() -> response.code
            response.hasSignal() -> 128 + response.signal
            else -> throw InvalidError("Unexpected exit status")
        }
    }

    private fun stdio(
        fileDescriptor: TaskCommandRouterOuterClass.TaskExecStdioFileDescriptor,
    ) = flow {
        commandRouter.execStdioRead(
            TaskCommandRouterOuterClass.TaskExecStdioReadRequest.newBuilder()
                .setTaskId(taskId)
                .setExecId(execId)
                .setOffset(0)
                .setFileDescriptor(fileDescriptor)
                .build(),
        ).collect { item ->
            emit(item.data.toByteArray())
        }
    }
}
