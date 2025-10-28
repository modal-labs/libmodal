// Package modal is a lightweight, idiomatic Go SDK for Modal.com.
//
// It mirrors the core feature-set of Modal’s Python SDK while feeling
// natural in Go:
//
//   - Spin up Sandboxes — fast, secure, ephemeral VMs for running code.
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
// At runtime the SDK resolves credentials in this order:
//
//  1. Environment variables
//     MODAL_TOKEN_ID, MODAL_TOKEN_SECRET, MODAL_ENVIRONMENT (optional)
//  2. A profile explicitly requested via `MODAL_PROFILE`
//  3. A profile marked `active = true` in `~/.modal.toml`
//
// See `config.go` for the resolution logic.
//
// For additional examples and language-parity tests, see
// https://github.com/modal-labs/libmodal/tree/main/modal-go.
package modal
