# Modal Go Library

[![Go Reference](https://pkg.go.dev/badge/github.com/modal-labs/libmodal/modal-go)](https://pkg.go.dev/github.com/modal-labs/libmodal/modal-go)
[![Build Status](https://github.com/modal-labs/libmodal/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/modal-labs/libmodal/actions?query=branch%3Amain)

The [Modal](https://modal.com/) Go SDK allows you to run Modal Functions and Sandboxes from Go applications.

## Documentation

See the [documentation and examples](https://github.com/modal-labs/libmodal?tab=readme-ov-file#go-modal-go) on GitHub.

## Requirements

Go 1.23 or higher.

## Installation

In a project using Go modules, just run:

```bash
go get -u github.com/modal-labs/libmodal/modal-go
```

Then, reference modal-go in a Go program with `import`:

```go
import (
	"github.com/modal-labs/libmodal/modal-go"
)
```
