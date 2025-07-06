require_relative 'test_helper'
require 'modal-rb'

class TestCls < Minitest::Test
  def setup
    super
    # Stub Pickle.dumps and Pickle.loads for these tests
    Modal::Pickle.stub :dumps, ->(obj) { obj.to_s.bytes.pack('C*') } do
      Modal::Pickle.stub :loads, ->(bytes) { bytes.force_encoding('UTF-8') } do
        yield
      end
    end
  end

  def test_cls_call
    app_name = "libmodal-test-support"
    cls_name = "EchoCls"
    function_id = "fu-echocls"
    expected_output = "output: hello"

    # Mock Cls.lookup (function_get)
    mock_modal_client.expect(:call, Modal::Proto::FunctionGetResponse.new(
      function_id: function_id,
      handle_metadata: Modal::Proto::FunctionHandleMetadata.new(
        method_handle_metadata: { "echo_string" => Modal::Proto::MethodHandleMetadata.new }
      )
    ), [:function_get, MiniTest::Any])

    # Mock instance creation (function_map and await_output)
    Modal::ControlPlaneInvocation.stub :create, ->(func_id, input, inv_type) {
      mock_invocation = MiniTest::Mock.new
      mock_invocation.expect(:function_call_id, "fc-echocls-123")
      mock_invocation.expect(:await_output, expected_output)
      mock_invocation
    } do
      cls = Modal::Cls.lookup(app_name, cls_name)
      instance = cls.instance
      function_ = instance.method("echo_string")
      result = function_.remote([], s: "hello")
      assert_equal expected_output, result
    end
    mock_modal_client.verify
  end

  def test_cls_call_parametrized
    app_name = "libmodal-test-support"
    cls_name = "EchoClsParametrized"
    service_function_id = "fu-echoparamcls"
    bound_function_id = "fu-echoparamcls-bound"
    expected_output = "output: hello-init"

    # Mock Cls.lookup (function_get)
    mock_modal_client.expect(:call, Modal::Proto::FunctionGetResponse.new(
      function_id: service_function_id,
      handle_metadata: Modal::Proto::FunctionHandleMetadata.new(
        class_parameter_info: Modal::Proto::ClassParameterInfo.new(
          format: Modal::Proto::ClassParameterInfo_ParameterSerializationFormat::PARAM_SERIALIZATION_FORMAT_PROTO,
          schema: [
            Modal::Proto::ClassParameterSpec.new(name: "name", type: Modal::Proto::ParameterType::PARAM_TYPE_STRING)
          ]
        ),
        method_handle_metadata: { "echo_parameter" => Modal::Proto::MethodHandleMetadata.new }
      )
    ), [:function_get, MiniTest::Any])

    # Mock instance creation (function_bind_params)
    mock_modal_client.expect(:call, Modal::Proto::FunctionBindParamsResponse.new(bound_function_id: bound_function_id), [:function_bind_params, MiniTest::Any])

    # Mock invocation for the bound function
    Modal::ControlPlaneInvocation.stub :create, ->(func_id, input, inv_type) {
      mock_invocation = MiniTest::Mock.new
      mock_invocation.expect(:function_call_id, "fc-echoparamcls-123")
      mock_invocation.expect(:await_output, expected_output)
      mock_invocation
    } do
      cls = Modal::Cls.lookup(app_name, cls_name)
      instance = cls.instance(name: "hello-init")
      function_ = instance.method("echo_parameter")
      result = function_.remote
      assert_equal expected_output, result
    end
    mock_modal_client.verify
  end

  def test_cls_not_found
    app_name = "libmodal-test-support"
    cls_name = "NotRealClassName"

    mock_modal_client.expect(:call, nil, [:function_get, MiniTest::Any]) do
      raise Modal::NotFoundError.new("Class '#{app_name}/#{cls_name}' not found")
    end

    assert_raises Modal::NotFoundError do
      Modal::Cls.lookup(app_name, cls_name)
    end
    mock_modal_client.verify
  end

  def test_cls_non_existent_method
    app_name = "libmodal-test-support"
    cls_name = "EchoCls"
    function_id = "fu-echocls"

    # Mock Cls.lookup (function_get)
    mock_modal_client.expect(:call, Modal::Proto::FunctionGetResponse.new(
      function_id: function_id,
      handle_metadata: Modal::Proto::FunctionHandleMetadata.new(
        method_handle_metadata: { "echo_string" => Modal::Proto::MethodHandleMetadata.new }
      )
    ), [:function_get, MiniTest::Any])

    cls = Modal::Cls.lookup(app_name, cls_name)
    instance = cls.instance
    assert_raises Modal::NotFoundError do
      instance.method("nonexistent")
    end
    mock_modal_client.verify
  end
end
