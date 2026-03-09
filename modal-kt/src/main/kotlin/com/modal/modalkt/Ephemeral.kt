package com.modal.modalkt

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

const val ephemeralObjectHeartbeatSleep: Long = 300_000

class EphemeralHeartbeatManager(
    private val heartbeatFn: suspend () -> Unit,
    private val sleepMs: Long = ephemeralObjectHeartbeatSleep,
    private val scope: CoroutineScope = CoroutineScope(SupervisorJob() + Dispatchers.IO),
) {
    private val job: Job

    init {
        job = scope.launch {
            while (true) {
                heartbeatFn()
                delay(sleepMs)
            }
        }
    }

    fun stop() {
        job.cancel()
    }
}
