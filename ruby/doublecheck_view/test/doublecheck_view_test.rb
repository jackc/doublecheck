require 'test_helper'

class DoublecheckViewTest < Minitest::Test
  def test_that_it_has_a_version_number
    refute_nil ::DoublecheckView::VERSION
  end
end
