package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ModalClient

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val image = modal.images.fromRegistry("python:3.13-slim")
    val sandbox = modal.sandboxes.create(app, image)

    val process = sandbox.exec(
        listOf(
            "python",
            "-c",
            """
import time
import sys
for i in range(10):
    print(i)
    print(i, file=sys.stderr)
""".trimIndent(),
        ),
    )
    println(process.stdout.readText())
    println(process.stderr.readText())
    println(process.wait())

    sandbox.terminate()
}
