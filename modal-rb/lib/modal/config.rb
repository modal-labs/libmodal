require 'toml-rb'
require 'fileutils'

module Modal
  module Config
    CONFIG_FILE = File.join(Dir.home, '.modal.toml')

    def self.read_config_file
      if File.exist?(CONFIG_FILE)
        TomlRB.parse(File.read(CONFIG_FILE))
      else
        {}
      end
    rescue StandardError => e
      raise "Failed to read or parse .modal.toml: #{e.message}"
    end

    def self.get_profile(profile_name = nil)
      config = read_config_file

      profile_name ||= ENV['MODAL_PROFILE']
      unless profile_name
        config.each do |name, data|
          if data['active']
            profile_name = name
            break
          end
        end
      end

      if profile_name && !config.key?(profile_name)
        raise "Profile \"#{profile_name}\" not found in .modal.toml. Please set the MODAL_PROFILE environment variable or specify a valid profile."
      end

      profile_data = profile_name ? config[profile_name] : {}

      server_url = ENV['MODAL_SERVER_URL'] || profile_data['server_url'] || 'https://api.modal.com'
      token_id = ENV['MODAL_TOKEN_ID'] || profile_data['token_id']
      token_secret = ENV['MODAL_TOKEN_SECRET'] || profile_data['token_secret']
      environment = ENV['MODAL_ENVIRONMENT'] || profile_data['environment']
      image_builder_version = ENV['MODAL_IMAGE_BUILDER_VERSION'] || profile_data['image_builder_version'] || '2024.10'

      unless token_id && token_secret
        raise "Profile \"#{profile_name}\" is missing token_id or token_secret. Please set them in .modal.toml or as environment variables."
      end

      {
        server_url: server_url,
        token_id: token_id,
        token_secret: token_secret,
        environment: environment,
        image_builder_version: image_builder_version
      }
    end

    def self.environment_name(environment = nil)
      environment || profile[:environment] || ""
    end

    def self.image_builder_version(version = nil)
      version || profile[:image_builder_version] || "2024.10"
    end

    @profile = get_profile
    def self.profile
      @profile
    end
  end
end
