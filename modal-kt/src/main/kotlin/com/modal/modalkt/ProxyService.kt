package com.modal.modalkt

import io.grpc.Status
import modal.client.*

data class ProxyFromNameParams(
    val environment: String? = null,
)

class ProxyService(
    private val client: ModalClient,
) {
    suspend fun fromName(
        name: String,
        params: ProxyFromNameParams = ProxyFromNameParams(),
    ): Proxy {
        try {
            val response = client.cpClient.proxyGet(
                ProxyGetRequest.newBuilder()
                    .setName(name)
                    .setEnvironmentName(client.environmentName(params.environment))
                    .build(),
            )
            val proxy = response.proxy
            if (proxy == null || proxy.proxyId.isEmpty()) {
                throw NotFoundError("Proxy '$name' not found")
            }
            return Proxy(proxy.proxyId)
        } catch (error: Throwable) {
            if (statusCode(error) == Status.Code.NOT_FOUND) {
                throw NotFoundError("Proxy '$name' not found")
            }
            throw error
        }
    }
}
