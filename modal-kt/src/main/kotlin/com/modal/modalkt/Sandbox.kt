package com.modal.modalkt

import modal.client.*
import modal.task_command_router.*

enum class StdioBehavior {
    PIPE,
    IGNORE,
}

enum class StreamMode {
    TEXT,
    BINARY,
}

data class SandboxCreateParams(
    val cpu: Double? = null,
    val cpuLimit: Double? = null,
    val memoryMiB: Int? = null,
    val memoryLimitMiB: Int? = null,
    val gpu: String? = null,
    val timeoutMs: Long? = null,
    val idleTimeoutMs: Long? = null,
    val workdir: String? = null,
    val command: List<String>? = null,
    val env: Map<String, String>? = null,
    val secrets: List<Secret> = emptyList(),
    val volumes: Map<String, Volume> = emptyMap(),
    val cloudBucketMounts: Map<String, CloudBucketMount> = emptyMap(),
    val pty: Boolean = false,
    val encryptedPorts: List<Int> = emptyList(),
    val h2Ports: List<Int> = emptyList(),
    val unencryptedPorts: List<Int> = emptyList(),
    val blockNetwork: Boolean = false,
    val cidrAllowlist: List<String>? = null,
    val cloud: String? = null,
    val regions: List<String> = emptyList(),
    val verbose: Boolean = false,
    val proxy: Proxy? = null,
    val name: String? = null,
    val experimentalOptions: Map<String, Any?> = emptyMap(),
    val customDomain: String? = null,
    val memory: Int? = null,
    val memoryLimit: Int? = null,
    val timeout: Long? = null,
    val idleTimeout: Long? = null,
)

data class SandboxExecParams(
    val mode: StreamMode = StreamMode.TEXT,
    val stdout: StdioBehavior = StdioBehavior.PIPE,
    val stderr: StdioBehavior = StdioBehavior.PIPE,
    val workdir: String? = null,
    val timeoutMs: Long? = null,
    val env: Map<String, String>? = null,
    val secrets: List<Secret> = emptyList(),
    val pty: Boolean = false,
    val timeout: Long? = null,
)

fun defaultSandboxPtyInfo(): PTYInfo {
    return PTYInfo.newBuilder()
        .setEnabled(true)
        .setWinszRows(24)
        .setWinszCols(80)
        .setEnvTerm("xterm-256color")
        .setEnvColorterm("truecolor")
        .setEnvTermProgram("")
        .setPtyType(PTYType.PTY_TYPE_SHELL)
        .setNoTerminateOnIdleStdin(true)
        .build()
}

fun validateExecArgs(args: List<String>) {
    val argMaxBytes = 1 shl 16
    var totalLength = 0
    for (arg in args) {
        totalLength += arg.length
    }
    if (totalLength > argMaxBytes) {
        throw InvalidError(
            "Total length of CMD arguments must be less than $argMaxBytes bytes (ARG_MAX). Got $totalLength bytes.",
        )
    }
}

