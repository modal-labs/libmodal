package com.modal.modalkt

import com.squareup.wire.ProtoAdapter
import io.grpc.CallOptions
import io.grpc.Channel
import io.grpc.ClientCall
import io.grpc.Metadata
import io.grpc.MethodDescriptor
import io.grpc.Status
import io.grpc.StatusException
import io.grpc.StatusRuntimeException
import java.io.ByteArrayInputStream
import java.io.InputStream
import java.util.concurrent.TimeUnit
import kotlin.coroutines.resume
import kotlin.coroutines.resumeWithException
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.callbackFlow
import kotlinx.coroutines.suspendCancellableCoroutine

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

private fun <T : Any> wireMarshaller(adapter: ProtoAdapter<T>): MethodDescriptor.Marshaller<T> {
    return object : MethodDescriptor.Marshaller<T> {
        override fun stream(value: T): InputStream {
            return ByteArrayInputStream(adapter.encode(value))
        }

        override fun parse(stream: InputStream): T {
            return stream.use { adapter.decode(it.readBytes()) }
        }
    }
}

private fun callOptions(timeoutMs: Long?): CallOptions {
    if (timeoutMs == null) {
        return CallOptions.DEFAULT
    }
    return CallOptions.DEFAULT.withDeadlineAfter(timeoutMs, TimeUnit.MILLISECONDS)
}

private fun <ReqT : Any, ResT : Any> methodDescriptor(
    serviceName: String,
    methodName: String,
    requestAdapter: ProtoAdapter<ReqT>,
    responseAdapter: ProtoAdapter<ResT>,
    type: MethodDescriptor.MethodType,
): MethodDescriptor<ReqT, ResT> {
    return MethodDescriptor.newBuilder<ReqT, ResT>()
        .setType(type)
        .setFullMethodName(MethodDescriptor.generateFullMethodName(serviceName, methodName))
        .setRequestMarshaller(wireMarshaller(requestAdapter))
        .setResponseMarshaller(wireMarshaller(responseAdapter))
        .build()
}

internal suspend fun <ReqT : Any, ResT : Any> unaryRpc(
    channel: Channel,
    serviceName: String,
    methodName: String,
    request: ReqT,
    requestAdapter: ProtoAdapter<ReqT>,
    responseAdapter: ProtoAdapter<ResT>,
    options: RpcOptions = RpcOptions(),
): ResT {
    val descriptor = methodDescriptor(
        serviceName = serviceName,
        methodName = methodName,
        requestAdapter = requestAdapter,
        responseAdapter = responseAdapter,
        type = MethodDescriptor.MethodType.UNARY,
    )
    val call = channel.newCall(descriptor, callOptions(options.timeoutMs))

    return suspendCancellableCoroutine { continuation ->
        var response: ResT? = null

        continuation.invokeOnCancellation {
            call.cancel("Cancelled", it)
        }

        call.start(
            object : ClientCall.Listener<ResT>() {
                override fun onMessage(message: ResT) {
                    response = message
                }

                override fun onClose(status: Status, trailers: Metadata) {
                    if (!status.isOk) {
                        continuation.resumeWithException(status.asRuntimeException(trailers))
                        return
                    }

                    val value = response
                    if (value == null) {
                        continuation.resumeWithException(
                            Status.INTERNAL
                                .withDescription("Missing unary gRPC response for $methodName")
                                .asRuntimeException(trailers),
                        )
                        return
                    }

                    continuation.resume(value)
                }
            },
            options.headers,
        )
        call.request(2)
        call.sendMessage(request)
        call.halfClose()
    }
}

internal fun <ReqT : Any, ResT : Any> serverStreamingRpc(
    channel: Channel,
    serviceName: String,
    methodName: String,
    request: ReqT,
    requestAdapter: ProtoAdapter<ReqT>,
    responseAdapter: ProtoAdapter<ResT>,
    options: RpcOptions = RpcOptions(),
): Flow<ResT> {
    val descriptor = methodDescriptor(
        serviceName = serviceName,
        methodName = methodName,
        requestAdapter = requestAdapter,
        responseAdapter = responseAdapter,
        type = MethodDescriptor.MethodType.SERVER_STREAMING,
    )
    val call = channel.newCall(descriptor, callOptions(options.timeoutMs))

    return callbackFlow {
        call.start(
            object : ClientCall.Listener<ResT>() {
                override fun onMessage(message: ResT) {
                    trySend(message)
                    call.request(1)
                }

                override fun onClose(status: Status, trailers: Metadata) {
                    if (status.isOk) {
                        close()
                    } else {
                        close(status.asRuntimeException(trailers))
                    }
                }
            },
            options.headers,
        )
        call.request(1)
        call.sendMessage(request)
        call.halfClose()

        awaitClose {
            call.cancel("Cancelled", null)
        }
    }
}
