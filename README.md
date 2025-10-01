# libmodal: [Modal](https://modal.com) SDKs for JavaScript / TypeScript and Go

[![Build Status](https://github.com/modal-labs/libmodal/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/modal-labs/libmodal/actions?query=branch%3Amain)
[![JS Reference Documentation](https://img.shields.io/static/v1?message=reference&logo=javascript&labelColor=5c5c5c&color=1182c3&logoColor=white&label=%20)](https://modal-labs.github.io/libmodal/)
[![JS npm Version](https://img.shields.io/npm/v/modal.svg)](https://www.npmjs.org/package/modal)
[![JS npm Downloads](https://img.shields.io/npm/dm/modal.svg)](https://www.npmjs.com/package/modal)
[![Go Reference Documentation](https://pkg.go.dev/badge/github.com/modal-labs/libmodal/modal-go)](https://pkg.go.dev/github.com/modal-labs/libmodal/modal-go)

**libmodal** (beta) provides convenient, on-demand access to serverless cloud compute on Modal from JavaScript/TypeScript and Go projects. Use it to safely run arbitrary code in Modal Sandboxes, call Modal Functions, and interact with Modal resources. For Python, see the main [Modal Python SDK](https://github.com/modal-labs/modal-client) instead.

We're working towards feature parity with the main Modal Python SDK, although defining Modal Functions will likely remain exclusive to Python.

## Documentation

See the main [Modal documentation](https://modal.com/docs/guide) and [user guides](https://modal.com/docs/guide) for high-level overviews. For details, see the API reference documentation in [JavaScript / TypeScript](https://modal-labs.github.io/libmodal/) and [Go](https://pkg.go.dev/github.com/modal-labs/libmodal/modal-go#section-documentation).

We also provide a number of examples:
- Call a deployed Function: [JS](./modal-js/examples/function-call.ts) / [Go](./modal-go/examples/function-call/main.go)
- Spawn a deployed Function: [JS](./modal-js/examples/function-spawn.ts) / [Go](./modal-go/examples/function-spawn/main.go)
- Call a deployed Cls: [JS](./modal-js/examples/cls-call.ts) / [Go](./modal-go/examples/cls-call/main.go)
- Call a deployed Cls, and override its options: [JS](./modal-js/examples/cls-call-with-options.ts) / [Go](./modal-go/examples/cls-call-with-options/main.go)
- Create a Sandbox: [JS](./modal-js/examples/sandbox.ts) / [Go](./modal-go/examples/sandbox/main.go)
- Create a named Sandbox: [JS](./modal-js/examples/sandbox-named.ts) / [Go](./modal-go/examples/sandbox-named/main.go)
- Create a Sandbox with GPU: [JS](./modal-js/examples/sandbox-gpu.ts) / [Go](./modal-go/examples/sandbox-gpu/main.go)
- Create a Sandbox using a private image from AWS ECR: [JS](./modal-js/examples/sandbox-private-image.ts) / [Go](./modal-go/examples/sandbox-private-image/main.go)
- Take a snapshot of the filesystem of a Sandbox: [JS](./modal-js/examples/sandbox-filesystem-snapshot.ts) / [Go](./modal-go/examples/sandbox-filesystem-snapshot/main.go)
- Execute Sandbox commands: [JS](./modal-js/examples/sandbox-exec.ts) / [Go](./modal-go/examples/sandbox-exec/main.go)
- Running a coding agent in a Sandbox: [JS](./modal-js/examples/sandbox-agent.ts) / [Go](./modal-go/examples/sandbox-agent/main.go)
- Check the status and exit code of a Sandbox: [JS](./modal-js/examples/sandbox-poll.ts) / [Go](./modal-go/examples/sandbox-poll/main.go)
- Access Sandbox filesystem: [JS](./modal-js/examples/sandbox-filesystem.ts) / [Go](./modal-go/examples/sandbox-filesystem/main.go)
- Expose ports on a Sandbox using Tunnels: [JS](./modal-js/examples/sandbox-tunnels.ts) / [Go](./modal-go/examples/sandbox-tunnels/main.go)
- Include Secrets in Sandbox: [JS](./modal-js/examples/sandbox-secrets.ts) / [Go](./modal-go/examples/sandbox-secrets/main.go)
- Mount a Volume to a Sandbox: [JS](./modal-js/examples/sandbox-volume.ts) / [Go](./modal-go/examples/sandbox-volume/main.go), and same but with an ephemeral Volume: [JS](./modal-js/examples/sandbox-volume-ephemeral.ts) / [Go](./modal-go/examples/sandbox-volume-ephemeral/main.go)
- Mount a cloud bucket to a Sandbox: [JS](./modal-js/examples/sandbox-cloud-bucket.ts) / [Go](./modal-go/examples/sandbox-cloud-bucket/main.go)
- Eagerly build an Image for a Sandbox: [JS](./modal-js/examples/sandbox-prewarm.ts) / [Go](./modal-go/examples/sandbox-prewarm/main.go)
- Building custom Images: [JS](./modal-js/examples/image-building.ts) / [Go](./modal-go/examples/image-building/main.go)

## Installation

First authenticate with Modal (see [Getting started](https://modal.com/docs/guide#getting-started)). Either sign in with the Modal CLI using `pip install modal && modal setup`, or in machine environments set these environment variables:

```bash
# Replace these with your actual token!
export MODAL_TOKEN_ID=ak-NOTAREALTOKENSTRINGXYZ
export MODAL_TOKEN_SECRET=as-FAKESECRETSTRINGABCDEF
```

### JavaScript / TypeScript (`modal-js/`)

Requires Node 22 or later. We bundle both ES Modules and CommonJS formats, so you can load the package with either `import` or `require()` in any project.

Install this in any server-side Node.js / Deno / Bun project.

```bash
npm install modal
```

npm package: https://www.npmjs.com/package/modal

### Go (`modal-go/`)

Requires Go 1.23 or later.

Install the latest version:

```bash
go get -u github.com/modal-labs/libmodal/modal-go
```

Import in your application:

```go
import "github.com/modal-labs/libmodal/modal-go"
```

Go package: https://pkg.go.dev/github.com/modal-labs/libmodal/modal-go

## Support

For usage questions and other support, please reach out on the [Modal Community Slack](https://modal.com/slack).

## Community SDKs

There are also open-source Modal libraries built and maintained by our community. These projects are not officially supported by Modal and we thus can't vouch for them, but feel free to explore and contribute.

- Ruby: [anthonycorletti/modal-rb](https://github.com/anthonycorletti/modal-rb)
