require_relative 'pickle'

module Modal
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
      function_put_inputs_item = Modal::Client::FunctionPutInputsItem.new(idx: 0, input: input)
      request = Modal::Client::FunctionMapRequest.new(
        function_id: function_id,
        function_call_type: Modal::Client::FunctionCallType::FUNCTION_CALL_TYPE_UNARY,
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

      retry_item = Modal::Client::FunctionRetryInputsItem.new(
        input_jwt: @input_jwt,
        input: @input,
        retry_count: retry_count
      )
      request = Modal::Client::FunctionRetryInputsRequest.new(
        function_call_jwt: @function_call_jwt,
        inputs: [retry_item]
      )
      function_retry_response = Modal.client.call(:function_retry_inputs, request)
      @input_jwt = function_retry_response.input_jwts[0]
    end

    private

    def poll_function_output(function_call_id, timeout_ms = nil)
      start_time = Time.now
      last_entry_id = ""

      loop do
        request = Modal::Client::FunctionGetOutputsRequest.new(
          function_call_id: function_call_id,
          timeout: 55.0, # seconds
          last_entry_id: last_entry_id,
          max_values: 1,
          clear_on_success: true,
          requested_at: Time.now.to_f
        )

        resp = Modal.client.call(:function_get_outputs, request)

        if resp.last_entry_id && !resp.last_entry_id.empty?
          last_entry_id = resp.last_entry_id
        end

        if resp.outputs && resp.outputs.any?
          output_item = resp.outputs.first
          if output_item.result
            return process_result(output_item.result, output_item.data_format)
          end
        end

        if resp.respond_to?(:num_unfinished_inputs) && resp.num_unfinished_inputs == 0
          return nil
        end

        if timeout_ms && (Time.now - start_time) * 1000 > timeout_ms
          raise FunctionTimeoutError.new("Function call timed out after #{timeout_ms}ms")
        end

        sleep(1)
      end
    rescue GRPC::BadStatus => e
      if e.code == GRPC::Core::StatusCodes::DEADLINE_EXCEEDED
        raise FunctionTimeoutError.new("Function call timed out.")
      else
        raise e
      end
    end

    def process_result(result, data_format)
      status = result.status.to_s.to_sym

      case status
      when :GENERIC_STATUS_SUCCESS
        if result.data && !result.data.empty?
          return Pickle.load(result.data)
        elsif result.data_blob_id && !result.data_blob_id.empty?
          return nil
        else
          return nil
        end
      when :GENERIC_STATUS_TIMEOUT
        raise FunctionTimeoutError.new(result.exception || "Function timed out")
      when :GENERIC_STATUS_INTERNAL_FAILURE
        raise InternalFailure.new(result.exception || "Internal failure")
      else
        error_msg = result.exception || "Unknown error (status: #{result.status})"
        raise RemoteError.new("Function execution failed: #{error_msg}")
      end
    end
  end
end
