package com.modal.modalkt

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
        .toByteArray()
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
            builder.stringValue = value
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
            builder.intValue = value.toLong()
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
            builder.boolValue = value
        }

        ParameterType.PARAM_TYPE_BYTES -> {
            val value = if (rawValue == null && parameterSpec.hasDefault) {
                parameterSpec.bytesDefault.toByteArray()
            } else {
                rawValue
            }
            if (value !is ByteArray) {
                throw InvalidError("Parameter '$name' must be a byte array")
            }
            builder.bytesValue = com.google.protobuf.ByteString.copyFrom(value)
        }

        else -> throw InvalidError("Unsupported parameter type: ${parameterSpec.type}")
    }

    return builder.build()
}
