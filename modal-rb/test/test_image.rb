require_relative 'test_helper'

class TestImage < Minitest::Test
  def test_image_from_registry
    app_id = "ap-testapp123"
    image_id = "im-testimage456"

    # Mock the image_get_or_create call to return a successful result
    mock_modal_client.expect(:call, Modal::Proto::ImageGetOrCreateResponse.new(
      image_id: image_id,
      result: Modal::Proto::GenericResult.new(status: Modal::Proto::GenericResult_GenericStatus::GENERIC_STATUS_SUCCESS)
    ), [:image_get_or_create, MiniTest::Any])

    # Mock the `sleep` call to prevent actual waiting in tests
    Modal::Image.stub :sleep, nil do
      app = Modal::App.new(app_id)
      image = app.image_from_registry("alpine:3.21")
      assert_equal image_id, image.image_id
    end
    mock_modal_client.verify
  end

  def test_image_from_aws_ecr
    app_id = "ap-testapp123"
    image_id = "im-ecrimage789"
    secret_id = "st-awssecret123"

    mock_secret = Modal::Secret.new(secret_id)

    # Mock the image_get_or_create call for AWS ECR
    mock_modal_client.expect(:call, Modal::Proto::ImageGetOrCreateResponse.new(
      image_id: image_id,
      result: Modal::Proto::GenericResult.new(status: Modal::Proto::GenericResult_GenericStatus::GENERIC_STATUS_SUCCESS)
    ), [:image_get_or_create, MiniTest::Any])

    Modal::Image.stub :sleep, nil do
      app = Modal::App.new(app_id)
      image = app.image_from_aws_ecr("my-private-registry/my-image:latest", mock_secret)
      assert_equal image_id, image.image_id
    end
    mock_modal_client.verify
  end

  def test_image_build_failure
    app_id = "ap-testapp123"
    image_id = "im-failedimage"

    # Mock the image_get_or_create call to return a failed result
    mock_modal_client.expect(:call, Modal::Proto::ImageGetOrCreateResponse.new(
      image_id: image_id,
      result: Modal::Proto::GenericResult.new(
        status: Modal::Proto::GenericResult_GenericStatus::GENERIC_STATUS_FAILURE,
        exception: "Build failed due to some error."
      )
    ), [:image_get_or_create, MiniTest::Any])

    Modal::Image.stub :sleep, nil do
      app = Modal::App.new(app_id)
      assert_raises(RuntimeError, "Image build for #{image_id} failed with the exception:\nBuild failed due to some error.") do
        app.image_from_registry("bad-image:latest")
      end
    end
    mock_modal_client.verify
  end
end
