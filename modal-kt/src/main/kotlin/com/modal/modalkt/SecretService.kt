package com.modal.modalkt

import io.grpc.Status
import modal.client.*

data class SecretFromNameParams(
    val environment: String? = null,
    val requiredKeys: List<String> = emptyList(),
)

data class SecretFromObjectParams(
    val environment: String? = null,
)

data class SecretDeleteParams(
    val environment: String? = null,
    val allowMissing: Boolean = false,
)

class SecretService(
    private val client: ModalClient,
) {
    suspend fun fromName(
        name: String,
        params: SecretFromNameParams = SecretFromNameParams(),
    ): Secret {
        try {
            val response = client.cpClient.secretGetOrCreate(
                SecretGetOrCreateRequest.newBuilder()
                    .setDeploymentName(name)
                    .setEnvironmentName(client.environmentName(params.environment))
                    .addAllRequiredKeys(params.requiredKeys)
                    .build(),
            )
            client.logger.debug("Retrieved Secret", "secret_id", response.secretId, "secret_name", name)
            return Secret(response.secretId, name)
        } catch (error: Throwable) {
            if (statusCode(error) == Status.Code.NOT_FOUND) {
                throw NotFoundError(statusMessage(error))
            }
            if (statusCode(error) == Status.Code.FAILED_PRECONDITION &&
                statusMessage(error).contains("Secret is missing key")
            ) {
                throw NotFoundError(statusMessage(error))
            }
            throw error
        }
    }

    suspend fun fromObject(
        entries: Map<String, *>,
        params: SecretFromObjectParams = SecretFromObjectParams(),
    ): Secret {
        val validatedEntries = linkedMapOf<String, String>()
        for ((key, value) in entries) {
            if (value !is String) {
                throw InvalidError(
                    "entries must be an object mapping string keys to string values, but got:\n$entries",
                )
            }
            validatedEntries[key] = value
        }

        val response = client.cpClient.secretGetOrCreate(
            SecretGetOrCreateRequest.newBuilder()
                .setObjectCreationType(ObjectCreationType.OBJECT_CREATION_TYPE_EPHEMERAL)
                .putAllEnvDict(validatedEntries)
                .setEnvironmentName(client.environmentName(params.environment))
                .build(),
        )
        client.logger.debug("Created ephemeral Secret", "secret_id", response.secretId)
        return Secret(response.secretId)
    }

    suspend fun delete(
        name: String,
        params: SecretDeleteParams = SecretDeleteParams(),
    ) {
        try {
            val secret = fromName(
                name,
                SecretFromNameParams(environment = params.environment),
            )
            client.cpClient.secretDelete(
                SecretDeleteRequest.newBuilder()
                    .setSecretId(secret.secretId)
                    .build(),
            )
            client.logger.debug("Deleted Secret", "secret_name", name, "secret_id", secret.secretId)
        } catch (error: Throwable) {
            val isMissing = error is NotFoundError || statusCode(error) == Status.Code.NOT_FOUND
            if (isMissing && params.allowMissing) {
                return
            }
            throw error
        }
    }
}

suspend fun mergeEnvIntoSecrets(
    client: ModalClient,
    env: Map<String, String>? = null,
    secrets: List<Secret>? = null,
): List<Secret> {
    val result = secrets?.toMutableList() ?: mutableListOf()
    if (env != null && env.isNotEmpty()) {
        result += client.secrets.fromObject(env)
    }
    return result
}
