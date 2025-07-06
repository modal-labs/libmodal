require 'digest'
require_relative 'pickle'

module Modal
  class Function_
    attr_reader :function_id, :method_name

    MAX_OBJECT_SIZE_BYTES = 2 * 1024 * 1024 # 2 MiB
    MAX_SYSTEM_RETRIES = 8

    def initialize(function_id, method_name = nil)
      @function_id = function_id
      @method_name = method_name
    end

    def self.lookup(app_name, name, options = {})
      environment = options[:environment] || Modal::Config.environment_name
      request = Modal::Client::FunctionGetRequest.new(
        app_name: app_name,
        object_tag: name,
        environment_name: Config.environment_name(environment)
      )
      resp = Modal.client.call(:function_get, request)
      new(resp.function_id)
    rescue NotFoundError
      raise NotFoundError.new("Function '#{app_name}/#{name}' not found")
    end

    def remote(args = [], kwargs = {})
      input = create_input(args, kwargs)
      invocation = ControlPlaneInvocation.create(@function_id, input, Modal::Client::FunctionCallInvocationType::FUNCTION_CALL_INVOCATION_TYPE_SYNC)

      retry_count = 0
      loop do
        begin
          return invocation.await_output
        rescue InternalFailure => e
          if retry_count <= MAX_SYSTEM_RETRIES
            invocation.retry(retry_count)
            retry_count += 1
          else
            raise e
          end
        end
      end
    end

    def spawn(args = [], kwargs = {})
      input = create_input(args, kwargs)
      invocation = ControlPlaneInvocation.create(@function_id, input, Modal::Client::FunctionCallInvocationType::FUNCTION_CALL_INVOCATION_TYPE_ASYNC)
      FunctionCall.new(invocation.function_call_id)
    end

    private

    def create_input(args, kwargs)
      # Create a proper Python tuple structure
      # The Python function expects to receive (*args, **kwargs)
      # So we need to create a tuple where:
      # - First element is a tuple of positional args
      # - Second element is a dict of keyword args

      # Convert Ruby array to Python tuple, Ruby hash to Python dict
      python_args = args.is_a?(Array) ? args : [args]
      python_kwargs = kwargs.is_a?(Hash) ? kwargs : {}

      # Create the payload as a tuple of (args_tuple, kwargs_dict)
      payload = Pickle.dumps([python_args, python_kwargs])
      args_blob_id = nil

      if payload.bytesize > MAX_OBJECT_SIZE_BYTES
        args_blob_id = blob_upload(payload)
      end

      Modal::Client::FunctionInput.new(
        args: args_blob_id ? nil : payload,
        args_blob_id: args_blob_id,
        data_format: Modal::Client::DataFormat::DATA_FORMAT_PICKLE,
        method_name: @method_name,
        final_input: false
      )
    end

    def blob_upload(data)
      content_md5 = Digest::MD5.base64digest(data)
      content_sha256 = Digest::SHA256.base64digest(data)
      content_length = data.bytesize

      request = Modal::Client::BlobCreateRequest.new(
        content_md5: content_md5,
        content_sha256_base64: content_sha256,
        content_length: content_length
      )
      resp = Modal.client.call(:blob_create, request)

      if resp.multipart
        raise "Function input size exceeds multipart upload threshold, unsupported by this SDK version"
      elsif resp.upload_url
        require 'net/http'
        require 'uri'

        uri = URI.parse(resp.upload_url)
        http = Net::HTTP.new(uri.host, uri.port)
        http.use_ssl = uri.scheme == 'https'

        req = Net::HTTP::Put.new(uri.request_uri)
        req['Content-Type'] = 'application/octet-stream'
        req['Content-MD5'] = content_md5
        req.body = data

        upload_resp = http.request(req)

        unless upload_resp.code.to_i >= 200 && upload_resp.code.to_i < 300
          raise "Failed blob upload: #{upload_resp.message}"
        end
        resp.blob_id
      else
        raise "Missing upload URL in BlobCreate response"
      end
    end
  end
end
