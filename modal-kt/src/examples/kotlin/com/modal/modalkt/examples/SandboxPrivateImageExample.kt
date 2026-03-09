package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ModalClient
import com.modal.modalkt.SandboxCreateParams
import com.modal.modalkt.SecretFromNameParams

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val secret = modal.secrets.fromName(
        "libmodal-aws-ecr-test",
        SecretFromNameParams(requiredKeys = listOf("AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY")),
    )
    val image = modal.images.fromAwsEcr(
        "459781239556.dkr.ecr.us-east-1.amazonaws.com/ecr-private-registry-test-7522615:python",
        secret,
    )
    val sandbox = modal.sandboxes.create(
        app,
        image,
        SandboxCreateParams(command = listOf("python", "-c", "import sys; sys.stdout.write(sys.stdin.read())")),
    )
    sandbox.stdin.writeText("this is input that should be mirrored by the Python one-liner")
    sandbox.stdin.close()
    println(sandbox.stdout.readText())
    sandbox.terminate()
}
