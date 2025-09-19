# libmodal: [Modal](https://modal.com) SDK Lite

Modal client libraries for JavaScript and Go. **(Alpha)**

This repository provides lightweight alternatives to the [Modal Python Library](https://github.com/modal-labs/modal-client). They let you start Sandboxes (secure VMs), call Modal Functions, and manage containers. However, they don't support deploying Modal Functions — those still need to be written in Python!

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

- [Call a deployed Function](./modal-js/examples/function-call.ts)
- [Spawn a deployed Function](./modal-js/examples/function-spawn.ts)
- [Call a deployed Cls](./modal-js/examples/cls-call.ts)
- [Call a deployed Cls, and override its options](./modal-js/examples/cls-call-with-options.ts)
- [Create a Sandbox](./modal-js/examples/sandbox.ts)
- [Create a named Sandbox](./modal-js/examples/sandbox-named.ts)
- [Create a Sandbox with GPU](./modal-js/examples/sandbox-gpu.ts)
- [Create a Sandbox using a private image from AWS ECR](./modal-js/examples/sandbox-private-image.ts)
- [Take a snapshot of the filesystem of a Sandbox](./modal-js/examples/sandbox-filesystem-snapshot.ts)
- [Execute Sandbox commands](./modal-js/examples/sandbox-exec.ts)
- [Running a coding agent in a Sandbox](./modal-js/examples/sandbox-agent.ts)
- [Check the status and exit code of a Sandbox](./modal-js/examples/sandbox-poll.ts)
- [Access Sandbox filesystem](./modal-js/examples/sandbox-filesystem.ts)
- [Expose ports on a Sandbox using Tunnels](./modal-js/examples/sandbox-tunnels.ts)
- [Include Secrets in Sandbox](./modal-js/examples/sandbox-secrets.ts)
- [Mount a Volume to a Sandbox](./modal-js/examples/sandbox-volume.ts), and same but with an [ephemeral Volume](./modal-js/examples/sandbox-volume-ephemeral.ts)
- [Mount a cloud bucket to a Sandbox](./modal-js/examples/sandbox-cloud-bucket.ts)
- [Eagarly build an Image for a Sandbox](./modal-js/examples/sandbox-prewarm.ts)
- [Building custom Images](./modal-js/examples/image-building.ts)

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

- [Call a deployed Function](./modal-go/examples/function-call/main.go)
- [Spawn a deployed Function](./modal-go/examples/function-spawn/main.go)
- [Call a deployed Cls](./modal-go/examples/cls-call/main.go)
- [Call a deployed Cls, and override its options](./modal-go/examples/cls-call-with-options/main.go)
- [Create a Sandbox](./modal-go/examples/sandbox/main.go)
- [Create a named Sandbox](./modal-go/examples/sandbox-named/main.go)
- [Create a Sandbox with GPU](./modal-go/examples/sandbox-gpu/main.go)
- [Create a Sandbox using a private image from AWS ECR](./modal-go/examples/sandbox-private-image/main.go)
- [Take a snapshot of the filesystem of a Sandbox](./modal-go/examples/sandbox-filesystem-snapshot/main.go)
- [Execute Sandbox commands](./modal-go/examples/sandbox-exec/main.go)
- [Running a coding agent in a Sandbox](./modal-go/examples/sandbox-agent/main.go)
- [Check the status and exit code of a Sandbox](./modal-go/examples/sandbox-poll/main.go)
- [Access Sandbox filesystem](./modal-go/examples/sandbox-filesystem/main.go)
- [Expose ports on a Sandbox using Tunnels](./modal-go/examples/sandbox-tunnels/main.go)
- [Include Secrets in Sandbox](./modal-go/examples/sandbox-secrets/main.go)
- [Mount a Volume to a Sandbox](./modal-go/examples/sandbox-volume/main.go), and same but with an [ephemeral Volume](./modal-go/examples/sandbox-volume-ephemeral/main.go)
- [Mount a cloud bucket to a Sandbox](./modal-go/examples/sandbox-cloud-bucket/main.go)
- [Eagarly build an Image for a Sandbox](./modal-go/examples/sandbox-prewarm/main.go)
- [Building custom Images](./modal-go/examples/image-building/main.go)

### Python

If you're using Python, please use the [Modal Python Library](https://github.com/modal-labs/modal-client), which is the main SDK and a separate project.

## Technical details

`libmodal` is a cross-language client SDK for Modal. However, it does not have all the features of the [Modal Python Library](https://github.com/modal-labs/modal-client). We hope to add more features over time, although defining Modal Functions will still be exclusively in Python.

### Tests

Tests are run against production, and you need to be authenticated with Modal to run them. See the [`test-support/`](./test-support) folder for details.

### Development principles

To keep complexity manageable, we try to maintain identical behavior across languages. This means:

- When merging a feature or change into `main`, update it for all languages simultaneously, with tests.
- Code structure should be similar between folders.
- Use a common set of gRPC primitives (retries, deadlines) and exceptions.
- Complex types like streams must behave as close as possible.
- Timeouts should use milliseconds in TypeScript, and `time.Duration` in Go.

## License

Code is released under [a permissive license](./LICENSE).


## Community SDKs

There are also open-source Modal libraries built and maintained by our community. These projects are not officially supported by Modal and we thus can't vouch for them, but feel free to explore and contribute.

- Ruby: [anthonycorletti/modal-rb](https://github.com/anthonycorletti/modal-rb)
