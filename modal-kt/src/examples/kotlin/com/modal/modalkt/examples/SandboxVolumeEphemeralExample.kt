package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ModalClient
import com.modal.modalkt.SandboxCreateParams

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val image = modal.images.fromRegistry("alpine:3.21")
    val volume = modal.volumes.ephemeral()

    val writer = modal.sandboxes.create(
        app,
        image,
        SandboxCreateParams(
            command = listOf("sh", "-c", "echo 'Hello from writer Sandbox!' > /mnt/volume/message.txt"),
            volumes = mapOf("/mnt/volume" to volume),
        ),
    )
    writer.wait()
    writer.terminate()

    val reader = modal.sandboxes.create(
        app,
        image,
        SandboxCreateParams(
            command = listOf("cat", "/mnt/volume/message.txt"),
            volumes = mapOf("/mnt/volume" to volume.readOnly()),
        ),
    )
    println(reader.stdout.readText())
    reader.terminate()
    volume.closeEphemeral()
}
