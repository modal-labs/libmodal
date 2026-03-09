package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ModalClient
import com.modal.modalkt.SandboxCreateConnectTokenParams
import java.net.URI
import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.http.HttpResponse

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val image = modal.images.fromRegistry("python:3.12-alpine")
    val sandbox = modal.sandboxes.create(app, image)
    val creds = sandbox.createConnectToken(SandboxCreateConnectTokenParams("abc"))

    val request = HttpRequest.newBuilder(URI(creds.url))
        .header("Authorization", "Bearer ${creds.token}")
        .GET()
        .build()
    val response = HttpClient.newHttpClient().send(request, HttpResponse.BodyHandlers.ofString())
    println(response.statusCode())
    println(response.body())
    sandbox.terminate()
}
