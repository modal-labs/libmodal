package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ImageDockerfileCommandsParams
import com.modal.modalkt.ModalClient
import com.modal.modalkt.SandboxCreateParams

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val secret = modal.secrets.fromObject(mapOf("CURL_VERSION" to "8.12.1-r1"))

    val image = modal.images
        .fromRegistry("alpine:3.21")
        .dockerfileCommands(
            listOf("RUN apk add --no-cache curl=\$CURL_VERSION"),
            ImageDockerfileCommandsParams(secrets = listOf(secret)),
        )
        .dockerfileCommands(listOf("ENV SERVER=ipconfig.me"))

    val sandbox = modal.sandboxes.create(
        app,
        image,
        SandboxCreateParams(command = listOf("sh", "-c", "curl -Ls \$SERVER")),
    )
    println("Created Sandbox with ID: ${sandbox.sandboxId}")
    println("Sandbox output: ${sandbox.stdout.readText()}")
    sandbox.terminate()
}
