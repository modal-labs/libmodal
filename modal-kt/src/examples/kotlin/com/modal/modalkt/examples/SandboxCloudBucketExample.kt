package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.CloudBucketMountCreateParams
import com.modal.modalkt.ModalClient
import com.modal.modalkt.SandboxCreateParams

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val image = modal.images.fromRegistry("alpine:3.21")
    val secret = modal.secrets.fromName("libmodal-aws-bucket-secret")

    val sandbox = modal.sandboxes.create(
        app,
        image,
        SandboxCreateParams(
            command = listOf("sh", "-c", "ls -la /mnt/s3-bucket"),
            cloudBucketMounts = mapOf(
                "/mnt/s3-bucket" to modal.cloudBucketMounts.create(
                    "my-s3-bucket",
                    CloudBucketMountCreateParams(
                        secret = secret,
                        keyPrefix = "data/",
                        readOnly = true,
                    ),
                ),
            ),
        ),
    )
    println(sandbox.stdout.readText())
    sandbox.terminate()
}
