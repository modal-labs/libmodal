require 'grpc'
require 'securerandom'

module Modal
  class ApiClient
    RETRYABLE_GRPC_STATUS_CODES = Set.new([
      GRPC::Core::StatusCodes::DEADLINE_EXCEEDED,
      GRPC::Core::StatusCodes::UNAVAILABLE,
      GRPC::Core::StatusCodes::CANCELLED,
      GRPC::Core::StatusCodes::INTERNAL,
      GRPC::Core::StatusCodes::UNKNOWN,
    ])

    def initialize
      @profile = Config.profile
      target, credentials = parse_server_url(@profile[:server_url])

      @stub = Modal::Client::ModalClient::Stub.new(
        target,
        credentials,
        channel_args: {
          'grpc.max_receive_message_length' => 100 * 1024 * 1024, # 100 MiB
          'grpc.max_send_message_length' => 100 * 1024 * 1024, # 100 MiB
        }
      )
    end

    def call(method_name, request_pb, options = {})
      retries = options[:retries] || 3
      base_delay = options[:base_delay] || 0.1 # seconds
      max_delay = options[:max_delay] || 1.0 # seconds
      delay_factor = options[:delay_factor] || 2
      timeout = options[:timeout] # milliseconds

      idempotency_key = SecureRandom.uuid
      attempt = 0

      loop do
        metadata = {
          'x-modal-client-type' => Modal::Client::ClientType::CLIENT_TYPE_LIBMODAL.to_s, # TODO: libmodal_rb!!!
          'x-modal-client-version' => '1.0.0',
          'x-modal-token-id' => @profile[:token_id],
          'x-modal-token-secret' => @profile[:token_secret],
          'x-idempotency-key' => idempotency_key,
          'x-retry-attempt' => attempt.to_s,
        }
        metadata['x-retry-delay'] = base_delay.to_s if attempt > 0

        call_options = { metadata: metadata }
        call_options[:deadline] = Time.now + timeout / 1000.0 if timeout

        begin
          response = @stub.send(method_name, request_pb, call_options)
          return response
        rescue GRPC::BadStatus => e
          if RETRYABLE_GRPC_STATUS_CODES.include?(e.code) && attempt < retries
            puts "Retrying #{method_name} due to #{e.code} (attempt #{attempt + 1}/#{retries})"
            sleep(base_delay)
            base_delay = [base_delay * delay_factor, max_delay].min
            attempt += 1
          else
            raise convert_grpc_error(e)
          end
        rescue StandardError => e
          raise e
        end
      end
    end

    private

    def parse_server_url(server_url)
      if server_url.start_with?('https://')
        target = server_url.sub('https://', '')
        credentials = GRPC::Core::ChannelCredentials.new
      elsif server_url.start_with?('http://')
        target = server_url.sub('http://', '')
        credentials = :this_channel_is_insecure
      else
        target = server_url
        credentials = GRPC::Core::ChannelCredentials.new
      end

      [target, credentials]
    end

    def convert_grpc_error(grpc_error)
      case grpc_error.code
      when GRPC::Core::StatusCodes::NOT_FOUND
        NotFoundError.new(grpc_error.details)
      when GRPC::Core::StatusCodes::FAILED_PRECONDITION
        if grpc_error.details.include?("Secret is missing key")
          NotFoundError.new(grpc_error.details)
        else
          InvalidError.new(grpc_error.details)
        end
      when GRPC::Core::StatusCodes::DEADLINE_EXCEEDED
        FunctionTimeoutError.new(grpc_error.details)
      when GRPC::Core::StatusCodes::INTERNAL, GRPC::Core::StatusCodes::UNKNOWN
        InternalFailure.new(grpc_error.details)
      else
        RemoteError.new(grpc_error.details)
      end
    end
  end

  @client = ApiClient.new
  def self.client
    @client
  end
end
