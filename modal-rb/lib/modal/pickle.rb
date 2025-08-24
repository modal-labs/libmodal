require 'pycall'
require 'json'

module Pickle
  class PickleError < StandardError; end

  def self.pickle_module
    @pickle_module ||= PyCall.import_module('pickle')
  end

  def self.json_module
    @json_module ||= PyCall.import_module('json')
  end

  def self.load(data)
    begin
      pickle_data = data.respond_to?(:read) ? data.read : data
      python_obj = pickle_module.loads(pickle_data)
      json_str = json_module.dumps(python_obj)
      JSON.parse(json_str.to_s)
    rescue => e
      raise PickleError, "Failed to load pickle data: #{e.message}"
    end
  end

  def self.dumps(obj)
    begin
      json_str = JSON.generate(obj)
      python_obj = json_module.loads(json_str)
      pickle_module.dumps(python_obj).to_s
    rescue => e
      raise PickleError, "Failed to dump object to pickle: #{e.message}"
    end
  end

  def self.dump(obj, file)
    begin
      pickled_data = dumps(obj)
      if file.respond_to?(:write)
        file.write(pickled_data)
      else
        File.open(file, 'wb') do |f|
          f.write(pickled_data)
        end
      end
    rescue => e
      raise PickleError, "Failed to dump object to file: #{e.message}"
    end
  end
end
