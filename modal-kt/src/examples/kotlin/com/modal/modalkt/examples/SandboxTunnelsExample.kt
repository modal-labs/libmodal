package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ModalClient
import com.modal.modalkt.SandboxCreateParams
import java.net.URI
import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.http.HttpResponse

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val image = modal.images.fromRegistry("python:3.12-alpine")
    val sandbox = modal.sandboxes.create(
        app,
        image,
        SandboxCreateParams(
            command = listOf("python3", "-m", "http.server", "8000"),
            encryptedPorts = listOf(8000),
            timeoutMs = 60_000,
            idleTimeoutMs = 30_000,
        ),
    )
    val tunnel = sandbox.tunnels()[8000] ?: error("missing tunnel")
    val request = HttpRequest.newBuilder(URI(tunnel.url)).GET().build()
    val response = HttpClient.newHttpClient().send(request, HttpResponse.BodyHandlers.ofString())
    println(response.body().take(500))
    sandbox.terminate()
}
