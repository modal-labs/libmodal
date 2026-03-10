package com.modal.modalkt

import modal.client.*
import okio.ByteString.Companion.toByteString

fun encodeParameterSet(
    schema: List<ClassParameterSpec>,
    params: Map<String, Any?>,
): ByteArray {
    val encoded = schema.map { spec ->
        encodeParameter(spec, params[spec.name])
    }.sortedBy { it.name }

    return ClassParameterSet.newBuilder()
        .addAllParameters(encoded)
        .build()
        .encode()
}

private fun encodeParameter(
    parameterSpec: ClassParameterSpec,
    rawValue: Any?,
): ClassParameterValue {
    val name = parameterSpec.name
    val builder = ClassParameterValue.newBuilder()
        .setName(name)
        .setType(parameterSpec.type)

    when (parameterSpec.type) {
        ParameterType.PARAM_TYPE_STRING -> {
            val value = if (rawValue == null && parameterSpec.hasDefault) {
                parameterSpec.stringDefault
            } else {
                rawValue
            }
            if (value !is String) {
                throw InvalidError("Parameter '$name' must be a string")
            }
            builder.setStringValue(value)
        }

        ParameterType.PARAM_TYPE_INT -> {
            val value = if (rawValue == null && parameterSpec.hasDefault) {
                parameterSpec.intDefault
            } else {
                rawValue
            }
            if (value !is Number) {
                throw InvalidError("Parameter '$name' must be an integer")
            }
            builder.setIntValue(value.toLong())
        }

        ParameterType.PARAM_TYPE_BOOL -> {
            val value = if (rawValue == null && parameterSpec.hasDefault) {
                parameterSpec.boolDefault
            } else {
                rawValue
            }
            if (value !is Boolean) {
                throw InvalidError("Parameter '$name' must be a boolean")
            }
            builder.setBoolValue(value)
        }

        ParameterType.PARAM_TYPE_BYTES -> {
            val value = if (rawValue == null && parameterSpec.hasDefault) {
                parameterSpec.bytesDefault?.toByteArray()
                    ?: throw InvalidError("Parameter '$name' is missing a default byte value")
            } else {
                rawValue
            }
            if (value !is ByteArray) {
                throw InvalidError("Parameter '$name' must be a byte array")
            }
            builder.setBytesValue(value.toByteString())
        }

        else -> throw InvalidError("Unsupported parameter type: ${parameterSpec.type}")
    }

    return builder.build()
}
