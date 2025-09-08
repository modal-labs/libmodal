// This example calls a Modal Cls defined in `libmodal_test_support.py`,
// and overrides the default options.

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

	cls, err := mc.Cls.Lookup(ctx, "libmodal-test-support", "EchoClsParametrized", nil)
	if err != nil {
		log.Fatalf("Failed to lookup Cls: %v", err)
	}

	instance, err := cls.Instance(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to create Cls instance: %v", err)
	}

	method, err := instance.Method("echo_env_var")
	if err != nil {
		log.Fatalf("Failed to access Cls method: %v", err)
	}

	secret, err := mc.Secrets.FromMap(ctx, map[string]string{
		"SECRET_MESSAGE": "hello, Secret",
	}, nil)
	if err != nil {
		log.Fatalf("Failed to create Secret: %v", err)
	}

	instanceWithOptions, err := cls.
		WithOptions(modal.ClsOptions{
			Secrets: []*modal.Secret{secret},
		}).
		WithConcurrency(modal.ClsConcurrencyOptions{MaxInputs: 1}).
		Instance(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to create Cls instance with options: %v", err)
	}

	methodWithOptions, err := instanceWithOptions.Method("echo_env_var")
	if err != nil {
		log.Fatalf("Failed to access Cls method with options: %v", err)
	}

	// Call the Cls function, without the Secret being set.
	result, err := method.Remote(ctx, []any{"SECRET_MESSAGE"}, nil)
	if err != nil {
		log.Fatalf("Failed to call Cls method: %v", err)
	}
	log.Println(result)

	// Call the Cls function with overrides, and confirm that the Secret is set.
	result, err = methodWithOptions.Remote(ctx, []any{"SECRET_MESSAGE"}, nil)
	if err != nil {
		log.Fatalf("Failed to call Cls method with options: %v", err)
	}
	log.Println(result)
}
