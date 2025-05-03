// This example calls a Modal Cls defined in `libmodal_test_support.py`.

package main

import (
	"context"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()
	app := "libmodal-test-support"
	clsName := "EchoCls"

	log.Printf("Calling %s.%s", app, clsName)
	cls, err := modal.ClsLookup(
		ctx,
		app, clsName, modal.LookupOptions{},
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

	result, err := function.Remote(nil, map[string]any{"s": "hello"})
	if err != nil {
		log.Fatalf("Failed to call Cls method: %v", err)
	}

	log.Printf("Function output: %v", result)
}
