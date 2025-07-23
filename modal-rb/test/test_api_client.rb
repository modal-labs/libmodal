require_relative 'test_helper'

class TestApiClient < Minitest::Test
  def test_api_client_initialization
    client = Modal::ApiClient.new
    assert_instance_of Modal::ApiClient, client
    assert_respond_to client, :call
    assert_respond_to client, :stub
    assert client.instance_variable_get(:@stub).is_a?(Modal::Client::ModalClient::Stub)
  end

  def test_mock_call
    with_mocked_client do
      mock_response = Minitest::Mock.new
      mock_response.expect(:status, 200)
      mock_response.expect(:body, { status: 'success', data: [] })

      mock_api_client.expect(:call, mock_response, [:app_list, Modal::Client::AppListRequest])

      response = mock_api_client.call(:app_list, Modal::Client::AppListRequest.new(
        environment_name: Modal::Config.environment_name
      ))

      assert_equal 200, response.status
      body = response.body
      assert_equal 'success', body[:status]
      assert_instance_of Array, body[:data]
      assert_empty body[:data]

      mock_api_client.verify
    end
  end
end
