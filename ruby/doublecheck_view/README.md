# Doublecheck View

doublecheck_view is a Ruby gem that makes it easy to validate doublecheck views
during automated tests. See https://github.com/jackc/doublecheck for more
information about the doublecheck view pattern.

## Installation

Add this line to your application's Gemfile:

```ruby
gem 'doublecheck_view'
```

And then execute:

    $ bundle

Or install it yourself as:

    $ gem install doublecheck_view

## Usage

You can check for errors manually at any time. The following code will select from every view in the `doublecheck` schema and `check_result` will only be valid if no rows are returned from any view.

```ruby
doublecheck = DoublecheckView::Doublecheck.new
check_result = doublecheck.check
assert check_result.valid?, check_result.errors
```

However, the preferred approach is to run doublecheck after every test that uses
the database.

```ruby
# Inside a MiniTest or ActiveSupport::TestCase class body

  def before_teardown
    doublecheck = DoublecheckView::Doublecheck.new
    check_result = doublecheck.check
    assert check_result.valid?, check_result.errors
    super
  end
```

Or for RSpec:

```ruby
RSpec.configure do |config|
  # ...
  config.before(:each) do
    doublecheck = DoublecheckView::Doublecheck.new
    check_result = doublecheck.check
    expect(check_result).to be_valid, "doublecheck views: #{check_result.errors}"
  end
end
```

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/jackc/doublecheck.


## License

The gem is available as open source under the terms of the [MIT License](http://opensource.org/licenses/MIT).

