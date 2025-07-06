require_relative 'invocation'

module Modal
  class FunctionCall
    attr_reader :function_call_id

    def initialize(function_call_id)
      @function_call_id = function_call_id
    end

    def self.from_id(function_call_id)
      new(function_call_id)
    end

    def get(options = {})
      timeout = options[:timeout]
      invocation = ControlPlaneInvocation.from_function_call_id(@function_call_id)
      invocation.await_output(timeout)
    end

    def cancel(options = {})
      terminate_containers = options[:terminate_containers] || false
      request = Modal::Proto::FunctionCallCancelRequest.new(
        function_call_id: @function_call_id,
        terminate_containers: terminate_containers
      )
      Modal.client.call(:function_call_cancel, request)
    end
  end
end
