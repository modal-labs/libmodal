package com.modal.modalkt.examples

import com.modal.modalkt.ModalClient

suspend fun main() {
    val modal = ModalClient()
    val function = modal.functions.fromName("libmodal-test-support", "echo_string")
    val stats = function.getCurrentStats()
    println("Function Statistics:")
    println("  Backlog: ${stats.backlog} inputs")
    println("  Total Runners: ${stats.numTotalRunners} containers")
}
