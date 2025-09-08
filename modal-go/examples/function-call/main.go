// This example calls a function defined in `libmodal_test_support.py`.

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

	ret, err := echo.Remote(ctx, []any{"Hello world!"}, nil)
	if err != nil {
		log.Fatalf("Failed to call Function: %v", err)
	}
	log.Println("Response:", ret)

	ret, err = echo.Remote(ctx, nil, map[string]any{"s": "Hello world!"})
	if err != nil {
		log.Fatalf("Failed to call Function with kwargs: %v", err)
	}
	log.Println("Response:", ret)
}
