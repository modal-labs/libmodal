// This example spawns a Function defined in `libmodal_test_support.py`, and
// later gets its outputs.

package main

import (
	"context"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()
	mc, err := modal.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	echo, err := mc.Functions.Lookup(ctx, "libmodal-test-support", "echo_string", nil)
	if err != nil {
		log.Fatalf("Failed to lookup Function: %v", err)
	}

	fc, err := echo.Spawn(ctx, nil, map[string]any{"s": "Hello world!"})
	if err != nil {
		log.Fatalf("Failed to spawn Function: %v", err)
	}

	ret, err := fc.Get(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to get Function results: %v", err)
	}
	log.Println("Response:", ret)
}
