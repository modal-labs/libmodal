package com.modal.modalkt

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class VersionTest {
    @Test
    fun versionConstantFormat() {
        assertTrue(getSdkVersion().matches(Regex("""^\d+\.\d+\.\d+(-dev\.\d+)?$""")))
    }

    @Test
    fun clientVersion() {
        val client = ModalClient(
            ModalClientParams(
                controlPlaneClient = MockControlPlaneClient(),
                authTokenProvider = MockControlPlaneClient(),
            ),
        )
        assertEquals(getSdkVersion(), client.version())
    }
}
