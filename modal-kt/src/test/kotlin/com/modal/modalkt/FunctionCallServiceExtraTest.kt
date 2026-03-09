package com.modal.modalkt

import kotlinx.coroutines.runBlocking
import modal.client.Api
import kotlin.test.Test
import kotlin.test.assertEquals

class FunctionCallServiceExtraTest {
    @Test
    fun functionCallServiceFromId() = runBlocking {
        val (client, _) = createMockModalClients()
        val functionCall = client.functionCalls.fromId("fc-123")
        assertEquals("fc-123", functionCall.functionCallId)
    }

    @Test
    fun functionCallCancel() = runBlocking {
        val (client, mock) = createMockModalClients()
        var seenTerminate: Boolean? = null
        mock.handleUnary("/FunctionCallCancel") { request ->
            request as Api.FunctionCallCancelRequest
            seenTerminate = request.terminateContainers
            com.google.protobuf.Empty.getDefaultInstance()
        }

        val functionCall = FunctionCall(client, "fc-123")
        functionCall.cancel(FunctionCallCancelParams(terminateContainers = true))
        assertEquals(true, seenTerminate)
    }
}
