package com.modal.modalkt

import java.util.ArrayDeque
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.emptyFlow
import modal.client.*
import modal.task_command_router.*

class MockControlPlaneClient : ControlPlaneClient {
    private val unaryHandlers = mutableMapOf<String, ArrayDeque<suspend (Any) -> Any>>()
    private val streamingHandlers = mutableMapOf<String, ArrayDeque<suspend (Any) -> Flow<Any>>>()

    fun handleUnary(
        method: String,
        handler: suspend (Any) -> Any,
    ) {
        val queue = unaryHandlers.getOrPut(method) { ArrayDeque() }
        queue.addLast(handler)
    }

    fun handleStreaming(
        method: String,
        handler: suspend (Any) -> Flow<Any>,
    ) {
        val queue = streamingHandlers.getOrPut(method) { ArrayDeque() }
        queue.addLast(handler)
    }

    fun assertExhausted() {
        val outstandingUnary = unaryHandlers.filterValues { it.isNotEmpty() }
        val outstandingStreaming = streamingHandlers.filterValues { it.isNotEmpty() }
        if (outstandingUnary.isNotEmpty() || outstandingStreaming.isNotEmpty()) {
            throw AssertionError(
                buildString {
                    append("Not all expected gRPC calls were made:")
                    for ((key, queue) in outstandingUnary) {
                        append("\n- $key: ${queue.size} expectation(s) remaining")
                    }
                    for ((key, queue) in outstandingStreaming) {
                        append("\n- $key: ${queue.size} expectation(s) remaining")
                    }
                },
            )
        }
    }

    private suspend fun unary(method: String, request: Any): Any {
        val queue = unaryHandlers[method] ?: error("Unexpected gRPC call: $method with request $request")
        if (queue.isEmpty()) {
            error("Unexpected gRPC call: $method with request $request")
        }
        val handler = queue.removeFirst()
        return handler(request)
    }

    @Suppress("UNCHECKED_CAST")
    private suspend fun <T> streaming(method: String, request: Any): Flow<T> {
        val queue = streamingHandlers[method] ?: return emptyFlow()
        if (queue.isEmpty()) {
            error("Unexpected gRPC call: $method with request $request")
        }
        val handler = queue.removeFirst()
        return handler(request) as Flow<T>
    }

    override suspend fun appGetOrCreate(request: AppGetOrCreateRequest): AppGetOrCreateResponse {
        return unary("/AppGetOrCreate", request) as AppGetOrCreateResponse
    }

    override suspend fun secretGetOrCreate(request: SecretGetOrCreateRequest): SecretGetOrCreateResponse {
        return unary("/SecretGetOrCreate", request) as SecretGetOrCreateResponse
    }

    override suspend fun secretDelete(request: SecretDeleteRequest) {
        unary("/SecretDelete", request)
    }

    override suspend fun volumeGetOrCreate(request: VolumeGetOrCreateRequest): VolumeGetOrCreateResponse {
        return unary("/VolumeGetOrCreate", request) as VolumeGetOrCreateResponse
    }

    override suspend fun volumeDelete(request: VolumeDeleteRequest) {
        unary("/VolumeDelete", request)
    }

    override suspend fun volumeHeartbeat(request: VolumeHeartbeatRequest) {
        unary("/VolumeHeartbeat", request)
    }

    override suspend fun proxyGet(request: ProxyGetRequest): ProxyGetResponse {
        return unary("/ProxyGet", request) as ProxyGetResponse
    }

    override suspend fun imageFromId(request: ImageFromIdRequest): ImageFromIdResponse {
        return unary("/ImageFromId", request) as ImageFromIdResponse
    }

    override suspend fun imageDelete(request: ImageDeleteRequest) {
        unary("/ImageDelete", request)
    }

    override suspend fun imageGetOrCreate(request: ImageGetOrCreateRequest): ImageGetOrCreateResponse {
        return unary("/ImageGetOrCreate", request) as ImageGetOrCreateResponse
    }

    override suspend fun imageJoinStreaming(request: ImageJoinStreamingRequest): Flow<ImageJoinStreamingResponse> {
        return streaming("/ImageJoinStreaming", request)
    }

    override suspend fun functionGet(request: FunctionGetRequest): FunctionGetResponse {
        return unary("/FunctionGet", request) as FunctionGetResponse
    }

    override suspend fun functionGetCurrentStats(request: FunctionGetCurrentStatsRequest): ProtoFunctionStats {
        return unary("/FunctionGetCurrentStats", request) as ProtoFunctionStats
    }

    override suspend fun functionUpdateSchedulingParams(request: FunctionUpdateSchedulingParamsRequest) {
        unary("/FunctionUpdateSchedulingParams", request)
    }

    override suspend fun functionBindParams(request: FunctionBindParamsRequest): FunctionBindParamsResponse {
        return unary("/FunctionBindParams", request) as FunctionBindParamsResponse
    }

    override suspend fun sandboxCreate(request: SandboxCreateRequest): SandboxCreateResponse {
        return unary("/SandboxCreate", request) as SandboxCreateResponse
    }

