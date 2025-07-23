$LOAD_PATH.unshift File.expand_path('../lib', __dir__)

require 'minitest/autorun'
require 'webmock/minitest'

require 'modal'

TEST_MODAL_TOML = File.join(Dir.home, '.modal.toml')

module TestHelper
  def self.included(base)
    base.class_eval do
      def setup
        setup_modal_test_environment
        super
      end

      def teardown
        cleanup_modal_test_environment
        super
      end
    end
  end

  private

  def setup_modal_test_environment
    WebMock.disable_net_connect!

    toml_content = if File.exist?(TEST_MODAL_TOML)
      existing_content = File.read(TEST_MODAL_TOML)
      if existing_content.include?('[modal-rb-test]')
        existing_content
      else
        existing_content + <<~TOML

        [modal-rb-test]
        server_url = "https://localhost:50051"
        token_id = "test-token-id"
        token_secret = "test-token-secret"
        environment = "modal-rb-test-env"
        active = true
        TOML
      end
    else
      <<~TOML

      [modal-rb-test]
      server_url = "https://localhost:50051"
      token_id = "test-token-id"
      token_secret = "test-token-secret"
      environment = "modal-rb-test-env"
      active = true
      TOML
    end

    File.write(TEST_MODAL_TOML, toml_content)

    Modal::Config.instance_variable_set(:@profile, Modal::Config.get_profile('modal-rb-test'))
  end

  def cleanup_modal_test_environment
    if File.exist?(TEST_MODAL_TOML)
      content = File.read(TEST_MODAL_TOML)
      updated_content = content.gsub(/\[modal-rb-test\].*?(?=\[|\z)/m, '').strip

      if updated_content.empty?
        FileUtils.rm_f(TEST_MODAL_TOML)
      else
        File.write(TEST_MODAL_TOML, updated_content)
      end
    end

    WebMock.allow_net_connect!
  end

  def mock_api_client
    @mock_api_client ||= Minitest::Mock.new
  end

  def with_mocked_client(&block)
    Modal.stub(:client, mock_api_client, &block)
  end
end

class Minitest::Test
  include TestHelper
end
