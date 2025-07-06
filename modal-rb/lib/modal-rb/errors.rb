module Modal
  class ModalError < StandardError; end

  class NotFoundError < ModalError; end

  class InvalidError < ModalError; end

  class FunctionTimeoutError < ModalError; end

  class InternalFailure < ModalError; end

  class RemoteError < ModalError; end

  class QueueFullError < ModalError; end

  class QueueEmptyError < ModalError; end

  class SandboxFilesystemError < ModalError; end
end
