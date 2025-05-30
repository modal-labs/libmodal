// This example calls a function defined in `libmodal_test_support.py`.

package main

import (
	"context"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()

	echo, err := modal.FunctionLookup(ctx, "libmodal-test-support", "echo_string", modal.LookupOptions{})
	if err != nil {
		log.Fatalf("Failed to lookup function: %v", err)
	}

	ret, err := echo.Remote([]any{"Hello world!"}, nil)
	if err != nil {
		log.Fatalf("Failed to call function: %v", err)
	}
	log.Println("Response:", ret)

	ret, err = echo.Remote(nil, map[string]any{"s": "Hello world!"})
	if err != nil {
		log.Fatalf("Failed to call function with kwargs: %v", err)
	}
	log.Println("Response:", ret)
}
