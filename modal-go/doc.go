// Package modal is a lightweight, idiomatic Go client for Modal.com.
//
// The library mirrors the core feature-set of Modal’s Python client while
// feeling natural in Go:
//
//   - Spin up sandboxes — fast, secure, ephemeral VMs for running code.
//   - Invoke Modal Functions and manage their inputs / outputs.
//   - Read, write, and list files in Modal Volumes.
//   - Create or inspect containers, streams, and logs.
//
// **What it does not do:** deploying Modal Functions. Deployment is still
// handled in Python; this package is for calling and orchestrating them
// from other projects.
//
// # Authentication
//
// At runtime the client resolves credentials in this order:
//
//  1. Environment variables
//     MODAL_TOKEN_ID, MODAL_TOKEN_SECRET, MODAL_ENVIRONMENT (optional)
//  2. A profile explicitly requested via `MODAL_PROFILE`
//  3. A profile marked `active = true` in `~/.modal.toml`
//
// See `config.go` for the resolution logic.
//
// # Stability
//
// `libmodal` is **alpha** software; the API may change without notice until
// a v1.0.0 release. Please pin versions and file issues generously.
//
// For additional examples and language-parity tests, see
// https://github.com/modal-labs/libmodal.

package modal
