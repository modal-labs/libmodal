package com.modal.modalkt

import kotlin.test.Test
import kotlin.test.assertContentEquals
import modal.client.*

class SerializationTest {
    @Test
    fun parameterSerialization() {
        var schema = listOf(
            ClassParameterSpec.newBuilder()
                .setName("foo")
                .setType(ParameterType.PARAM_TYPE_STRING)
                .build(),
            ClassParameterSpec.newBuilder()
                .setName("i")
                .setType(ParameterType.PARAM_TYPE_INT)
                .build(),
        )
        val values = mapOf<String, Any?>("i" to 5, "foo" to "bar")

        var serialized = encodeParameterSet(schema, values)
        var expected = byteArrayOf(
            10, 12, 10, 3, 102, 111, 111, 16, 1, 26, 3, 98, 97, 114, 10, 7, 10, 1, 105,
            16, 2, 32, 5,
        )
        assertContentEquals(expected, serialized)

        schema = listOf(schema[1], schema[0])
        serialized = encodeParameterSet(schema, values)
        assertContentEquals(expected, serialized)

        schema = listOf(
            ClassParameterSpec.newBuilder()
                .setName("x")
                .setType(ParameterType.PARAM_TYPE_BYTES)
                .setHasDefault(true)
                .setBytesDefault(com.google.protobuf.ByteString.copyFrom(byteArrayOf(0)))
                .build(),
        )
        serialized = encodeParameterSet(schema, emptyMap())
        expected = byteArrayOf(10, 8, 10, 1, 120, 16, 4, 50, 1, 0)
        assertContentEquals(expected, serialized)
    }
}
