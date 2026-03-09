package com.modal.modalkt.examples

import com.modal.modalkt.ModalClient

suspend fun main() {
    val modal = ModalClient()
    val cls = modal.cls.fromName("libmodal-test-support", "EchoCls")
    val instance = cls.instance()
    val method = instance.method("echo_string")

    println(method.remote(listOf("Hello world!")))
    println(method.remote(kwargs = mapOf("s" to "Hello world!")))
}
