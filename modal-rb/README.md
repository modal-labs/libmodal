# modal-rb

modal-rb is the official Ruby client library for interacting with Modal, a cloud platform for running your code at scale. This library allows you to define, deploy, and manage Modal applications, images, secrets, functions, and sandboxes directly from your Ruby applications.

##  Table of Contents
1. [Features](#features)
1. [Installation](#installation)
1. [Configuration](#configuration)
1. [Usage](#usage)
    1. [Connecting to Modal](#connecting-to-modal)
    1. [Managing Apps](#managing-apps)
    1. [Working with Images](#working-with-images)
    1. [Handling Secrets](#handling-secrets)
    1. [Calling Functions](#calling-functions)
    1. [Using Sandboxes](#using-sandboxes)
    1. [Queues](#queues)
    1. [Classes (CLSs)](#classes-clss)
1. [Contributing](#contributing)
1. [License](#license)

### Features
App Management: Create, lookup, and manage Modal applications.

Image Building: Define and build Docker images for your Modal functions, including support for public registries and AWS ECR.

Secret Management: Securely manage environment variables and credentials.

Function Invocation: Call Modal functions synchronously and asynchronously.

Sandbox Interaction: Create and interact with isolated execution environments (sandboxes) for running arbitrary commands, including stdin/stdout/stderr and filesystem access.

Queueing: Utilize distributed queues for inter-service communication.

Classes (CLSs): Interact with Modal deployed classes and their methods.

Error Handling: Custom error classes for better debugging.

### Installation
Prerequisites:

- Ruby (3.0 or later recommended)
- Bundler

Add to your Gemfile:

```ruby
# Gemfile
source "https://rubygems.org"

gem "modal-rb", "~> 0.1.0" # Or the latest version
gem "grpc", "~> 1.60"
gem "google-protobuf", "~> 3.25"
gem "toml-rb", "~> 2.2" # For config parsing
```

Install Gems:

```
bundle install
```

Generate Protobufs:
This library relies on gRPC protobuf definitions. You need to generate the Ruby classes from the .proto files. Ensure you have the modal-client repository cloned as a sibling directory or adjust Rakefile accordingly.

```
bundle exec rake generate_protos
```

This command will create the necessary Ruby files in proto/modal_proto.

### Configuration
modal-rb reads configuration from ~/.modal.toml and environment variables.

Example `~/.modal.toml`

```
[default]
server_url = "https://api.modal.com:443"
token_id = "your-modal-token-id"
token_secret = "your-modal-token-secret"
environment = "prod"
active = true

[dev]
server_url = "http://localhost:50051"
token_id = "your-dev-token-id"
token_secret = "your-dev-token-secret"
environment = "dev"
```

Environment Variables

You can override configuration values using environment variables:
```
MODAL_PROFILE: Specifies the profile to use from ~/.modal.toml (e.g., dev).

MODAL_SERVER_URL: Overrides server_url.

MODAL_TOKEN_ID: Overrides token_id.

MODAL_TOKEN_SECRET: Overrides token_secret.

MODAL_ENVIRONMENT: Overrides environment.

MODAL_IMAGE_BUILDER_VERSION: Overrides image_builder_version.
```

### Usage
#### Connecting to Modal

The Modal::Client is automatically initialized based on your configuration. You can access it via Modal.client.

```
require 'modal-rb'

# The client is automatically configured on load
# Modal.client

#### Managing Apps

require 'modal-rb'

# Lookup an existing app
app = Modal::App.lookup("my-existing-app")
puts "Found app: #{app.app_id}"

# Create an app if it doesn't exist
new_app = Modal::App.lookup("my-new-app", create_if_missing: true)
puts "Created/Found new app: #{new_app.app_id}"
```

#### Working with Images

```
require 'modal-rb'

# Assume 'my_app' is an instance of Modal::App
my_app = Modal::App.lookup("my-app", create_if_missing: true)

# Create an image from a public registry
image = my_app.image_from_registry("python:3.10-slim-buster")
puts "Created image from registry: #{image.image_id}"

# Create an image from AWS ECR using a Modal Secret
# First, ensure you have a secret configured in Modal for AWS ECR access
# secret = Modal::Secret.from_name("my-aws-ecr-secret")
# ecr_image = my_app.image_from_aws_ecr("123456789012.dkr.ecr.us-east-1.amazonaws.com/my-repo:latest", secret)
# puts "Created image from AWS ECR: #{ecr_image.image_id}"
```

#### Handling Secrets

```
require 'modal-rb'

# Lookup a secret by name
secret = Modal::Secret.from_name("my-api-keys")
puts "Found secret: #{secret.secret_id}"

# Lookup a secret and ensure specific keys are present
begin
  secret_with_keys = Modal::Secret.from_name("my-api-keys", required_keys: ["API_KEY", "SECRET_TOKEN"])
  puts "Found secret with required keys: #{secret_with_keys.secret_id}"
rescue Modal::NotFoundError => e
  puts "Error: #{e.message}"
end
```

#### Calling Functions

```
require 'modal-rb'

# Assume 'my_app' is an instance of Modal::App
my_app = Modal::App.lookup("my-app", create_if_missing: true)

# Lookup a deployed function
# In a real scenario, this function would be deployed on Modal
# For example, if you have a Python app `my_app.py` with `@stub.function()`
# def my_function(x): return x * 2
#
# function_ = Modal::Function_.lookup("my-app", "my_function")

# Example of a mock function call (replace with actual lookup)
# For demonstration, we'll simulate a function that echoes its input
class MockFunction
  def initialize(id); @function_id = id; end
  def remote(args, kwargs)
    puts "Simulating remote call with args: #{args}, kwargs: #{kwargs}"
    # In a real scenario, this would involve gRPC calls to Modal
    if kwargs[:s]
      "output: #{kwargs[:s]}"
    elsif args.first
      "output: #{args.first}"
    else
      "output: no input"
    end
  end
  def spawn(args, kwargs)
    puts "Simulating spawn call with args: #{args}, kwargs: #{kwargs}"
    # Return a mock FunctionCall object
    Modal::FunctionCall.from_id("fc-mock-spawn-123")
  end
end

# Replace with actual lookup once you have deployed functions
# function_ = Modal::Function_.lookup("my-app", "my_function")
function_ = MockFunction.new("fu-mock-123")

# Synchronous call
result = function_.remote([], s: "hello from Ruby")
puts "Function result: #{result}" # Expected: "output: hello from Ruby"

# Asynchronous call
function_call = function_.spawn([], s: "async hello")
puts "Function call ID: #{function_call.function_call_id}"

# Get result of asynchronous call
async_result = function_call.get
puts "Async function result: #{async_result}" # Expected: "output: async hello"

# Cancel an asynchronous call
# function_call.cancel
```

#### Using Sandboxes

```
require 'modal-rb'

# Assume 'my_app' is an instance of Modal::App
my_app = Modal::App.lookup("my-app", create_if_missing: true)
image = my_app.image_from_registry("debian:stable-slim")

# Create a sandbox
sandbox = my_app.create_sandbox(image, command: ["sleep", "60"])
puts "Created sandbox: #{sandbox.sandbox_id}"

begin
  # Execute a command in the sandbox
  process = sandbox.exec(["echo", "Hello from sandbox!"])
  puts "Sandbox stdout: #{process.stdout.read_text}"
  puts "Sandbox stderr: #{process.stderr.read_text}"
  exit_code = process.wait
  puts "Sandbox command exited with code: #{exit_code}"

  # Write to stdin and read from stdout (e.g., using 'cat')
  cat_process = sandbox.exec(["cat"])
  cat_process.stdin.write_text("This is stdin input.\n")
  cat_process.stdin.close
  puts "Cat stdout: #{cat_process.stdout.read_text}"
  cat_process.wait

  # Filesystem operations
  file_handle = sandbox.open("/tmp/test_file.txt", "w")
  file_handle.write("Hello from Ruby to file!")
  file_handle.close

  read_handle = sandbox.open("/tmp/test_file.txt", "r")
  content = read_handle.read.force_encoding('UTF-8') # Ensure proper encoding
  puts "File content: #{content}"
  read_handle.close

rescue Modal::ModalError => e
  puts "Error interacting with sandbox: #{e.message}"
ensure
  # Terminate the sandbox when done
  sandbox.terminate
  puts "Sandbox terminated."
end
```

#### Queues

```
require 'modal-rb'

# Create an ephemeral queue
queue = Modal::Queue.ephemeral
puts "Created ephemeral queue: #{queue.queue_id}"

# Put items
queue.put("hello")
queue.put_many([1, 2, 3])
puts "Queue length: #{queue.len}"

# Get items
item1 = queue.get # "hello" (from Pickle mock)
puts "Got item: #{item1}"

items = queue.get_many(2) # [1, 2] (from Pickle mock)
puts "Got items: #{items}"

# Iterate over queue (simplified, actual streaming would be more complex)
puts "Iterating over remaining items:"
queue.iterate.each do |item|
  puts "- #{item}"
end

# Close ephemeral queue
queue.close_ephemeral
puts "Ephemeral queue closed."

# Lookup a named queue (create if missing)
named_queue = Modal::Queue.lookup("my-named-queue", create_if_missing: true)
puts "Named queue ID: #{named_queue.queue_id}"
# You can now use named_queue like the ephemeral queue
```

#### Classes (CLSs)

```
require 'modal-rb'

# Assume 'my_app' is an instance of Modal::App
my_app = Modal::App.lookup("my-app", create_if_missing: true)

# Lookup a deployed class
# For example, if you have a Python app `my_app.py` with `@stub.cls()`
# class MyClass:
#     def __init__(self, name): self.name = name
#     @modal.method()
#     def greet(self): return f"Hello, {self.name}!"
#
# my_cls = Modal::Cls.lookup("my-app", "MyClass")

# Example of a mock class (replace with actual lookup)
class MockCls
  def initialize(id); @service_function_id = id; end
  def instance(params = {})
    mock_method = Class.new do
      define_method(:remote) do |args, kwargs|
        "output: #{params[:name] || 'default'}"
      end
    end.new
    Modal::ClsInstance.new("greet" => mock_method)
  end
end

# Replace with actual lookup once you have deployed classes
# my_cls = Modal::Cls.lookup("my-app", "MyClass")
my_cls = MockCls.new("fu-mock-cls-123")


# Create an instance of the class (with optional parameters)
instance = my_cls.instance(name: "Ruby User")

# Call a method on the instance
result = instance.method("greet").remote
puts "Class method result: #{result}" # Expected: "output: Ruby User"
```

### Contributing
See [DEVELOPING.md](./DEVELOPING.md) for details on setting up your development environment, running tests, and contributing to the project. Contributions are welcome!

### License
See [LICENSE](../LICENSE)

