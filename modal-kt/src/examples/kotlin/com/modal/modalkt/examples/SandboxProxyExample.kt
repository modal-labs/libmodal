package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ModalClient
import com.modal.modalkt.ProxyFromNameParams
import com.modal.modalkt.SandboxCreateParams

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val image = modal.images.fromRegistry("alpine/curl:8.14.1")
    val proxy = modal.proxies.fromName("libmodal-test-proxy", ProxyFromNameParams(environment = "libmodal"))
    val sandbox = modal.sandboxes.create(app, image, SandboxCreateParams(proxy = proxy))
    val process = sandbox.exec(listOf("curl", "-s", "ifconfig.me"))
    println(process.stdout.readText())
    sandbox.terminate()
}
