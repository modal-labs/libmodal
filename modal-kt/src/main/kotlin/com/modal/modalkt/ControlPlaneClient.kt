package com.modal.modalkt

import io.grpc.ManagedChannel
import io.grpc.Metadata
import io.grpc.netty.shaded.io.grpc.netty.NettyChannelBuilder
import modal.client.Api
import modal.client.ModalClientGrpcKt
import java.net.URI
import java.util.concurrent.TimeUnit

interface ControlPlaneClient : AuthTokenProvider, TaskRouterAccessProvider {
    suspend fun appGetOrCreate(request: Api.AppGetOrCreateRequest): Api.AppGetOrCreateResponse

    suspend fun secretGetOrCreate(request: Api.SecretGetOrCreateRequest): Api.SecretGetOrCreateResponse

    suspend fun secretDelete(request: Api.SecretDeleteRequest)

    suspend fun volumeGetOrCreate(request: Api.VolumeGetOrCreateRequest): Api.VolumeGetOrCreateResponse

    suspend fun volumeDelete(request: Api.VolumeDeleteRequest)

    suspend fun volumeHeartbeat(request: Api.VolumeHeartbeatRequest)

    suspend fun proxyGet(request: Api.ProxyGetRequest): Api.ProxyGetResponse

    suspend fun imageFromId(request: Api.ImageFromIdRequest): Api.ImageFromIdResponse

    suspend fun imageDelete(request: Api.ImageDeleteRequest)

    suspend fun imageGetOrCreate(request: Api.ImageGetOrCreateRequest): Api.ImageGetOrCreateResponse

    suspend fun imageJoinStreaming(request: Api.ImageJoinStreamingRequest): kotlinx.coroutines.flow.Flow<Api.ImageJoinStreamingResponse>

    suspend fun functionGet(request: Api.FunctionGetRequest): Api.FunctionGetResponse

    suspend fun functionGetCurrentStats(request: Api.FunctionGetCurrentStatsRequest): Api.FunctionStats

    suspend fun functionUpdateSchedulingParams(request: Api.FunctionUpdateSchedulingParamsRequest)

    suspend fun functionBindParams(request: Api.FunctionBindParamsRequest): Api.FunctionBindParamsResponse

    suspend fun sandboxCreate(request: Api.SandboxCreateRequest): Api.SandboxCreateResponse

    suspend fun sandboxWait(request: Api.SandboxWaitRequest): Api.SandboxWaitResponse

    suspend fun sandboxGetFromName(request: Api.SandboxGetFromNameRequest): Api.SandboxGetFromNameResponse

    suspend fun sandboxList(request: Api.SandboxListRequest): Api.SandboxListResponse

    suspend fun sandboxTerminate(request: Api.SandboxTerminateRequest): Api.SandboxTerminateResponse

    suspend fun sandboxGetTunnels(request: Api.SandboxGetTunnelsRequest): Api.SandboxGetTunnelsResponse

    suspend fun sandboxCreateConnectToken(
        request: Api.SandboxCreateConnectTokenRequest,
    ): Api.SandboxCreateConnectTokenResponse

    suspend fun sandboxTagsSet(request: Api.SandboxTagsSetRequest)

    suspend fun sandboxTagsGet(request: Api.SandboxTagsGetRequest): Api.SandboxTagsGetResponse

    suspend fun queueGetOrCreate(request: Api.QueueGetOrCreateRequest): Api.QueueGetOrCreateResponse

    suspend fun queueDelete(request: Api.QueueDeleteRequest)

    suspend fun queueHeartbeat(request: Api.QueueHeartbeatRequest)

    suspend fun queueGet(request: Api.QueueGetRequest): Api.QueueGetResponse

    suspend fun queuePut(request: Api.QueuePutRequest)

    suspend fun queueLen(request: Api.QueueLenRequest): Api.QueueLenResponse

    suspend fun queueNextItems(request: Api.QueueNextItemsRequest): Api.QueueNextItemsResponse

    suspend fun queueClear(request: Api.QueueClearRequest)

    suspend fun functionMap(request: Api.FunctionMapRequest): Api.FunctionMapResponse

    suspend fun functionGetOutputs(request: Api.FunctionGetOutputsRequest): Api.FunctionGetOutputsResponse

