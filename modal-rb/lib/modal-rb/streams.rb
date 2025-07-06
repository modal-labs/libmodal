require 'stringio'

module Modal
  class ModalReadStream
    def initialize(source_iterable)
      @source_iterable = source_iterable
    end

    def read_text
      chunks = []
      @source_iterable.each do |bytes|
        chunks << bytes.force_encoding('UTF-8') # Assume UTF-8 for text
      end
      chunks.join('')
    end

    def read_bytes
      chunks = []
      @source_iterable.each do |bytes|
        chunks << bytes
      end
      chunks.join('').bytes.pack('C*') # Concatenate and return as binary string
    end

    # Allow iteration over the stream
    def each(&block)
      @source_iterable.each(&block)
    end
  end

  class ModalWriteStream
    def initialize(sink_writable)
      @sink_writable = sink_writable
    end

    def write_text(text)
      @sink_writable.write(text.bytes.pack('C*')) # Convert string to bytes
    end

    def write_bytes(bytes)
      @sink_writable.write(bytes)
    end

    def close
      @sink_writable.close if @sink_writable.respond_to?(:close)
    end
  end
end
