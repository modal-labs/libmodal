package main

import (
	"context"
	"log"
	"time"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()
	mc, err := modal.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	app, err := mc.Apps.FromName(ctx, "libmodal-example", &modal.AppFromNameParams{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to get or create App: %v", err)
	}

	// Create a Sandbox with Python's built-in HTTP server
	image := mc.Images.FromRegistry("python:3.12-alpine", nil)

	sb, err := mc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Command:        []string{"python3", "-m", "http.server", "8000"},
		EncryptedPorts: []int{8000},
		Timeout:        1 * time.Minute,
		IdleTimeout:    30 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create Sandbox: %v", err)
	}
	defer func() {
		if err := sb.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", sb.SandboxID, err)
		}
	}()

	creds, err := sb.CreateConnectToken(ctx, &modal.SandboxCreateConnectToken{UserMetadata: "abc"})
	if err != nil {
		log.Fatalf("Failed to create connect token: %v", err)
	}
	log.Printf("Got credentials: %v\n", creds.Token)
}
