package com.modal.modalkt

data class Secret(
    val secretId: String,
)

data class Proxy(
    val proxyId: String,
)

data class Volume(
    val volumeId: String,
    val isReadOnly: Boolean = false,
) {
    fun readOnly(): Volume {
        return copy(isReadOnly = true)
    }

    fun closeEphemeral() {
    }
}

data class App(
    val appId: String,
    val name: String? = null,
)

data class Tunnel(
    val host: String,
    val port: Int,
    val unencryptedHost: String? = null,
    val unencryptedPort: Int? = null,
) {
    val url: String
        get() {
            if (port == 443) {
                return "https://$host"
            }
            return "https://$host:$port"
        }

    val tlsSocket: Pair<String, Int>
        get() = host to port

    val tcpSocket: Pair<String, Int>
        get() {
            val rawHost = unencryptedHost
            val rawPort = unencryptedPort
            if (rawHost == null || rawPort == null) {
                throw InvalidError("This tunnel is not configured for unencrypted TCP.")
            }
            return rawHost to rawPort
        }
}
