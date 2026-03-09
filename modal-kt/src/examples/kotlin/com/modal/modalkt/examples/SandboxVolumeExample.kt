package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ModalClient
import com.modal.modalkt.SandboxCreateParams
import com.modal.modalkt.VolumeFromNameParams

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val image = modal.images.fromRegistry("alpine:3.21")
    val volume = modal.volumes.fromName("libmodal-example-volume", VolumeFromNameParams(createIfMissing = true))

    val writer = modal.sandboxes.create(
        app,
        image,
        SandboxCreateParams(
            command = listOf("sh", "-c", "echo 'Hello from writer Sandbox!' > /mnt/volume/message.txt"),
            volumes = mapOf("/mnt/volume" to volume),
        ),
    )
    writer.wait()

    val reader = modal.sandboxes.create(
        app,
        image,
        SandboxCreateParams(volumes = mapOf("/mnt/volume" to volume.readOnly())),
    )
    val process = reader.exec(listOf("cat", "/mnt/volume/message.txt"))
    println(process.stdout.readText())
    writer.terminate()
    reader.terminate()
}
