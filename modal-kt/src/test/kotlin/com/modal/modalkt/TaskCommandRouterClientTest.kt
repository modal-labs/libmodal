package com.modal.modalkt

import io.grpc.Status
import io.grpc.StatusException
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertNull
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking
import modal.client.*

class TaskCommandRouterClientTest {
    private val logger = RecordingLogger()

    @Test
    fun parseJwtExpirationWithValidJwt() {
        val exp = System.currentTimeMillis() / 1000 + 3600
        val jwt = mockJwt(exp.toString())
        assertEquals(exp, parseJwtExpiration(jwt, logger))
    }

    @Test
    fun parseJwtExpirationWithoutExpClaim() {
        assertNull(parseJwtExpiration(mockJwt(null), logger))
    }

    @Test
    fun parseJwtExpirationWithMalformedJwt() {
        assertNull(parseJwtExpiration("only.two", logger))
    }

    @Test
    fun parseJwtExpirationWithInvalidBase64() {
        assertNull(parseJwtExpiration("invalid.!!!invalid!!!.signature", logger))
        assertTrue(logger.warnMessages.isNotEmpty())
    }

    @Test
    fun callWithRetriesOnTransientErrorsSucceeds() = runBlocking {
        var calls = 0
        val result = callWithRetriesOnTransientErrors(
            func = {
            calls += 1
            "success"
            },
        )

        assertEquals("success", result)
        assertEquals(1, calls)
    }

    @Test
    fun callWithRetriesOnTransientErrorsRetriesRetryableCodes() = runBlocking {
        for (code in listOf(
            Status.Code.DEADLINE_EXCEEDED,
            Status.Code.UNAVAILABLE,
            Status.Code.CANCELLED,
            Status.Code.INTERNAL,
            Status.Code.UNKNOWN,
        )) {
            var calls = 0
            val result = callWithRetriesOnTransientErrors(
                func = {
                    calls += 1
                    if (calls == 1) {
                        throw StatusException(Status.fromCode(code).withDescription("retry"))
                    }
                    "success"
                },
                baseDelayMs = 1,
            )
            assertEquals("success", result)
            assertEquals(2, calls)
        }
    }

    @Test
    fun callWithRetriesOnTransientErrorsRejectsNonRetryable() = runBlocking {
        val error = assertFailsWith<StatusException> {
            callWithRetriesOnTransientErrors(
                func = {
                    throw StatusException(Status.INVALID_ARGUMENT.withDescription("invalid"))
                },
                baseDelayMs = 1,
            )
        }
        assertEquals(Status.Code.INVALID_ARGUMENT, error.status.code)
    }

    @Test
    fun callWithRetriesOnTransientErrorsRejectsMaxRetriesExceeded() = runBlocking {
        var calls = 0
        assertFailsWith<StatusException> {
            callWithRetriesOnTransientErrors(
                func = {
                    calls += 1
                    throw StatusException(Status.UNAVAILABLE.withDescription("unavailable"))
                },
                baseDelayMs = 1,
                maxRetries = 3,
            )
        }
        assertEquals(4, calls)
    }

    @Test
    fun callWithRetriesOnTransientErrorsRejectsDeadlineExceeded() = runBlocking {
        assertFailsWith<InvalidError> {
            callWithRetriesOnTransientErrors(
                func = {
                    throw StatusException(Status.UNAVAILABLE.withDescription("unavailable"))
                },
                baseDelayMs = 100,
                deadlineMs = System.currentTimeMillis() + 50,
            )
        }
    }

    @Test
    fun refreshJwtRecoversAfterTransientFailure() = runBlocking {
        val provider = object : TaskRouterAccessProvider {
            var calls = 0

            override suspend fun taskGetCommandRouterAccess(
                request: TaskGetCommandRouterAccessRequest,
            ): TaskGetCommandRouterAccessResponse {
                calls += 1
                if (calls == 1) {
                    throw IllegalStateException("Transient network error")
                }
                return TaskGetCommandRouterAccessResponse.newBuilder()
                    .setUrl("https://example.com")
                    .setJwt(mockJwt((System.currentTimeMillis() / 1000 + 3600).toString()))
                    .build()
            }
        }

        val client = TaskCommandRouterClientImpl(
            serverClient = provider,
            taskId = "test-task",
            serverUrl = "https://example.com",
            jwt = mockJwt("0"),
            logger = logger,
        )

        assertFailsWith<IllegalStateException> {
            client.refreshJwt()
        }
        client.refreshJwt()
    }

    private fun mockJwt(exp: String?): String {
        val header = base64Url("""{"alg":"HS256","typ":"JWT"}""")
        val payload = if (exp == null) {
            base64Url("{}")
        } else {
            base64Url("""{"exp":$exp}""")
        }
        return "$header.$payload.fake-signature"
    }

    private fun base64Url(value: String): String {
        return java.util.Base64.getUrlEncoder()
            .withoutPadding()
            .encodeToString(value.toByteArray())
    }

    private class RecordingLogger : Logger {
        val warnMessages = mutableListOf<String>()

        override fun debug(message: String, vararg args: Any?) {
        }

        override fun info(message: String, vararg args: Any?) {
        }

        override fun warn(message: String, vararg args: Any?) {
            warnMessages += message
        }

        override fun error(message: String, vararg args: Any?) {
        }
    }
}
