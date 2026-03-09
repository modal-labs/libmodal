package com.modal.modalkt

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class RetriesTest {
    @Test
    fun parseRetries() {
        val retries = parseRetries(3)!!
        assertEquals(3, retries.maxRetries)
        assertEquals(1.0, retries.backoffCoefficient)
        assertEquals(1000, retries.initialDelayMs)
    }

    @Test
    fun retriesConstructor() {
        val retries = Retries(
            maxRetries = 2,
            backoffCoefficient = 2.0,
            initialDelayMs = 2000,
            maxDelayMs = 5000,
        )
        assertEquals(2, retries.maxRetries)
        assertEquals(2.0, retries.backoffCoefficient)
        assertEquals(2000, retries.initialDelayMs)
        assertEquals(5000, retries.maxDelayMs)

        assertFailsWith<InvalidError> { Retries(maxRetries = -1) }
        assertFailsWith<InvalidError> { Retries(maxRetries = 0, backoffCoefficient = 0.9) }
        assertFailsWith<InvalidError> { Retries(maxRetries = 0, initialDelayMs = 61_000) }
        assertFailsWith<InvalidError> { Retries(maxRetries = 0, maxDelayMs = 500) }
    }
}
