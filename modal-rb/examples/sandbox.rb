require 'modal'

def main
  app = Modal::App.lookup("my-modal-app")
  puts "Found app: #{app.app_id}"
  image = app.image_from_registry("alpine:latest")
  puts "Using image: #{image.image_id}"
  sandbox = app.create_sandbox(image, timeout: 30000, cpu: 0.5, memory: 256, command: ["cat"])
  puts "Created sandbox with ID: #{sandbox.sandbox_id}"
  sandbox.stdin.write("Hello from Modal!")
  sandbox.stdin.close
  output = sandbox.stdout.read
  puts "Sandbox output: #{output}"
end

main
