package main

import (
	"context"
	"io"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()

	app, err := modal.AppLookup(ctx, "libmodal-example", &modal.LookupOptions{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to lookup or create App: %v", err)
	}
	image := modal.NewImageFromRegistry("alpine:3.21", nil)

	secret, err := modal.SecretFromName(context.Background(), "libmodal-test-secret", &modal.SecretFromNameOptions{RequiredKeys: []string{"c"}})
	if err != nil {
		log.Fatalf("Failed finding a Secret: %v", err)
	}

	ephemeralSecret, err := modal.SecretFromMap(ctx, map[string]string{
		"d": "123",
	}, nil)
	if err != nil {
		log.Fatalf("Failed creating ephemeral Secret: %v", err)
	}

	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command: []string{"sh", "-lc", "printenv | grep -E '^c|d='"}, Secrets: []*modal.Secret{secret, ephemeralSecret},
	})
	if err != nil {
		log.Fatalf("Failed to create Sandbox: %v", err)
	}
	log.Printf("Sandbox created: %s\n", sb.SandboxId)

	output, err := io.ReadAll(sb.Stdout)
	if err != nil {
		log.Fatalf("Failed to read output: %v", err)
	}
	log.Printf("Sandbox environment variables from Secrets:\n%v", string(output))
}
