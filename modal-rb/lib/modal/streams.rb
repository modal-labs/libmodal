require 'stringio'

module Modal
  class ModalReadStream
    def initialize(source_iterable)
      @source_iterable = source_iterable
    end

    def read
      read_text
    end

    def read_text
      chunks = []
      @source_iterable.each do |bytes|
        chunks << bytes.dup.force_encoding('UTF-8')
      end
      chunks.join('')
    end

    def read_bytes
      chunks = []
      @source_iterable.each do |bytes|
        chunks << bytes.dup
      end
      chunks.join('').bytes.pack('C*')
    end

    def each(&block)
      @source_iterable.each(&block)
    end

    def close
      @source_iterable.close if @source_iterable.respond_to?(:close)
    end
  end

  class ModalWriteStream
    def initialize(sink_writable)
      @sink_writable = sink_writable
    end


    def write(data)
      if data.is_a?(String)
        if data.encoding == Encoding::BINARY
          write_bytes(data)
        else
          write_text(data)
        end
      else

        write_bytes(data.to_s)
      end
    end

    def write_text(text)
      @sink_writable.write(text)
    end

    def write_bytes(bytes)
      @sink_writable.write(bytes)
    end

    def close
      @sink_writable.close if @sink_writable.respond_to?(:close)
    end

    def flush
      @sink_writable.flush if @sink_writable.respond_to?(:flush)
    end
  end
end
