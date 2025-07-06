module Modal
  # Interface for Invocation (ControlPlaneInvocation or InputPlaneInvocation)
  module Invocation
    def await_output(timeout = nil); end
    def retry(retry_count); end
  end

  class ControlPlaneInvocation
    include Invocation

    attr_reader :function_call_id

    def initialize(function_call_id, input = nil, function_call_jwt = nil, input_jwt = nil)
      @function_call_id = function_call_id
      @input = input
      @function_call_jwt = function_call_jwt
      @input_jwt = input_jwt
    end

    def self.create(function_id, input, invocation_type)
      function_put_inputs_item = Modal::Proto::FunctionPutInputsItem.new(idx: 0, input: input)
      request = Modal::Proto::FunctionMapRequest.new(
        function_id: function_id,
        function_call_type: Modal::Proto::FunctionCallType::FUNCTION_CALL_TYPE_UNARY,
        function_call_invocation_type: invocation_type,
        pipelined_inputs: [function_put_inputs_item]
      )
      function_map_response = Modal.client.call(:function_map, request)

      new(
        function_map_response.function_call_id,
        input,
        function_map_response.function_call_jwt,
        function_map_response.pipelined_inputs[0].input_jwt
      )
    end

    def self.from_function_call_id(function_call_id)
      new(function_call_id)
    end

    def await_output(timeout = nil)
      poll_function_output(@function_call_id, timeout)
    end

    def retry(retry_count)
      unless @input
        raise "Cannot retry function invocation - input missing"
      end

      retry_item = Modal::Proto::FunctionRetryInputsItem.new(
        input_jwt: @input_jwt,
        input: @input,
        retry_count: retry_count
      )
      request = Modal::Proto::FunctionRetryInputsRequest.new(
        function_call_jwt: @function_call_jwt,
        inputs: [retry_item]
      )
      function_retry_response = Modal.client.call(:function_retry_inputs, request)
      @input_jwt = function_retry_response.input_jwts[0]
    end

    private

    # Polls for function output. This is a simplified version.
    def poll_function_output(function_call_id, timeout_ms = nil)
      # This is a simplified polling loop.
      # In a real implementation, you'd manage polling intervals and check for
      # specific output events or completion status.
      # The `FunctionGetOutputsResponse` would contain `completed` and `output`.
      start_time = Time.now
      loop do
        request = Modal::Proto::FunctionGetOutputsRequest.new(
          function_call_id: function_call_id,
          timeout: 55 # seconds, placeholder
        )
        resp = Modal.client.call(:function_get_outputs, request)

        if resp.completed
          if resp.output
            return Pickle.loads(resp.output)
          else
            return nil # Or raise an error if output is expected but missing
          end
        end

        if timeout_ms && (Time.now - start_time) * 1000 > timeout_ms
          raise FunctionTimeoutError.new("Function call timed out after #{timeout_ms}ms")
        end

        sleep(1) # Poll every second
      end
    rescue GRPC::BadStatus => e
      if e.code == GRPC::Core::StatusCodes::DEADLINE_EXCEEDED
        raise FunctionTimeoutError.new("Function call timed out.")
      else
        raise e # Re-raise other gRPC errors
      end
    end
  end
end
