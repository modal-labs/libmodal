package com.modal.modalkt.examples

import com.modal.modalkt.ModalClient
import com.modal.modalkt.ModalClientParams
import io.grpc.ClientCall
import io.grpc.ClientInterceptor
import io.grpc.ForwardingClientCall
import io.grpc.ForwardingClientCallListener
import io.grpc.Metadata
import io.grpc.MethodDescriptor

suspend fun main() {
    val interceptor = object : ClientInterceptor {
        override fun <ReqT : Any?, RespT : Any?> interceptCall(
            method: MethodDescriptor<ReqT, RespT>,
            callOptions: io.grpc.CallOptions,
            next: io.grpc.Channel,
        ): ClientCall<ReqT, RespT> {
            val start = System.currentTimeMillis()
            val call = next.newCall(method, callOptions)
            return object : ForwardingClientCall.SimpleForwardingClientCall<ReqT, RespT>(call) {
                override fun start(responseListener: Listener<RespT>, headers: Metadata) {
                    super.start(
                        object : ForwardingClientCallListener.SimpleForwardingClientCallListener<RespT>(responseListener) {
                            override fun onClose(status: io.grpc.Status, trailers: Metadata) {
                                val duration = System.currentTimeMillis() - start
                                println("[TELEMETRY] method=${method.fullMethodName} duration=${duration}ms status=${status.code}")
                                super.onClose(status, trailers)
                            }
                        },
                        headers,
                    )
                }
            }
        }
    }

    val modal = ModalClient(
        ModalClientParams(
            grpcInterceptors = listOf(interceptor),
        ),
    )
    val function = modal.functions.fromName("libmodal-test-support", "echo_string")
    println(function.remote(kwargs = mapOf("s" to "Hello from telemetry example!")))
}
