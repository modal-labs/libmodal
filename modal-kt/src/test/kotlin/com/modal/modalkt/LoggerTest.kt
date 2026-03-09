package com.modal.modalkt

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class LoggerTest {
    @Test
    fun parseLogLevelValues() {
        assertEquals(LogLevel.DEBUG, parseLogLevel("debug"))
        assertEquals(LogLevel.DEBUG, parseLogLevel("DEBUG"))
        assertEquals(LogLevel.WARN, parseLogLevel("warning"))
        assertEquals(LogLevel.WARN, parseLogLevel("WARNING"))
        assertEquals(LogLevel.WARN, parseLogLevel(""))
    }

    @Test
    fun parseLogLevelRejectsInvalidValues() {
        val error = assertFailsWith<InvalidError> {
            parseLogLevel("invalid")
        }
        assertEquals(
            "Invalid log level value: \"invalid\" (must be debug, info, warn, or error)",
            error.message,
        )
    }

    @Test
    fun createLoggerReturnsDefaultLoggerWhenNoCustomLoggerProvided() {
        val logger = createLogger(null, "debug")
        assertTrue(logger is DefaultLogger)
    }

    @Test
    fun createLoggerAppliesLevelFilteringToCustomLogger() {
        val logger = RecordingLogger()
        val filtered = createLogger(logger, "warn")

        filtered.debug("debug")
        filtered.info("info")
        filtered.warn("warn")
        filtered.error("error")

        assertFalse(logger.messages.contains("debug"))
        assertFalse(logger.messages.contains("info"))
        assertTrue(logger.messages.contains("warn"))
        assertTrue(logger.messages.contains("error"))
    }

    @Test
    fun createLoggerPassesArgumentsToCustomLogger() {
        val logger = RecordingLogger()
        val filtered = createLogger(logger, "debug")

        filtered.debug("message", "key1", "value1", "key2", 123)

        assertEquals(listOf("message", "key1", "value1", "key2", 123), logger.lastArgs)
    }

    private class RecordingLogger : Logger {
        val messages = mutableListOf<String>()
        var lastArgs: List<Any?> = emptyList()

        override fun debug(message: String, vararg args: Any?) {
            messages += message
            lastArgs = listOf(message) + args.toList()
        }

        override fun info(message: String, vararg args: Any?) {
            messages += message
        }

        override fun warn(message: String, vararg args: Any?) {
            messages += message
        }

        override fun error(message: String, vararg args: Any?) {
            messages += message
        }
    }
}
