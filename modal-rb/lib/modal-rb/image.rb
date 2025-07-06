module Modal
  class Image
    attr_reader :image_id

    def initialize(image_id)
      @image_id = image_id
    end

    def self.from_registry_internal(app_id, tag, image_registry_config = nil)
      request = Modal::Proto::ImageGetOrCreateRequest.new(
        app_id: app_id,
        image: Modal::Proto::Image.new(
          dockerfile_commands: ["FROM #{tag}"],
          image_registry_config: image_registry_config
        ),
        namespace: Modal::Proto::DeploymentNamespace::DEPLOYMENT_NAMESPACE_WORKSPACE,
        builder_version: Config.image_builder_version
      )

      resp = Modal.client.call(:image_get_or_create, request)

      result = nil
      metadata = nil

      if resp.result && resp.result.status
        result = resp.result
        metadata = resp.metadata
      else
        result_joined = nil
        loop do
          # This part would require a streaming gRPC call, which is more complex in Ruby.
          # For simplicity, this is a simplified poll. In a real implementation, you'd
          # need to handle the streaming response from imageJoinStreaming.
          # For now, we'll just assume the first response has the result or poll.
          # A proper implementation would involve iterating over `client.image_join_streaming`.
          # For this base, we'll just simulate a wait.
          puts "Waiting for image build for #{resp.image_id}..."
          sleep(5) # Simulate polling

          # In a real scenario, you'd make another `image_join_streaming` call here
          # and check `item.result`.
          # For now, we'll just assume success after some wait if not already built.
          # This is a simplification and will not work for actual streaming logs.
          if resp.result && resp.result.status
            result_joined = resp.result
            metadata = resp.metadata
            break
          end
        end
        result = result_joined
      end

      # Note: metadata is currently unused, similar to modal-js
      _ = metadata

      case result.status
      when Modal::Proto::GenericResult_GenericStatus::GENERIC_STATUS_FAILURE
        raise "Image build for #{resp.image_id} failed with the exception:\n#{result.exception}"
      when Modal::Proto::GenericResult_GenericStatus::GENERIC_STATUS_TERMINATED
        raise "Image build for #{resp.image_id} terminated due to external shut-down. Please try again."
      when Modal::Proto::GenericResult_GenericStatus::GENERIC_STATUS_TIMEOUT
        raise "Image build for #{resp.image_id} timed out. Please try again with a larger timeout parameter."
      when Modal::Proto::GenericResult_GenericStatus::GENERIC_STATUS_SUCCESS
        new(resp.image_id)
      else
        raise "Image build for #{resp.image_id} failed with unknown status: #{result.status}"
      end
    end
  end
end
