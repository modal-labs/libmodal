require_relative 'test_helper'

class TestQueue < Minitest::Test
  def setup
    super
    # Stub Pickle.dumps and Pickle.loads for these tests
    Modal::Pickle.stub :dumps, ->(obj) { obj.to_s.bytes.pack('C*') } do
      Modal::Pickle.stub :loads, ->(bytes) { bytes.force_encoding('UTF-8') } do
        yield
      end
    end
  end

  def test_queue_ephemeral
    queue_id = "qu-ephemeral123"
    mock_modal_client.expect(:call, Modal::Proto::QueueGetOrCreateResponse.new(queue_id: queue_id), [:queue_get_or_create, MiniTest::Any])
    mock_modal_client.expect(:call, Modal::Proto::QueuePutItemsResponse.new, [:queue_put_items, MiniTest::Any])
    mock_modal_client.expect(:call, Modal::Proto::QueueLenResponse.new(len: 1), [:queue_len, MiniTest::Any])
    mock_modal_client.expect(:call, Modal::Proto::QueueGetItemsResponse.new(items: [Modal::Proto::QueueItem.new(value: "123".bytes.pack('C*'))]), [:queue_get_items, MiniTest::Any])
    mock_modal_client.expect(:call, Modal::Proto::QueueDeleteResponse.new, [:queue_delete, MiniTest::Any])

    queue = Modal::Queue.ephemeral
    assert_equal queue_id, queue.queue_id
    queue.put(123)
    assert_equal 1, queue.len
    assert_equal "123", queue.get # Note: Pickle mock returns string
    queue.close_ephemeral
    mock_modal_client.verify
  end

  def test_queue_invalid_name
    ["has space", "has/slash", "a" * 65].each do |name|
      assert_raises Modal::InvalidError do
        Modal::Queue.lookup(name)
      end
    end
  end

  def test_queue_put_and_get_many
    queue_id = "qu-many123"
    mock_modal_client.expect(:call, Modal::Proto::QueueGetOrCreateResponse.new(queue_id: queue_id), [:queue_get_or_create, MiniTest::Any])
    mock_modal_client.expect(:call, Modal::Proto::QueuePutItemsResponse.new, [:queue_put_items, MiniTest::Any])
    mock_modal_client.expect(:call, Modal::Proto::QueueLenResponse.new(len: 3), [:queue_len, MiniTest::Any])
    mock_modal_client.expect(:call, Modal::Proto::QueueGetItemsResponse.new(items: [
      Modal::Proto::QueueItem.new(value: "1".bytes.pack('C*')),
      Modal::Proto::QueueItem.new(value: "2".bytes.pack('C*')),
      Modal::Proto::QueueItem.new(value: "3".bytes.pack('C*'))
    ]), [:queue_get_items, MiniTest::Any])
    mock_modal_client.expect(:call, Modal::Proto::QueueDeleteResponse.new, [:queue_delete, MiniTest::Any])


    queue = Modal::Queue.ephemeral
    queue.put_many([1, 2, 3])
    assert_equal 3, queue.len
    assert_equal ["1", "2", "3"], queue.get_many(3) # Note: Pickle mock returns strings
    queue.close_ephemeral
    mock_modal_client.verify
  end

  def test_queue_iterate
    queue_id = "qu-iterate123"
    mock_modal_client.expect(:call, Modal::Proto::QueueGetOrCreateResponse.new(queue_id: queue_id), [:queue_get_or_create, MiniTest::Any])
    mock_modal_client.expect(:call, Modal::Proto::QueuePutItemsResponse.new, [:queue_put_items, MiniTest::Any])

    # Mock the iterate calls
    mock_modal_client.expect(:call, Modal::Proto::QueueNextItemsResponse.new(
      items: [
        Modal::Proto::QueueItem.new(value: "0".bytes.pack('C*'), entry_id: "0-0"),
        Modal::Proto::QueueItem.new(value: "1".bytes.pack('C*'), entry_id: "0-1")
      ],
      entry_id: "0-1"
    ), [:queue_next_items, MiniTest::Any])
    mock_modal_client.expect(:call, Modal::Proto::QueueNextItemsResponse.new(
      items: [
        Modal::Proto::QueueItem.new(value: "2".bytes.pack('C*'), entry_id: "0-2")
      ],
      entry_id: "0-2"
    ), [:queue_next_items, MiniTest::Any])
    mock_modal_client.expect(:call, Modal::Proto::QueueNextItemsResponse.new(items: [], eof: true), [:queue_next_items, MiniTest::Any])
    mock_modal_client.expect(:call, Modal::Proto::QueueDeleteResponse.new, [:queue_delete, MiniTest::Any])


    queue = Modal::Queue.ephemeral
    queue.put_many([0, 1, 2])

    results = []
    queue.iterate(item_poll_timeout: 100).each do |item| # Small timeout for test
      results << item
    end
    assert_equal ["0", "1", "2"], results # Note: Pickle mock returns strings
    queue.close_ephemeral
    mock_modal_client.verify
  end
end
