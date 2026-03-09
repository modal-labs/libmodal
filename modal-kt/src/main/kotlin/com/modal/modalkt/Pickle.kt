package com.modal.modalkt

import net.razorvine.pickle.Pickler
import net.razorvine.pickle.Unpickler

internal object PickleCodec {
    fun encode(value: Any?): ByteArray {
        return Pickler().dumps(value)
    }

    fun decode(bytes: ByteArray): Any? {
        return Unpickler().loads(bytes)
    }
}
