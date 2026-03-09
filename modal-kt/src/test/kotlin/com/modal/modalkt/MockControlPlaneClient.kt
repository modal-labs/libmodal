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
