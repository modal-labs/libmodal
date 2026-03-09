package com.modal.modalkt

import io.grpc.ManagedChannel
import io.grpc.Metadata
import io.grpc.ClientInterceptor
import io.grpc.ClientInterceptors
import io.grpc.netty.shaded.io.grpc.netty.NettyChannelBuilder
import modal.client.ModalClientGrpcKt
import java.net.URI
import java.util.concurrent.TimeUnit

interface ControlPlaneClient : AuthTokenProvider, TaskRouterAccessProvider {
    suspend fun appGetOrCreate(request: AppGetOrCreateRequest): AppGetOrCreateResponse

    suspend fun secretGetOrCreate(request: SecretGetOrCreateRequest): SecretGetOrCreateResponse

    suspend fun secretDelete(request: SecretDeleteRequest)

    suspend fun volumeGetOrCreate(request: VolumeGetOrCreateRequest): VolumeGetOrCreateResponse

    suspend fun volumeDelete(request: VolumeDeleteRequest)

    suspend fun volumeHeartbeat(request: VolumeHeartbeatRequest)

    suspend fun proxyGet(request: ProxyGetRequest): ProxyGetResponse

    suspend fun imageFromId(request: ImageFromIdRequest): ImageFromIdResponse

    suspend fun imageDelete(request: ImageDeleteRequest)

    suspend fun imageGetOrCreate(request: ImageGetOrCreateRequest): ImageGetOrCreateResponse

    suspend fun imageJoinStreaming(request: ImageJoinStreamingRequest): kotlinx.coroutines.flow.Flow<ImageJoinStreamingResponse>

    suspend fun functionGet(request: FunctionGetRequest): FunctionGetResponse

    suspend fun functionGetCurrentStats(request: FunctionGetCurrentStatsRequest): ProtoFunctionStats

    suspend fun functionUpdateSchedulingParams(request: FunctionUpdateSchedulingParamsRequest)

    suspend fun functionBindParams(request: FunctionBindParamsRequest): FunctionBindParamsResponse

    suspend fun sandboxCreate(request: SandboxCreateRequest): SandboxCreateResponse

    suspend fun sandboxWait(request: SandboxWaitRequest): SandboxWaitResponse

    suspend fun sandboxGetTaskId(request: SandboxGetTaskIdRequest): SandboxGetTaskIdResponse

    suspend fun containerFilesystemExec(
        request: ContainerFilesystemExecRequest,
    ): ContainerFilesystemExecResponse

    suspend fun containerFilesystemExecGetOutput(
        request: ContainerFilesystemExecGetOutputRequest,
    ): kotlinx.coroutines.flow.Flow<FilesystemRuntimeOutputBatch>

    suspend fun sandboxSnapshotFs(request: SandboxSnapshotFsRequest): SandboxSnapshotFsResponse

    suspend fun sandboxGetFromName(request: SandboxGetFromNameRequest): SandboxGetFromNameResponse

    suspend fun sandboxList(request: SandboxListRequest): SandboxListResponse

    suspend fun sandboxTerminate(request: SandboxTerminateRequest): SandboxTerminateResponse

    suspend fun sandboxGetTunnels(request: SandboxGetTunnelsRequest): SandboxGetTunnelsResponse

    suspend fun sandboxCreateConnectToken(
        request: SandboxCreateConnectTokenRequest,
    ): SandboxCreateConnectTokenResponse

    suspend fun sandboxStdinWrite(request: SandboxStdinWriteRequest)

    suspend fun sandboxTagsSet(request: SandboxTagsSetRequest)

    suspend fun sandboxTagsGet(request: SandboxTagsGetRequest): SandboxTagsGetResponse

    suspend fun queueGetOrCreate(request: QueueGetOrCreateRequest): QueueGetOrCreateResponse

    suspend fun queueDelete(request: QueueDeleteRequest)

    suspend fun queueHeartbeat(request: QueueHeartbeatRequest)

    suspend fun queueGet(request: QueueGetRequest): QueueGetResponse

    suspend fun queuePut(request: QueuePutRequest)

    suspend fun queueLen(request: QueueLenRequest): QueueLenResponse

    suspend fun queueNextItems(request: QueueNextItemsRequest): QueueNextItemsResponse

    suspend fun queueClear(request: QueueClearRequest)

    suspend fun functionMap(request: FunctionMapRequest): FunctionMapResponse

    suspend fun functionGetOutputs(request: FunctionGetOutputsRequest): FunctionGetOutputsResponse

