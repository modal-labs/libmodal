package com.modal.modalkt

import kotlin.test.Test
import kotlin.test.assertFailsWith

class ValidationTest {
    @Test
    fun checkForRenamedParamsRejectsOldNames() {
        val error = assertFailsWith<InvalidError> {
            checkForRenamedParams(mapOf("timeout" to 5000), mapOf("timeout" to "timeoutMs"))
        }
        kotlin.test.assertEquals(
            "Parameter 'timeout' has been renamed to 'timeoutMs'.",
            error.message,
        )
    }

    @Test
    fun checkForRenamedParamsAllowsCurrentNames() {
        checkForRenamedParams(mapOf("timeoutMs" to 5000), mapOf("timeout" to "timeoutMs"))
        checkForRenamedParams(null, mapOf("timeout" to "timeoutMs"))
        checkForRenamedParams(emptyMap(), mapOf("timeout" to "timeoutMs"))
    }

    @Test
    fun modalClientRejectsLegacyTimeoutParameter() {
        assertFailsWith<InvalidError> {
            ModalClient(ModalClientParams(timeout = 5000))
        }
    }
}
