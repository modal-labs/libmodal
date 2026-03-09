package com.modal.modalkt

import com.upokecenter.cbor.CBORObject
import com.upokecenter.cbor.CBORType

internal fun cborEncode(value: Any?): ByteArray {
    return CBORObject.FromObject(value).EncodeToBytes()
}

internal fun cborDecode(data: ByteArray): Any? {
    val obj = CBORObject.DecodeFromBytes(data)
    return when (obj.type) {
        CBORType.TextString -> obj.AsString()
        CBORType.Boolean -> obj.AsBoolean()
        CBORType.ByteString -> obj.GetByteString()
        else -> obj.ToObject(Any::class.java)
    }
}
