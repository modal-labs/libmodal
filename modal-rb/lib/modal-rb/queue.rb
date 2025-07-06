module Modal
  class Queue
    QUEUE_INITIAL_PUT_BACKOFF = 0.1 # seconds
    QUEUE_DEFAULT_PARTITION_TTL = 24 * 3600 # seconds

    attr_reader :queue_id

    def initialize(queue_id)
      @queue_id = queue_id
    end

    def self.ephemeral
      request = Modal::Proto::QueueGetOrCreateRequest.new(
        object_creation_type: Modal::Proto::ObjectCreationType::OBJECT_CREATION_TYPE_EPHEMERAL_AUTO_TEARDOWN
      )
      resp = Modal.client.call(:queue_get_or_create, request)
      new(resp.queue_id)
    end

    def self.lookup(name, options = {})
      create_if_missing = options[:create_if_missing] || false
      environment = options[:environment]

      request = Modal::Proto::QueueGetOrCreateRequest.new(
        deployment_name: name,
        environment_name: Config.environment_name(environment),
        object_creation_type: create_if_missing ? Modal::Proto::ObjectCreationType::OBJECT_CREATION_TYPE_CREATE_IF_MISSING : Modal::Proto::ObjectCreationType::OBJECT_CREATION_TYPE_UNSPECIFIED
      )

      resp = Modal.client.call(:queue_get_or_create, request)
      new(resp.queue_id)
    rescue StandardError => e
      # Basic validation, similar to JS client's test
      if name.include?(" ") || name.include?("/") || name.length > 64
        raise InvalidError.new("Invalid queue name: '#{name}'")
      end
      raise e # Re-raise original error if not a known invalid name pattern
    end

    def close_ephemeral
      request = Modal::Proto::QueueDeleteRequest.new(queue_id: @queue_id)
      Modal.client.call(:queue_delete, request)
    end

    def put(item, options = {})
      put_many([item], options)
    end

    def put_many(items, options = {})
      serialized_items = items.map { |i| Pickle.dumps(i) }
      request = Modal::Proto::QueuePutItemsRequest.new(
        queue_id: @queue_id,
        items: serialized_items.map { |data| Modal::Proto::QueueItem.new(value: data) },
        partition_key: Queue.validate_partition_key(options[:partition]),
        timeout: options[:timeout] ? options[:timeout] / 1000 : nil # Convert ms to seconds
      )
      Modal.client.call(:queue_put_items, request)
    rescue GRPC::BadStatus => e
      if e.code == GRPC::Core::StatusCodes::RESOURCE_EXHAUSTED
        raise QueueFullError.new("Queue is full: #{e.details}")
      else
        raise e
      end
    end

    def get(options = {})
      get_many(1, options).first
    end

    def get_many(n, options = {})
      request = Modal::Proto::QueueGetItemsRequest.new(
        queue_id: @queue_id,
        n: n,
        partition_key: Queue.validate_partition_key(options[:partition]),
        timeout: options[:timeout] ? options[:timeout] / 1000 : nil # Convert ms to seconds
      )
      resp = Modal.client.call(:queue_get_items, request)
      resp.items.map { |item| Pickle.loads(item.value) }
    rescue GRPC::BadStatus => e
      if e.code == GRPC::Core::StatusCodes::UNAVAILABLE || e.code == GRPC::Core::StatusCodes::DEADLINE_EXCEEDED
        raise QueueEmptyError.new("Queue is empty or timed out: #{e.details}")
      else
        raise e
      end
    end

    def len(options = {})
      if options[:partition] && options[:total]
        raise InvalidError.new("Partition must be null when requesting total length.")
      end

      request = Modal::Proto::QueueLenRequest.new(
        queue_id: @queue_id,
        partition_key: Queue.validate_partition_key(options[:partition]),
        total: options[:total] || false
      )
      resp = Modal.client.call(:queue_len, request)
      resp.len
    end

    def iterate(options = {})
      # This is a simplified iterator.
      # A full implementation would involve streaming gRPC calls and managing `last_entry_id`.
      # For now, it fetches all available items up to a timeout.
      last_entry_id = nil
      item_poll_timeout = options[:item_poll_timeout] || 0 # milliseconds
      validated_partition_key = Queue.validate_partition_key(options[:partition])

      enum_for(:each_item_in_queue, last_entry_id, item_poll_timeout, validated_partition_key)
    end

    private

    def each_item_in_queue(last_entry_id, item_poll_timeout, validated_partition_key)
      fetch_deadline = Time.now + item_poll_timeout / 1000.0 # Convert ms to seconds
      max_poll_duration = 30.0 # seconds

      loop do
        poll_duration = [0.0, [max_poll_duration, fetch_deadline - Time.now].min].max

        request = Modal::Proto::QueueNextItemsRequest.new(
          queue_id: @queue_id,
          partition_key: validated_partition_key,
          item_poll_timeout: poll_duration,
          last_entry_id: last_entry_id || ""
        )
        response = Modal.client.call(:queue_next_items, request)

        if response.items && response.items.any?
          response.items.each do |item|
            yield Pickle.loads(item.value)
            last_entry_id = item.entry_id
          end
          fetch_deadline = Time.now + item_poll_timeout / 1000.0
        elsif Time.now > fetch_deadline
          break
        end
      end
    end

    def self.validate_partition_key(key)
      key || "" # Return empty string if nil, as per JS client's behavior
    end
  end
end
