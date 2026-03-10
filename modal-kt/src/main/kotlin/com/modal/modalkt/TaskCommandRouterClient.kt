package com.modal.modalkt

import io.grpc.ManagedChannel
import io.grpc.Metadata
import io.grpc.netty.shaded.io.grpc.netty.NettyChannelBuilder
import java.net.URI
import java.util.Base64
import java.util.concurrent.TimeUnit
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import modal.client.*
import modal.task_command_router.*

interface TaskRouterAccessProvider {
    suspend fun taskGetCommandRouterAccess(
        request: TaskGetCommandRouterAccessRequest,
    ): TaskGetCommandRouterAccessResponse
}

interface TaskCommandRouter {
    suspend fun execStart(request: TaskExecStartRequest)

    suspend fun execStdinWrite(request: TaskExecStdinWriteRequest)

    suspend fun execWait(
        request: TaskExecWaitRequest,
    ): TaskExecWaitResponse

    suspend fun execStdioRead(
        request: TaskExecStdioReadRequest,
    ): Flow<TaskExecStdioReadResponse>

    suspend fun mountDirectory(request: TaskMountDirectoryRequest)

    suspend fun snapshotDirectory(
        request: TaskSnapshotDirectoryRequest,
    ): TaskSnapshotDirectoryResponse

    fun close()
}

fun parseJwtExpiration(jwtToken: String, logger: Logger): Long? {
    return try {
        val parts = jwtToken.split(".")
        if (parts.size != 3) {
            return null
        }
        val payload = Base64.getUrlDecoder().decode(parts[1])
        val json = payload.toString(Charsets.UTF_8)
        Regex("\"exp\"\\s*:\\s*(\\d+)").find(json)?.groupValues?.get(1)?.toLongOrNull()
    } catch (error: IllegalArgumentException) {
        logger.warn("Failed to parse JWT expiration", "error", error)
        null
    }
}

suspend fun <T> callWithRetriesOnTransientErrors(
    func: suspend () -> T,
    baseDelayMs: Long = 10,
    delayFactor: Int = 2,
    maxRetries: Int? = 10,
    deadlineMs: Long? = null,
    isClosed: (() -> Boolean)? = null,
): T {
    var delayMs = baseDelayMs
    var retries = 0

    while (true) {
        if (deadlineMs != null && System.currentTimeMillis() >= deadlineMs) {
            throw InvalidError("Deadline exceeded")
        }

        try {
            return func()
        } catch (error: Throwable) {
            if (statusCode(error) == io.grpc.Status.Code.CANCELLED && isClosed?.invoke() == true) {
                throw ClientClosedError()
            }
            val retryable = isRetryableGrpc(error)
            val canRetry = maxRetries == null || retries < maxRetries
            if (!retryable || !canRetry) {
                throw error
            }
            if (deadlineMs != null && System.currentTimeMillis() + delayMs >= deadlineMs) {
                throw InvalidError("Deadline exceeded")
            }
            delay(delayMs)
            delayMs *= delayFactor
            retries += 1
        }
    }
}

