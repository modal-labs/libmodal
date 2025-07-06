require 'modal'

def main
  app = Modal::App.lookup("sandbox-filesystem-example", create_if_missing: true)
  puts "Found app: #{app.app_id}"

  image = app.image_from_registry("alpine:latest")
  puts "Using image: #{image.image_id}"

  sandbox = app.create_sandbox(image,
    timeout: 60000,
    cpu: 0.5,
    memory: 256,
    command: ["sleep", "100"]  # Keep it alive for 100 seconds
  )
  puts "Created sandbox with ID: #{sandbox.sandbox_id}"

  # Wait a moment for the sandbox to be ready
  sleep(2)

  # Create and write to a file
  puts "\n=== Writing to a file ==="
  file = sandbox.open("/tmp/hello.txt", "w")
  file.write("Hello from Modal Ruby!\n")
  file.write("This is line 2\n")
  file.flush
  file.close
  puts "✓ Created and wrote to /tmp/hello.txt"

  # Read the file back
  puts "\n=== Reading the file back ==="
  file = sandbox.open("/tmp/hello.txt", "r")
  content = file.read
  puts "File content:"
  puts content
  file.close

  # Create a directory and file structure
  puts "\n=== Creating directory structure ==="

  # Use exec to create directories (since we don't have mkdir method yet)
  result = sandbox.exec(["mkdir", "-p", "/tmp/mydir/subdir"])
  exit_code = result.wait
  puts "✓ Created directories (exit code: #{exit_code})"

  # Create a file in the subdirectory
  file = sandbox.open("/tmp/mydir/subdir/data.txt", "w")
  file.write("Data in subdirectory\n")
  file.close
  puts "✓ Created /tmp/mydir/subdir/data.txt"

  # List directory contents
  puts "\n=== Listing directory contents ==="
  result = sandbox.exec(["ls", "-la", "/tmp/mydir/subdir"])
  output = result.stdout.read
  puts "Directory listing:"
  puts output
  result.wait

  # Append to a file
  puts "\n=== Appending to file ==="
  file = sandbox.open("/tmp/hello.txt", "a")
  file.write("This is an appended line\n")
  file.close

  # Read it back
  file = sandbox.open("/tmp/hello.txt", "r")
  updated_content = file.read
  puts "Updated file content:"
  puts updated_content
  file.close

  # Work with binary data
  puts "\n=== Working with binary data ==="
  binary_data = [0x48, 0x65, 0x6c, 0x6c, 0x6f].pack('C*')  # "Hello" in binary
  file = sandbox.open("/tmp/binary.dat", "wb")
  file.write(binary_data)
  file.close

  # Read binary data back
  file = sandbox.open("/tmp/binary.dat", "rb")
  read_binary = file.read
  puts "Binary data read back: #{read_binary.unpack('C*').inspect}"
  file.close

  # Use the filesystem with a process
  puts "\n=== Using filesystem with processes ==="

  # Create a script file
  script_content = <<~SCRIPT
    #!/bin/sh
    echo "Script started"
    echo "Current directory: $(pwd)"
    echo "Files in /tmp:"
    ls -la /tmp/
    echo "Script finished"
  SCRIPT

  file = sandbox.open("/tmp/myscript.sh", "w")
  file.write(script_content)
  file.close

  # Make it executable and run it
  sandbox.exec(["chmod", "+x", "/tmp/myscript.sh"]).wait
  result = sandbox.exec(["/tmp/myscript.sh"])
  script_output = result.stdout.read
  puts "Script output:"
  puts script_output
  result.wait

  puts "\n=== Filesystem example completed ==="

  # Clean up
  sandbox.terminate
  puts "✓ Sandbox terminated"

rescue => e
  puts "Error: #{e.message}"
  puts e.backtrace if ENV['DEBUG']
ensure
  begin
    sandbox&.terminate
  rescue => e
    puts "Failed to terminate sandbox: #{e.message}"
  end
end

main
