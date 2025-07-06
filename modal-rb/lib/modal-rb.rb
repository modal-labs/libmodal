require 'modal-rb/version'

module Modal
  require 'modal-rb/modal_proto/options_pb'
  require 'modal-rb/modal_proto/api_pb'
  require 'modal-rb/modal_proto/api_services_pb'

  require 'modal-rb/config'

  require 'modal-rb/api_client'
  require 'modal-rb/app'
  require 'modal-rb/errors'
  require 'modal-rb/image'
  require 'modal-rb/secret'
  require 'modal-rb/function'
  require 'modal-rb/function_call'
  require 'modal-rb/sandbox'
  require 'modal-rb/sandbox_filesystem'
  require 'modal-rb/streams'
  require 'modal-rb/pickle' # placeholder
end
