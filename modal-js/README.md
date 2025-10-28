# Modal JavaScript SDK

[![Build Status](https://github.com/modal-labs/libmodal/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/modal-labs/libmodal/actions?query=branch%3Amain)
[![JS Reference Documentation](https://img.shields.io/badge/docs-reference-blue)](https://modal-labs.github.io/libmodal/)
[![npm Version](https://img.shields.io/npm/v/modal.svg)](https://www.npmjs.org/package/modal)
[![npm Downloads](https://img.shields.io/npm/dm/modal.svg)](https://www.npmjs.com/package/modal)

The [Modal](https://modal.com/) JavaScript SDK provides convenient, on-demand access to serverless cloud compute on Modal from JS/TS projects. Use it to safely run arbitrary code in Modal Sandboxes, call Modal Functions, and interact with Modal resources.

It comes with built-in TypeScript type definitions.

We're approaching feature parity with the main [Modal Python SDK](https://github.com/modal-labs/modal-client), although defining Modal Functions will likely remain exclusive to Python.

## Installation

Install this in any server-side Node.js / Deno / Bun project.

```bash
npm install modal
```

npm package: https://www.npmjs.com/package/modal

## Requirements

Node 22 or later. We bundle both ES Modules and CommonJS formats, so you can load the package with either `import` or `require()` in any project.

## Documentation

See the main [Modal documentation](https://modal.com/docs/guide) and [user guides](https://modal.com/docs/guide) for high-level overviews. For details, see the [API reference documentation for for JS](https://modal-labs.github.io/libmodal/).

We also provide a number of examples:

- [Call a deployed Function](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/function-call.ts)
- [Spawn a deployed Function](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/function-spawn.ts)
- [Call a deployed Cls](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/cls-call.ts)
- [Call a deployed Cls, and override its options](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/cls-call-with-options.ts)
- [Create a Sandbox](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/sandbox.ts)
- [Create a named Sandbox](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/sandbox-named.ts)
- [Create a Sandbox with GPU](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/sandbox-gpu.ts)
- [Create a Sandbox using a private image from AWS ECR](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/sandbox-private-image.ts)
- [Take a snapshot of the filesystem of a Sandbox](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/sandbox-filesystem-snapshot.ts)
- [Execute Sandbox commands](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/sandbox-exec.ts)
- [Running a coding agent in a Sandbox](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/sandbox-agent.ts)
- [Check the status and exit code of a Sandbox](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/sandbox-poll.ts)
- [Access Sandbox filesystem](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/sandbox-filesystem.ts)
- [Expose ports on a Sandbox using Tunnels](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/sandbox-tunnels.ts)
- [Include Secrets in Sandbox](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/sandbox-secrets.ts)
- [Mount a Volume to a Sandbox](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/sandbox-volume.ts), and same but [with an ephemeral Volume](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/sandbox-volume-ephemeral.ts)
- [Mount a cloud bucket to a Sandbox](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/sandbox-cloud-bucket.ts)
- [Eagerly build an Image for a Sandbox](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/sandbox-prewarm.ts)
- [Building custom Images](https://github.com/modal-labs/libmodal/blob/main/modal-js/examples/image-building.ts)

### Authenticating with Modal

You also need to authenticate with Modal (see [Getting started](https://modal.com/docs/guide#getting-started)). Either sign in with the Modal CLI using `pip install modal && modal setup`, or in machine environments set these environment variables:

```bash
# Replace these with your actual token!
export MODAL_TOKEN_ID=ak-NOTAREALTOKENSTRINGXYZ
export MODAL_TOKEN_SECRET=as-FAKESECRETSTRINGABCDEF
```

## Support

For usage questions and other support, please reach out on the [Modal Community Slack](https://modal.com/slack).
