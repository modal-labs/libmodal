require_relative 'test_helper'

class TestFunction < Minitest::Test
  def setup
    super
    # Stub Pickle.dumps and Pickle.loads for these tests
    Modal::Pickle.stub :dumps, ->(obj) { obj.to_s.bytes.pack('C*') } do
      Modal::Pickle.stub :loads, ->(bytes) { bytes.force_encoding('UTF-8') } do
        yield
      end
    end
  end

  def test_function_call
    function_id = "fu-testfunc123"
    app_name = "libmodal-test-support"
    function_name = "echo_string"
    expected_output = "output: hello"
    function_call_id = "fc-123"

    # Mock Function.lookup
    mock_modal_client.expect(:call, Modal::Proto::FunctionGetResponse.new(function_id: function_id), [:function_get, MiniTest::Any])

    # Mock ControlPlaneInvocation.create
    Modal::ControlPlaneInvocation.stub :create, ->(func_id, input, inv_type) {
      mock_invocation = MiniTest::Mock.new
      mock_invocation.expect(:function_call_id, function_call_id)
      mock_invocation.expect(:await_output, expected_output)
      mock_invocation
    } do
      function_ = Modal::Function_.lookup(app_name, function_name)
      result_kwargs = function_.remote([], s: "hello")
      assert_equal expected_output, result_kwargs

      # Reset mock for args call
      mock_modal_client.expect(:call, Modal::Proto::FunctionGetResponse.new(function_id: function_id), [:function_get, MiniTest::Any])
      Modal::ControlPlaneInvocation.stub :create, ->(func_id, input, inv_type) {
        mock_invocation = MiniTest::Mock.new
        mock_invocation.expect(:function_call_id, function_call_id)
        mock_invocation.expect(:await_output, expected_output)
        mock_invocation
      } do
        function_ = Modal::Function_.lookup(app_name, function_name)
        result_args = function_.remote(["hello"])
        assert_equal expected_output, result_args
      end
    end
    mock_modal_client.verify
  end

  def test_function_spawn
    function_id = "fu-testfuncspawn"
    app_name = "libmodal-test-support"
    function_name = "echo_string"
    function_call_id = "fc-spawn-123"

    # Mock Function.lookup
    mock_modal_client.expect(:call, Modal::Proto::FunctionGetResponse.new(function_id: function_id), [:function_get, MiniTest::Any])

    # Mock ControlPlaneInvocation.create
    Modal::ControlPlaneInvocation.stub :create, ->(func_id, input, inv_type) {
      mock_invocation = MiniTest::Mock.new
      mock_invocation.expect(:function_call_id, function_call_id)
      mock_invocation
    } do
      function_ = Modal::Function_.lookup(app_name, function_name)
      function_call = function_.spawn([], s: "hello")
      assert_equal function_call_id, function_call.function_call_id
    end
    mock_modal_client.verify
  end

  def test_function_not_found
    app_name = "libmodal-test-support"
    function_name = "not_a_real_function"

    mock_modal_client.expect(:call, nil, [:function_get, MiniTest::Any]) do
      raise Modal::NotFoundError.new("Function '#{app_name}/#{function_name}' not found")
    end

    assert_raises Modal::NotFoundError do
      Modal::Function_.lookup(app_name, function_name)
    end
    mock_modal_client.verify
  end

  def test_function_call_large_input
    function_id = "fu-largeinput"
    app_name = "libmodal-test-support"
    function_name = "bytelength"
    test_len = 3 * 1000 * 1000 # More than 2 MiB
    input_data = "x" * test_len # Ruby string will be converted to bytes

    # Mock Function.lookup
    mock_modal_client.expect(:call, Modal::Proto::FunctionGetResponse.new(function_id: function_id), [:function_get, MiniTest::Any])

    # Mock blob_create
    mock_modal_client.expect(:call, Modal::Proto::BlobCreateResponse.new(
      blob_id: "bl-largeblob123",
      upload_url: "http://localhost:8080/upload/blob"
    ), [:blob_create, MiniTest::Any])

    # Mock HTTP PUT request for blob upload
    stub_request(:put, "http://localhost:8080/upload/blob").to_return(status: 200, body: "")

    # Mock ControlPlaneInvocation.create
    Modal::ControlPlaneInvocation.stub :create, ->(func_id, input, inv_type) {
      mock_invocation = MiniTest::Mock.new
      mock_invocation.expect(:function_call_id, "fc-largeinput-123")
      mock_invocation.expect(:await_output, test_len)
      mock_invocation
    } do
      function_ = Modal::Function_.lookup(app_name, function_name)
      result = function_.remote([input_data])
      assert_equal test_len, result
    end
    mock_modal_client.verify
  end
end
