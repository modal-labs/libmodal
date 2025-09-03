// This example calls a Modal Cls defined in `libmodal_test_support.py`.

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

	// Lookup a deployed Cls.
	cls, err := mc.Cls.Lookup(ctx, "libmodal-test-support", "EchoCls", nil)
	if err != nil {
		log.Fatalf("Failed to lookup Cls: %v", err)
	}

	instance, err := cls.Instance(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to create Cls instance: %v", err)
	}

	function, err := instance.Method("echo_string")
	if err != nil {
		log.Fatalf("Failed to access Cls method: %v", err)
	}

	// Call the Cls function with args.
	result, err := function.Remote(ctx, []any{"Hello world!"}, nil)
	if err != nil {
		log.Fatalf("Failed to call Cls method: %v", err)
	}
	log.Println("Response:", result)

	// Call the Cls function with kwargs.
	result, err = function.Remote(ctx, nil, map[string]any{"s": "Hello world!"})
	if err != nil {
		log.Fatalf("Failed to call Cls method: %v", err)
	}
	log.Println("Response:", result)
}
