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

	image := mc.Images.FromRegistry("nvidia/cuda:12.4.0-devel-ubuntu22.04", nil)

	sb, err := mc.Sandboxes.Create(ctx, *app, *image, &modal.SandboxCreateParams{
		GPU: "A10G",
	})
	if err != nil {
		log.Fatalf("Failed to create Sandbox: %v", err)
	}
	log.Printf("Started Sandbox with A10G GPU: %s", sb.SandboxID)
	defer func() {
		if err := sb.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", sb.SandboxID, err)
		}
	}()

	log.Println("Running `nvidia-smi` in Sandbox:")

	p, err := sb.Exec(ctx, []string{"nvidia-smi"}, nil)
	if err != nil {
		log.Fatalf("Failed to execute nvidia-smi in Sandbox: %v", err)
	}

	output, err := io.ReadAll(p.Stdout)
	if err != nil {
		log.Fatalf("Failed to read stdout: %v", err)
	}

	log.Printf("%s", string(output))
}
