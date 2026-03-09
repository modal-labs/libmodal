package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ModalClient
import com.modal.modalkt.SandboxCreateParams

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val image = modal.images.fromRegistry("alpine:3.21")

    val sandbox = modal.sandboxes.create(
        app,
        image,
        SandboxCreateParams(command = listOf("cat")),
    )
    println("Sandbox: ${sandbox.sandboxId}")

    val sandboxFromId = modal.sandboxes.fromId(sandbox.sandboxId)
    println("Queried Sandbox from ID: ${sandboxFromId.sandboxId}")

    sandbox.stdin.writeText("this is input that should be mirrored by cat")
    sandbox.stdin.close()
    println("output: ${sandbox.stdout.readText()}")

    sandbox.terminate()
}
