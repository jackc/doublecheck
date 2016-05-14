require 'test_helper'

class DoublecheckTestDoublecheckTest < Minitest::Test
  def test_it_gets_views
    expected_views = ["syntax error", "with_multiple_errors", "without_errors"]
    assert_equal expected_views, DoublecheckTest::Doublecheck.new.views
  end

  def test_check_without_errors_is_valid
    check_result = DoublecheckTest::Doublecheck.new.check views: ["without_errors"]
    assert check_result.valid?
  end

  def test_check_with_errors
    check_result = DoublecheckTest::Doublecheck.new.check
    refute check_result.valid?
  end
end
