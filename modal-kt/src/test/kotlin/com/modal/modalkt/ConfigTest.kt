package com.modal.modalkt

import com.github.stefanbirkner.systemlambda.SystemLambda.withEnvironmentVariable
import kotlin.test.Test
import kotlin.test.assertEquals

class ConfigTest {
    @Test
    fun getConfigPathWithEnvVar() {
        val customPath = "/custom/path/to/config.toml"
        withEnvironmentVariable("MODAL_CONFIG_PATH", customPath).execute {
            assertEquals(customPath, configFilePath())
        }
    }

    @Test
    fun getConfigPathWithoutEnvVar() {
        withEnvironmentVariable("MODAL_CONFIG_PATH", "").execute {
            val expectedPath = java.nio.file.Path.of(
                System.getProperty("user.home"),
                ".modal.toml",
            ).toString()
            assertEquals(expectedPath, configFilePath())
        }
    }
}
