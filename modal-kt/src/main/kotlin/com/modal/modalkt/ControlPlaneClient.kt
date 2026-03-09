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
