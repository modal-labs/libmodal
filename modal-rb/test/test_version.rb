require_relative 'test_helper'

class TestVersion < Minitest::Test
  def test_version_is_modal_rb
    assert Modal::VERSION
  end
end
