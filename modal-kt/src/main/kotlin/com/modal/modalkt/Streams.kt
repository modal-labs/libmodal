package com.modal.modalkt

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.collect
import java.io.ByteArrayOutputStream

class ModalReadStream(
    private val supplier: suspend () -> Flow<ByteArray>,
) {
    private var consumed = false

    suspend fun readText(): String {
        if (consumed) {
            return ""
        }
        consumed = true
        return readBytesInternal().toString(Charsets.UTF_8)
    }

    suspend fun readBytes(): ByteArray {
        if (consumed) {
            return ByteArray(0)
        }
        consumed = true
        return readBytesInternal()
    }

    fun asFlow(): Flow<ByteArray> {
        return kotlinx.coroutines.flow.flow {
            if (consumed) {
                return@flow
            }
            consumed = true
            supplier().collect { emit(it) }
        }
    }

    private suspend fun readBytesInternal(): ByteArray {
        val output = ByteArrayOutputStream()
        supplier().collect { chunk ->
            output.write(chunk)
        }
        return output.toByteArray()
    }
}

class ModalWriteStream(
    private val writeBlock: suspend (ByteArray) -> Unit,
    private val closeBlock: suspend () -> Unit,
) {
    suspend fun writeText(text: String) {
        writeBlock(text.toByteArray())
    }

    suspend fun writeBytes(bytes: ByteArray) {
        writeBlock(bytes)
    }

    suspend fun close() {
        closeBlock()
    }
}
