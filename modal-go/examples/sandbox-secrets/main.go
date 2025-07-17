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
		log.Fatalf("Failed to lookup or create app: %v", err)
	}

	image, err := app.ImageFromRegistry("alpine:3.21", nil)
	if err != nil {
		log.Fatalf("Failed to create image from registry: %v", err)
	}
	secret, err := modal.SecretFromName(context.Background(), "libmodal-test-secret", &modal.SecretFromNameOptions{RequiredKeys: []string{"c"}})
	if err != nil {
		log.Fatalf("Failed finding a secret: %v", err)
	}

	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command: []string{"printenv", "c"}, Secrets: []*modal.Secret{secret},
	})
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	log.Printf("Sandbox created: %s\n", sb.SandboxId)

	output, err := io.ReadAll(sb.Stdout)
	if err != nil {
		log.Fatalf("Failed to read output: %v", err)
	}
	log.Printf("Environment variable c: %v", string(output))
}
