require_relative 'test_helper'

class TestApp < Minitest::Test
  def test_app_lookup
    app_id = "ap-testapp123"
    mock_modal_client.expect(:call, Modal::Client::AppGetOrCreateResponse.new(app_id: app_id), [:app_get_or_create, Minitest::Mock::ANY])

    app = Modal::App.lookup("libmodal-test", create_if_missing: true)
    assert_equal app_id, app.app_id
    mock_modal_client.verify
  end

  def test_app_lookup_not_found
    mock_modal_client.expect(:call, [:app_get_or_create, :any]) do
      raise Modal::NotFoundError.new("App 'nonexistent-app' not found")
    end

    assert_raises Modal::NotFoundError do
      Modal::App.lookup("nonexistent-app")
    end
    mock_modal_client.verify
  end

  def test_create_sandbox
    image_id = "im-testimage123"
    sandbox_id = "sb-testsandbox123"
    app_id = "ap-testapp123"

    mock_app = Modal::App.new(app_id)
    mock_image = Modal::Image.new(image_id)

    mock_modal_client.expect(:call, Modal::Client::SandboxCreateResponse.new(sandbox_id: sandbox_id), [:sandbox_create, Minitest::Mock::ANY])

    sandbox = mock_app.create_sandbox(mock_image)
    assert_equal sandbox_id, sandbox.sandbox_id
    mock_modal_client.verify
  end
end
