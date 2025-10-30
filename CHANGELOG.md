# Changelog

Both client libraries are pre-1.0, and they have separate versioning.

## Unreleased

- All Go SDK functions that take a Context will respect the timeout of the context.

## modal-js/v0.5.0, modal-go/v0.5.0

The first beta release of the Modal SDKs for JS and Go (graduating from alpha). See the [Migration Guide](./MIGRATION-GUIDE.md) for a detailed list of breaking changes.

- The SDKs now expose a central Modal Client object as the main entry point for interacting with Modal resources.
- The interface for working with Modal object instances (Functions, Sandboxes, Images, etc.) is largely the same as before, with some naming changes.
- Calling deployed Functions and classes now uses a new protocol for payload serialization which requires the deployed apps to use the Modal Python SDK 1.2 or newer.
- Internally removed the global client (and config/profile data in global scope), moving all that to the Client type.
- Consistent parameter naming across both SDKs: all `Options` structs/interfaces renamed to `Params`.
- JS-specific changes:
  - Added explicit unit suffixes to all parameters that represent durations (in milliseconds, suffixed with `Ms`) or memory amounts (in MiB, suffixed with `MiB`).
- Go-specific changes:
  - Changed how we do context passing, so contexts now only affect the current operation and are not used for lifecycle management of the created resources.
  - All `Params` structs are now passed as pointers for consistency and to support optional parameters.
  - Field names follow Go casing conventions (e.g., `Id` → `ID`, `Url` → `URL`, `TokenId` → `TokenID`).
  - Added explicit unit suffixes to all parameters that represent memory amounts (in MiB, suffixed with `MiB`).

Additional new features:
- Added support for setting CPU and memory limits when creating Sandboxes and Cls instances.

## modal-js/v0.3.25, modal-go/v0.0.25

- Fixed a bug in modal-js related to unpickling objects from Python (Function calls, Queues, etc.), where integers between 32678 and 65535 were incorrectly decoded as signed integers.
- Internal updates for how authentication tokens are handled for input plane clients.

## modal-js/v0.3.24, modal-go/v0.0.24

- Added `env` parameters to several methods, as a convenience for passing environment variables into Sandboxes, etc.
- Added `Sandbox.getTags()` (JS) and `Sandbox.GetTags()` (Go).

## modal-js/v0.3.23, modal-go/v0.0.23

- Added support for PTYs in Sandboxes.

## modal-js/v0.3.22, modal-go/v0.0.22

- Added `Image.dockerfileCommands()` (JS) and `ImageDockerfileCommands()` (Go).

## modal-js/v0.3.21, modal-go/v0.0.21

- Added support for setting idle timeout when creating Sandboxes.

## modal-js/v0.3.20, modal-go/v0.0.20

- Added `Image.delete()` (JS) and `ImageDelete()` (Go).
- Changed `Image.fromId()` (JS) and `NewImageFromId()` (Go) to throw a `NotFoundError` if the Image does not exist. Note that the signature of `NewImageFromId()` has changed.

## modal-js/v0.3.19, modal-go/v0.0.19

- `Sandbox.exec` in JS now correctly accepts a list of Secrets.

## modal-js/v0.3.18, modal-go/v0.0.18

