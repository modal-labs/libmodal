package com.modal.modalkt

open class ModalException(message: String) : RuntimeException(message)

class FunctionTimeoutError(message: String) : ModalException(message)

class RemoteError(message: String) : ModalException(message)

class InternalFailure(message: String) : ModalException(message)

class NotFoundError(message: String) : ModalException(message)

class AlreadyExistsError(message: String) : ModalException(message)

class InvalidError(message: String) : ModalException(message)

class QueueEmptyError(message: String) : ModalException(message)

class QueueFullError(message: String) : ModalException(message)

class SandboxFilesystemError(message: String) : ModalException(message)

class SandboxTimeoutError(message: String = "Sandbox operation timed out") : ModalException(message)

class ClientClosedError(
    message: String = "Unable to perform operation on a detached sandbox",
) : ModalException(message)
