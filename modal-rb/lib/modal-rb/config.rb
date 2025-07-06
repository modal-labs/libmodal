module Modal
  class Config
    def self.profile
      {
        server_url: ENV['MODAL_SERVER_URL'] || 'https://api.modal.com',
        token_id: ENV['MODAL_TOKEN_ID'],
        token_secret: ENV['MODAL_TOKEN_SECRET']
      }
    end

    def self.environment_name(environment = nil)
      environment || ENV['MODAL_ENVIRONMENT'] || 'main'
    end

    def self.image_builder_version
      ENV['MODAL_IMAGE_BUILDER_VERSION'] || '2023.12'
    end
  end
end
