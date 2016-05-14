require 'test_helper'

class DoublecheckTestTest < Minitest::Test
  def test_that_it_has_a_version_number
    refute_nil ::DoublecheckTest::VERSION
  end
end
