package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ModalClient
import com.modal.modalkt.SandboxCreateParams

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val image = modal.images.fromRegistry("nvidia/cuda:12.4.0-devel-ubuntu22.04")
    val sandbox = modal.sandboxes.create(app, image, SandboxCreateParams(gpu = "A10G"))
    try {
        val process = sandbox.exec(listOf("nvidia-smi"))
        println(process.stdout.readText())
    } finally {
        sandbox.terminate()
    }
}
