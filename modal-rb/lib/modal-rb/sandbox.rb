require_relative 'sandbox_filesystem'
require_relative 'streams'

module Modal
  class Sandbox
    attr_reader :sandbox_id, :stdin, :stdout, :stderr
    attr_accessor :task_id # Used internally for filesystem operations

    def initialize(sandbox_id)
      @sandbox_id = sandbox_id
      @task_id = nil # Will be set after exec or other operations

      # Initialize streams (mocked for now, actual streaming requires more complex gRPC handling)
      @stdin = ModalWriteStream.new(SandboxInputStream.new(sandbox_id))
      @stdout = ModalReadStream.new(SandboxOutputStream.new(sandbox_id, Modal::Proto::FileDescriptor::FILE_DESCRIPTOR_STDOUT))
      @stderr = ModalReadStream.new(SandboxOutputStream.new(sandbox_id, Modal::Proto::FileDescriptor::FILE_DESCRIPTOR_STDERR))
    end

    def exec(command, options = {})
      mode = options[:mode] || "text"
      stdout_behavior = options[:stdout] || "pipe"
      stderr_behavior = options[:stderr] || "pipe"

      request = Modal::Proto::ContainerExecRequest.new(
        sandbox_id: @sandbox_id,
        command: command,
        # stdin_behavior: Modal::Proto::StdioBehavior::STDIO_BEHAVIOR_PIPE, # Not explicitly set in JS example
        stdout_behavior: case stdout_behavior
                         when "pipe" then Modal::Proto::StdioBehavior::STDIO_BEHAVIOR_PIPE
                         when "ignore" then Modal::Proto::StdioBehavior::STDIO_BEHAVIOR_IGNORE
                         end,
        stderr_behavior: case stderr_behavior
                         when "pipe" then Modal::Proto::StdioBehavior::STDIO_BEHAVIOR_PIPE
                         when "ignore" then Modal::Proto::StdioBehavior::STDIO_BEHAVIOR_IGNORE
                         end,
        get_raw_bytes: mode == "binary"
      )

      resp = Modal.client.call(:container_exec, request)
      @task_id = resp.task_id # Set task_id for filesystem operations

      ContainerProcess.new(resp.exec_id, mode)
    end

    def open(path, mode)
      request = Modal::Proto::ContainerFilesystemExecRequest.new(
        file_open_request: Modal::Proto::FileOpenRequest.new(
          path: path,
          mode: mode
        ),
        task_id: @task_id # Use the task_id from a previous exec or set it if opening first
      )
      result = run_filesystem_exec(request)
      SandboxFile.new(result[:response].response.file_open_response.file_descriptor, @task_id)
    end

    def terminate
      request = Modal::Proto::SandboxTerminateRequest.new(sandbox_id: @sandbox_id)
      Modal.client.call(:sandbox_terminate, request)
    end

    def wait
      loop do
        request = Modal::Proto::SandboxWaitRequest.new(
          sandbox_id: @sandbox_id,
          timeout: 55 # seconds
        )
        resp = Modal.client.call(:sandbox_wait, request)
        if resp.completed
          return resp.exit_code || 0
        end
        sleep(1) # Poll every second
      end
    end

    private

    # Helper to run filesystem exec requests and handle output.
    # This is a simplified version, actual streaming would be more complex.
    def run_filesystem_exec(request)
      response = Modal.client.call(:container_filesystem_exec, request)
      chunks = []
      retries = 10
      completed = false

      while !completed && retries > 0
        begin
          output_request = Modal::Proto::ContainerFilesystemExecGetOutputRequest.new(
            exec_id: response.exec_id,
            timeout: 55 # seconds
          )
          # This would ideally be a streaming call. For simplicity, we'll assume a single response for now.
          batch = Modal.client.call(:container_filesystem_exec_get_output, output_request)

          chunks << batch.output if batch.output
          if batch.eof
            completed = true
            break
          end
          if batch.error
            raise SandboxFilesystemError.new(batch.error.error_message)
          end
        rescue GRPC::BadStatus => e
          if Modal.client.class.const_get(:RETRYABLE_GRPC_STATUS_CODES).include?(e.code) && retries > 0
            retries -= 1
            sleep(0.5) # Small backoff
          else
            raise e
          end
        end
      end
      { chunks: chunks, response: response }
    end
  end

  class ContainerProcess
    attr_reader :exec_id, :stdin, :stdout, :stderr

    def initialize(exec_id, mode)
      @exec_id = exec_id
      @stdin = ModalWriteStream.new(ContainerProcessInputStream.new(exec_id))

      if mode == "text"
        @stdout = ModalReadStream.new(ContainerProcessOutputStream.new(exec_id, Modal::Proto::FileDescriptor::FILE_DESCRIPTOR_STDOUT, true))
        @stderr = ModalReadStream.new(ContainerProcessOutputStream.new(exec_id, Modal::Proto::FileDescriptor::FILE_DESCRIPTOR_STDERR, true))
      else
        @stdout = ModalReadStream.new(ContainerProcessOutputStream.new(exec_id, Modal::Proto::FileDescriptor::FILE_DESCRIPTOR_STDOUT, false))
        @stderr = ModalReadStream.new(ContainerProcessOutputStream.new(exec_id, Modal::Proto::FileDescriptor::FILE_DESCRIPTOR_STDERR, false))
      end
    end

    def wait
      loop do
        request = Modal::Proto::ContainerExecWaitRequest.new(
          exec_id: @exec_id,
          timeout: 55 # seconds
        )
        resp = Modal.client.call(:container_exec_wait, request)
        if resp.completed
          return resp.exit_code || 0
        end
        sleep(1) # Poll every second
      end
    end
  end

  # Internal stream classes (simplified mocks for now)
  class SandboxInputStream
    def initialize(sandbox_id)
      @sandbox_id = sandbox_id
      @index = 1
    end

    def write(chunk)
      request = Modal::Proto::SandboxStdinWriteRequest.new(
        sandbox_id: @sandbox_id,
        input: chunk.bytes.pack('C*'), # Convert to bytes
        index: @index
      )
      Modal.client.call(:sandbox_stdin_write, request)
      @index += 1
    end

    def close
      request = Modal::Proto::SandboxStdinWriteRequest.new(
        sandbox_id: @sandbox_id,
        index: @index,
        eof: true
      )
      Modal.client.call(:sandbox_stdin_write, request)
    end
  end

  class SandboxOutputStream
    def initialize(sandbox_id, file_descriptor)
      @sandbox_id = sandbox_id
      @file_descriptor = file_descriptor
      @last_entry_id = "0-0"
    end

    # This is a simplified `each` method for the stream.
    # In a real implementation, this would handle streaming gRPC responses.
    def each
      loop do
        request = Modal::Proto::SandboxGetLogsRequest.new(
          sandbox_id: @sandbox_id,
          file_descriptor: @file_descriptor,
          timeout: 55, # seconds
          last_entry_id: @last_entry_id
        )
        resp = Modal.client.call(:sandbox_get_logs, request)

        resp.items.each do |item|
          yield item.data.bytes.pack('C*') # Yield raw bytes, TextDecoderStream will handle text
        end
        @last_entry_id = resp.entry_id

        if resp.eof
          break
        end
        sleep(0.1) # Simulate polling for new data
      end
    end
  end

  class ContainerProcessInputStream
    def initialize(exec_id)
      @exec_id = exec_id
      @message_index = 1
    end

    def write(chunk)
      request = Modal::Proto::ContainerExecPutInputRequest.new(
        exec_id: @exec_id,
        input: Modal::Proto::ContainerExecInput.new(
          message: chunk.bytes.pack('C*'), # Convert to bytes
          message_index: @message_index
        )
      )
      Modal.client.call(:container_exec_put_input, request)
      @message_index += 1
    end

    def close
      request = Modal::Proto::ContainerExecPutInputRequest.new(
        exec_id: @exec_id,
        input: Modal::Proto::ContainerExecInput.new(
          message_index: @message_index,
          eof: true
        )
      )
      Modal.client.call(:container_exec_put_input, request)
    end
  end

  class ContainerProcessOutputStream
    def initialize(exec_id, file_descriptor, decode_text)
      @exec_id = exec_id
      @file_descriptor = file_descriptor
      @decode_text = decode_text
      @last_batch_index = 0
    end

    def each
      loop do
        request = Modal::Proto::ContainerExecGetOutputRequest.new(
          exec_id: @exec_id,
          file_descriptor: @file_descriptor,
          timeout: 55, # seconds
          get_raw_bytes: !@decode_text,
          last_batch_index: @last_batch_index
        )
        resp = Modal.client.call(:container_exec_get_output, request)

        resp.items.each do |item|
          yield item.message_bytes
        end
        @last_batch_index = resp.batch_index

        if resp.exit_code
          break
        end
        sleep(0.1) # Simulate polling
      end
    end
  end
end
