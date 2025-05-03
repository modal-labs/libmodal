# libmodal: [Modal](https://modal.com) SDK Lite

Modal client libraries for JavaScript and Go. **(Alpha)**

This repository provides lightweight alternatives to the [Modal Python Library](https://github.com/modal-labs/modal-client). They let you start sandboxes (secure VMs), call Modal Functions, read or edit Volumes, and manage containers. However, they don't support deploying Modal Functions — those still need to be written in Python!

Each language in this repository has a library with similar features and API, so you can use Modal from any project.

## Setup

Make sure you've authenticated with Modal. You can either sign in with the Modal CLI `pip install modal && modal setup`, or in machine environments, set the following environment variables on your app:

```bash
# Replace these with your actual token!
export MODAL_TOKEN_ID=ak-NOTAREALTOKENSTRINGXYZ
export MODAL_TOKEN_SECRET=as-FAKESECRETSTRINGABCDEF
```

Then you're ready to add the Modal SDK to your project.

### JavaScript (`modal-js/`)

Install this in any server-side Node.js / Deno / Bun project.

```bash
npm install modal
```

Examples:

- [Call a deployed function](./modal-js/examples/function-call.ts)
- [Call a deployed cls](./modal-js/examples/cls-call.ts)
- [Create a sandbox](./modal-js/examples/sandbox.ts)
- [Execute sandbox commands](./modal-js/examples/sandbox-exec.ts)

### Go (`modal-go/`)

First, use `go get` to install the latest version of the library.

```bash
go get -u github.com/modal-labs/libmodal/modal-go
```

Next, include Modal in your application:

```go
import "github.com/modal-labs/libmodal/modal-go"
```

Examples:

- [Call a deployed function](./modal-go/examples/function-call/main.go)
- [Call a deployed cls](./modal-go/examples/cls-call/main.go)
- [Create a sandbox](./modal-go/examples/sandbox/main.go)
- [Execute sandbox commands](./modal-go/examples/sandbox-exec/main.go)

### Python

If you're using Python, please use the [Modal Python Library](https://github.com/modal-labs/modal-client), which is the main SDK and a separate project.

## Technical details

`libmodal` is a cross-language client SDK for Modal. However, it does not have all the features of the [Modal Python Library](https://github.com/modal-labs/modal-client). We hope to add more features over time, although defining Modal Functions will still be exclusively in Python.

### Tests

Tests are run against production, and you need to be authenticated with Modal to run them. See the [`test-support/`](./test-support) folder for details.

### Development principles

To keep complexity manageable, we try to maintain identical behavior across languages. This means:

- When adding a feature, add it to all languages simultaneously, with tests.
- Code structure should be similar between folders.
- Use a common set of gRPC primitives (retries, deadlines) and exceptions.
- Complex types like streams must behave as close as possible.

## License

Code is released under the [MIT license](./LICENSE).
