package com.modal.modalkt

import com.github.stefanbirkner.systemlambda.SystemLambda.withEnvironmentVariable
import com.squareup.wire.ProtoAdapter
import io.grpc.ClientCall
import io.grpc.ClientInterceptor
import io.grpc.ForwardingClientCall
import io.grpc.ForwardingClientCallListener
import io.grpc.Metadata
import io.grpc.MethodDescriptor
import io.grpc.Server
import io.grpc.ServerServiceDefinition
import io.grpc.netty.shaded.io.grpc.netty.NettyServerBuilder
import io.grpc.stub.ServerCalls
import java.io.ByteArrayInputStream
import java.io.InputStream
import java.util.concurrent.atomic.AtomicBoolean
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFails
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking
import modal.client.*

class ClientRuntimeTest {
    @Test
    fun modalClientWithCustomInterceptor() = runBlocking {
        val firstCalled = AtomicBoolean(false)
        val secondCalled = AtomicBoolean(false)
        val firstMethod = mutableListOf<String>()
        val secondMethod = mutableListOf<String>()

        val firstInterceptor = object : ClientInterceptor {
            override fun <ReqT : Any?, RespT : Any?> interceptCall(
                method: MethodDescriptor<ReqT, RespT>,
                callOptions: io.grpc.CallOptions,
                next: io.grpc.Channel,
            ): ClientCall<ReqT, RespT> {
                firstCalled.set(true)
                firstMethod += method.fullMethodName
                return next.newCall(method, callOptions)
            }
        }
        val secondInterceptor = object : ClientInterceptor {
            override fun <ReqT : Any?, RespT : Any?> interceptCall(
                method: MethodDescriptor<ReqT, RespT>,
                callOptions: io.grpc.CallOptions,
                next: io.grpc.Channel,
            ): ClientCall<ReqT, RespT> {
                secondCalled.set(true)
                secondMethod += method.fullMethodName
                return next.newCall(method, callOptions)
            }
        }

        val server = testServer(sleepMillis = 0)
        try {
            withEnvironmentVariable("MODAL_SERVER_URL", "http://127.0.0.1:${server.port}").execute {
                runBlocking {
                    val client = ModalClient(
                        ModalClientParams(
                            tokenId = "test-token-id",
                            tokenSecret = "test-token-secret",
                            grpcInterceptors = listOf(firstInterceptor, secondInterceptor),
                            environment = "test",
                        ),
                    )
                    try {
                        client.apps.fromName("test-app")
                    } finally {
                        client.close()
                    }
                }
            }
        } finally {
            server.shutdownNow()
        }

        assertTrue(firstCalled.get())
        assertTrue(secondCalled.get())
        assertTrue(firstMethod.first().contains("ModalClient/"))
        assertTrue(secondMethod.first().contains("ModalClient/"))
    }

    @Test
    fun clientRespectsTimeout() = runBlocking {
        val server = testServer(sleepMillis = 100)
        try {
            withEnvironmentVariable("MODAL_SERVER_URL", "http://127.0.0.1:${server.port}").execute {
                runBlocking {
                    val client = ModalClient(
                        ModalClientParams(
                            tokenId = "test-token-id",
                            tokenSecret = "test-token-secret",
                            environment = "test",
                            timeoutMs = 10,
                        ),
                    )

                    assertFails {
                        client.apps.fromName("test-app")
                    }
                    client.close()
                }
            }
        } finally {
            server.shutdownNow()
        }
    }

    private fun testServer(sleepMillis: Long): Server {
        val service = ServerServiceDefinition.builder("modal.client.ModalClient")
            .addMethod(
                unaryMethod(
                    methodName = "AuthTokenGet",
                    requestAdapter = AuthTokenGetRequest.ADAPTER,
                    responseAdapter = AuthTokenGetResponse.ADAPTER,
                ),
                ServerCalls.asyncUnaryCall { _: AuthTokenGetRequest, responseObserver ->
                    responseObserver.onNext(
                        AuthTokenGetResponse.newBuilder()
                            .setToken("x.eyJleHAiOjk5OTk5OTk5OTl9.x")
                            .build(),
                    )
                    responseObserver.onCompleted()
                },
            )
            .addMethod(
                unaryMethod(
                    methodName = "AppGetOrCreate",
                    requestAdapter = AppGetOrCreateRequest.ADAPTER,
                    responseAdapter = AppGetOrCreateResponse.ADAPTER,
                ),
                ServerCalls.asyncUnaryCall { request: AppGetOrCreateRequest, responseObserver ->
                    if (sleepMillis > 0) {
                        Thread.sleep(sleepMillis)
                    }
                    responseObserver.onNext(
                        AppGetOrCreateResponse.newBuilder()
                            .setAppId(request.appName)
                            .build(),
                    )
                    responseObserver.onCompleted()
                },
            )
            .build()
        val server = NettyServerBuilder.forPort(0)
            .addService(service)
            .build()
        server.start()
        return server
    }

    private fun <ReqT : Any, ResT : Any> unaryMethod(
        methodName: String,
        requestAdapter: ProtoAdapter<ReqT>,
        responseAdapter: ProtoAdapter<ResT>,
    ): MethodDescriptor<ReqT, ResT> {
        return MethodDescriptor.newBuilder<ReqT, ResT>()
            .setType(MethodDescriptor.MethodType.UNARY)
            .setFullMethodName(MethodDescriptor.generateFullMethodName("modal.client.ModalClient", methodName))
            .setRequestMarshaller(wireMarshaller(requestAdapter))
            .setResponseMarshaller(wireMarshaller(responseAdapter))
            .build()
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
}