    override suspend fun sandboxGetLogs(request: SandboxGetLogsRequest): Flow<TaskLogsBatch> {
        return streaming("/SandboxGetLogs", request)
    }

    override suspend fun sandboxWait(request: SandboxWaitRequest): SandboxWaitResponse {
        return unary("/SandboxWait", request) as SandboxWaitResponse
    }

    override suspend fun sandboxGetTaskId(request: SandboxGetTaskIdRequest): SandboxGetTaskIdResponse {
        return unary("/SandboxGetTaskId", request) as SandboxGetTaskIdResponse
    }

    override suspend fun containerFilesystemExec(
        request: ContainerFilesystemExecRequest,
    ): ContainerFilesystemExecResponse {
        return unary("/ContainerFilesystemExec", request) as ContainerFilesystemExecResponse
    }

    override suspend fun containerFilesystemExecGetOutput(
        request: ContainerFilesystemExecGetOutputRequest,
    ): Flow<FilesystemRuntimeOutputBatch> {
        return streaming("/ContainerFilesystemExecGetOutput", request)
    }

    override suspend fun sandboxSnapshotFs(request: SandboxSnapshotFsRequest): SandboxSnapshotFsResponse {
        return unary("/SandboxSnapshotFs", request) as SandboxSnapshotFsResponse
    }

    override suspend fun sandboxGetFromName(request: SandboxGetFromNameRequest): SandboxGetFromNameResponse {
        return unary("/SandboxGetFromName", request) as SandboxGetFromNameResponse
    }

    override suspend fun sandboxList(request: SandboxListRequest): SandboxListResponse {
        return unary("/SandboxList", request) as SandboxListResponse
    }

    override suspend fun sandboxTerminate(request: SandboxTerminateRequest): SandboxTerminateResponse {
        return unary("/SandboxTerminate", request) as SandboxTerminateResponse
    }

    override suspend fun sandboxGetTunnels(request: SandboxGetTunnelsRequest): SandboxGetTunnelsResponse {
        return unary("/SandboxGetTunnels", request) as SandboxGetTunnelsResponse
    }

    override suspend fun sandboxCreateConnectToken(
        request: SandboxCreateConnectTokenRequest,
    ): SandboxCreateConnectTokenResponse {
        return unary("/SandboxCreateConnectToken", request) as SandboxCreateConnectTokenResponse
    }

    override suspend fun sandboxStdinWrite(request: SandboxStdinWriteRequest) {
        unary("/SandboxStdinWrite", request)
    }

    override suspend fun sandboxTagsSet(request: SandboxTagsSetRequest) {
        unary("/SandboxTagsSet", request)
    }

    override suspend fun sandboxTagsGet(request: SandboxTagsGetRequest): SandboxTagsGetResponse {
        return unary("/SandboxTagsGet", request) as SandboxTagsGetResponse
    }

    override suspend fun queueGetOrCreate(request: QueueGetOrCreateRequest): QueueGetOrCreateResponse {
        return unary("/QueueGetOrCreate", request) as QueueGetOrCreateResponse
    }

    override suspend fun queueDelete(request: QueueDeleteRequest) {
        unary("/QueueDelete", request)
    }

    override suspend fun queueHeartbeat(request: QueueHeartbeatRequest) {
        unary("/QueueHeartbeat", request)
    }

    override suspend fun queueGet(request: QueueGetRequest): QueueGetResponse {
        return unary("/QueueGet", request) as QueueGetResponse
    }

    override suspend fun queuePut(request: QueuePutRequest) {
        unary("/QueuePut", request)
    }

    override suspend fun queueLen(request: QueueLenRequest): QueueLenResponse {
        return unary("/QueueLen", request) as QueueLenResponse
    }

    override suspend fun queueNextItems(request: QueueNextItemsRequest): QueueNextItemsResponse {
        return unary("/QueueNextItems", request) as QueueNextItemsResponse
    }

    override suspend fun queueClear(request: QueueClearRequest) {
        unary("/QueueClear", request)
    }

    override suspend fun functionMap(request: FunctionMapRequest): FunctionMapResponse {
        return unary("/FunctionMap", request) as FunctionMapResponse
    }

    override suspend fun functionGetOutputs(request: FunctionGetOutputsRequest): FunctionGetOutputsResponse {
        return unary("/FunctionGetOutputs", request) as FunctionGetOutputsResponse
    }

    override suspend fun functionRetryInputs(request: FunctionRetryInputsRequest): FunctionRetryInputsResponse {
        return unary("/FunctionRetryInputs", request) as FunctionRetryInputsResponse
    }

    override suspend fun functionCallCancel(request: FunctionCallCancelRequest) {
        unary("/FunctionCallCancel", request)
    }

    override suspend fun attemptStart(request: AttemptStartRequest): AttemptStartResponse {
        return unary("/AttemptStart", request) as AttemptStartResponse
    }

    override suspend fun attemptAwait(request: AttemptAwaitRequest): AttemptAwaitResponse {
        return unary("/AttemptAwait", request) as AttemptAwaitResponse
    }

