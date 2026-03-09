package com.modal.modalkt

fun checkForRenamedParams(
    params: Map<String, Any?>?,
    renamed: Map<String, String>,
) {
    if (params == null) {
        return
    }

    for ((oldName, newName) in renamed) {
        if (params.containsKey(oldName)) {
            throw InvalidError("Parameter '$oldName' has been renamed to '$newName'.")
        }
    }
}
