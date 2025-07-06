require_relative 'test_helper'

class TestFunctionCall < Minitest::Test
  def setup
    super
    # Stub Pickle.loads for these tests
    Modal::Pickle.stub :loads, ->(bytes) { bytes.force_encoding('UTF-8') } do
      yield
    end
  end

  def test_function_call_get_and_cancel
    function_call_id = "fc-test123"
    expected_result = "output: hello"

    # Mock invocation for get
    mock_invocation_get = MiniTest::Mock.new
    mock_invocation_get.expect(:await_output, expected_result)
    Modal::ControlPlaneInvocation.stub :from_function_call_id, mock_invocation_get do
      function_call = Modal::FunctionCall.from_id(function_call_id)
      result = function_call.get
      assert_equal expected_result, result
    end
    mock_invocation_get.verify

    # Mock cancellation
    mock_modal_client.expect(:call, Modal::Proto::FunctionCallCancelResponse.new, [:function_call_cancel, MiniTest::Any])
    function_call = Modal::FunctionCall.from_id(function_call_id)
    function_call.cancel
    mock_modal_client.verify
  end

  def test_function_call_timeout
    function_call_id = "fc-timeout"

    # Mock invocation for timeout
    mock_invocation_timeout = MiniTest::Mock.new
    mock_invocation_timeout.expect(:await_output, nil) do
      raise Modal::FunctionTimeoutError.new("Function call timed out after 1000ms")
    end
    Modal::ControlPlaneInvocation.stub :from_function_call_id, mock_invocation_timeout do
      function_call = Modal::FunctionCall.from_id(function_call_id)
      assert_raises Modal::FunctionTimeoutError do
        function_call.get(timeout: 1000)
      end
    end
    mock_invocation_timeout.verify
  end
end
