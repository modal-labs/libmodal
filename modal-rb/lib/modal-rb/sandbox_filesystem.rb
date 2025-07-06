module Modal
  class SandboxFile
    attr_reader :file_descriptor, :task_id

    def initialize(file_descriptor, task_id)
      @file_descriptor = file_descriptor
      @task_id = task_id
    end

    def read
      request = Modal::Proto::ContainerFilesystemExecRequest.new(
        file_read_request: Modal::Proto::FileReadRequest.new(
          file_descriptor: @file_descriptor
        ),
        task_id: @task_id
      )
      resp = Modal.client.call(:container_filesystem_exec, request)
      # This is a simplified read. In a real scenario, you'd handle streaming chunks.
      # Assuming `resp.file_read_response.data` contains the full content for simplicity.
      resp.file_read_response.data.bytes.pack('C*') # Return as binary string
    end

    def write(data)
      request = Modal::Proto::ContainerFilesystemExecRequest.new(
        file_write_request: Modal::Proto::FileWriteRequest.new(
          file_descriptor: @file_descriptor,
          data: data.bytes.pack('C*') # Convert to bytes
        ),
        task_id: @task_id
      )
      Modal.client.call(:container_filesystem_exec, request)
    end

    def flush
      request = Modal::Proto::ContainerFilesystemExecRequest.new(
        file_flush_request: Modal::Proto::FileFlushRequest.new(
          file_descriptor: @file_descriptor
        ),
        task_id: @task_id
      )
      Modal.client.call(:container_filesystem_exec, request)
    end

    def close
      request = Modal::Proto::ContainerFilesystemExecRequest.new(
        file_close_request: Modal::Proto::FileCloseRequest.new(
          file_descriptor: @file_descriptor
        ),
        task_id: @task_id
      )
      Modal.client.call(:container_filesystem_exec, request)
    end
  end
end
