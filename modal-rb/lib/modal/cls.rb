module Modal
  class Cls
    attr_reader :service_function_id, :schema, :method_names

    def initialize(service_function_id, schema, method_names)
      @service_function_id = service_function_id
      @schema = schema
      @method_names = method_names
    end

    def self.lookup(app_name, name, options = {})
      service_function_name = "#{name}.*"
      request = Modal::Client::FunctionGetRequest.new(
        app_name: app_name,
        object_tag: service_function_name,
        environment_name: Config.environment_name(options[:environment])
      )

      service_function = Modal.client.call(:function_get, request)
      parameter_info = service_function.handle_metadata&.class_parameter_info
      schema = parameter_info&.schema || []

      if schema.any? && parameter_info.format != Modal::Client::ClassParameterInfo_ParameterSerializationFormat::PARAM_SERIALIZATION_FORMAT_PROTO
        raise "Unsupported parameter format: #{parameter_info.format}"
      end

      method_names = if service_function.handle_metadata&.method_handle_metadata
                       service_function.handle_metadata.method_handle_metadata.keys
                     else
                       raise "Cls requires Modal deployments using client v0.67 or later."
                     end

      new(service_function.function_id, schema, method_names)
    rescue NotFoundError => e
      raise NotFoundError.new("Class '#{app_name}/#{name}' not found")
    end

    def instance(params = {})
      function_id = if @schema.empty?
                      @service_function_id
                    else
                      bind_parameters(params)
                    end

      methods = {}
      @method_names.each do |name|
        methods[name] = Function_.new(function_id, name)
      end
      ClsInstance.new(methods)
    end

    private

    def bind_parameters(params)
      serialized_params = encode_parameter_set(@schema, params)
      request = Modal::Client::FunctionBindParamsRequest.new(
        function_id: @service_function_id,
        serialized_params: serialized_params
      )
      bind_resp = Modal.client.call(:function_bind_params, request)
      bind_resp.bound_function_id
    end

    def encode_parameter_set(schema, params)
      encoded_params = schema.map do |param_spec|
        encode_parameter(param_spec, params[param_spec.name.to_sym])
      end

      encoded_params.sort_by!(&:name)
      Modal::Client::ClassParameterSet.encode(parameters: encoded_params).to_proto
    end

    def encode_parameter(param_spec, value)
      name = param_spec.name
      param_type = param_spec.type
      param_value = Modal::Client::ClassParameterValue.new(name: name, type: param_type)

      case param_type
      when Modal::Client::ParameterType::PARAM_TYPE_STRING
        if value.nil? && param_spec.has_default
          value = param_spec.string_default || ""
        end
        unless value.is_a?(String)
          raise "Parameter '#{name}' must be a string"
        end
        param_value.string_value = value
      when Modal::Client::ParameterType::PARAM_TYPE_INT
        if value.nil? && param_spec.has_default
          value = param_spec.int_default || 0
        end
        unless value.is_a?(Integer)
          raise "Parameter '#{name}' must be an integer"
        end
        param_value.int_value = value
      when Modal::Client::ParameterType::PARAM_TYPE_BOOL
        if value.nil? && param_spec.has_default
          value = param_spec.bool_default || false
        end
        unless [true, false].include?(value)
          raise "Parameter '#{name}' must be a boolean"
        end
        param_value.bool_value = value
      when Modal::Client::ParameterType::PARAM_TYPE_BYTES
        if value.nil? && param_spec.has_default
          value = param_spec.bytes_default || ""
        end
        unless value.is_a?(String)
          raise "Parameter '#{name}' must be a byte array (String in Ruby)"
        end
        param_value.bytes_value = value.bytes.pack('C*')
      else
        raise "Unsupported parameter type: #{param_type}"
      end
      param_value
    end
  end

  class ClsInstance
    def initialize(methods)
      @methods = methods
    end

    def method(name)
      func = @methods[name.to_s]
      unless func
        raise NotFoundError.new("Method '#{name}' not found on class")
      end
      func
    end
  end
end
