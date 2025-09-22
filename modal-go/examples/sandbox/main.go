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

	app, err := mc.Apps.FromName(ctx, "libmodal-example", &modal.AppFromNameParams{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to get or create App: %v", err)
	}

	image := mc.Images.FromRegistry("alpine:3.21", nil)

	sb, err := mc.Sandboxes.Create(ctx, *app, *image, &modal.SandboxCreateParams{
		Command: []string{"cat"},
	})
	if err != nil {
		log.Fatalf("Failed to create Sandbox: %v", err)
	}
	log.Printf("sandbox: %s\n", sb.SandboxID)
	defer func() {
		if err := sb.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", sb.SandboxID, err)
		}
	}()

	sbFromID, err := mc.Sandboxes.FromID(ctx, sb.SandboxID)
	if err != nil {
		log.Fatalf("Failed to get Sandbox with ID: %v", err)
	}
	log.Printf("Queried Sandbox with ID: %v", sbFromID.SandboxID)

	_, err = sb.Stdin.Write([]byte("this is input that should be mirrored by cat"))
	if err != nil {
		log.Fatalf("Failed to write to Sandbox stdin: %v", err)
	}
	err = sb.Stdin.Close()
	if err != nil {
		log.Fatalf("Failed to close Sandbox stdin: %v", err)
	}

	output, err := io.ReadAll(sb.Stdout)
	if err != nil {
		log.Fatalf("Failed to read from Sandbox stdout: %v", err)
	}

	log.Printf("output: %s\n", string(output))
}
