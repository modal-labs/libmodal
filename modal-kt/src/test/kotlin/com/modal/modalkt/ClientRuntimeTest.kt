package com.modal.modalkt

import io.grpc.ClientCall
import io.grpc.ClientInterceptor
import io.grpc.ForwardingClientCall
import io.grpc.ForwardingClientCallListener
import io.grpc.Metadata
import io.grpc.MethodDescriptor
import io.grpc.Server
import io.grpc.netty.shaded.io.grpc.netty.NettyServerBuilder
import io.grpc.stub.StreamObserver
import com.github.stefanbirkner.systemlambda.SystemLambda.withEnvironmentVariable
import kotlinx.coroutines.runBlocking
import modal.client.Api
import modal.client.ModalClientGrpc
import java.util.concurrent.atomic.AtomicBoolean
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFails
import kotlin.test.assertTrue

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
        val service = object : ModalClientGrpc.ModalClientImplBase() {
            override fun authTokenGet(
                request: Api.AuthTokenGetRequest,
                responseObserver: StreamObserver<Api.AuthTokenGetResponse>,
            ) {
                responseObserver.onNext(
                    Api.AuthTokenGetResponse.newBuilder()
                        .setToken("x.eyJleHAiOjk5OTk5OTk5OTl9.x")
                        .build(),
                )
                responseObserver.onCompleted()
            }

            override fun appGetOrCreate(
                request: Api.AppGetOrCreateRequest,
                responseObserver: StreamObserver<Api.AppGetOrCreateResponse>,
            ) {
                if (sleepMillis > 0) {
                    Thread.sleep(sleepMillis)
                }
                responseObserver.onNext(
                    Api.AppGetOrCreateResponse.newBuilder()
                        .setAppId(request.appName)
                        .build(),
                )
                responseObserver.onCompleted()
            }
        }
        val server = NettyServerBuilder.forPort(0)
            .addService(service)
            .build()
        server.start()
        return server
    }
}