- Added `Image.build` (JS) and `Image.Build` (Go).
- Added `Image.fromId` (JS) / `NewImageFromId` (Go).
- Operations on an ehpemeral Queue after having called `CloseEhpemeral()` will now explicitly fail in Go.
- Added support for instantiating a Cls with custom options, using `Cls.withOptions()`/`.withConcurrency()`/`.withBatching()` (JS) / `Cls.WithOptions()`/`.WithConcurrency()`/`.WithBatching()` (Go).
- Added support for [Named Sandboxes](https://modal.com/docs/guide/sandbox#named-sandboxes) (examples in [JS](./modal-js/examples/sandbox-named.ts) and [Go](./modal-go/examples/sandbox-named/main.go)).
- Added support for `Volume.ephemeral()` (JS) / `VolumeEphemeral()` (Go).

## modal-js/v0.3.17, modal-go/v0.0.17

- Added support for more parameters to `Sandbox.create()`:
  - `blockNetwork`: Whether to block all network access from the Sandbox.
  - `cidrAllowlist`: List of CIDRs the Sandbox is allowed to access.
  - `gpu`: GPU reservation for the Sandbox (e.g. "A100", "T4:2", "A100-80GB:4").
  - `cloud`: Cloud provider to run the Sandbox on.
  - `regions`: Region(s) to run the Sandbox on.
  - `verbose`: Enable verbose logging.
  - `proxy`: Connect a Modal Proxy to a Sandbox.
  - `workdir`: Set the working directory.
- Added support for mounting `CloudBucketMount`s to Sandboxes.
- Added top level for Image objects that are lazy. The Images are built when creating a Sandbox.
  - `Image.fromRegistry` in typescript and `NewImageFromRegistry` in golang.
  - `Image.fromAwsEcr` in typescript and `NewImageFromAwsEcr` in golang.
  - `Image.fromGcpArtifactRegistry` in typescript and `NewImageFromGcpArtifactRegistry` in golang.
- Added `Secret.fromObject()` (JS) / `SecretFromMap()` (Go) to create a Secret from key-value pairs (like `from_dict()` in Python).
- Added `name` field to `App`s, `Sandbox`es, `Secret`s, `Volume`s, and `Queue`s.
- Added support for `Function.getCurrentStats()` (JS) / `Function.GetCurrentStats()` (Go).
- Added support for `Function.updateAutoscaler()` (JS) / `Function.UpdateAutoscaler()` (Go).
- Added support for `Function.getWebURL()` (JS) / `Function.GetWebURL()` (Go).
- Added support for `Volume.readOnly()` (JS) / `Volume.ReadOnly()` (Go).
- Added support for setting tags on Sandboxes, and for listing Sandboxes (by tag).

## modal-js/v0.3.16, modal-go/v0.0.16

- Added support for getting Sandboxes from an ID.

## modal-js/v0.3.15, modal-go/v0.0.15

- Added support for snapshotting the filesystem of a Sandbox.
- Added support for polling Sandboxes to check if they are still running, or get the exit code.
- Added support to execute commands in Sandboxes with Secrets.
- Added support for creating Sandboxes with Secrets.

## modal-js/v0.3.14, modal-go/v0.0.14

- Added support for setting up Tunnels to expose live TCP ports for Sandboxes.

## modal-js/v0.3.13, modal-go/v0.0.13

- Fixed calls of Cls with experimental `input_plane_region` option.
- (Go) Removed `Function.InputPlaneURL` from being exposed as public API.

## modal-js/v0.3.12, modal-go/v0.0.12

- Added support for passing a Secret to `imageFromRegistry()` (JS) / `ImageFromRegistry()` (Go) to pull images from private registries.
- Added support for creating Images from Google Artifact Registry with `imageFromGcpArtifactRegistry()` (JS) / `ImageFromGcpArtifactRegistry()` (Go).
- Added experimental support for calling remote Functions deployed with the `input_plane_region` option in Python.

## modal-js/v0.3.11, modal-go/v0.0.11

- Added `InitializeClient()` (Go) / `initializeClient()` (JS) to initialize the client at runtime with credentials.
- Client libraries no longer panic at startup if no token ID / Secret is provided. Instead, they will throw an error when trying to use the client.

## modal-js/v0.3.10, modal-go/v0.0.10

- Added `workdir` and `timeout` options to `ExecOptions` for Sandbox processes.

## modal-js/v0.3.9, modal-go/v0.0.9

- Added support for Sandbox filesystem.

## modal-js/v0.3.8

- Added support for CommonJS format / `require()`. Previously, modal-js only supported ESM `import`.

## modal-js/v0.3.7, modal-go/v0.0.8

- Added support for creating Images from AWS ECR with `App.imageFromAwsEcr()` (JS) / `App.ImageFromAwsEcr()` (Go).
- Added support for accessing Modal Secrets with `Secret.fromName()` (JS) / `modal.SecretFromName()` (Go).
- Fixed serialization of some pickled objects (negative ints, dicts) in modal-js.

## modal-js/v0.3.6, modal-go/v0.0.7

- Added support for the `Queue` object to manage distributed FIFO queues.
  - Queues have a similar interface as Python, with `put()` and `get()` being the primary methods.
  - You can put structured objects onto Queues, with limited support for the pickle format.
- Added `InvalidError`, `QueueEmptyError`, and `QueueFullError` to support Queues.
- Fixed a bug in `modal-js` that produced incorrect bytecode for bytes objects.
- Options in the Go SDK now take pointer types, and can be `nil` for default values.

## modal-js/v0.3.5, modal-go/v0.0.6

- Added support for spawning Functions with `Function_.spawn()` (JS) / `Function.Spawn()` (Go).

## modal-js/v0.3.4, modal-go/v0.0.5

- Added feature for looking up and calling remote classes via the `Cls` object.
- (Go) Removed the initial `ctx context.Context` argument from `Function.Remote()`.

## modal-js/v0.3.3, modal-go/v0.0.4

- Support calling remote Functions with arguments greater than 2 MiB in byte payload size.

## modal-js/v0.3.2, modal-go/v0.0.3

- First public release
- Basic `Function`, `Sandbox`, `Image`, and `ContainerProcess` support
