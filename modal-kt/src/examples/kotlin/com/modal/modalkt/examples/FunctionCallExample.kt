package com.modal.modalkt.examples

import com.modal.modalkt.ModalClient

suspend fun main() {
    val modal = ModalClient()
    val function = modal.functions.fromName("libmodal-test-support", "echo_string")

    val resultArgs = function.remote(listOf("Hello world!"))
    println(resultArgs)

    val resultKwargs = function.remote(kwargs = mapOf("s" to "Hello world!"))
    println(resultKwargs)
}
