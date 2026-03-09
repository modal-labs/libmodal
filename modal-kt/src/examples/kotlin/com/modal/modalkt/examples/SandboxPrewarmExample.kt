package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ModalClient
import com.modal.modalkt.SandboxCreateParams

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val image = modal.images.fromRegistry("alpine:3.21").build(app)
    println(image.imageId)
    val image2 = modal.images.fromId(image.imageId)
    val sandbox = modal.sandboxes.create(app, image2, SandboxCreateParams(command = listOf("cat")))
    println(sandbox.sandboxId)
    sandbox.terminate()
}
