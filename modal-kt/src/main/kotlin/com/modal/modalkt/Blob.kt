package com.modal.modalkt

import java.net.URI
import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.http.HttpResponse
import java.security.MessageDigest
import java.util.Base64
import modal.client.*

private val httpClient: HttpClient = HttpClient.newHttpClient()

suspend fun blobUpload(
    client: ControlPlaneClient,
    data: ByteArray,
): String {
    val contentMd5 = base64Digest("MD5", data)
    val contentSha256 = base64Digest("SHA-256", data)

    val response = client.blobCreate(
        BlobCreateRequest.newBuilder()
            .setContentMd5(contentMd5)
            .setContentSha256Base64(contentSha256)
            .setContentLength(data.size.toLong())
            .build(),
    )

    if (response.hasMultipart()) {
        throw InvalidError("Function input size exceeds multipart upload threshold, unsupported by this SDK version")
    }
    if (!response.hasUploadUrl()) {
        throw InvalidError("Missing upload URL in BlobCreate response")
    }

    val request = HttpRequest.newBuilder(URI(response.uploadUrl))
        .PUT(HttpRequest.BodyPublishers.ofByteArray(data))
        .header("Content-Type", "application/octet-stream")
        .header("Content-MD5", contentMd5)
        .build()
    val uploadResponse = httpClient.send(request, HttpResponse.BodyHandlers.discarding())
    if (uploadResponse.statusCode() !in 200..299) {
        throw InvalidError("Failed blob upload: ${uploadResponse.statusCode()}")
    }
    return response.blobId
}

suspend fun blobDownload(
    client: ControlPlaneClient,
    blobId: String,
): ByteArray {
    val response = client.blobGet(
        BlobGetRequest.newBuilder()
            .setBlobId(blobId)
            .build(),
    )
    val request = HttpRequest.newBuilder(URI(response.downloadUrl))
        .GET()
        .build()
    val downloadResponse = httpClient.send(request, HttpResponse.BodyHandlers.ofByteArray())
    if (downloadResponse.statusCode() !in 200..299) {
        throw InvalidError("Failed to download blob: ${downloadResponse.statusCode()}")
    }
    return downloadResponse.body()
}

private fun base64Digest(algorithm: String, data: ByteArray): String {
    return Base64.getEncoder().encodeToString(
        MessageDigest.getInstance(algorithm).digest(data),
    )
}
