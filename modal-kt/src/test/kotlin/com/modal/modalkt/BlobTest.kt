package com.modal.modalkt

import com.sun.net.httpserver.HttpServer
import java.net.InetSocketAddress
import java.util.concurrent.atomic.AtomicReference
import kotlin.test.Test
import kotlin.test.assertContentEquals
import kotlin.test.assertEquals
import kotlinx.coroutines.runBlocking
import modal.client.*

class BlobTest {
    @Test
    fun blobUploadUsesProvidedUrl() = runBlocking {
        val received = AtomicReference<ByteArray>()
        val server = HttpServer.create(InetSocketAddress(0), 0)
        server.createContext("/upload") { exchange ->
            received.set(exchange.requestBody.readBytes())
            exchange.sendResponseHeaders(200, -1)
            exchange.close()
        }
        server.start()
        try {
            val (client, mock) = createMockModalClients()
            val uploadUrl = "http://127.0.0.1:${server.address.port}/upload"
            mock.handleUnary("/BlobCreate") { request ->
                request as BlobCreateRequest
                BlobCreateResponse.newBuilder()
                    .setBlobId("bl-123")
                    .setUploadUrl(uploadUrl)
                    .build()
            }

            val data = ByteArray(32) { 7 }
            val blobId = blobUpload(client.cpClient, data)
            assertEquals("bl-123", blobId)
            assertContentEquals(data, received.get())
        } finally {
            server.stop(0)
        }
    }

    @Test
    fun blobDownloadFetchesBytes() = runBlocking {
        val data = "hello".toByteArray()
        val server = HttpServer.create(InetSocketAddress(0), 0)
        server.createContext("/download") { exchange ->
            exchange.sendResponseHeaders(200, data.size.toLong())
            exchange.responseBody.use { it.write(data) }
        }
        server.start()
        try {
            val (client, mock) = createMockModalClients()
            val downloadUrl = "http://127.0.0.1:${server.address.port}/download"
            mock.handleUnary("/BlobGet") {
                BlobGetResponse.newBuilder()
                    .setDownloadUrl(downloadUrl)
                    .build()
            }

            val downloaded = blobDownload(client.cpClient, "bl-123")
            assertContentEquals(data, downloaded)
        } finally {
            server.stop(0)
        }
    }
}
