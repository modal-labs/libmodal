$LOAD_PATH.unshift(File.expand_path('../lib', __dir__))
$LOAD_PATH.unshift(File.expand_path('../lib/modal-rb', __dir__))

require 'minitest/autorun'
require 'webmock/minitest'
require 'mocha/minitest'
require 'minitest/mock' # Ensures MiniTest::Mock is loaded

# Add lib directory to load path so 'modal-rb' can be found

# --- CRITICAL CHANGE HERE ---
# Load your main gem library *before* any test definitions
# This ensures Modal and Modal::Proto are defined early.
require 'modal-rb'

# --- END CRITICAL CHANGE ---

# Set up a dummy .modal.toml for testing purposes
# This file will be created/deleted during tests
TEST_MODAL_TOML = File.join(Dir.home, '.modal.toml')

module TestHelper
  def setup
    # Create a dummy .modal.toml for tests
    File.write(TEST_MODAL_TOML, <<~TOML
      [default]
      server_url = "http://localhost:50051"
      token_id = "test-token-id"
      token_secret = "test-token-secret"
      environment = "test"
      active = true
    TOML
    )

    # Reload config to pick up test profile
    Modal::Config.instance_variable_set(:@profile, Modal::Config.profile)

    # Stub the gRPC client for all tests
    # This is crucial for unit testing without a running Modal backend
    Modal.stub :client, mock_modal_client do
      super
    end
  end

  def teardown
    # Clean up the dummy .modal.toml
    FileUtils.rm_f(TEST_MODAL_TOML)
    super
  end

  # A mock gRPC client for testing purposes
  def mock_modal_client
    @mock_client ||= Minitest::Mock.new
    # Define common mock behaviors here if applicable
    # e.g., @mock_client.expect(:call, ..., [:app_get_or_create, ...])
    @mock_client
  end
end

# Include the helper module in all Minitest::Test classes
Minitest::Test.send(:include, TestHelper)
