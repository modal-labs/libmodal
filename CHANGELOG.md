# Changelog

Both client libraries are pre-1.0, and they have separate versioning.

## Unreleased

- Added support for the `Queue` object to manage distributed FIFO queues.
  - Queues have a similar interface as Python, with `put()` and `get()` being the primary methods.
  - You can put structured objects onto Queues, with limited support for the pickle format.
- Added `InvalidError`, `QueueEmptyError`, and `QueueFullError` to support Queues.
- Fixed a bug in `modal-js` that produced incorrect bytecode for bytes objects.
- Options in the Go SDK now take pointer types, and can be `nil` for default values.

## modal-js/v0.3.5, modal-go/v0.0.6

- Added support for spawning functions with `Function_.spawn()` (TS) / `Function.Spawn()` (Go).

## modal-js/v0.3.4, modal-go/v0.0.5

- Added feature for looking up and calling remote classes via the `Cls` object.
- (Go) Removed the initial `ctx context.Context` argument from `Function.Remote()`.

## modal-js/v0.3.3, modal-go/v0.0.4

- Support calling remote functions with arguments greater than 2 MiB in byte payload size.

## modal-js/v0.3.2, modal-go/v0.0.3

- First public release
- Basic `Function`, `Sandbox`, `Image`, and `ContainerProcess` support
