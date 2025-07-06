require_relative 'test_helper'

class TestSandboxFilesystem < Minitest::Test
  def setup
    super
    # Stub sandbox creation and termination to focus on filesystem ops
    @app_id = "ap-fs-test"
    @image_id = "im-fs-test"
    @sandbox_id = "sb-fs-test"
    @task_id = "tk-fs-test" # Task ID for filesystem ops

    mock_app = Modal::App.new(@app_id)
    mock_image = Modal::Image.new(@image_id)

    mock_modal_client.expect(:call, Modal::Proto::SandboxCreateResponse.new(sandbox_id: @sandbox_id), [:sandbox_create, MiniTest::Any])
    @sandbox = mock_app.create_sandbox(mock_image)
    @sandbox.task_id = @task_id # Manually set task_id for filesystem ops

    # Stub exec to get a task_id, if not already done
    mock_modal_client.expect(:call, Modal::Proto::ContainerExecResponse.new(exec_id: "ex-dummy", task_id: @task_id), [:container_exec, MiniTest::Any])
    @sandbox.exec(["true"]) # Run a dummy command to set task_id

    # Stub termination to ensure it's called at the end of the test
    mock_modal_client.expect(:call, Modal::Proto::SandboxTerminateResponse.new, [:sandbox_terminate, MiniTest::Any])
  end

  def teardown
    @sandbox.terminate
    super
  end

  def mock_filesystem_exec_response(file_descriptor: nil, data: nil, eof: true, error: nil)
    response_pb = nil
    if file_descriptor
      response_pb = Modal::Proto::ContainerFilesystemExecResponse.new(
        file_open_response: Modal::Proto::FileOpenResponse.new(file_descriptor: file_descriptor)
      )
    elsif data
      response_pb = Modal::Proto::ContainerFilesystemExecResponse.new(
        file_read_response: Modal::Proto::FileReadResponse.new(data: data)
      )
    end

    batch_pb = Modal::Proto::ContainerFilesystemExecGetOutputResponse.new(
      output: data, # For read, this is the data
      eof: eof,
      error: error ? Modal::Proto::GrpcError.new(error_message: error) : nil
    )
    # The `run_filesystem_exec` helper in sandbox.rb expects a single response from `container_filesystem_exec_get_output`
    # and then iterates. So we'll mock the `container_filesystem_exec_get_output` directly.
    mock_modal_client.expect(:call, response_pb, [:container_filesystem_exec, MiniTest::Any])
    mock_modal_client.expect(:call, batch_pb, [:container_filesystem_exec_get_output, MiniTest::Any])
  end

  def test_write_and_read_binary_file
    test_data = "binary_content".bytes.pack('C*')
    file_descriptor = "fd-bin-1"

    mock_filesystem_exec_response(file_descriptor: file_descriptor) # Open
    mock_filesystem_exec_response(data: nil) # Write
    mock_filesystem_exec_response(data: test_data) # Read
    mock_filesystem_exec_response(data: nil) # Close

    write_handle = @sandbox.open("/tmp/test.bin", "w")
    write_handle.write(test_data)
    write_handle.close

    read_handle = @sandbox.open("/tmp/test.bin", "r")
    read_data = read_handle.read
    assert_equal test_data, read_data
    read_handle.close

    mock_modal_client.verify
  end

  def test_append_to_file_binary
    initial_data = "initial".bytes.pack('C*')
    append_data = "appended".bytes.pack('C*')
    expected_data = (initial_data.bytes + append_data.bytes).pack('C*')
    file_descriptor = "fd-append-1"

    mock_filesystem_exec_response(file_descriptor: file_descriptor) # Open 'w'
    mock_filesystem_exec_response(data: nil) # Write initial
    mock_filesystem_exec_response(data: nil) # Close

    mock_filesystem_exec_response(file_descriptor: file_descriptor) # Open 'a'
    mock_filesystem_exec_response(data: nil) # Write appended
    mock_filesystem_exec_response(data: nil) # Close

    mock_filesystem_exec_response(file_descriptor: file_descriptor) # Open 'r'
    mock_filesystem_exec_response(data: expected_data) # Read final
    mock_filesystem_exec_response(data: nil) # Close

    write_handle = @sandbox.open("/tmp/append.txt", "w")
    write_handle.write(initial_data)
    write_handle.close

    append_handle = @sandbox.open("/tmp/append.txt", "a")
    append_handle.write(append_data)
    append_handle.close

    read_handle = @sandbox.open("/tmp/append.txt", "r")
    content = read_handle.read
    assert_equal expected_data, content
    read_handle.close

    mock_modal_client.verify
  end

  def test_file_handle_flush
    test_data = "flush_data".bytes.pack('C*')
    file_descriptor = "fd-flush-1"

    mock_filesystem_exec_response(file_descriptor: file_descriptor) # Open 'w'
    mock_filesystem_exec_response(data: nil) # Write
    mock_filesystem_exec_response(data: nil) # Flush
    mock_filesystem_exec_response(data: nil) # Close

    mock_filesystem_exec_response(file_descriptor: file_descriptor) # Open 'r'
    mock_filesystem_exec_response(data: test_data) # Read
    mock_filesystem_exec_response(data: nil) # Close

    handle = @sandbox.open("/tmp/flush.txt", "w")
    handle.write(test_data)
    handle.flush
    handle.close

    read_handle = @sandbox.open("/tmp/flush.txt", "r")
    content = read_handle.read
    assert_equal test_data, content
    read_handle.close

    mock_modal_client.verify
  end
end
