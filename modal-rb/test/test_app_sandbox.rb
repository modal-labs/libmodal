require_relative 'test_helper'

class TestAppSandbox < Minitest::Test
  def test_app_lookup_with_create_if_missing
    with_mocked_client do
      app_name = "test_app_#{Time.now.to_i}"
      mock_response = Minitest::Mock.new
      mock_response.expect(:app_id, "app_123")
      expected_request = Modal::Client::AppGetOrCreateRequest.new(
        app_name: app_name,
        environment_name: Modal::Config.environment_name,
      )
      mock_api_client.expect(:call, mock_response, [:app_get_or_create, expected_request])
      response = mock_api_client.call(:app_get_or_create, expected_request)
      assert_equal "app_123", response.app_id
      mock_api_client.verify
    end
  end

  def test_app_lookup_without_create_if_missing
    with_mocked_client do
      app_name = "existing_app"
      mock_response = Minitest::Mock.new
      mock_response.expect(:app_id, "app_456")
      expected_request = Modal::Client::AppGetOrCreateRequest.new(
        app_name: app_name,
        environment_name: Modal::Config.environment_name,
      )
      mock_api_client.expect(:call, mock_response, [:app_get_or_create, expected_request])
      response = mock_api_client.call(:app_get_or_create, expected_request)
      assert_equal "app_456", response.app_id
      mock_api_client.verify
    end
  end
end
