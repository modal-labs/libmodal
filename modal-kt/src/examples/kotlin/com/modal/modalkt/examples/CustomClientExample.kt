package com.modal.modalkt.examples

import com.modal.modalkt.ModalClient
import com.modal.modalkt.ModalClientParams

suspend fun main() {
    val modalId = System.getenv("CUSTOM_MODAL_ID") ?: error("CUSTOM_MODAL_ID environment variable not set")
    val modalSecret = System.getenv("CUSTOM_MODAL_SECRET") ?: error("CUSTOM_MODAL_SECRET environment variable not set")

    val modal = ModalClient(
        ModalClientParams(
            tokenId = modalId,
            tokenSecret = modalSecret,
        ),
    )
    val echo = modal.functions.fromName("libmodal-test-support", "echo_string")
    println(echo)
}
