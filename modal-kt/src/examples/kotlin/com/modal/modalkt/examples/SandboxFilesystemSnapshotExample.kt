package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ModalClient

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val image = modal.images.fromRegistry("alpine:3.21")

    val sandbox = modal.sandboxes.create(app, image)
    sandbox.exec(listOf("mkdir", "-p", "/app/data")).wait()
    sandbox.exec(listOf("sh", "-c", "echo 'This file was created in the first Sandbox' > /app/data/info.txt")).wait()

    val snapshotImage = sandbox.snapshotFilesystem()
    println(snapshotImage.imageId)
    sandbox.terminate()

    val sandbox2 = modal.sandboxes.create(app, snapshotImage)
    val proc = sandbox2.exec(listOf("cat", "/app/data/info.txt"))
    println(proc.stdout.readText())
    sandbox2.terminate()
}
