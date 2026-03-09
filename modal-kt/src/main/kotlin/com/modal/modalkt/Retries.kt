package com.modal.modalkt

data class Retries(
    val maxRetries: Int,
    val backoffCoefficient: Double = 2.0,
    val initialDelayMs: Long = 1_000,
    val maxDelayMs: Long = 60_000,
) {
    init {
        if (maxRetries !in 0..10) {
            throw InvalidError("Invalid maxRetries: $maxRetries. Must be between 0 and 10.")
        }
        if (backoffCoefficient < 1.0 || backoffCoefficient > 10.0) {
            throw InvalidError(
                "Invalid backoffCoefficient: $backoffCoefficient. Must be between 1.0 and 10.0",
            )
        }
        if (initialDelayMs < 0 || initialDelayMs > 60_000) {
            throw InvalidError(
                "Invalid initialDelayMs: $initialDelayMs. Must be between 0 and 60000 ms.",
            )
        }
        if (maxDelayMs < 1_000 || maxDelayMs > 60_000) {
            throw InvalidError(
                "Invalid maxDelayMs: $maxDelayMs. Must be between 1000 and 60000 ms.",
            )
        }
    }
}

fun parseRetries(retries: Any?): Retries? {
    if (retries == null) {
        return null
    }
    if (retries is Int) {
        if (retries !in 0..10) {
            throw InvalidError("Retries parameter must be an integer between 0 and 10. Found: $retries")
        }
        return Retries(
            maxRetries = retries,
            backoffCoefficient = 1.0,
            initialDelayMs = 1_000,
        )
    }
    if (retries is Retries) {
        return retries
    }
    throw InvalidError(
        "Retries parameter must be an integer or instance of Retries. Found: ${retries::class.qualifiedName}.",
    )
}
