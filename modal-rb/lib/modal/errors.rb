module Modal
  class ModalError < StandardError; end
  class FunctionTimeoutError < ModalError; end
  class RemoteError < ModalError; end
  class InternalFailure < ModalError; end
  class NotFoundError < ModalError; end
  class InvalidError < ModalError; end
  class QueueEmptyError < ModalError; end
  class QueueFullError < ModalError; end
  class SandboxFilesystemError < ModalError; end
end
