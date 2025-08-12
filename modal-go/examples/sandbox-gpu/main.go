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

	image, err := app.ImageFromRegistry("nvidia/cuda:12.4.0-devel-ubuntu22.04", nil)
	if err != nil {
		log.Fatalf("Failed to create image from registry: %v", err)
	}

	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{
		GPU: "A10G",
	})
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	log.Printf("Started sandbox with A10G GPU: %s", sb.SandboxId)
	defer sb.Terminate()

	log.Println("Running `nvidia-smi` in sandbox:")

	p, err := sb.Exec([]string{"nvidia-smi"}, modal.ExecOptions{})
	if err != nil {
		log.Fatalf("Failed to execute nvidia-smi in sandbox: %v", err)
	}

	output, err := io.ReadAll(p.Stdout)
	if err != nil {
		log.Fatalf("Failed to read stdout: %v", err)
	}

	log.Printf("%s", string(output))
}