    override suspend fun attemptRetry(request: AttemptRetryRequest): AttemptRetryResponse {
        return unary("/AttemptRetry", request) as AttemptRetryResponse
    }

    override suspend fun blobCreate(request: BlobCreateRequest): BlobCreateResponse {
        return unary("/BlobCreate", request) as BlobCreateResponse
    }

    override suspend fun blobGet(request: BlobGetRequest): BlobGetResponse {
        return unary("/BlobGet", request) as BlobGetResponse
    }

    override suspend fun authTokenGet(request: AuthTokenGetRequest): AuthTokenGetResponse {
        return unary("/AuthTokenGet", request) as AuthTokenGetResponse
    }

    override suspend fun taskGetCommandRouterAccess(
        request: TaskGetCommandRouterAccessRequest,
    ): TaskGetCommandRouterAccessResponse {
        return unary("/TaskGetCommandRouterAccess", request) as TaskGetCommandRouterAccessResponse
    }

    override fun close() {
    }
}

fun createMockModalClients(
    backgroundScope: kotlinx.coroutines.CoroutineScope? = null,
    ephemeralHeartbeatSleepMs: Long = ephemeralObjectHeartbeatSleep,
): Pair<ModalClient, MockControlPlaneClient> {
    val mock = MockControlPlaneClient()
    val client = ModalClient(
        ModalClientParams(
            controlPlaneClient = mock,
            authTokenProvider = mock,
            tokenId = "test-token-id",
            tokenSecret = "test-token-secret",
            backgroundScope = backgroundScope,
            ephemeralHeartbeatSleepMs = ephemeralHeartbeatSleepMs,
        ),
    )
    return client to mock
}

class MockTaskCommandRouterClient : TaskCommandRouter {
    private val unaryHandlers = mutableMapOf<String, ArrayDeque<suspend (Any) -> Any>>()
    private val streamingHandlers = mutableMapOf<String, ArrayDeque<suspend (Any) -> Flow<Any>>>()

    fun handleUnary(method: String, handler: suspend (Any) -> Any) {
        unaryHandlers.getOrPut(method) { ArrayDeque() }.addLast(handler)
    }

    fun handleStreaming(method: String, handler: suspend (Any) -> Flow<Any>) {
        streamingHandlers.getOrPut(method) { ArrayDeque() }.addLast(handler)
    }

    private suspend fun unary(method: String, request: Any): Any {
        val queue = unaryHandlers[method] ?: error("Unexpected router call: $method with request $request")
        if (queue.isEmpty()) {
            error("Unexpected router call: $method with request $request")
        }
        return queue.removeFirst()(request)
    }

    @Suppress("UNCHECKED_CAST")
    private suspend fun <T> streaming(method: String, request: Any): Flow<T> {
        val queue = streamingHandlers[method] ?: error("Unexpected router call: $method with request $request")
        if (queue.isEmpty()) {
            error("Unexpected router call: $method with request $request")
        }
        return queue.removeFirst()(request) as Flow<T>
    }

    override suspend fun execStart(request: modal.task_command_router.TaskExecStartRequest) {
        unary("/TaskExecStart", request)
    }

    override suspend fun execStdinWrite(request: modal.task_command_router.TaskExecStdinWriteRequest) {
        unary("/TaskExecStdinWrite", request)
    }

    override suspend fun execWait(
        request: modal.task_command_router.TaskExecWaitRequest,
    ): modal.task_command_router.TaskExecWaitResponse {
        return unary("/TaskExecWait", request) as modal.task_command_router.TaskExecWaitResponse
    }

    override suspend fun execStdioRead(
        request: modal.task_command_router.TaskExecStdioReadRequest,
    ): Flow<modal.task_command_router.TaskExecStdioReadResponse> {
        return streaming("/TaskExecStdioRead", request)
    }

    override suspend fun mountDirectory(
        request: modal.task_command_router.TaskMountDirectoryRequest,
    ) {
        unary("/TaskMountDirectory", request)
    }

    override suspend fun snapshotDirectory(
        request: modal.task_command_router.TaskSnapshotDirectoryRequest,
    ): modal.task_command_router.TaskSnapshotDirectoryResponse {
        return unary("/TaskSnapshotDirectory", request) as modal.task_command_router.TaskSnapshotDirectoryResponse
    }

    override fun close() {
    }
}

fun createMockModalClientsWithTaskRouter(
    backgroundScope: kotlinx.coroutines.CoroutineScope? = null,
    ephemeralHeartbeatSleepMs: Long = ephemeralObjectHeartbeatSleep,
): Triple<ModalClient, MockControlPlaneClient, MockTaskCommandRouterClient> {
    val mock = MockControlPlaneClient()
    val router = MockTaskCommandRouterClient()
    val client = ModalClient(
        ModalClientParams(
            controlPlaneClient = mock,
            authTokenProvider = mock,
            tokenId = "test-token-id",
            tokenSecret = "test-token-secret",
            backgroundScope = backgroundScope,
            ephemeralHeartbeatSleepMs = ephemeralHeartbeatSleepMs,
            taskCommandRouterFactory = { _, _ -> router },
        ),
    )
    return Triple(client, mock, router)
}
