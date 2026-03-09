package com.modal.modalkt

import kotlinx.coroutines.delay
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import modal.client.Api
import java.util.Base64

interface TaskRouterAccessProvider {
    suspend fun taskGetCommandRouterAccess(
        request: Api.TaskGetCommandRouterAccessRequest,
    ): Api.TaskGetCommandRouterAccessResponse
}

fun parseJwtExpiration(jwtToken: String, logger: Logger): Long? {
    return try {
        val parts = jwtToken.split(".")
        if (parts.size != 3) {
            return null
        }

        val payload = Base64.getUrlDecoder().decode(parts[1])
        val json = payload.toString(Charsets.UTF_8)
        val exp = Regex("\"exp\"\\s*:\\s*(\\d+)").find(json)?.groupValues?.get(1)?.toLongOrNull()
        exp
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
) {
    private val refreshMutex = Mutex()
    private var jwtExpiration: Long? = parseJwtExpiration(jwt, logger)
    private var closed: Boolean = false

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
                Api.TaskGetCommandRouterAccessRequest.newBuilder()
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

    fun close() {
        closed = true
    }
}
