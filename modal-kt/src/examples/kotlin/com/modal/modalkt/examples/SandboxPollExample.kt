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
        SandboxCreateParams(command = listOf("sh", "-c", "read line; exit 42")),
    )

    println(sandbox.poll())
    sandbox.stdin.writeText("hello, goodbye\n")
    sandbox.stdin.close()
    println(sandbox.wait())
    println(sandbox.poll())
}
