// This example calls a function defined in `libmodal_test_support.py`.

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

	ret, err := echo.Remote([]any{"Hello world!"}, nil)
	if err != nil {
		fmt.Errorf("Failed to call function: %w", err)
	}
	log.Println("Response:", ret)

	ret, err = echo.Remote(nil, map[string]any{"s": "Hello world!"})
	if err != nil {
		fmt.Errorf("Failed to call function with kwargs: %w", err)
	}
	log.Println("Response:", ret)
}