fun buildSandboxCreateRequestProto(
    appId: String,
    imageId: String,
    params: SandboxCreateParams = SandboxCreateParams(),
): SandboxCreateRequest {
    checkForRenamedParams(
        mapOf(
            "memory" to params.memory,
            "memoryLimit" to params.memoryLimit,
            "timeout" to params.timeout,
            "idleTimeout" to params.idleTimeout,
        ).filterValues { it != null },
        mapOf(
            "memory" to "memoryMiB",
            "memoryLimit" to "memoryLimitMiB",
            "timeout" to "timeoutMs",
            "idleTimeout" to "idleTimeoutMs",
        ),
    )

    val timeoutMs = params.timeoutMs
    if (timeoutMs != null) {
        if (timeoutMs <= 0) {
            throw InvalidError("timeoutMs must be positive, got $timeoutMs")
        }
        if (timeoutMs % 1000 != 0L) {
            throw InvalidError("timeoutMs must be a multiple of 1000ms, got $timeoutMs")
        }
    }

    val idleTimeoutMs = params.idleTimeoutMs
    if (idleTimeoutMs != null) {
        if (idleTimeoutMs <= 0) {
            throw InvalidError("idleTimeoutMs must be positive, got $idleTimeoutMs")
        }
        if (idleTimeoutMs % 1000 != 0L) {
            throw InvalidError("idleTimeoutMs must be a multiple of 1000ms, got $idleTimeoutMs")
        }
    }

    val workdir = params.workdir
    if (workdir != null && !workdir.startsWith("/")) {
        throw InvalidError("workdir must be an absolute path, got: $workdir")
    }

    val gpuConfig = parseGpuConfig(params.gpu)
    val resourcesBuilder = Resources.newBuilder()
        .setGpuConfig(gpuConfig)

    val cpu = params.cpu
    val cpuLimit = params.cpuLimit
    if (cpu == null && cpuLimit != null) {
        throw InvalidError("must also specify cpu when cpuLimit is specified")
    }
    if (cpu != null) {
        if (cpu <= 0.0) {
            throw InvalidError("cpu ($cpu) must be a positive number")
        }
        resourcesBuilder.milliCpu = (cpu * 1000).toInt()
        if (cpuLimit != null) {
            if (cpuLimit < cpu) {
                throw InvalidError("cpu ($cpu) cannot be higher than cpuLimit ($cpuLimit)")
            }
            resourcesBuilder.milliCpuMax = (cpuLimit * 1000).toInt()
        }
    }

    val memoryMiB = params.memoryMiB
    val memoryLimitMiB = params.memoryLimitMiB
    if (memoryMiB == null && memoryLimitMiB != null) {
        throw InvalidError("must also specify memoryMiB when memoryLimitMiB is specified")
    }
    if (memoryMiB != null) {
        if (memoryMiB <= 0) {
            throw InvalidError("the memoryMiB request ($memoryMiB) must be a positive number")
        }
        resourcesBuilder.memoryMb = memoryMiB
        if (memoryLimitMiB != null) {
            if (memoryLimitMiB < memoryMiB) {
                throw InvalidError(
                    "the memoryMiB request ($memoryMiB) cannot be higher than memoryLimitMiB ($memoryLimitMiB)",
                )
            }
            resourcesBuilder.memoryMbMax = memoryLimitMiB
        }
    }

    val networkAccessBuilder = NetworkAccess.newBuilder()
    val cidrAllowlist = params.cidrAllowlist
    if (params.blockNetwork) {
        if (!cidrAllowlist.isNullOrEmpty()) {
            throw InvalidError("cidrAllowlist cannot be used when blockNetwork is enabled")
        }
        networkAccessBuilder.networkAccessType = NetworkAccessType.BLOCKED
    } else if (!cidrAllowlist.isNullOrEmpty()) {
        networkAccessBuilder.networkAccessType = NetworkAccessType.ALLOWLIST
        networkAccessBuilder.addAllAllowedCidrs(cidrAllowlist)
    } else {
        networkAccessBuilder.networkAccessType = NetworkAccessType.OPEN
    }

    val experimentalOptions = mutableMapOf<String, Boolean>()
    for ((name, value) in params.experimentalOptions) {
        if (value !is Boolean) {
            throw InvalidError("experimental option '$name' must be a boolean, got $value")
        }
        experimentalOptions[name] = value
    }

    val definitionBuilder = SandboxProto.newBuilder()
        .setImageId(imageId)
        .setTimeoutSecs((timeoutMs ?: 300_000L).div(1000).toInt())
        .setNetworkAccess(networkAccessBuilder.build())
        .setResources(resourcesBuilder.build())
        .setVerbose(params.verbose)
        .setCloudProviderStr(params.cloud.orEmpty())
        .setOpenPorts(
            PortSpecs.newBuilder()
                .addAllPorts(buildOpenPorts(params))
                .build(),
        )
        .addAllSecretIds(params.secrets.map { it.secretId })
        .addAllVolumeMounts(
            params.volumes.map { (mountPath, volume) ->
                VolumeMount.newBuilder()
                    .setVolumeId(volume.volumeId)
                    .setMountPath(mountPath)
                    .setAllowBackgroundCommits(true)
                    .setReadOnly(volume.isReadOnly)
                    .build()
            },
        )
        .addAllCloudBucketMounts(
            params.cloudBucketMounts.map { (mountPath, mount) ->
                mount.toProto(mountPath)
            },
        )
        .putAllExperimentalOptions(experimentalOptions)
        .addAllEntrypointArgs(params.command ?: emptyList())

    if (idleTimeoutMs != null) {
        definitionBuilder.idleTimeoutSecs = (idleTimeoutMs / 1000).toInt()
    }
    if (workdir != null) {
        definitionBuilder.workdir = workdir
    }
    if (params.pty) {
        definitionBuilder.ptyInfo = defaultSandboxPtyInfo()
    }
    if (params.regions.isNotEmpty()) {
        definitionBuilder.schedulerPlacement = SchedulerPlacement.newBuilder()
            .addAllRegions(params.regions)
            .build()
    }
    if (params.proxy != null) {
        definitionBuilder.proxyId = params.proxy.proxyId
    }
    if (params.name != null) {
        definitionBuilder.name = params.name
    }
    if (params.customDomain != null) {
        definitionBuilder.customDomain = params.customDomain
    }

    return SandboxCreateRequest.newBuilder()
        .setAppId(appId)
        .setDefinition(definitionBuilder.build())
        .build()
}

