require_relative 'test_helper'

class TestSecret < Minitest::Test
  def test_secret_from_name
    secret_id = "st-testsecret123"
    mock_modal_client.expect(:call, Modal::Proto::SecretGetOrCreateResponse.new(secret_id: secret_id), [:secret_get_or_create, MiniTest::Any])

    secret = Modal::Secret.from_name("libmodal-test-secret")
    assert_equal secret_id, secret.secret_id
    assert_match(/^st-/, secret.secret_id)
    mock_modal_client.verify
  end

  def test_secret_from_name_not_found
    mock_modal_client.expect(:call, nil, [:secret_get_or_create, MiniTest::Any]) do
      raise Modal::NotFoundError.new("Secret 'missing-secret' not found")
    end

    assert_raises Modal::NotFoundError do
      Modal::Secret.from_name("missing-secret")
    end
    mock_modal_client.verify
  end

  def test_secret_from_name_with_required_keys
    secret_id = "st-testsecret-with-keys"
    mock_modal_client.expect(:call, Modal::Proto::SecretGetOrCreateResponse.new(secret_id: secret_id), [:secret_get_or_create, MiniTest::Any])

    secret = Modal::Secret.from_name("libmodal-test-secret", required_keys: ["a", "b", "c"])
    assert_equal secret_id, secret.secret_id
    mock_modal_client.verify
  end

  def test_secret_from_name_missing_required_key
    mock_modal_client.expect(:call, nil, [:secret_get_or_create, MiniTest::Any]) do
      raise Modal::NotFoundError.new("Secret is missing key(s): missing-key")
    end

    assert_raises Modal::NotFoundError do
      Modal::Secret.from_name("libmodal-test-secret", required_keys: ["a", "b", "c", "missing-key"])
    end
    mock_modal_client.verify
  end
end
