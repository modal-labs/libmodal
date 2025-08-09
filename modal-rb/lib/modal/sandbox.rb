require_relative 'sandbox_filesystem'
require_relative 'streams'
require 'ostruct'

module Modal
  class Sandbox
    attr_reader :sandbox_id, :stdin, :stdout, :stderr
    attr_accessor :task_id

    def initialize(sandbox_id)
      @sandbox_id = sandbox_id
      @task_id = nil

      @stdin = ModalWriteStream.new(SandboxInputStream.new(sandbox_id))
      @stdout = ModalReadStream.new(SandboxOutputStream.new(sandbox_id, Modal::Client::FileDescriptor::FILE_DESCRIPTOR_STDOUT))
      @stderr = ModalReadStream.new(SandboxOutputStream.new(sandbox_id, Modal::Client::FileDescriptor::FILE_DESCRIPTOR_STDERR))
    end

    def exec(command, options = {})
      ensure_task_id

      workdir = options[:workdir]
      timeout_secs = options[:timeout] ? options[:timeout] / 1000 : 0

      request = Modal::Client::ContainerExecRequest.new(
        task_id: @task_id,
        command: command,
        workdir: workdir,
        timeout_secs: timeout_secs
      )

      resp = Modal.client.call(:container_exec, request)
      ContainerProcess.new(resp.exec_id, "text")
    end

    def open(path, mode)
      ensure_task_id

      request = Modal::Client::ContainerFilesystemExecRequest.new(
        file_open_request: Modal::Client::ContainerFileOpenRequest.new(
          path: path,
          mode: mode
        ),
        task_id: @task_id
      )
      resp = run_filesystem_exec(request)
      SandboxFile.new(resp.response.file_open_response.file_descriptor, @task_id)
    end

    def terminate
      request = Modal::Client::SandboxTerminateRequest.new(sandbox_id: @sandbox_id)
      Modal.client.call(:sandbox_terminate, request)
    end

    def wait
      loop do
        request = Modal::Client::SandboxWaitRequest.new(
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
    def run_filesystem_exec(request)
      response = Modal.client.call(:container_filesystem_exec, request)
      if response.respond_to?(:file_descriptor) && response.file_descriptor
        return OpenStruct.new(
          response: OpenStruct.new(
            file_open_response: OpenStruct.new(
              file_descriptor: response.file_descriptor
            )
          )
        )
      end

      exec_id = response.exec_id
      retries = 10

      while retries > 0
        begin
          output_request = Modal::Client::ContainerFilesystemExecGetOutputRequest.new(
            exec_id: exec_id,
            timeout: 10
          )

          stream = Modal.client.call(:container_filesystem_exec_get_output, output_request)

          stream.each do |batch|
            if batch.respond_to?(:error) && batch.error
              raise SandboxFilesystemError.new(batch.error.error_message)
            end

            if batch.respond_to?(:eof) && batch.eof
              return OpenStruct.new(response: response)
            end
          end

          retries -= 1
          sleep(0.1)

        rescue GRPC::BadStatus => e
          if e.code == GRPC::Core::StatusCodes::DEADLINE_EXCEEDED
            retries -= 1
            next
          else
            raise e
          end
        end
      end

      raise SandboxFilesystemError.new("Filesystem operation timed out")
    end

    def ensure_task_id
      return if @task_id

      request = Modal::Client::SandboxGetTaskIdRequest.new(
        sandbox_id: @sandbox_id,
        wait_until_ready: true
      )
      resp = Modal.client.call(:sandbox_get_task_id, request)

      if resp.task_id && !resp.task_id.empty?
        @task_id = resp.task_id
      else
        raise "Sandbox #{@sandbox_id} does not have a task ID, it may not be running"
      end
    end
  end

  class ContainerProcess
    attr_reader :exec_id, :stdin, :stdout, :stderr

    def initialize(exec_id, mode)
      @exec_id = exec_id
      @stdin = ModalWriteStream.new(ContainerProcessInputStream.new(exec_id))

      if mode == "text"
        @stdout = ModalReadStream.new(ContainerProcessOutputStream.new(exec_id, Modal::Client::FileDescriptor::FILE_DESCRIPTOR_STDOUT, true))
        @stderr = ModalReadStream.new(ContainerProcessOutputStream.new(exec_id, Modal::Client::FileDescriptor::FILE_DESCRIPTOR_STDERR, true))
      else
        @stdout = ModalReadStream.new(ContainerProcessOutputStream.new(exec_id, Modal::Client::FileDescriptor::FILE_DESCRIPTOR_STDOUT, false))
        @stderr = ModalReadStream.new(ContainerProcessOutputStream.new(exec_id, Modal::Client::FileDescriptor::FILE_DESCRIPTOR_STDERR, false))
      end
    end

    def wait
      loop do
        request = Modal::Client::ContainerExecWaitRequest.new(
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

  class SandboxInputStream
    def initialize(sandbox_id)
      @sandbox_id = sandbox_id
      @index = 1
    end

    def write(chunk)
      request = Modal::Client::SandboxStdinWriteRequest.new(
        sandbox_id: @sandbox_id,
        input: chunk.bytes.pack('C*'), # Convert to bytes
        index: @index
      )
      Modal.client.call(:sandbox_stdin_write, request)
      @index += 1
    end

    def close
      request = Modal::Client::SandboxStdinWriteRequest.new(
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
      @last_entry_id = ""
      @data_collected = []
      @finished = false
    end

    def each
      return enum_for(:each) unless block_given?

      return if @finished


      # make one call and collect all data until EOF
      request = Modal::Client::SandboxGetLogsRequest.new(
        sandbox_id: @sandbox_id,
        file_descriptor: @file_descriptor,
        timeout: 10, # Give it more time to get all the data
        last_entry_id: @last_entry_id
      )

      begin
        resp = Modal.client.call(:sandbox_get_logs, request)

        # Process the entire streaming response
        resp.each do |batch|

          # Update last_entry_id
          if batch.respond_to?(:entry_id) && batch.entry_id && !batch.entry_id.empty?
            @last_entry_id = batch.entry_id
          end

          # Collect data from this batch
          if batch.respond_to?(:items) && batch.items
            batch.items.each do |item|
              if item.respond_to?(:data) && item.data && !item.data.empty?
                @data_collected << item.data
              end
            end
          end

          # Check for EOF
          if batch.respond_to?(:eof) && batch.eof
            @finished = true
            break
          end
        end

      rescue GRPC::BadStatus => e
        if e.code == GRPC::Core::StatusCodes::DEADLINE_EXCEEDED
          @finished = true
        else
          raise e
        end
      end

      # Yield all collected data
      @data_collected.each { |data| yield data }

    end
  end

  class ContainerProcessInputStream
    def initialize(exec_id)
      @exec_id = exec_id
      @message_index = 1
    end

    def write(chunk)
      request = Modal::Client::ContainerExecPutInputRequest.new(
        exec_id: @exec_id,
        input: Modal::Client::ContainerExecInput.new(
          message: chunk.bytes.pack('C*'), # Convert to bytes
          message_index: @message_index
        )
      )
      Modal.client.call(:container_exec_put_input, request)
      @message_index += 1
    end

    def close
      request = Modal::Client::ContainerExecPutInputRequest.new(
        exec_id: @exec_id,
        input: Modal::Client::ContainerExecInput.new(
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
      @finished = false
    end

    def each
      return enum_for(:each) unless block_given?
      return if @finished


      begin
        request = Modal::Client::ContainerExecGetOutputRequest.new(
          exec_id: @exec_id,
          file_descriptor: @file_descriptor,
          timeout: 55,
          get_raw_bytes: true,
          last_batch_index: @last_batch_index
        )

        stream = Modal.client.call(:container_exec_get_output, request)

        stream.each do |batch|
          @last_batch_index = batch.batch_index if batch.respond_to?(:batch_index)
          if batch.respond_to?(:items) && batch.items
            batch.items.each do |item|
              if item.message_bytes && !item.message_bytes.empty?
                yield item.message_bytes
              end
            end
          end

          if (batch.respond_to?(:has_exit_code) && batch.has_exit_code) || batch.items.empty?
            break
          end
        end

      rescue GRPC::BadStatus => e
        if e.code == GRPC::Core::StatusCodes::DEADLINE_EXCEEDED
        else
          raise e
        end
      end

      @finished = true
    end
  end
end
