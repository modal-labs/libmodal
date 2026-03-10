package com.modal.modalkt

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertNotNull
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking
import modal.client.*

class ClsServiceTest {
    private fun mockFunctionProto(): FunctionGetResponse {
        return FunctionGetResponse.newBuilder()
            .setFunctionId("fid")
            .setHandleMetadata(
                FunctionHandleMetadata.newBuilder()
                    .putMethodHandleMetadata("echo_string", FunctionHandleMetadata.getDefaultInstance())
                    .setClassParameterInfo(
                        ClassParameterInfo.newBuilder()
                            .addAllSchema(emptyList())
                            .build(),
                    )
                    .build(),
            )
            .build()
    }

    @Test
    fun clsWithOptionsStacking() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/FunctionGet") { mockFunctionProto() }
        mock.handleUnary("/FunctionBindParams") { request ->
            request as FunctionBindParamsRequest
            assertEquals("fid", request.functionId)
            val options = requireNotNull(request.functionOptions)
            val resources = requireNotNull(options.resources)
            assertEquals(60, options.timeoutSecs)
            assertEquals(250, resources.milliCpu)
            assertEquals(256, resources.memoryMb)
            assertTrue(options.hasResources())
            assertEquals(listOf("sec-1"), options.secretIdsList)
            assertEquals(true, options.replaceSecretIds)
            assertEquals(true, options.replaceVolumeMounts)
            assertEquals("/mnt/test", options.volumeMountsList.first().mountPath)
            FunctionBindParamsResponse.newBuilder()
                .setBoundFunctionId("fid-1")
                .build()
        }

        val cls = client.cls.fromName("libmodal-test-support", "EchoCls")
        val instance = cls
            .withOptions(ClsWithOptionsParams(timeoutMs = 45_000, cpu = 0.25))
            .withOptions(ClsWithOptionsParams(timeoutMs = 60_000, memoryMiB = 256, gpu = "T4"))
            .withOptions(
                ClsWithOptionsParams(
                    secrets = listOf(Secret("sec-1")),
                    volumes = mapOf("/mnt/test" to Volume("vol-1")),
                ),
            )
            .instance()
        assertNotNull(instance)
    }

    @Test
    fun clsWithConcurrencyAndBatchingChaining() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/FunctionGet") { mockFunctionProto() }
        mock.handleUnary("/FunctionBindParams") { request ->
            request as FunctionBindParamsRequest
            val options = requireNotNull(request.functionOptions)
            assertEquals(60, options.timeoutSecs)
            assertEquals(10, options.maxConcurrentInputs)
            assertEquals(11, options.batchMaxSize)
            assertEquals(12, options.batchLingerMs)
            FunctionBindParamsResponse.newBuilder()
                .setBoundFunctionId("fid-1")
                .build()
        }

        val cls = client.cls.fromName("libmodal-test-support", "EchoCls")
        val instance = cls
            .withOptions(ClsWithOptionsParams(timeoutMs = 60_000))
            .withConcurrency(ClsWithConcurrencyParams(maxInputs = 10))
            .withBatching(ClsWithBatchingParams(maxBatchSize = 11, waitMs = 12))
            .instance()
        assertNotNull(instance)
    }

    @Test
    fun clsWithOptionsRetries() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/FunctionGet") { mockFunctionProto() }
        mock.handleUnary("/FunctionBindParams") { request ->
            request as FunctionBindParamsRequest
            val policy = requireNotNull(requireNotNull(request.functionOptions).retryPolicy)
            assertEquals(3, policy.retries)
            assertEquals(1.0f, policy.backoffCoefficient)
            assertEquals(1000, policy.initialDelayMs)
            assertEquals(60_000, policy.maxDelayMs)
            FunctionBindParamsResponse.newBuilder().setBoundFunctionId("fid-1").build()
        }
        client.cls.fromName("libmodal-test-support", "EchoCls")
            .withOptions(ClsWithOptionsParams(retries = 3))
            .instance()
    }

    @Test
    fun clsWithOptionsInvalidValues() = runBlocking {
        val (client, mock) = createMockModalClients()
        mock.handleUnary("/FunctionGet") { mockFunctionProto() }
        val cls = client.cls.fromName("libmodal-test-support", "EchoCls")

        assertFailsWith<InvalidError> {
            cls.withOptions(ClsWithOptionsParams(timeoutMs = 1500)).instance()
        }
        assertFailsWith<InvalidError> {
            cls.withOptions(ClsWithOptionsParams(scaledownWindowMs = 2500)).instance()
        }
    }
}
