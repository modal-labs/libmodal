package com.modal.modalkt

import com.squareup.wire.ProtoAdapter
import io.grpc.ClientInterceptor
import io.grpc.ClientInterceptors
import io.grpc.ManagedChannel
import io.grpc.Metadata
import io.grpc.netty.shaded.io.grpc.netty.NettyChannelBuilder
import java.net.URI
import java.util.concurrent.TimeUnit
import modal.client.*

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
    private val serviceName = "modal.client.ModalClient"
    private val channel: ManagedChannel = buildChannel(profile)
    private val rpcChannel = if (interceptors.isEmpty()) {
        channel
    } else {
        ClientInterceptors.intercept(channel, interceptors)
    }
    private val authTokenManager = AuthTokenManager(this, logger)

    override suspend fun appGetOrCreate(request: AppGetOrCreateRequest): AppGetOrCreateResponse {
        return unaryCall("AppGetOrCreate", request, AppGetOrCreateRequest.ADAPTER, AppGetOrCreateResponse.ADAPTER)
    }

    override suspend fun secretGetOrCreate(request: SecretGetOrCreateRequest): SecretGetOrCreateResponse {
        return unaryCall("SecretGetOrCreate", request, SecretGetOrCreateRequest.ADAPTER, SecretGetOrCreateResponse.ADAPTER)
    }

    override suspend fun secretDelete(request: SecretDeleteRequest) {
        unaryCall("SecretDelete", request, SecretDeleteRequest.ADAPTER, ProtoAdapter.EMPTY)
    }

    override suspend fun volumeGetOrCreate(request: VolumeGetOrCreateRequest): VolumeGetOrCreateResponse {
        return unaryCall("VolumeGetOrCreate", request, VolumeGetOrCreateRequest.ADAPTER, VolumeGetOrCreateResponse.ADAPTER)
    }

    override suspend fun volumeDelete(request: VolumeDeleteRequest) {
        unaryCall("VolumeDelete", request, VolumeDeleteRequest.ADAPTER, ProtoAdapter.EMPTY)
    }

    override suspend fun volumeHeartbeat(request: VolumeHeartbeatRequest) {
        unaryCall("VolumeHeartbeat", request, VolumeHeartbeatRequest.ADAPTER, ProtoAdapter.EMPTY)
    }

    override suspend fun proxyGet(request: ProxyGetRequest): ProxyGetResponse {
        return unaryCall("ProxyGet", request, ProxyGetRequest.ADAPTER, ProxyGetResponse.ADAPTER)
    }

    override suspend fun imageFromId(request: ImageFromIdRequest): ImageFromIdResponse {
        return unaryCall("ImageFromId", request, ImageFromIdRequest.ADAPTER, ImageFromIdResponse.ADAPTER)
    }

    override suspend fun imageDelete(request: ImageDeleteRequest) {
        unaryCall("ImageDelete", request, ImageDeleteRequest.ADAPTER, ProtoAdapter.EMPTY)
    }

    override suspend fun imageGetOrCreate(request: ImageGetOrCreateRequest): ImageGetOrCreateResponse {
        return unaryCall("ImageGetOrCreate", request, ImageGetOrCreateRequest.ADAPTER, ImageGetOrCreateResponse.ADAPTER)
    }

    override suspend fun imageJoinStreaming(
        request: ImageJoinStreamingRequest,
    ): kotlinx.coroutines.flow.Flow<ImageJoinStreamingResponse> {
        return streamingCall(
            methodName = "ImageJoinStreaming",
            request = request,
            requestAdapter = ImageJoinStreamingRequest.ADAPTER,
            responseAdapter = ImageJoinStreamingResponse.ADAPTER,
        )
    }

    override suspend fun functionGet(request: FunctionGetRequest): FunctionGetResponse {
        return unaryCall("FunctionGet", request, FunctionGetRequest.ADAPTER, FunctionGetResponse.ADAPTER)
    }

    override suspend fun functionGetCurrentStats(request: FunctionGetCurrentStatsRequest): ProtoFunctionStats {
        return unaryCall("FunctionGetCurrentStats", request, FunctionGetCurrentStatsRequest.ADAPTER, modal.client.FunctionStats.ADAPTER)
    }

    override suspend fun functionUpdateSchedulingParams(request: FunctionUpdateSchedulingParamsRequest) {
        unaryCall("FunctionUpdateSchedulingParams", request, FunctionUpdateSchedulingParamsRequest.ADAPTER, ProtoAdapter.EMPTY)
    }

    override suspend fun functionBindParams(request: FunctionBindParamsRequest): FunctionBindParamsResponse {
        return unaryCall("FunctionBindParams", request, FunctionBindParamsRequest.ADAPTER, FunctionBindParamsResponse.ADAPTER)
    }

    override suspend fun sandboxCreate(request: SandboxCreateRequest): SandboxCreateResponse {
        return unaryCall("SandboxCreate", request, SandboxCreateRequest.ADAPTER, SandboxCreateResponse.ADAPTER)
    }

    override suspend fun sandboxWait(request: SandboxWaitRequest): SandboxWaitResponse {
        return unaryCall("SandboxWait", request, SandboxWaitRequest.ADAPTER, SandboxWaitResponse.ADAPTER)
    }

    override suspend fun sandboxGetTaskId(request: SandboxGetTaskIdRequest): SandboxGetTaskIdResponse {
        return unaryCall("SandboxGetTaskId", request, SandboxGetTaskIdRequest.ADAPTER, SandboxGetTaskIdResponse.ADAPTER)
    }

    override suspend fun containerFilesystemExec(
        request: ContainerFilesystemExecRequest,
    ): ContainerFilesystemExecResponse {
        return unaryCall(
            "ContainerFilesystemExec",
            request,
            ContainerFilesystemExecRequest.ADAPTER,
            ContainerFilesystemExecResponse.ADAPTER,
        )
    }

    override suspend fun containerFilesystemExecGetOutput(
        request: ContainerFilesystemExecGetOutputRequest,
    ): kotlinx.coroutines.flow.Flow<FilesystemRuntimeOutputBatch> {
        return streamingCall(
            methodName = "ContainerFilesystemExecGetOutput",
            request = request,
            requestAdapter = ContainerFilesystemExecGetOutputRequest.ADAPTER,
            responseAdapter = FilesystemRuntimeOutputBatch.ADAPTER,
        )
    }

    override suspend fun sandboxSnapshotFs(request: SandboxSnapshotFsRequest): SandboxSnapshotFsResponse {
        return unaryCall("SandboxSnapshotFs", request, SandboxSnapshotFsRequest.ADAPTER, SandboxSnapshotFsResponse.ADAPTER)
    }

    override suspend fun sandboxGetFromName(request: SandboxGetFromNameRequest): SandboxGetFromNameResponse {
        return unaryCall("SandboxGetFromName", request, SandboxGetFromNameRequest.ADAPTER, SandboxGetFromNameResponse.ADAPTER)
    }

    override suspend fun sandboxList(request: SandboxListRequest): SandboxListResponse {
        return unaryCall("SandboxList", request, SandboxListRequest.ADAPTER, SandboxListResponse.ADAPTER)
    }

    override suspend fun sandboxTerminate(request: SandboxTerminateRequest): SandboxTerminateResponse {
        return unaryCall("SandboxTerminate", request, SandboxTerminateRequest.ADAPTER, SandboxTerminateResponse.ADAPTER)
    }

    override suspend fun sandboxGetTunnels(request: SandboxGetTunnelsRequest): SandboxGetTunnelsResponse {
        return unaryCall("SandboxGetTunnels", request, SandboxGetTunnelsRequest.ADAPTER, SandboxGetTunnelsResponse.ADAPTER)
    }

    override suspend fun sandboxCreateConnectToken(
        request: SandboxCreateConnectTokenRequest,
    ): SandboxCreateConnectTokenResponse {
        return unaryCall(
            "SandboxCreateConnectToken",
            request,
            SandboxCreateConnectTokenRequest.ADAPTER,
            SandboxCreateConnectTokenResponse.ADAPTER,
        )
    }

    override suspend fun sandboxStdinWrite(request: SandboxStdinWriteRequest) {
        unaryCall("SandboxStdinWrite", request, SandboxStdinWriteRequest.ADAPTER, ProtoAdapter.EMPTY)
    }

    override suspend fun sandboxTagsSet(request: SandboxTagsSetRequest) {
        unaryCall("SandboxTagsSet", request, SandboxTagsSetRequest.ADAPTER, ProtoAdapter.EMPTY)
    }

    override suspend fun sandboxTagsGet(request: SandboxTagsGetRequest): SandboxTagsGetResponse {
        return unaryCall("SandboxTagsGet", request, SandboxTagsGetRequest.ADAPTER, SandboxTagsGetResponse.ADAPTER)
    }

    override suspend fun queueGetOrCreate(request: QueueGetOrCreateRequest): QueueGetOrCreateResponse {
        return unaryCall("QueueGetOrCreate", request, QueueGetOrCreateRequest.ADAPTER, QueueGetOrCreateResponse.ADAPTER)
    }

    override suspend fun queueDelete(request: QueueDeleteRequest) {
        unaryCall("QueueDelete", request, QueueDeleteRequest.ADAPTER, ProtoAdapter.EMPTY)
    }

    override suspend fun queueHeartbeat(request: QueueHeartbeatRequest) {
        unaryCall("QueueHeartbeat", request, QueueHeartbeatRequest.ADAPTER, ProtoAdapter.EMPTY)
    }

    override suspend fun queueGet(request: QueueGetRequest): QueueGetResponse {
        return unaryCall("QueueGet", request, QueueGetRequest.ADAPTER, QueueGetResponse.ADAPTER)
    }

    override suspend fun queuePut(request: QueuePutRequest) {
        unaryCall("QueuePut", request, QueuePutRequest.ADAPTER, ProtoAdapter.EMPTY)
    }

    override suspend fun queueLen(request: QueueLenRequest): QueueLenResponse {
        return unaryCall("QueueLen", request, QueueLenRequest.ADAPTER, QueueLenResponse.ADAPTER)
    }

    override suspend fun queueNextItems(request: QueueNextItemsRequest): QueueNextItemsResponse {
        return unaryCall("QueueNextItems", request, QueueNextItemsRequest.ADAPTER, QueueNextItemsResponse.ADAPTER)
    }

    override suspend fun queueClear(request: QueueClearRequest) {
        unaryCall("QueueClear", request, QueueClearRequest.ADAPTER, ProtoAdapter.EMPTY)
    }

    override suspend fun functionMap(request: FunctionMapRequest): FunctionMapResponse {
        return unaryCall("FunctionMap", request, FunctionMapRequest.ADAPTER, FunctionMapResponse.ADAPTER)
    }

    override suspend fun functionGetOutputs(request: FunctionGetOutputsRequest): FunctionGetOutputsResponse {
        return unaryCall("FunctionGetOutputs", request, FunctionGetOutputsRequest.ADAPTER, FunctionGetOutputsResponse.ADAPTER)
    }

    override suspend fun functionRetryInputs(request: FunctionRetryInputsRequest): FunctionRetryInputsResponse {
        return unaryCall("FunctionRetryInputs", request, FunctionRetryInputsRequest.ADAPTER, FunctionRetryInputsResponse.ADAPTER)
    }

    override suspend fun functionCallCancel(request: FunctionCallCancelRequest) {
        unaryCall("FunctionCallCancel", request, FunctionCallCancelRequest.ADAPTER, ProtoAdapter.EMPTY)
    }

    override suspend fun sandboxGetLogs(request: SandboxGetLogsRequest): kotlinx.coroutines.flow.Flow<TaskLogsBatch> {
        return streamingCall(
            methodName = "SandboxGetLogs",
            request = request,
            requestAdapter = SandboxGetLogsRequest.ADAPTER,
            responseAdapter = TaskLogsBatch.ADAPTER,
        )
    }

    override suspend fun attemptStart(request: AttemptStartRequest): AttemptStartResponse {
        return unaryCall("AttemptStart", request, AttemptStartRequest.ADAPTER, AttemptStartResponse.ADAPTER)
    }

    override suspend fun attemptAwait(request: AttemptAwaitRequest): AttemptAwaitResponse {
        return unaryCall("AttemptAwait", request, AttemptAwaitRequest.ADAPTER, AttemptAwaitResponse.ADAPTER)
    }

    override suspend fun attemptRetry(request: AttemptRetryRequest): AttemptRetryResponse {
        return unaryCall("AttemptRetry", request, AttemptRetryRequest.ADAPTER, AttemptRetryResponse.ADAPTER)
    }

    override suspend fun blobCreate(request: BlobCreateRequest): BlobCreateResponse {
        return unaryCall("BlobCreate", request, BlobCreateRequest.ADAPTER, BlobCreateResponse.ADAPTER)
    }

    override suspend fun blobGet(request: BlobGetRequest): BlobGetResponse {
        return unaryCall("BlobGet", request, BlobGetRequest.ADAPTER, BlobGetResponse.ADAPTER)
    }

    override suspend fun authTokenGet(request: AuthTokenGetRequest): AuthTokenGetResponse {
        return unaryCall("AuthTokenGet", request, AuthTokenGetRequest.ADAPTER, AuthTokenGetResponse.ADAPTER, includeAuthToken = false)
    }

    override suspend fun taskGetCommandRouterAccess(
        request: TaskGetCommandRouterAccessRequest,
    ): TaskGetCommandRouterAccessResponse {
        return unaryCall(
            "TaskGetCommandRouterAccess",
            request,
            TaskGetCommandRouterAccessRequest.ADAPTER,
            TaskGetCommandRouterAccessResponse.ADAPTER,
        )
    }

    override fun close() {
        channel.shutdownNow()
    }

    private suspend fun <RequestT : Any, ResponseT : Any> unaryCall(
        methodName: String,
        request: RequestT,
        requestAdapter: ProtoAdapter<RequestT>,
        responseAdapter: ProtoAdapter<ResponseT>,
        includeAuthToken: Boolean = true,
    ): ResponseT {
        return callWithRetriesOnTransientErrors(
            func = {
                unaryRpc(
                    channel = rpcChannel,
                    serviceName = serviceName,
                    methodName = methodName,
                    request = request,
                    requestAdapter = requestAdapter,
                    responseAdapter = responseAdapter,
                    options = RpcOptions(
                        timeoutMs = defaultTimeoutMs,
                        headers = authHeaders(includeAuthToken),
                    ),
                )
            },
        )
    }

    private suspend fun <RequestT : Any, ResponseT : Any> streamingCall(
        methodName: String,
        request: RequestT,
        requestAdapter: ProtoAdapter<RequestT>,
        responseAdapter: ProtoAdapter<ResponseT>,
        includeAuthToken: Boolean = true,
    ): kotlinx.coroutines.flow.Flow<ResponseT> {
        return serverStreamingRpc(
            channel = rpcChannel,
            serviceName = serviceName,
            methodName = methodName,
            request = request,
            requestAdapter = requestAdapter,
            responseAdapter = responseAdapter,
            options = RpcOptions(
                timeoutMs = defaultTimeoutMs,
                headers = authHeaders(includeAuthToken),
            ),
        )
    }

    private suspend fun authHeaders(includeAuthToken: Boolean): Metadata {
        val headers = Metadata()
        headers.put(key("x-modal-client-type"), ClientType.CLIENT_TYPE_LIBMODAL_JS.value.toString())
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
