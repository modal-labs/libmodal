package com.modal.modalkt

import java.util.Base64
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import modal.client.*

const val REFRESH_WINDOW_SECONDS: Long = 5 * 60
const val DEFAULT_EXPIRY_OFFSET_SECONDS: Long = 20 * 60

interface AuthTokenProvider {
    suspend fun authTokenGet(request: AuthTokenGetRequest): AuthTokenGetResponse
}

class AuthTokenManager(
    private val client: AuthTokenProvider,
    private val logger: Logger,
) {
    private val mutex = Mutex()
    private var currentToken: String = ""
    private var tokenExpirySeconds: Long = 0

    suspend fun getToken(): String {
        if (currentToken.isEmpty() || isExpired()) {
            return lockedRefreshToken()
        }

        if (needsRefresh()) {
            try {
                lockedRefreshToken()
            } catch (error: Throwable) {
                logger.error("refreshing auth token", "error", error)
            }
        }

        return currentToken
    }

    fun getCurrentToken(): String {
        return currentToken
    }

    fun setToken(token: String, expirySeconds: Long) {
        currentToken = token
        tokenExpirySeconds = expirySeconds
    }

    fun isExpired(nowSeconds: Long = nowSeconds()): Boolean {
        return nowSeconds >= tokenExpirySeconds
    }

    private fun needsRefresh(nowSeconds: Long = nowSeconds()): Boolean {
        return nowSeconds >= tokenExpirySeconds - REFRESH_WINDOW_SECONDS
    }

    private suspend fun lockedRefreshToken(): String {
        mutex.withLock {
            if (currentToken.isNotEmpty() && !needsRefresh()) {
                return currentToken
            }

            fetchToken()
            return currentToken
        }
    }

    private suspend fun fetchToken() {
        val response = client.authTokenGet(AuthTokenGetRequest.getDefaultInstance())
        val token = response.token
        if (token.isNullOrEmpty()) {
            throw InvalidError(
                "Internal error: did not receive auth token from server, please contact Modal support",
            )
        }

        currentToken = token
        val expiration = decodeJwtExpiry(token)
        tokenExpirySeconds = if (expiration > 0) {
            expiration
        } else {
            logger.warn("x-modal-auth-token does not contain exp field")
            nowSeconds() + DEFAULT_EXPIRY_OFFSET_SECONDS
        }

        val expiresIn = tokenExpirySeconds - nowSeconds()
        val refreshIn = tokenExpirySeconds - nowSeconds() - REFRESH_WINDOW_SECONDS
        logger.debug(
            "Fetched auth token",
            "expires_in",
            "${expiresIn}s",
            "refresh_in",
            "${refreshIn}s",
        )
    }

    private fun decodeJwtExpiry(token: String): Long {
        return try {
            val parts = token.split(".")
            if (parts.size != 3) {
                return 0
            }

            val payload = parts[1]
            val decoded = Base64.getUrlDecoder().decode(payload)
            val content = decoded.toString(Charsets.UTF_8)
            val expMatch = Regex("\"exp\"\\s*:\\s*(\\d+)").find(content)
            expMatch?.groupValues?.get(1)?.toLongOrNull() ?: 0
        } catch (_: IllegalArgumentException) {
            0
        }
    }

    private fun nowSeconds(): Long {
        return System.currentTimeMillis() / 1000
    }
}
