$LOAD_PATH.unshift File.expand_path('../../lib', __FILE__)
require 'doublecheck_view'

require 'yaml'
require 'active_record'
database_config = YAML.load_file(File.expand_path("../database.yml", __FILE__))
ActiveRecord::Base.establish_connection database_config["test"]

require 'minitest/autorun'
