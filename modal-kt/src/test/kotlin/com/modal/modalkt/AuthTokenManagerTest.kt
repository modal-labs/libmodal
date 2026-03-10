package com.modal.modalkt

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertFalse
import kotlin.test.assertTrue
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.runBlocking
import modal.client.*

class AuthTokenManagerTest {
    @Test
    fun decodeJwtAndCacheToken() = runBlocking {
        val now = System.currentTimeMillis() / 1000
        val token = createTestJwt(now + 1800)
        val provider = TestAuthTokenProvider(token)
        val manager = AuthTokenManager(provider, newLogger())

        val result = manager.getToken()

        assertEquals(token, result)
        assertEquals(token, manager.getCurrentToken())
    }

    @Test
    fun lazyFetchOnlyHappensOnce() = runBlocking {
        val token = createTestJwt(System.currentTimeMillis() / 1000 + 3600)
        val provider = TestAuthTokenProvider(token)
        val manager = AuthTokenManager(provider, newLogger())

        assertEquals(token, manager.getToken())
        assertEquals(token, manager.getToken())
        assertEquals(1, provider.calls)
    }

    @Test
    fun expiredStateIsComputedCorrectly() {
        val manager = AuthTokenManager(TestAuthTokenProvider(""), newLogger())
        val now = System.currentTimeMillis() / 1000

        manager.setToken(createTestJwt(now + 3600), now + 3600)
        assertFalse(manager.isExpired(now))

        manager.setToken(createTestJwt(now - 60), now - 60)
        assertTrue(manager.isExpired(now))
    }

    @Test
    fun expiredTokenRefreshes() = runBlocking {
        val now = System.currentTimeMillis() / 1000
        val provider = TestAuthTokenProvider(createTestJwt(now + 3600))
        val manager = AuthTokenManager(provider, newLogger())
        manager.setToken(createTestJwt(now - 60), now - 60)

        assertEquals(provider.token, manager.getToken())
    }

    @Test
    fun concurrentGetsUseSingleRefresh() = runBlocking {
        val token = createTestJwt(System.currentTimeMillis() / 1000 + 3600)
        val provider = TestAuthTokenProvider(token)
        val manager = AuthTokenManager(provider, newLogger())

        val results = List(3) {
            async { manager.getToken() }
        }.awaitAll()

        assertEquals(listOf(token, token, token), results)
        assertEquals(1, provider.calls)
    }

    @Test
    fun emptyResponseFails() = runBlocking {
        val manager = AuthTokenManager(TestAuthTokenProvider(""), newLogger())

        assertFailsWith<InvalidError> {
            manager.getToken()
        }
    }

    private class TestAuthTokenProvider(var token: String) : AuthTokenProvider {
        var calls: Int = 0

        override suspend fun authTokenGet(request: AuthTokenGetRequest): AuthTokenGetResponse {
            calls += 1
            return AuthTokenGetResponse.newBuilder()
                .setToken(token)
                .build()
        }
    }

    private fun createTestJwt(expiry: Long): String {
        val header = base64Url("""{"alg":"HS256","typ":"JWT"}""")
        val payload = base64Url("""{"exp":$expiry}""")
        return "$header.$payload.signature"
    }

    private fun base64Url(input: String): String {
        return java.util.Base64.getUrlEncoder()
            .withoutPadding()
            .encodeToString(input.toByteArray())
    }
}
