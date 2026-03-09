package com.modal.modalkt.examples

import com.modal.modalkt.ModalClient

suspend fun main() {
    val modal = ModalClient()
    val function = modal.functions.fromName("libmodal-test-support", "echo_string")
    val call = function.spawn(kwargs = mapOf("s" to "Hello world!"))
    println(call.get())
}
