package com.modal.modalkt

import java.util.logging.Level

enum class LogLevel {
    DEBUG,
    INFO,
    WARN,
    ERROR,
}

interface Logger {
    fun debug(message: String, vararg args: Any?)

    fun info(message: String, vararg args: Any?)

    fun warn(message: String, vararg args: Any?)

    fun error(message: String, vararg args: Any?)
}

fun parseLogLevel(level: String): LogLevel {
    val normalized = level.trim().lowercase()
    if (normalized.isEmpty()) {
        return LogLevel.WARN
    }
    if (normalized == "warning") {
        return LogLevel.WARN
    }

    return when (normalized) {
        "debug" -> LogLevel.DEBUG
        "info" -> LogLevel.INFO
        "warn" -> LogLevel.WARN
        "error" -> LogLevel.ERROR
        else -> throw InvalidError(
            "Invalid log level value: \"$level\" (must be debug, info, warn, or error)",
        )
    }
}

class DefaultLogger(
    private val delegate: java.util.logging.Logger = java.util.logging.Logger.getLogger("com.modal.modalkt"),
) : Logger {
    override fun debug(message: String, vararg args: Any?) {
        delegate.log(Level.FINE, format(message, args))
    }

    override fun info(message: String, vararg args: Any?) {
        delegate.log(Level.INFO, format(message, args))
    }

    override fun warn(message: String, vararg args: Any?) {
        delegate.log(Level.WARNING, format(message, args))
    }

    override fun error(message: String, vararg args: Any?) {
        delegate.log(Level.SEVERE, format(message, args))
    }

    private fun format(message: String, args: Array<out Any?>): String {
        if (args.isEmpty()) {
            return message
        }

        val suffix = args.joinToString(separator = " ") { value ->
            value?.toString() ?: "null"
        }
        return "$message $suffix"
    }
}

private class FilteredLogger(
    private val delegate: Logger,
    private val threshold: LogLevel,
) : Logger {
    override fun debug(message: String, vararg args: Any?) {
        if (threshold.ordinal <= LogLevel.DEBUG.ordinal) {
            delegate.debug(message, *args)
        }
    }

    override fun info(message: String, vararg args: Any?) {
        if (threshold.ordinal <= LogLevel.INFO.ordinal) {
            delegate.info(message, *args)
        }
    }

    override fun warn(message: String, vararg args: Any?) {
        if (threshold.ordinal <= LogLevel.WARN.ordinal) {
            delegate.warn(message, *args)
        }
    }

    override fun error(message: String, vararg args: Any?) {
        if (threshold.ordinal <= LogLevel.ERROR.ordinal) {
            delegate.error(message, *args)
        }
    }
}

fun createLogger(logger: Logger? = null, logLevel: String = ""): Logger {
    val threshold = parseLogLevel(logLevel)
    val baseLogger = logger ?: DefaultLogger()

    if (logger == null) {
        return baseLogger
    }

    return FilteredLogger(baseLogger, threshold)
}

fun newLogger(logLevel: String = ""): Logger {
    return createLogger(null, logLevel)
}
