package com.modal.modalkt.examples

import com.modal.modalkt.ClsWithConcurrencyParams
import com.modal.modalkt.ClsWithOptionsParams
import com.modal.modalkt.ModalClient

suspend fun main() {
    val modal = ModalClient()
    val cls = modal.cls.fromName("libmodal-test-support", "EchoClsParametrized")
    val instance = cls.instance()
    val method = instance.method("echo_env_var")

    val secret = modal.secrets.fromObject(mapOf("SECRET_MESSAGE" to "hello, Secret"))
    val instanceWithOptions = cls
        .withOptions(ClsWithOptionsParams(secrets = listOf(secret)))
        .withConcurrency(ClsWithConcurrencyParams(maxInputs = 1))
        .instance()
    val methodWithOptions = instanceWithOptions.method("echo_env_var")

    println(method.remote(listOf("SECRET_MESSAGE")))
    println(methodWithOptions.remote(listOf("SECRET_MESSAGE")))
}
