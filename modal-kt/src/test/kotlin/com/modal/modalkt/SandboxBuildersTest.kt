package com.modal.modalkt

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertFailsWith
import kotlin.test.assertNull
import kotlin.test.assertTrue
import modal.client.Api
import modal.task_command_router.TaskCommandRouterOuterClass

class SandboxBuildersTest {
    @Test
    fun parseGpuConfigValues() {
        assertEquals(Api.GPUConfig.getDefaultInstance(), parseGpuConfig(null))
        assertEquals("T4", parseGpuConfig("T4").gpuType)
        assertEquals(1, parseGpuConfig("A10G").count)
        assertEquals(3, parseGpuConfig("A100-80GB:3").count)
        assertEquals("A100", parseGpuConfig("a100:4").gpuType)

        assertFailsWith<InvalidError> { parseGpuConfig("T4:invalid") }
        assertFailsWith<InvalidError> { parseGpuConfig("T4:") }
        assertFailsWith<InvalidError> { parseGpuConfig("T4:0") }
        assertFailsWith<InvalidError> { parseGpuConfig("T4:-1") }
    }

    @Test
    fun buildSandboxCreateRequestWithoutPty() {
        val request = buildSandboxCreateRequestProto("app-123", "img-456")
        val definition = request.definition

        assertFalse(definition.hasPtyInfo())
    }

    @Test
    fun buildSandboxCreateRequestWithPty() {
        val request = buildSandboxCreateRequestProto(
            "app-123",
            "img-456",
            SandboxCreateParams(pty = true),
        )

        val ptyInfo = request.definition.ptyInfo
        assertTrue(ptyInfo.enabled)
        assertEquals(24, ptyInfo.winszRows)
        assertEquals(80, ptyInfo.winszCols)
        assertEquals("xterm-256color", ptyInfo.envTerm)
        assertEquals("truecolor", ptyInfo.envColorterm)
        assertEquals(Api.PTYInfo.PTYType.PTY_TYPE_SHELL, ptyInfo.ptyType)
    }

    @Test
    fun buildSandboxCreateRequestWithCpuAndMemory() {
        val request = buildSandboxCreateRequestProto(
            "app-123",
            "img-456",
            SandboxCreateParams(
                cpu = 2.0,
                cpuLimit = 4.5,
                memoryMiB = 1024,
                memoryLimitMiB = 2048,
            ),
        )

        val resources = request.definition.resources
        assertEquals(2000, resources.milliCpu)
        assertEquals(4500, resources.milliCpuMax)
        assertEquals(1024, resources.memoryMb)
        assertEquals(2048, resources.memoryMbMax)
    }

    @Test
    fun buildSandboxCreateRequestRejectsInvalidLimits() {
        assertFailsWith<InvalidError> {
            buildSandboxCreateRequestProto(
                "app-123",
                "img-456",
                SandboxCreateParams(cpu = 4.0, cpuLimit = 2.0),
            )
        }
        assertFailsWith<InvalidError> {
            buildSandboxCreateRequestProto(
                "app-123",
                "img-456",
                SandboxCreateParams(cpuLimit = 4.0),
            )
        }
        assertFailsWith<InvalidError> {
            buildSandboxCreateRequestProto(
                "app-123",
                "img-456",
                SandboxCreateParams(memoryMiB = 2048, memoryLimitMiB = 1024),
            )
        }
        assertFailsWith<InvalidError> {
            buildSandboxCreateRequestProto(
                "app-123",
                "img-456",
                SandboxCreateParams(memoryLimitMiB = 2048),
            )
        }
    }

    @Test
    fun buildSandboxCreateRequestDefaults() {
        val request = buildSandboxCreateRequestProto("app-123", "img-456")
        val definition = request.definition

        assertEquals(300, definition.timeoutSecs)
        assertTrue(definition.entrypointArgsList.isEmpty())
        assertEquals(Api.NetworkAccess.NetworkAccessType.OPEN, definition.networkAccess.networkAccessType)
        assertTrue(definition.networkAccess.allowedCidrsList.isEmpty())
        assertFalse(definition.verbose)
        assertEquals("", definition.cloudProviderStr)
        assertEquals(0, definition.resources.milliCpu)
        assertEquals(0, definition.resources.memoryMb)
        assertFalse(definition.hasIdleTimeoutSecs())
        assertFalse(definition.hasWorkdir())
        assertFalse(definition.hasSchedulerPlacement())
        assertFalse(definition.hasProxyId())
        assertTrue(definition.volumeMountsList.isEmpty())
        assertTrue(definition.cloudBucketMountsList.isEmpty())
        assertTrue(definition.secretIdsList.isEmpty())
        assertTrue(definition.openPorts.portsList.isEmpty())
        assertFalse(definition.hasName())
    }

