// This example spawns a function defined in `libmodal_test_support.py`, and
// later gets its outputs.

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()

	echo, err := modal.FunctionLookup(ctx, "libmodal-test-support", "echo_string", modal.LookupOptions{})
	if err != nil {
		fmt.Errorf("Failed to lookup function: %w", err)
	}

	fc, err := echo.Spawn(nil, map[string]any{"s": "Hello world!"})
	if err != nil {
		fmt.Errorf("Failed to spawn function: %w", err)
	}

	ret, err := fc.Get(modal.FunctionCallGetOptions{})
	if err != nil {
		fmt.Errorf("Failed to get function results: %w", err)
	}
	log.Println("Response:", ret)
}
