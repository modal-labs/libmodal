package com.modal.modalkt

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.emptyFlow
import modal.client.Api
import java.util.ArrayDeque

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

    override suspend fun appGetOrCreate(request: Api.AppGetOrCreateRequest): Api.AppGetOrCreateResponse {
        return unary("/AppGetOrCreate", request) as Api.AppGetOrCreateResponse
    }

    override suspend fun secretGetOrCreate(request: Api.SecretGetOrCreateRequest): Api.SecretGetOrCreateResponse {
        return unary("/SecretGetOrCreate", request) as Api.SecretGetOrCreateResponse
    }

    override suspend fun secretDelete(request: Api.SecretDeleteRequest) {
        unary("/SecretDelete", request)
    }

    override suspend fun volumeGetOrCreate(request: Api.VolumeGetOrCreateRequest): Api.VolumeGetOrCreateResponse {
        return unary("/VolumeGetOrCreate", request) as Api.VolumeGetOrCreateResponse
    }

    override suspend fun volumeDelete(request: Api.VolumeDeleteRequest) {
        unary("/VolumeDelete", request)
    }

    override suspend fun volumeHeartbeat(request: Api.VolumeHeartbeatRequest) {
        unary("/VolumeHeartbeat", request)
    }

    override suspend fun proxyGet(request: Api.ProxyGetRequest): Api.ProxyGetResponse {
        return unary("/ProxyGet", request) as Api.ProxyGetResponse
    }

    override suspend fun imageFromId(request: Api.ImageFromIdRequest): Api.ImageFromIdResponse {
        return unary("/ImageFromId", request) as Api.ImageFromIdResponse
    }

    override suspend fun imageDelete(request: Api.ImageDeleteRequest) {
        unary("/ImageDelete", request)
    }

    override suspend fun imageGetOrCreate(request: Api.ImageGetOrCreateRequest): Api.ImageGetOrCreateResponse {
        return unary("/ImageGetOrCreate", request) as Api.ImageGetOrCreateResponse
    }

    override suspend fun imageJoinStreaming(request: Api.ImageJoinStreamingRequest): Flow<Api.ImageJoinStreamingResponse> {
        return streaming("/ImageJoinStreaming", request)
    }

    override suspend fun functionGet(request: Api.FunctionGetRequest): Api.FunctionGetResponse {
        return unary("/FunctionGet", request) as Api.FunctionGetResponse
    }

    override suspend fun functionGetCurrentStats(request: Api.FunctionGetCurrentStatsRequest): Api.FunctionStats {
        return unary("/FunctionGetCurrentStats", request) as Api.FunctionStats
    }

    override suspend fun functionUpdateSchedulingParams(request: Api.FunctionUpdateSchedulingParamsRequest) {
        unary("/FunctionUpdateSchedulingParams", request)
    }

    override suspend fun functionBindParams(request: Api.FunctionBindParamsRequest): Api.FunctionBindParamsResponse {
        return unary("/FunctionBindParams", request) as Api.FunctionBindParamsResponse
    }

    override suspend fun sandboxCreate(request: Api.SandboxCreateRequest): Api.SandboxCreateResponse {
        return unary("/SandboxCreate", request) as Api.SandboxCreateResponse
    }

    override suspend fun sandboxWait(request: Api.SandboxWaitRequest): Api.SandboxWaitResponse {
        return unary("/SandboxWait", request) as Api.SandboxWaitResponse
    }

    override suspend fun sandboxGetFromName(request: Api.SandboxGetFromNameRequest): Api.SandboxGetFromNameResponse {
        return unary("/SandboxGetFromName", request) as Api.SandboxGetFromNameResponse
    }

    override suspend fun sandboxList(request: Api.SandboxListRequest): Api.SandboxListResponse {
        return unary("/SandboxList", request) as Api.SandboxListResponse
    }

    override suspend fun sandboxTerminate(request: Api.SandboxTerminateRequest): Api.SandboxTerminateResponse {
        return unary("/SandboxTerminate", request) as Api.SandboxTerminateResponse
    }

    override suspend fun sandboxGetTunnels(request: Api.SandboxGetTunnelsRequest): Api.SandboxGetTunnelsResponse {
        return unary("/SandboxGetTunnels", request) as Api.SandboxGetTunnelsResponse
    }

    override suspend fun sandboxCreateConnectToken(
        request: Api.SandboxCreateConnectTokenRequest,
    ): Api.SandboxCreateConnectTokenResponse {
        return unary("/SandboxCreateConnectToken", request) as Api.SandboxCreateConnectTokenResponse
    }

    override suspend fun sandboxTagsSet(request: Api.SandboxTagsSetRequest) {
        unary("/SandboxTagsSet", request)
    }

    override suspend fun sandboxTagsGet(request: Api.SandboxTagsGetRequest): Api.SandboxTagsGetResponse {
        return unary("/SandboxTagsGet", request) as Api.SandboxTagsGetResponse
    }

    override suspend fun queueGetOrCreate(request: Api.QueueGetOrCreateRequest): Api.QueueGetOrCreateResponse {
        return unary("/QueueGetOrCreate", request) as Api.QueueGetOrCreateResponse
    }

    override suspend fun queueDelete(request: Api.QueueDeleteRequest) {
        unary("/QueueDelete", request)
    }

    override suspend fun queueHeartbeat(request: Api.QueueHeartbeatRequest) {
        unary("/QueueHeartbeat", request)
    }

    override suspend fun queueGet(request: Api.QueueGetRequest): Api.QueueGetResponse {
        return unary("/QueueGet", request) as Api.QueueGetResponse
    }

    override suspend fun queuePut(request: Api.QueuePutRequest) {
        unary("/QueuePut", request)
    }

    override suspend fun queueLen(request: Api.QueueLenRequest): Api.QueueLenResponse {
        return unary("/QueueLen", request) as Api.QueueLenResponse
    }

    override suspend fun queueNextItems(request: Api.QueueNextItemsRequest): Api.QueueNextItemsResponse {
        return unary("/QueueNextItems", request) as Api.QueueNextItemsResponse
    }

    override suspend fun queueClear(request: Api.QueueClearRequest) {
        unary("/QueueClear", request)
    }

    override suspend fun functionMap(request: Api.FunctionMapRequest): Api.FunctionMapResponse {
        return unary("/FunctionMap", request) as Api.FunctionMapResponse
    }

    override suspend fun functionGetOutputs(request: Api.FunctionGetOutputsRequest): Api.FunctionGetOutputsResponse {
        return unary("/FunctionGetOutputs", request) as Api.FunctionGetOutputsResponse
    }

    override suspend fun functionRetryInputs(request: Api.FunctionRetryInputsRequest): Api.FunctionRetryInputsResponse {
        return unary("/FunctionRetryInputs", request) as Api.FunctionRetryInputsResponse
    }

    override suspend fun functionCallCancel(request: Api.FunctionCallCancelRequest) {
        unary("/FunctionCallCancel", request)
    }

    override suspend fun authTokenGet(request: Api.AuthTokenGetRequest): Api.AuthTokenGetResponse {
        return unary("/AuthTokenGet", request) as Api.AuthTokenGetResponse
    }

    override suspend fun taskGetCommandRouterAccess(
        request: Api.TaskGetCommandRouterAccessRequest,
    ): Api.TaskGetCommandRouterAccessResponse {
        return unary("/TaskGetCommandRouterAccess", request) as Api.TaskGetCommandRouterAccessResponse
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
