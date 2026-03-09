package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ModalClient

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val baseImage = modal.images.fromRegistry("alpine:3.21").dockerfileCommands(listOf("RUN apk add --no-cache git"))

    val sandbox = modal.sandboxes.create(app, baseImage)
    sandbox.exec(listOf("git", "clone", "https://github.com/modal-labs/libmodal.git", "/repo")).wait()
    val snapshot = sandbox.snapshotDirectory("/repo")
    println(snapshot.imageId)
    sandbox.terminate()

    val sandbox2 = modal.sandboxes.create(app, baseImage)
    sandbox2.exec(listOf("mkdir", "-p", "/repo")).wait()
    sandbox2.mountImage("/repo", snapshot)
    val proc = sandbox2.exec(listOf("ls", "/repo"))
    println(proc.stdout.readText())
    sandbox2.terminate()
}
