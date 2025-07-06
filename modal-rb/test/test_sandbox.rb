require_relative 'test_helper'

class TestSandbox < Minitest::Test
  def setup
    super
    # Stub stream reading/writing to avoid actual stream processing in tests
    Modal::SandboxInputStream.any_instance.stubs(:write)
    Modal::SandboxInputStream.any_instance.stubs(:close)
    Modal::SandboxOutputStream.any_instance.stubs(:each).returns(Enumerator.new {}.each) # Empty enumerator
    Modal::ContainerProcessInputStream.any_instance.stubs(:write)
    Modal::ContainerProcessInputStream.any_instance.stubs(:close)
    Modal::ContainerProcessOutputStream.any_instance.stubs(:each).returns(Enumerator.new {}.each) # Empty enumerator
  end

  def test_create_one_sandbox
    app_id = "ap-testapp123"
    image_id = "im-testimage123"
    sandbox_id = "sb-testsandbox123"

    mock_app = Modal::App.new(app_id)
    mock_image = Modal::Image.new(image_id)

    mock_modal_client.expect(:call, Modal::Proto::SandboxCreateResponse.new(sandbox_id: sandbox_id), [:sandbox_create, MiniTest::Any])
    mock_modal_client.expect(:call, Modal::Proto::SandboxTerminateResponse.new, [:sandbox_terminate, MiniTest::Any])
    mock_modal_client.expect(:call, Modal::Proto::SandboxWaitResponse.new(completed: true, exit_code: 0), [:sandbox_wait, MiniTest::Any])

    sandbox = mock_app.create_sandbox(mock_image)
    assert_equal sandbox_id, sandbox.sandbox_id
    sandbox.terminate
    assert_equal 0, sandbox.wait
    mock_modal_client.verify
  end

  def test_pass_cat_to_stdin
    app_id = "ap-testapp123"
    image_id = "im-testimage123"
    sandbox_id = "sb-testsandbox123"
    exec_id = "ex-testexec123"
    input_text = "this is input that should be mirrored by cat"

    mock_app = Modal::App.new(app_id)
    mock_image = Modal::Image.new(image_id)

    # Mock sandbox creation
    mock_modal_client.expect(:call, Modal::Proto::SandboxCreateResponse.new(sandbox_id: sandbox_id), [:sandbox_create, MiniTest::Any])

    # Mock exec command
    mock_modal_client.expect(:call, Modal::Proto::ContainerExecResponse.new(exec_id: exec_id, task_id: "tk-123"), [:container_exec, MiniTest::Any])

    # Mock stdin write (actual stream write is stubbed, so just expect the gRPC call)
    Modal::SandboxInputStream.any_instance.unstub(:write) # Unstub to allow expectation
    Modal::SandboxInputStream.any_instance.unstub(:close) # Unstub to allow expectation
    mock_modal_client.expect(:call, Modal::Proto::SandboxStdinWriteResponse.new, [:sandbox_stdin_write, MiniTest::Any])
    mock_modal_client.expect(:call, Modal::Proto::SandboxStdinWriteResponse.new, [:sandbox_stdin_write, MiniTest::Any]) # For close

    # Mock stdout read (actual stream read is stubbed, so provide mock data)
    Modal::SandboxOutputStream.any_instance.unstub(:each) # Unstub to allow iteration
    Modal::SandboxOutputStream.any_instance.stubs(:each).returns(
      Enumerator.new do |y|
        y << input_text.bytes.pack('C*') # Yield bytes
      end.each
    )

    # Mock sandbox termination
    mock_modal_client.expect(:call, Modal::Proto::SandboxTerminateResponse.new, [:sandbox_terminate, MiniTest::Any])

    sandbox = mock_app.create_sandbox(mock_image, command: ["cat"])
    sandbox.stdin.write_text(input_text)
    sandbox.stdin.close
    assert_equal input_text, sandbox.stdout.read_text
    sandbox.terminate
    mock_modal_client.verify
  end

  def test_ignore_large_stdout
    app_id = "ap-testapp123"
    image_id = "im-testimage123"
    sandbox_id = "sb-testsandbox123"
    exec_id = "ex-ignorestdout"

    mock_app = Modal::App.new(app_id)
    mock_image = Modal::Image.new(image_id)

    # Mock sandbox creation
    mock_modal_client.expect(:call, Modal::Proto::SandboxCreateResponse.new(sandbox_id: sandbox_id), [:sandbox_create, MiniTest::Any])

    # Mock exec command with stdout: "ignore"
    mock_modal_client.expect(:call, Modal::Proto::ContainerExecResponse.new(exec_id: exec_id, task_id: "tk-123"), [:container_exec, MiniTest::Any])

    # Mock container_exec_wait
    mock_modal_client.expect(:call, Modal::Proto::ContainerExecWaitResponse.new(completed: true, exit_code: 0), [:container_exec_wait, MiniTest::Any])

    # Mock sandbox termination
    mock_modal_client.expect(:call, Modal::Proto::SandboxTerminateResponse.new, [:sandbox_terminate, MiniTest::Any])

    sandbox = mock_app.create_sandbox(mock_image)
    process = sandbox.exec(["python", "-c", "print(\"a\" * 1_000_000)"], stdout: "ignore")

    # The stdout stream should be empty because it's ignored
    assert_equal "", process.stdout.read_text
    assert_equal 0, process.wait
    sandbox.terminate
    mock_modal_client.verify
  end
end