    suspend fun functionRetryInputs(request: FunctionRetryInputsRequest): FunctionRetryInputsResponse

    suspend fun functionCallCancel(request: FunctionCallCancelRequest)

    suspend fun sandboxGetLogs(request: SandboxGetLogsRequest): kotlinx.coroutines.flow.Flow<TaskLogsBatch>

    suspend fun attemptStart(request: AttemptStartRequest): AttemptStartResponse

    suspend fun attemptAwait(request: AttemptAwaitRequest): AttemptAwaitResponse

    suspend fun attemptRetry(request: AttemptRetryRequest): AttemptRetryResponse

    suspend fun blobCreate(request: BlobCreateRequest): BlobCreateResponse

    suspend fun blobGet(request: BlobGetRequest): BlobGetResponse

    fun close()
}

class GrpcControlPlaneClient(
    private val profile: Profile,
    private val logger: Logger,
    interceptors: List<ClientInterceptor> = emptyList(),
    private val defaultTimeoutMs: Long? = null,
) : ControlPlaneClient {
    private val channel: ManagedChannel = buildChannel(profile)
    private val baseStub = ModalClientGrpcKt.ModalClientCoroutineStub(
        if (interceptors.isEmpty()) channel else ClientInterceptors.intercept(channel, interceptors),
    )
    private val authTokenManager = AuthTokenManager(this, logger)

    override suspend fun appGetOrCreate(request: AppGetOrCreateRequest): AppGetOrCreateResponse {
        return unaryCall(request) { stub, headers -> stub.appGetOrCreate(request, headers) }
    }

    override suspend fun secretGetOrCreate(request: SecretGetOrCreateRequest): SecretGetOrCreateResponse {
        return unaryCall(request) { stub, headers -> stub.secretGetOrCreate(request, headers) }
    }

    override suspend fun secretDelete(request: SecretDeleteRequest) {
        unaryCall(request) { stub, headers -> stub.secretDelete(request, headers) }
    }

    override suspend fun volumeGetOrCreate(request: VolumeGetOrCreateRequest): VolumeGetOrCreateResponse {
        return unaryCall(request) { stub, headers -> stub.volumeGetOrCreate(request, headers) }
    }

    override suspend fun volumeDelete(request: VolumeDeleteRequest) {
        unaryCall(request) { stub, headers -> stub.volumeDelete(request, headers) }
    }

    override suspend fun volumeHeartbeat(request: VolumeHeartbeatRequest) {
        unaryCall(request) { stub, headers -> stub.volumeHeartbeat(request, headers) }
    }

    override suspend fun proxyGet(request: ProxyGetRequest): ProxyGetResponse {
        return unaryCall(request) { stub, headers -> stub.proxyGet(request, headers) }
    }

    override suspend fun imageFromId(request: ImageFromIdRequest): ImageFromIdResponse {
        return unaryCall(request) { stub, headers -> stub.imageFromId(request, headers) }
    }

    override suspend fun imageDelete(request: ImageDeleteRequest) {
        unaryCall(request) { stub, headers -> stub.imageDelete(request, headers) }
    }

    override suspend fun imageGetOrCreate(request: ImageGetOrCreateRequest): ImageGetOrCreateResponse {
        return unaryCall(request) { stub, headers -> stub.imageGetOrCreate(request, headers) }
    }

    override suspend fun imageJoinStreaming(
        request: ImageJoinStreamingRequest,
    ): kotlinx.coroutines.flow.Flow<ImageJoinStreamingResponse> {
        val headers = authHeaders(includeAuthToken = true)
        return baseStub.imageJoinStreaming(request, headers)
    }

    override suspend fun functionGet(request: FunctionGetRequest): FunctionGetResponse {
        return unaryCall(request) { stub, headers -> stub.functionGet(request, headers) }
    }

    override suspend fun functionGetCurrentStats(request: FunctionGetCurrentStatsRequest): ProtoFunctionStats {
        return unaryCall(request) { stub, headers -> stub.functionGetCurrentStats(request, headers) }
    }

    override suspend fun functionUpdateSchedulingParams(request: FunctionUpdateSchedulingParamsRequest) {
        unaryCall(request) { stub, headers -> stub.functionUpdateSchedulingParams(request, headers) }
    }

    override suspend fun functionBindParams(request: FunctionBindParamsRequest): FunctionBindParamsResponse {
        return unaryCall(request) { stub, headers -> stub.functionBindParams(request, headers) }
    }

    override suspend fun sandboxCreate(request: SandboxCreateRequest): SandboxCreateResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxCreate(request, headers) }
    }

    override suspend fun sandboxWait(request: SandboxWaitRequest): SandboxWaitResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxWait(request, headers) }
    }

    override suspend fun sandboxGetTaskId(request: SandboxGetTaskIdRequest): SandboxGetTaskIdResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxGetTaskId(request, headers) }
    }

    override suspend fun containerFilesystemExec(
        request: ContainerFilesystemExecRequest,
    ): ContainerFilesystemExecResponse {
        return unaryCall(request) { stub, headers -> stub.containerFilesystemExec(request, headers) }
    }

    override suspend fun containerFilesystemExecGetOutput(
        request: ContainerFilesystemExecGetOutputRequest,
    ): kotlinx.coroutines.flow.Flow<FilesystemRuntimeOutputBatch> {
        val headers = authHeaders(includeAuthToken = true)
        return baseStub.containerFilesystemExecGetOutput(request, headers)
    }

    override suspend fun sandboxSnapshotFs(request: SandboxSnapshotFsRequest): SandboxSnapshotFsResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxSnapshotFs(request, headers) }
    }

    override suspend fun sandboxGetFromName(request: SandboxGetFromNameRequest): SandboxGetFromNameResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxGetFromName(request, headers) }
    }

    override suspend fun sandboxList(request: SandboxListRequest): SandboxListResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxList(request, headers) }
    }

    override suspend fun sandboxTerminate(request: SandboxTerminateRequest): SandboxTerminateResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxTerminate(request, headers) }
    }

    override suspend fun sandboxGetTunnels(request: SandboxGetTunnelsRequest): SandboxGetTunnelsResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxGetTunnels(request, headers) }
    }

    override suspend fun sandboxCreateConnectToken(
        request: SandboxCreateConnectTokenRequest,
    ): SandboxCreateConnectTokenResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxCreateConnectToken(request, headers) }
    }

    override suspend fun sandboxStdinWrite(request: SandboxStdinWriteRequest) {
        unaryCall(request) { stub, headers -> stub.sandboxStdinWrite(request, headers) }
    }

    override suspend fun sandboxTagsSet(request: SandboxTagsSetRequest) {
        unaryCall(request) { stub, headers -> stub.sandboxTagsSet(request, headers) }
    }

    override suspend fun sandboxTagsGet(request: SandboxTagsGetRequest): SandboxTagsGetResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxTagsGet(request, headers) }
    }

    override suspend fun queueGetOrCreate(request: QueueGetOrCreateRequest): QueueGetOrCreateResponse {
        return unaryCall(request) { stub, headers -> stub.queueGetOrCreate(request, headers) }
    }

    override suspend fun queueDelete(request: QueueDeleteRequest) {
        unaryCall(request) { stub, headers -> stub.queueDelete(request, headers) }
    }

    override suspend fun queueHeartbeat(request: QueueHeartbeatRequest) {
        unaryCall(request) { stub, headers -> stub.queueHeartbeat(request, headers) }
    }

    override suspend fun queueGet(request: QueueGetRequest): QueueGetResponse {
        return unaryCall(request) { stub, headers -> stub.queueGet(request, headers) }
    }

    override suspend fun queuePut(request: QueuePutRequest) {
        unaryCall(request) { stub, headers -> stub.queuePut(request, headers) }
    }

    override suspend fun queueLen(request: QueueLenRequest): QueueLenResponse {
        return unaryCall(request) { stub, headers -> stub.queueLen(request, headers) }
    }

    override suspend fun queueNextItems(request: QueueNextItemsRequest): QueueNextItemsResponse {
        return unaryCall(request) { stub, headers -> stub.queueNextItems(request, headers) }
    }

    override suspend fun queueClear(request: QueueClearRequest) {
        unaryCall(request) { stub, headers -> stub.queueClear(request, headers) }
    }

    override suspend fun functionMap(request: FunctionMapRequest): FunctionMapResponse {
        return unaryCall(request) { stub, headers -> stub.functionMap(request, headers) }
    }

    override suspend fun functionGetOutputs(request: FunctionGetOutputsRequest): FunctionGetOutputsResponse {
        return unaryCall(request) { stub, headers -> stub.functionGetOutputs(request, headers) }
    }

    override suspend fun functionRetryInputs(request: FunctionRetryInputsRequest): FunctionRetryInputsResponse {
        return unaryCall(request) { stub, headers -> stub.functionRetryInputs(request, headers) }
    }

    override suspend fun functionCallCancel(request: FunctionCallCancelRequest) {
        unaryCall(request) { stub, headers -> stub.functionCallCancel(request, headers) }
    }

    override suspend fun sandboxGetLogs(request: SandboxGetLogsRequest): kotlinx.coroutines.flow.Flow<TaskLogsBatch> {
        val headers = authHeaders(includeAuthToken = true)
        return baseStub.sandboxGetLogs(request, headers)
    }

    override suspend fun attemptStart(request: AttemptStartRequest): AttemptStartResponse {
        return unaryCall(request) { stub, headers -> stub.attemptStart(request, headers) }
    }

    override suspend fun attemptAwait(request: AttemptAwaitRequest): AttemptAwaitResponse {
        return unaryCall(request) { stub, headers -> stub.attemptAwait(request, headers) }
    }

    override suspend fun attemptRetry(request: AttemptRetryRequest): AttemptRetryResponse {
        return unaryCall(request) { stub, headers -> stub.attemptRetry(request, headers) }
    }

    override suspend fun blobCreate(request: BlobCreateRequest): BlobCreateResponse {
        return unaryCall(request) { stub, headers -> stub.blobCreate(request, headers) }
    }

    override suspend fun blobGet(request: BlobGetRequest): BlobGetResponse {
        return unaryCall(request) { stub, headers -> stub.blobGet(request, headers) }
    }

    override suspend fun authTokenGet(request: AuthTokenGetRequest): AuthTokenGetResponse {
        return unaryCall(request, includeAuthToken = false) { stub, headers ->
            stub.authTokenGet(request, headers)
        }
    }

    override suspend fun taskGetCommandRouterAccess(
        request: TaskGetCommandRouterAccessRequest,
    ): TaskGetCommandRouterAccessResponse {
        return unaryCall(request) { stub, headers ->
            stub.taskGetCommandRouterAccess(request, headers)
        }
    }

    override fun close() {
        channel.shutdownNow()
    }

    private suspend fun <RequestT, ResponseT> unaryCall(
        request: RequestT,
        includeAuthToken: Boolean = true,
        call: suspend (
            stub: ModalClientGrpcKt.ModalClientCoroutineStub,
            headers: Metadata,
        ) -> ResponseT,
    ): ResponseT {
        return callWithRetriesOnTransientErrors(
            func = {
                val headers = authHeaders(includeAuthToken)
                val stub = if (defaultTimeoutMs != null) {
                    baseStub.withDeadlineAfter(defaultTimeoutMs, TimeUnit.MILLISECONDS)
                } else {
                    baseStub
                }
                call(stub, headers)
            },
        )
    }

    private suspend fun authHeaders(includeAuthToken: Boolean): Metadata {
        val headers = Metadata()
        headers.put(key("x-modal-client-type"), ClientType.CLIENT_TYPE_LIBMODAL_JS_VALUE.toString())
        headers.put(key("x-modal-client-version"), "1.0.0")
        headers.put(key("x-modal-libmodal-version"), "modal-kt/${getSdkVersion()}")

        val tokenId = profile.tokenId
        val tokenSecret = profile.tokenSecret
        if (!tokenId.isNullOrEmpty()) {
            headers.put(key("x-modal-token-id"), tokenId)
        }
        if (!tokenSecret.isNullOrEmpty()) {
            headers.put(key("x-modal-token-secret"), tokenSecret)
        }
        if (includeAuthToken && !tokenId.isNullOrEmpty() && !tokenSecret.isNullOrEmpty()) {
            val authToken = authTokenManager.getToken()
            headers.put(key("x-modal-auth-token"), authToken)
        }
        return headers
    }

    private fun buildChannel(profile: Profile): ManagedChannel {
        val uri = URI(profile.serverUrl)
        val host = uri.host ?: throw InvalidError("Invalid Modal server URL: ${profile.serverUrl}")
        val port = if (uri.port == -1) 443 else uri.port
        val builder = NettyChannelBuilder.forAddress(host, port)
            .keepAliveTime(30, TimeUnit.SECONDS)
            .keepAliveTimeout(10, TimeUnit.SECONDS)
            .keepAliveWithoutCalls(true)
            .maxInboundMessageSize(100 * 1024 * 1024)

        if (isLocalhost(profile)) {
            builder.usePlaintext()
        } else {
            builder.useTransportSecurity()
        }

        return builder.build()
    }

    private fun key(name: String): Metadata.Key<String> {
        return Metadata.Key.of(name, Metadata.ASCII_STRING_MARSHALLER)
    }
}
