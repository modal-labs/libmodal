# Developing

We welcome contributions to `modal-rb`! This guide will help you set up your development environment, understand the project structure, and contribute effectively.

## Table of Contents

1. [Getting Started](#getting-started)
    1. [Prerequisites](#prerequisites)
    1. [Setting up the Development Environment](#set)
1. [Protobuf Generation](protobuf-generation)
1. [Running Tests](running-tests)
1. [Code Style and Linting](code-style-and-linting)
1. [Submitting Changes](submitting-changes)
1. [License](license)

### Getting Started
#### Prerequisites

Before you begin, ensure you have the following installed:

- Ruby: Version 3.0 or later. We recommend using a Ruby version manager like rbenv or rvm to manage your Ruby installations.

- Bundler: The Ruby gem dependency manager. Install with gem install bundler.

- grpc-tools: Required for generating Ruby protobuf bindings. Install with gem install grpc-tools.

- Git: For version control.

- modal-client repository: The Ruby client uses protobuf definitions from the modal-client repository. You should clone this repository as a sibling directory to modal-rb (e.g., ../modal-client). If your setup differs, you'll need to adjust the PROTO_DIR and MODAL_CLIENT_PROTO_PATH variables in the Rakefile.

#### Setting up the Development Environment

1. Clone the repository:

    ```
    git clone https://github.com/modal-labs/libmodal.git
    cd modal-rb
    ```

1. Install dependencies:
    ```
    bundle install
    ```

    This will install all the gems listed in the Gemfile, including development and test dependencies.

1. Generate Protobufs:

    The Ruby client communicates with the Modal API using gRPC, which relies on Protocol Buffers. You need to generate the Ruby classes from the .proto files.

    ```
    bundle exec rake generate_protos
    ```

    This will create Ruby files in the proto/modal_proto directory. These files should be committed to the repository.

### Protobuf Generation
The Rakefile contains a `generate_protos` task that uses `grpc_tools_ruby_protoc` to compile the `.proto` files from `modal-client/modal_proto` into Ruby classes.

To regenerate protobufs:

```
bundle exec rake generate_protos
```

Always run this command after pulling changes from modal-client that modify .proto files!

### Running Tests

modal-rb uses Minitest for its test suite.

To run all tests:

```
bundle exec rake test
```

To run a specific test file:

```
bundle exec ruby test/test_app.rb
```

#### Mocking gRPC Calls

Tests extensively use `mocha/minitest` and `webmock/minitest` to mock gRPC and HTTP calls to the Modal API. This allows for fast, isolated, and reliable unit tests without needing a running Modal backend.

`mock_modal_client`: A helper method in test/test_helper.rb that provides a MiniTest::Mock instance for stubbing gRPC client calls.

`WebMock.disable_net_connect!`: By default, WebMock prevents any actual HTTP connections. If you need to make real HTTP requests in a test (e.g., for testing blob uploads), you can selectively enable them.

### Code Style and Linting
We use RuboCop to enforce code style and best practices.

To run RuboCop:

```
bundle exec rubocop
```

Please ensure your code passes RuboCop checks before submitting a pull request. You can automatically fix some offenses:

```
bundle exec rubocop -a
```

### Submitting Changes
Fork the repository on GitHub.

Create a new branch for your feature or bug fix:

git checkout -b feature/your-feature-name

Make your changes.

Write tests for your changes. Ensure all existing tests pass.

Run RuboCop and fix any style issues.

Commit your changes with a clear and concise commit message.

Push your branch to your forked repository.

Open a Pull Request against the main branch of the modal-rb repository.

Provide a descriptive title and detailed description of your changes.

Reference any related issues.

### License

See [LICENSE](../LICENSE)

