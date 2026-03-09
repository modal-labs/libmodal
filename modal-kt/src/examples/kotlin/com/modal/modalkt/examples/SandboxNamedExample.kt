package com.modal.modalkt.examples

import com.modal.modalkt.AlreadyExistsError
import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ModalClient
import com.modal.modalkt.SandboxCreateParams

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val image = modal.images.fromRegistry("alpine:3.21")
    val sandboxName = "libmodal-example-named-sandbox"

    val sandbox = modal.sandboxes.create(
        app,
        image,
        SandboxCreateParams(name = sandboxName, command = listOf("cat")),
    )
    println("Created Sandbox with name: $sandboxName")

    try {
        modal.sandboxes.create(
            app,
            image,
            SandboxCreateParams(name = sandboxName, command = listOf("cat")),
        )
    } catch (error: AlreadyExistsError) {
        println(error.message)
    }

    val sandboxFromName = modal.sandboxes.fromName("libmodal-example", sandboxName)
    sandboxFromName.stdin.writeText("hello, named Sandbox")
    sandboxFromName.stdin.close()
    println(sandboxFromName.stdout.readText())
    sandbox.terminate()
}
