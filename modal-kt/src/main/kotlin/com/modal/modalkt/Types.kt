package com.modal.modalkt

data class App(
    val appId: String,
    val name: String? = null,
)

data class Secret(
    val secretId: String,
    val name: String? = null,
)

data class Proxy(
    val proxyId: String,
)

class Volume(
    val volumeId: String,
    val name: String? = null,
    val isReadOnly: Boolean = false,
    private val ephemeralHeartbeatManager: EphemeralHeartbeatManager? = null,
) {
    fun readOnly(): Volume {
        return Volume(volumeId, name, true, ephemeralHeartbeatManager)
    }

    fun closeEphemeral() {
        if (ephemeralHeartbeatManager == null) {
            throw InvalidError("Volume is not ephemeral.")
        }
        ephemeralHeartbeatManager.stop()
    }
}

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