class TaskCommandRouterClientImpl(
    private val serverClient: TaskRouterAccessProvider,
    private val taskId: String,
    private val serverUrl: String,
    private var jwt: String,
    private val logger: Logger,
) : TaskCommandRouter {
    private val serviceName = "modal.task_command_router.TaskCommandRouter"
    private val refreshMutex = Mutex()
    private var jwtExpiration: Long? = parseJwtExpiration(jwt, logger)
    private var closed: Boolean = false
    private val channel: ManagedChannel = buildChannel(serverUrl)

    companion object {
        suspend fun tryInit(
            serverClient: TaskRouterAccessProvider,
            taskId: String,
            logger: Logger,
            profile: Profile,
        ): TaskCommandRouterClientImpl? {
            val response = serverClient.taskGetCommandRouterAccess(
                TaskGetCommandRouterAccessRequest.newBuilder()
                    .setTaskId(taskId)
                    .build(),
            )
            return TaskCommandRouterClientImpl(
                serverClient = serverClient,
                taskId = taskId,
                serverUrl = response.url,
                jwt = response.jwt,
                logger = logger,
            )
        }
    }

    override suspend fun execStart(request: TaskExecStartRequest) {
        callWithAuthRetry {
            unaryRpc(
                channel = channel,
                serviceName = serviceName,
                methodName = "TaskExecStart",
                request = request,
                requestAdapter = TaskExecStartRequest.ADAPTER,
                responseAdapter = TaskExecStartResponse.ADAPTER,
                options = RpcOptions(headers = authHeaders()),
            )
        }
    }

    override suspend fun execStdinWrite(request: TaskExecStdinWriteRequest) {
        callWithAuthRetry {
            unaryRpc(
                channel = channel,
                serviceName = serviceName,
                methodName = "TaskExecStdinWrite",
                request = request,
                requestAdapter = TaskExecStdinWriteRequest.ADAPTER,
                responseAdapter = TaskExecStdinWriteResponse.ADAPTER,
                options = RpcOptions(headers = authHeaders()),
            )
        }
    }

    override suspend fun execWait(
        request: TaskExecWaitRequest,
    ): TaskExecWaitResponse {
        return callWithAuthRetry {
            unaryRpc(
                channel = channel,
                serviceName = serviceName,
                methodName = "TaskExecWait",
                request = request,
                requestAdapter = TaskExecWaitRequest.ADAPTER,
                responseAdapter = TaskExecWaitResponse.ADAPTER,
                options = RpcOptions(headers = authHeaders()),
            )
        }
    }

    override suspend fun execStdioRead(
        request: TaskExecStdioReadRequest,
    ): Flow<TaskExecStdioReadResponse> {
        return callWithAuthRetry {
            serverStreamingRpc(
                channel = channel,
                serviceName = serviceName,
                methodName = "TaskExecStdioRead",
                request = request,
                requestAdapter = TaskExecStdioReadRequest.ADAPTER,
                responseAdapter = TaskExecStdioReadResponse.ADAPTER,
                options = RpcOptions(headers = authHeaders()),
            )
        }
    }

    override suspend fun mountDirectory(request: TaskMountDirectoryRequest) {
        callWithAuthRetry {
            unaryRpc(
                channel = channel,
                serviceName = serviceName,
                methodName = "TaskMountDirectory",
                request = request,
                requestAdapter = TaskMountDirectoryRequest.ADAPTER,
                responseAdapter = com.squareup.wire.ProtoAdapter.EMPTY,
                options = RpcOptions(headers = authHeaders()),
            )
        }
    }

    override suspend fun snapshotDirectory(
        request: TaskSnapshotDirectoryRequest,
    ): TaskSnapshotDirectoryResponse {
        return callWithAuthRetry {
            unaryRpc(
                channel = channel,
                serviceName = serviceName,
                methodName = "TaskSnapshotDirectory",
                request = request,
                requestAdapter = TaskSnapshotDirectoryRequest.ADAPTER,
                responseAdapter = TaskSnapshotDirectoryResponse.ADAPTER,
                options = RpcOptions(headers = authHeaders()),
            )
        }
    }

    suspend fun refreshJwt() {
        refreshMutex.withLock {
            if (closed) {
                return
            }
            val currentExpiration = jwtExpiration
            if (currentExpiration != null) {
                val nowSeconds = System.currentTimeMillis() / 1000
                if (currentExpiration - nowSeconds > 30) {
                    logger.debug(
                        "Skipping JWT refresh because expiration is far enough in the future",
                        "task_id",
                        taskId,
                    )
                    return
                }
            }

            val response = serverClient.taskGetCommandRouterAccess(
                TaskGetCommandRouterAccessRequest.newBuilder()
                    .setTaskId(taskId)
                    .build(),
            )
            if (response.url != serverUrl) {
                throw InvalidError("Task router URL changed during session")
            }
            jwt = response.jwt
            jwtExpiration = parseJwtExpiration(response.jwt, logger)
        }
    }

    override fun close() {
        closed = true
        channel.shutdownNow()
    }

    private suspend fun <T> callWithAuthRetry(block: suspend () -> T): T {
        try {
            return block()
        } catch (error: Throwable) {
            if (statusCode(error) == io.grpc.Status.Code.UNAUTHENTICATED) {
                refreshJwt()
                return block()
            }
            throw error
        }
    }

    private fun authHeaders(): Metadata {
        val metadata = Metadata()
        metadata.put(
            Metadata.Key.of("authorization", Metadata.ASCII_STRING_MARSHALLER),
            "Bearer $jwt",
        )
        return metadata
    }

    private fun buildChannel(urlString: String): ManagedChannel {
        val url = URI.create(urlString)
        val builder = NettyChannelBuilder.forAddress(
            url.host,
            if (url.port == -1) 443 else url.port,
        )
            .keepAliveTime(30, TimeUnit.SECONDS)
            .keepAliveTimeout(10, TimeUnit.SECONDS)
            .keepAliveWithoutCalls(true)
            .maxInboundMessageSize(100 * 1024 * 1024)

        if (url.host == "localhost" || url.host == "127.0.0.1") {
            builder.usePlaintext()
        } else {
            builder.useTransportSecurity()
        }
        return builder.build()
    }
}
