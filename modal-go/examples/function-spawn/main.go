// This example spawns a function defined in `libmodal_test_support.py`, and
// later gets its outputs.

package main

import (
	"context"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()

	echo, err := modal.FunctionLookup(ctx, "libmodal-test-support", "echo_string", nil)
	if err != nil {
		log.Fatalf("Failed to lookup function: %v", err)
	}

	fc, err := echo.Spawn(nil, map[string]any{"s": "Hello world!"})
	if err != nil {
		log.Fatalf("Failed to spawn function: %v", err)
	}

	ret, err := fc.Get(nil)
	if err != nil {
		log.Fatalf("Failed to get function results: %v", err)
	}
	log.Println("Response:", ret)
}
