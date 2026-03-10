package com.modal.modalkt

import io.grpc.Status
import modal.client.*

data class VolumeFromNameParams(
    val environment: String? = null,
    val createIfMissing: Boolean = false,
)

data class VolumeEphemeralParams(
    val environment: String? = null,
)

data class VolumeDeleteParams(
    val environment: String? = null,
    val allowMissing: Boolean = false,
)

class VolumeService(
    private val client: ModalClient,
) {
    suspend fun fromName(
        name: String,
        params: VolumeFromNameParams = VolumeFromNameParams(),
    ): Volume {
        try {
            val response = client.cpClient.volumeGetOrCreate(
                VolumeGetOrCreateRequest.newBuilder()
                    .setDeploymentName(name)
                    .setEnvironmentName(client.environmentName(params.environment))
                    .setObjectCreationType(
                        if (params.createIfMissing) {
                            ObjectCreationType.OBJECT_CREATION_TYPE_CREATE_IF_MISSING
                        } else {
                            ObjectCreationType.OBJECT_CREATION_TYPE_UNSPECIFIED
                        },
                    )
                    .build(),
            )
            client.logger.debug("Retrieved Volume", "volume_id", response.volumeId, "volume_name", name)
            return Volume(response.volumeId, name)
        } catch (error: Throwable) {
            if (statusCode(error) == Status.Code.NOT_FOUND) {
                throw NotFoundError(statusMessage(error))
            }
            throw error
        }
    }

    suspend fun ephemeral(
        params: VolumeEphemeralParams = VolumeEphemeralParams(),
    ): Volume {
        val response = client.cpClient.volumeGetOrCreate(
            VolumeGetOrCreateRequest.newBuilder()
                .setObjectCreationType(ObjectCreationType.OBJECT_CREATION_TYPE_EPHEMERAL)
                .setEnvironmentName(client.environmentName(params.environment))
                .build(),
        )
        client.logger.debug("Created ephemeral Volume", "volume_id", response.volumeId)
        val heartbeatManager = EphemeralHeartbeatManager(
            heartbeatFn = {
                client.cpClient.volumeHeartbeat(
                    VolumeHeartbeatRequest.newBuilder()
                        .setVolumeId(response.volumeId)
                        .build(),
                )
            },
            sleepMs = client.ephemeralHeartbeatSleepMs,
            scope = client.backgroundScope,
        )
        return Volume(response.volumeId, null, false, heartbeatManager)
    }

    suspend fun delete(
        name: String,
        params: VolumeDeleteParams = VolumeDeleteParams(),
    ) {
        try {
            val volume = fromName(
                name,
                VolumeFromNameParams(environment = params.environment),
            )
            client.cpClient.volumeDelete(
                VolumeDeleteRequest.newBuilder()
                    .setVolumeId(volume.volumeId)
                    .build(),
            )
            client.logger.debug("Deleted Volume", "volume_name", name, "volume_id", volume.volumeId)
        } catch (error: Throwable) {
            val isMissing = error is NotFoundError || statusCode(error) == Status.Code.NOT_FOUND
            if (isMissing && params.allowMissing) {
                return
            }
            throw error
        }
    }
}