    @Test
    fun sandboxInvalidTimeouts() {
        assertFailsWith<InvalidError> {
            buildSandboxCreateRequestProto(
                "app-123",
                "img-456",
                SandboxCreateParams(timeoutMs = 0),
            )
        }
        assertFailsWith<InvalidError> {
            buildSandboxCreateRequestProto(
                "app-123",
                "img-456",
                SandboxCreateParams(timeoutMs = 1500),
            )
        }
        assertFailsWith<InvalidError> {
            buildSandboxCreateRequestProto(
                "app-123",
                "img-456",
                SandboxCreateParams(idleTimeoutMs = 2500),
            )
        }
    }

    @Test
    fun validateExecArgsRespectsLimit() {
        validateExecArgs(listOf("echo", "hello"))
        validateExecArgs(listOf("a".repeat((1 shl 16) - 10)))

        assertFailsWith<InvalidError> {
            validateExecArgs(listOf("a".repeat((1 shl 16) + 1)))
        }
    }

    @Test
    fun buildTaskExecStartRequestDefaults() {
        val request = buildTaskExecStartRequestProto("task-123", "exec-456", listOf("bash"))

        assertEquals("task-123", request.taskId)
        assertEquals("exec-456", request.execId)
        assertEquals(listOf("bash"), request.commandArgsList)
        assertEquals(
            TaskCommandRouterOuterClass.TaskExecStdoutConfig.TASK_EXEC_STDOUT_CONFIG_PIPE,
            request.stdoutConfig,
        )
        assertEquals(
            TaskCommandRouterOuterClass.TaskExecStderrConfig.TASK_EXEC_STDERR_CONFIG_PIPE,
            request.stderrConfig,
        )
        assertFalse(request.hasTimeoutSecs())
        assertFalse(request.hasWorkdir())
        assertTrue(request.secretIdsList.isEmpty())
        assertFalse(request.hasPtyInfo())
        assertFalse(request.runtimeDebug)
    }

    @Test
    fun buildTaskExecStartRequestVariants() {
        val ignored = buildTaskExecStartRequestProto(
            "task-123",
            "exec-456",
            listOf("bash"),
            SandboxExecParams(stdout = StdioBehavior.IGNORE, stderr = StdioBehavior.IGNORE),
        )
        assertEquals(
            TaskCommandRouterOuterClass.TaskExecStdoutConfig.TASK_EXEC_STDOUT_CONFIG_DEVNULL,
            ignored.stdoutConfig,
        )
        assertEquals(
            TaskCommandRouterOuterClass.TaskExecStderrConfig.TASK_EXEC_STDERR_CONFIG_DEVNULL,
            ignored.stderrConfig,
        )

        val withPty = buildTaskExecStartRequestProto(
            "task-123",
            "exec-456",
            listOf("bash"),
            SandboxExecParams(pty = true, workdir = "/tmp", timeoutMs = 5000),
        )
        assertTrue(withPty.hasPtyInfo())
        assertEquals("/tmp", withPty.workdir)
        assertEquals(5, withPty.timeoutSecs)
    }

    @Test
    fun buildTaskExecStartRequestRejectsInvalidTimeouts() {
        assertFailsWith<InvalidError> {
            buildTaskExecStartRequestProto(
                "task-123",
                "exec-456",
                listOf("bash"),
                SandboxExecParams(timeoutMs = 0),
            )
        }
        assertFailsWith<InvalidError> {
            buildTaskExecStartRequestProto(
                "task-123",
                "exec-456",
                listOf("bash"),
                SandboxExecParams(timeoutMs = 1500),
            )
        }
    }
}
