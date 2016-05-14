# Doublecheck Test

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

```ruby
doublecheck = DoublecheckView::Doublecheck.new
check_result = doublecheck.check
unless check_result.valid?
  puts check_result.errors
end
```

At the

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/jackc/doublecheck.


## License

The gem is available as open source under the terms of the [MIT License](http://opensource.org/licenses/MIT).

