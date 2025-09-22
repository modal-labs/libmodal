package main

import (
	"context"
	"io"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()
	mc, err := modal.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	app, err := mc.Apps.Lookup(ctx, "libmodal-example", &modal.LookupOptions{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to lookup or create App: %v", err)
	}
	image := mc.Images.FromRegistry("alpine:3.21", nil)

	secret, err := mc.Secrets.FromName(ctx, "libmodal-test-secret", &modal.SecretFromNameOptions{RequiredKeys: []string{"c"}})
	if err != nil {
		log.Fatalf("Failed finding a Secret: %v", err)
	}

	ephemeralSecret, err := mc.Secrets.FromMap(ctx, map[string]string{
		"d": "123",
	}, nil)
	if err != nil {
		log.Fatalf("Failed creating ephemeral Secret: %v", err)
	}

	sb, err := mc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateOptions{
		Command: []string{"sh", "-lc", "printenv | grep -E '^c|d='"}, Secrets: []*modal.Secret{secret, ephemeralSecret},
	})
	if err != nil {
		log.Fatalf("Failed to create Sandbox: %v", err)
	}
	log.Printf("Sandbox created: %s\n", sb.SandboxId)
	defer func() {
		if err := sb.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", sb.SandboxId, err)
		}
	}()

	output, err := io.ReadAll(sb.Stdout)
	if err != nil {
		log.Fatalf("Failed to read output: %v", err)
	}
	log.Printf("Sandbox environment variables from Secrets:\n%v", string(output))
}