private fun buildOpenPorts(params: SandboxCreateParams): List<PortSpec> {
    val ports = mutableListOf<PortSpec>()

    for (port in params.encryptedPorts) {
        ports += PortSpec.newBuilder()
            .setPort(port)
            .setUnencrypted(false)
            .build()
    }
    for (port in params.h2Ports) {
        ports += PortSpec.newBuilder()
            .setPort(port)
            .setUnencrypted(false)
            .setTunnelType(TunnelType.TUNNEL_TYPE_H2)
            .build()
    }
    for (port in params.unencryptedPorts) {
        ports += PortSpec.newBuilder()
            .setPort(port)
            .setUnencrypted(true)
            .build()
    }

    return ports
}

fun buildTaskExecStartRequestProto(
    taskId: String,
    execId: String,
    command: List<String>,
    params: SandboxExecParams = SandboxExecParams(),
): TaskExecStartRequest {
    checkForRenamedParams(
        mapOf("timeout" to params.timeout).filterValues { it != null },
        mapOf("timeout" to "timeoutMs"),
    )

    val timeoutMs = params.timeoutMs
    if (timeoutMs != null) {
        if (timeoutMs <= 0) {
            throw InvalidError("timeoutMs must be positive, got $timeoutMs")
        }
        if (timeoutMs % 1000 != 0L) {
            throw InvalidError("timeoutMs must be a multiple of 1000ms, got $timeoutMs")
        }
    }

    val stdoutConfig = when (params.stdout) {
        StdioBehavior.PIPE -> TaskExecStdoutConfig.TASK_EXEC_STDOUT_CONFIG_PIPE
        StdioBehavior.IGNORE -> TaskExecStdoutConfig.TASK_EXEC_STDOUT_CONFIG_DEVNULL
    }
    val stderrConfig = when (params.stderr) {
        StdioBehavior.PIPE -> TaskExecStderrConfig.TASK_EXEC_STDERR_CONFIG_PIPE
        StdioBehavior.IGNORE -> TaskExecStderrConfig.TASK_EXEC_STDERR_CONFIG_DEVNULL
    }

    val builder = TaskExecStartRequest.newBuilder()
        .setTaskId(taskId)
        .setExecId(execId)
        .addAllCommandArgs(command)
        .setStdoutConfig(stdoutConfig)
        .setStderrConfig(stderrConfig)
        .addAllSecretIds(params.secrets.map { it.secretId })
        .setRuntimeDebug(false)

    if (timeoutMs != null) {
        builder.timeoutSecs = (timeoutMs / 1000).toInt()
    }
    if (params.workdir != null) {
        builder.workdir = params.workdir
    }
    if (params.pty) {
        builder.ptyInfo = defaultSandboxPtyInfo()
    }

    return builder.build()
}
