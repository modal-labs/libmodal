require_relative 'test_helper'

class TestConfig < Minitest::Test
  def test_profile_is_modal_rb_test
    assert_equal 'modal-rb-test-env', Modal::Config.environment_name
  end

  def test_get_profile
    profile = Modal::Config.get_profile('modal-rb-test')
    assert_equal 'https://localhost:50051', profile[:server_url]
    assert_equal 'test-token-id', profile[:token_id]
    assert_equal 'test-token-secret', profile[:token_secret]
    assert_equal 'modal-rb-test-env', profile[:environment];
  end


  def test_get_profile_not_found
    assert_raises(RuntimeError) do
      Modal::Config.get_profile('non-existent-profile')
    end
  end
end
