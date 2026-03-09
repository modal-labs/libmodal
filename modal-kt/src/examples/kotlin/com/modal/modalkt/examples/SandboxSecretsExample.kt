package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ModalClient
import com.modal.modalkt.SandboxCreateParams
import com.modal.modalkt.SecretFromNameParams

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val image = modal.images.fromRegistry("alpine:3.21")
    val secret = modal.secrets.fromName("libmodal-test-secret", SecretFromNameParams(requiredKeys = listOf("c")))
    val ephemeralSecret = modal.secrets.fromObject(mapOf("d" to "123"))
    val sandbox = modal.sandboxes.create(
        app,
        image,
        SandboxCreateParams(
            command = listOf("sh", "-lc", "printenv | grep -E '^c|d='"),
            secrets = listOf(secret, ephemeralSecret),
        ),
    )
    println(sandbox.stdout.readText())
    sandbox.terminate()
}
