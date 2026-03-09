package com.modal.modalkt

import io.grpc.Status
import modal.client.Api

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
