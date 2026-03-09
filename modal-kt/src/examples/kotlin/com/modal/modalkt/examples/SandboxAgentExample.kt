package com.modal.modalkt.examples

import com.modal.modalkt.AppFromNameParams
import com.modal.modalkt.ImageDockerfileCommandsParams
import com.modal.modalkt.ModalClient
import com.modal.modalkt.SandboxExecParams

suspend fun main() {
    val modal = ModalClient()
    val app = modal.apps.fromName("libmodal-example", AppFromNameParams(createIfMissing = true))
    val image = modal.images
        .fromRegistry("alpine:3.21")
        .dockerfileCommands(
            listOf(
                "RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep",
                "RUN curl -fsSL https://claude.ai/install.sh | bash",
                "ENV PATH=/root/.local/bin:\$PATH USE_BUILTIN_RIPGREP=0",
            ),
        )

    val sandbox = modal.sandboxes.create(app, image)
    try {
        sandbox.exec(listOf("git", "clone", "https://github.com/modal-labs/libmodal", "/repo")).wait()
        val secret = modal.secrets.fromName(
            "libmodal-anthropic-secret",
            com.modal.modalkt.SecretFromNameParams(requiredKeys = listOf("ANTHROPIC_API_KEY")),
        )
        val process = sandbox.exec(
            listOf(
                "claude",
                "-p",
                "Summarize what this repository is about. Don't modify any code or files.",
            ),
            SandboxExecParams(
                pty = true,
                workdir = "/repo",
                secrets = listOf(secret),
            ),
        )
        process.wait()
        println(process.stdout.readText())
    } finally {
        sandbox.terminate()
    }
}
