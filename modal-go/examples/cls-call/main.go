// This example calls a Modal Cls defined in `libmodal_test_support.py`.

package main

import (
	"context"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()

	// Lookup a deployed Cls.
	cls, err := modal.ClsLookup(
		ctx,
		"libmodal-test-support", "EchoCls", modal.LookupOptions{},
	)
	if err != nil {
		log.Fatalf("Failed to lookup Cls: %v", err)
	}

	instance, err := cls.Instance(nil)
	if err != nil {
		log.Fatalf("Failed to create Cls instance: %v", err)
	}

	function, err := instance.Method("echo_string")
	if err != nil {
		log.Fatalf("Failed to access Cls method: %v", err)
	}

	// Call the Cls function with args.
	result, err := function.Remote([]any{"Hello world!"}, nil)
	if err != nil {
		log.Fatalf("Failed to call Cls method: %v", err)
	}
	log.Println("Response:", result)

	// Call the Cls function with kwargs.
	result, err = function.Remote(nil, map[string]any{"s": "Hello world!"})
	if err != nil {
		log.Fatalf("Failed to call Cls method: %v", err)
	}
	log.Println("Response:", result)
}
