module Modal
  class Secret
    attr_reader :secret_id

    def initialize(secret_id)
      @secret_id = secret_id
    end

    def self.from_name(name, options = {})
      environment = options[:environment]
      required_keys = options[:required_keys] || []

      request = Modal::Proto::SecretGetOrCreateRequest.new(
        deployment_name: name,
        namespace: Modal::Proto::DeploymentNamespace::DEPLOYMENT_NAMESPACE_WORKSPACE,
        environment_name: Config.environment_name(environment),
        required_keys: required_keys
      )

      resp = Modal.client.call(:secret_get_or_create, request)
      new(resp.secret_id)
    rescue NotFoundError, InvalidError => e
      # Re-raise specific errors for better clarity, similar to JS client
      if e.message.include?("Secret is missing key")
        raise NotFoundError.new(e.message)
      else
        raise NotFoundError.new("Secret '#{name}' not found")
      end
    end
  end
end
