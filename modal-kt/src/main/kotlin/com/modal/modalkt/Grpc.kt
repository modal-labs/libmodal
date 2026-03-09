package com.modal.modalkt

import io.grpc.Metadata
import io.grpc.Status
import io.grpc.StatusException
import io.grpc.StatusRuntimeException

internal data class RpcOptions(
    val timeoutMs: Long? = null,
    val headers: Metadata = Metadata(),
)

internal fun statusCode(error: Throwable): Status.Code? {
    if (error is StatusException) {
        return error.status.code
    }
    if (error is StatusRuntimeException) {
        return error.status.code
    }
    return null
}

internal fun statusMessage(error: Throwable): String {
    if (error is StatusException) {
        return error.status.description ?: error.message.orEmpty()
    }
    if (error is StatusRuntimeException) {
        return error.status.description ?: error.message.orEmpty()
    }
    return error.message.orEmpty()
}

internal fun isRetryableGrpc(error: Throwable): Boolean {
    return when (statusCode(error)) {
        Status.Code.DEADLINE_EXCEEDED,
        Status.Code.UNAVAILABLE,
        Status.Code.CANCELLED,
        Status.Code.INTERNAL,
        Status.Code.UNKNOWN,
        -> true

        else -> false
    }
}
