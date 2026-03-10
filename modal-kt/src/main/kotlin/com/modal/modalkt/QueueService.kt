package com.modal.modalkt

import io.grpc.Status
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import modal.client.*
import okio.ByteString
import okio.ByteString.Companion.toByteString

private const val queueInitialPutBackoffMs: Long = 100
private const val queueDefaultPartitionTtlMs: Long = 24 * 3600 * 1000

data class QueueFromNameParams(
    val environment: String? = null,
    val createIfMissing: Boolean = false,
)

data class QueueDeleteParams(
    val environment: String? = null,
    val allowMissing: Boolean = false,
)

data class QueueEphemeralParams(
    val environment: String? = null,
)

data class QueueClearParams(
    val partition: String? = null,
    val all: Boolean = false,
)

data class QueueGetParams(
    val timeoutMs: Long? = null,
    val partition: String? = null,
)

data class QueuePutParams(
    val timeoutMs: Long? = null,
    val partition: String? = null,
    val partitionTtlMs: Long? = null,
)

data class QueueLenParams(
    val partition: String? = null,
    val total: Boolean = false,
)

data class QueueIterateParams(
    val itemPollTimeoutMs: Long = 0,
    val partition: String? = null,
)

class QueueService(
    private val client: ModalClient,
) {
    suspend fun ephemeral(
        params: QueueEphemeralParams = QueueEphemeralParams(),
    ): Queue {
        val response = client.cpClient.queueGetOrCreate(
            QueueGetOrCreateRequest.newBuilder()
                .setObjectCreationType(ObjectCreationType.OBJECT_CREATION_TYPE_EPHEMERAL)
                .setEnvironmentName(client.environmentName(params.environment))
                .build(),
        )
        client.logger.debug("Created ephemeral Queue", "queue_id", response.queueId)
        val heartbeatManager = EphemeralHeartbeatManager(
            heartbeatFn = {
                client.cpClient.queueHeartbeat(
                    QueueHeartbeatRequest.newBuilder()
                        .setQueueId(response.queueId)
                        .build(),
                )
            },
            sleepMs = client.ephemeralHeartbeatSleepMs,
            scope = client.backgroundScope,
        )
        return Queue(client, response.queueId, null, heartbeatManager)
    }

    suspend fun fromName(
        name: String,
        params: QueueFromNameParams = QueueFromNameParams(),
    ): Queue {
        validateQueueName(name)
        try {
            val response = client.cpClient.queueGetOrCreate(
                QueueGetOrCreateRequest.newBuilder()
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
            client.logger.debug("Retrieved Queue", "queue_id", response.queueId, "queue_name", name)
            return Queue(client, response.queueId, name)
        } catch (error: Throwable) {
            if (statusCode(error) == Status.Code.NOT_FOUND) {
                throw NotFoundError(statusMessage(error))
            }
            throw error
        }
    }

    suspend fun delete(
        name: String,
        params: QueueDeleteParams = QueueDeleteParams(),
    ) {
        try {
            val queue = fromName(
                name,
                QueueFromNameParams(environment = params.environment),
            )
            client.cpClient.queueDelete(
                QueueDeleteRequest.newBuilder()
                    .setQueueId(queue.queueId)
                    .build(),
            )
            client.logger.debug("Deleted Queue", "queue_name", name, "queue_id", queue.queueId)
        } catch (error: Throwable) {
            val isMissing = error is NotFoundError || statusCode(error) == Status.Code.NOT_FOUND
            if (isMissing && params.allowMissing) {
                return
            }
            throw error
        }
    }

    private fun validateQueueName(name: String) {
        if (name.contains(" ") || name.contains("/") || name.length > 64) {
            throw InvalidError("Invalid queue name: $name")
        }
    }
}

class Queue(
    private val client: ModalClient,
    val queueId: String,
    val name: String? = null,
    private val ephemeralHeartbeatManager: EphemeralHeartbeatManager? = null,
) {
    suspend fun clear(params: QueueClearParams = QueueClearParams()) {
        if (params.partition != null && params.all) {
            throw InvalidError("Partition must be null when requesting to clear all.")
        }
        client.cpClient.queueClear(
            QueueClearRequest.newBuilder()
                .setQueueId(queueId)
                .setPartitionKey(validatePartitionKey(params.partition))
                .setAllPartitions(params.all)
                .build(),
        )
    }

    suspend fun get(params: QueueGetParams = QueueGetParams()): Any? {
        return getMany(1, params).first()
    }

    suspend fun getMany(
        n: Int,
        params: QueueGetParams = QueueGetParams(),
    ): List<Any?> {
        val partitionKey = validatePartitionKey(params.partition)
        val start = System.currentTimeMillis()
        var pollTimeoutMs = 50_000L
        if (params.timeoutMs != null) {
            pollTimeoutMs = minOf(pollTimeoutMs, params.timeoutMs)
        }

        while (true) {
            val response = client.cpClient.queueGet(
                QueueGetRequest.newBuilder()
                    .setQueueId(queueId)
                    .setPartitionKey(partitionKey)
                    .setTimeout(pollTimeoutMs.toFloat() / 1000f)
                    .setNValues(n)
                    .build(),
            )
            if (response.valuesCount > 0) {
                return response.valuesList.map { PickleCodec.decode(it.toByteArray()) }
            }
            if (params.timeoutMs != null) {
                val remaining = params.timeoutMs - (System.currentTimeMillis() - start)
                if (remaining <= 0) {
                    throw QueueEmptyError("Queue $queueId did not return values within ${params.timeoutMs}ms.")
                }
                pollTimeoutMs = minOf(pollTimeoutMs, remaining)
            }
        }
    }

    suspend fun put(
        value: Any?,
        params: QueuePutParams = QueuePutParams(),
    ) {
        putMany(listOf(value), params)
    }

    suspend fun putMany(
        values: List<Any?>,
        params: QueuePutParams = QueuePutParams(),
    ) {
        val encoded = values.map { PickleCodec.encode(it).toByteString() }
        val partitionKey = validatePartitionKey(params.partition)
        var delayMs = queueInitialPutBackoffMs
        val deadline = params.timeoutMs?.let { System.currentTimeMillis() + it }

        while (true) {
            try {
                client.cpClient.queuePut(
                    QueuePutRequest.newBuilder()
                        .setQueueId(queueId)
                        .addAllValues(encoded)
                        .setPartitionKey(partitionKey)
                        .setPartitionTtlSeconds(((params.partitionTtlMs ?: queueDefaultPartitionTtlMs) / 1000).toInt())
                        .build(),
                )
                return
            } catch (error: Throwable) {
                if (statusCode(error) == Status.Code.RESOURCE_EXHAUSTED) {
                    delayMs = minOf(delayMs * 2, 30_000)
                    if (deadline != null) {
                        val remaining = deadline - System.currentTimeMillis()
                        if (remaining <= 0) {
                            throw QueueFullError("Put failed on $queueId.")
                        }
                        delayMs = minOf(delayMs, remaining)
                    }
                    delay(delayMs)
                    continue
                }
                throw error
            }
        }
    }

    suspend fun len(params: QueueLenParams = QueueLenParams()): Int {
        if (params.partition != null && params.total) {
            throw InvalidError("Partition must be null when requesting total length.")
        }
        val response = client.cpClient.queueLen(
            QueueLenRequest.newBuilder()
                .setQueueId(queueId)
                .setPartitionKey(validatePartitionKey(params.partition))
                .setTotal(params.total)
                .build(),
        )
        return response.len
    }

    fun iterate(params: QueueIterateParams = QueueIterateParams()): Flow<Any?> = flow {
        var lastEntryId = ""
        val partitionKey = validatePartitionKey(params.partition)
        val maxPollDurationMs = 30_000L
        var fetchDeadline = System.currentTimeMillis() + params.itemPollTimeoutMs

        while (true) {
            val pollDurationMs = maxOf(0L, minOf(maxPollDurationMs, fetchDeadline - System.currentTimeMillis()))
            val response = client.cpClient.queueNextItems(
                QueueNextItemsRequest.newBuilder()
                    .setQueueId(queueId)
                    .setPartitionKey(partitionKey)
                    .setItemPollTimeout(pollDurationMs.toFloat() / 1000f)
                    .setLastEntryId(lastEntryId)
                    .build(),
            )

            if (response.itemsCount > 0) {
                for (item in response.itemsList) {
                    emit(PickleCodec.decode(item.value.toByteArray()))
                    lastEntryId = item.entryId
                }
                fetchDeadline = System.currentTimeMillis() + params.itemPollTimeoutMs
            } else if (System.currentTimeMillis() >= fetchDeadline) {
                break
            }
        }
    }

    fun closeEphemeral() {
        if (ephemeralHeartbeatManager == null) {
            throw InvalidError("Queue is not ephemeral.")
        }
        ephemeralHeartbeatManager.stop()
    }

    private fun validatePartitionKey(partition: String?): ByteString {
        if (partition == null) {
            return ByteString.EMPTY
        }
        val bytes = partition.toByteArray()
        if (bytes.isEmpty() || bytes.size > 64) {
            throw InvalidError("Queue partition key must be between 1 and 64 bytes.")
        }
        return bytes.toByteString()
    }
}
