package main

import (
	"context"
	"fmt"
	"log"
	"reflect"

	modal "github.com/modal-labs/libmodal/modal-go"
)

func main() {
	// Initialize the Modal client
	ctx := context.Background()

	// Look up a function that supports CBOR
	function, err := modal.FunctionLookup(ctx, "libmodal-test-support", "identity", nil)
	if err != nil {
		log.Fatalf("Function lookup failed: %v", err)
	}

	// Call the function with CBOR encoding
	// This will error if the remote function doesn't support CBOR input format
	result, err := function.Remote([]any{[]int{2, 3}}, nil)
	if err != nil {
		log.Fatalf("Function call failed: %v", err)
	}

	// Print detailed result information
	fmt.Printf("Result: %v\n", result)
	fmt.Printf("Result type: %T\n", result)

	// If result is a slice/array, print details about each element
	if reflect.TypeOf(result).Kind() == reflect.Slice {
		resultSlice := reflect.ValueOf(result)
		fmt.Printf("Result length: %d\n", resultSlice.Len())
		for i := 0; i < resultSlice.Len(); i++ {
			elem := resultSlice.Index(i).Interface()
			fmt.Printf("  [%d]: %v (type: %T)\n", i, elem, elem)
		}
	}

	// Also test with kwargs
	result2, err := function.Remote(nil, map[string]any{"s": "Hello CBOR with kwargs!"})
	if err != nil {
		log.Fatalf("Function call with kwargs failed: %v", err)
	}

	// Print detailed result information for kwargs call
	fmt.Printf("Result with kwargs: %v\n", result2)
	fmt.Printf("Result with kwargs type: %T\n", result2)

	// If result2 is a slice/array, print details about each element
	if reflect.TypeOf(result2).Kind() == reflect.Slice {
		resultSlice := reflect.ValueOf(result2)
		fmt.Printf("Result with kwargs length: %d\n", resultSlice.Len())
		for i := 0; i < resultSlice.Len(); i++ {
			elem := resultSlice.Index(i).Interface()
			fmt.Printf("  [%d]: %v (type: %T)\n", i, elem, elem)
		}
	}
}