    suspend fun functionRetryInputs(request: Api.FunctionRetryInputsRequest): Api.FunctionRetryInputsResponse

    suspend fun functionCallCancel(request: Api.FunctionCallCancelRequest)

    fun close()
}

class GrpcControlPlaneClient(
    private val profile: Profile,
    private val logger: Logger,
) : ControlPlaneClient {
    private val channel: ManagedChannel = buildChannel(profile)
    private val baseStub = ModalClientGrpcKt.ModalClientCoroutineStub(channel)
    private val authTokenManager = AuthTokenManager(this, logger)

    override suspend fun appGetOrCreate(request: Api.AppGetOrCreateRequest): Api.AppGetOrCreateResponse {
        return unaryCall(request) { stub, headers -> stub.appGetOrCreate(request, headers) }
    }

    override suspend fun secretGetOrCreate(request: Api.SecretGetOrCreateRequest): Api.SecretGetOrCreateResponse {
        return unaryCall(request) { stub, headers -> stub.secretGetOrCreate(request, headers) }
    }

    override suspend fun secretDelete(request: Api.SecretDeleteRequest) {
        unaryCall(request) { stub, headers -> stub.secretDelete(request, headers) }
    }

    override suspend fun volumeGetOrCreate(request: Api.VolumeGetOrCreateRequest): Api.VolumeGetOrCreateResponse {
        return unaryCall(request) { stub, headers -> stub.volumeGetOrCreate(request, headers) }
    }

    override suspend fun volumeDelete(request: Api.VolumeDeleteRequest) {
        unaryCall(request) { stub, headers -> stub.volumeDelete(request, headers) }
    }

    override suspend fun volumeHeartbeat(request: Api.VolumeHeartbeatRequest) {
        unaryCall(request) { stub, headers -> stub.volumeHeartbeat(request, headers) }
    }

    override suspend fun proxyGet(request: Api.ProxyGetRequest): Api.ProxyGetResponse {
        return unaryCall(request) { stub, headers -> stub.proxyGet(request, headers) }
    }

    override suspend fun imageFromId(request: Api.ImageFromIdRequest): Api.ImageFromIdResponse {
        return unaryCall(request) { stub, headers -> stub.imageFromId(request, headers) }
    }

    override suspend fun imageDelete(request: Api.ImageDeleteRequest) {
        unaryCall(request) { stub, headers -> stub.imageDelete(request, headers) }
    }

    override suspend fun imageGetOrCreate(request: Api.ImageGetOrCreateRequest): Api.ImageGetOrCreateResponse {
        return unaryCall(request) { stub, headers -> stub.imageGetOrCreate(request, headers) }
    }

    override suspend fun imageJoinStreaming(
        request: Api.ImageJoinStreamingRequest,
    ): kotlinx.coroutines.flow.Flow<Api.ImageJoinStreamingResponse> {
        val headers = authHeaders(includeAuthToken = true)
        return baseStub.imageJoinStreaming(request, headers)
    }

    override suspend fun functionGet(request: Api.FunctionGetRequest): Api.FunctionGetResponse {
        return unaryCall(request) { stub, headers -> stub.functionGet(request, headers) }
    }

    override suspend fun functionGetCurrentStats(request: Api.FunctionGetCurrentStatsRequest): Api.FunctionStats {
        return unaryCall(request) { stub, headers -> stub.functionGetCurrentStats(request, headers) }
    }

    override suspend fun functionUpdateSchedulingParams(request: Api.FunctionUpdateSchedulingParamsRequest) {
        unaryCall(request) { stub, headers -> stub.functionUpdateSchedulingParams(request, headers) }
    }

    override suspend fun functionBindParams(request: Api.FunctionBindParamsRequest): Api.FunctionBindParamsResponse {
        return unaryCall(request) { stub, headers -> stub.functionBindParams(request, headers) }
    }

    override suspend fun sandboxCreate(request: Api.SandboxCreateRequest): Api.SandboxCreateResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxCreate(request, headers) }
    }

    override suspend fun sandboxWait(request: Api.SandboxWaitRequest): Api.SandboxWaitResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxWait(request, headers) }
    }

    override suspend fun sandboxGetFromName(request: Api.SandboxGetFromNameRequest): Api.SandboxGetFromNameResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxGetFromName(request, headers) }
    }

    override suspend fun sandboxList(request: Api.SandboxListRequest): Api.SandboxListResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxList(request, headers) }
    }

    override suspend fun sandboxTerminate(request: Api.SandboxTerminateRequest): Api.SandboxTerminateResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxTerminate(request, headers) }
    }

    override suspend fun sandboxGetTunnels(request: Api.SandboxGetTunnelsRequest): Api.SandboxGetTunnelsResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxGetTunnels(request, headers) }
    }

    override suspend fun sandboxCreateConnectToken(
        request: Api.SandboxCreateConnectTokenRequest,
    ): Api.SandboxCreateConnectTokenResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxCreateConnectToken(request, headers) }
    }

    override suspend fun sandboxTagsSet(request: Api.SandboxTagsSetRequest) {
        unaryCall(request) { stub, headers -> stub.sandboxTagsSet(request, headers) }
    }

    override suspend fun sandboxTagsGet(request: Api.SandboxTagsGetRequest): Api.SandboxTagsGetResponse {
        return unaryCall(request) { stub, headers -> stub.sandboxTagsGet(request, headers) }
    }

    override suspend fun queueGetOrCreate(request: Api.QueueGetOrCreateRequest): Api.QueueGetOrCreateResponse {
        return unaryCall(request) { stub, headers -> stub.queueGetOrCreate(request, headers) }
    }

    override suspend fun queueDelete(request: Api.QueueDeleteRequest) {
        unaryCall(request) { stub, headers -> stub.queueDelete(request, headers) }
    }

    override suspend fun queueHeartbeat(request: Api.QueueHeartbeatRequest) {
        unaryCall(request) { stub, headers -> stub.queueHeartbeat(request, headers) }
    }

    override suspend fun queueGet(request: Api.QueueGetRequest): Api.QueueGetResponse {
        return unaryCall(request) { stub, headers -> stub.queueGet(request, headers) }
    }

    override suspend fun queuePut(request: Api.QueuePutRequest) {
        unaryCall(request) { stub, headers -> stub.queuePut(request, headers) }
    }

    override suspend fun queueLen(request: Api.QueueLenRequest): Api.QueueLenResponse {
        return unaryCall(request) { stub, headers -> stub.queueLen(request, headers) }
    }

    override suspend fun queueNextItems(request: Api.QueueNextItemsRequest): Api.QueueNextItemsResponse {
        return unaryCall(request) { stub, headers -> stub.queueNextItems(request, headers) }
    }

    override suspend fun queueClear(request: Api.QueueClearRequest) {
        unaryCall(request) { stub, headers -> stub.queueClear(request, headers) }
    }

    override suspend fun functionMap(request: Api.FunctionMapRequest): Api.FunctionMapResponse {
        return unaryCall(request) { stub, headers -> stub.functionMap(request, headers) }
    }

    override suspend fun functionGetOutputs(request: Api.FunctionGetOutputsRequest): Api.FunctionGetOutputsResponse {
        return unaryCall(request) { stub, headers -> stub.functionGetOutputs(request, headers) }
    }

    override suspend fun functionRetryInputs(request: Api.FunctionRetryInputsRequest): Api.FunctionRetryInputsResponse {
        return unaryCall(request) { stub, headers -> stub.functionRetryInputs(request, headers) }
    }

    override suspend fun functionCallCancel(request: Api.FunctionCallCancelRequest) {
        unaryCall(request) { stub, headers -> stub.functionCallCancel(request, headers) }
    }

    override suspend fun authTokenGet(request: Api.AuthTokenGetRequest): Api.AuthTokenGetResponse {
        return unaryCall(request, includeAuthToken = false) { stub, headers ->
            stub.authTokenGet(request, headers)
        }
    }

    override suspend fun taskGetCommandRouterAccess(
        request: Api.TaskGetCommandRouterAccessRequest,
    ): Api.TaskGetCommandRouterAccessResponse {
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
                call(baseStub, headers)
            },
        )
    }

    private suspend fun authHeaders(includeAuthToken: Boolean): Metadata {
        val headers = Metadata()
        headers.put(key("x-modal-client-type"), Api.ClientType.CLIENT_TYPE_LIBMODAL_JS_VALUE.toString())
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
