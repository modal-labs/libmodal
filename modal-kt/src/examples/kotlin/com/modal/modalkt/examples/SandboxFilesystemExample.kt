package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ModalClient

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val image = modal.images.fromRegistry("alpine:3.21")
    val sandbox = modal.sandboxes.create(app, image)

    val writeHandle = sandbox.open("/tmp/example.txt", "w")
    writeHandle.write("Hello, Modal filesystem!\n".toByteArray())
    writeHandle.close()

    val readHandle = sandbox.open("/tmp/example.txt", "r")
    println(String(readHandle.read()))
    readHandle.close()

    sandbox.terminate()
}
