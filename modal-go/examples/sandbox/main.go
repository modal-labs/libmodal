package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()

	app, err := modal.AppLookup(ctx, "libmodal-example", modal.LookupOptions{CreateIfMissing: true})
	if err != nil {
		fmt.Errorf("Failed to lookup or create app: %w", err)
	}

	image, err := app.ImageFromRegistry("alpine:3.21")
	if err != nil {
		fmt.Errorf("Failed to create image from registry: %w", err)
	}

	sb, err := app.CreateSandbox(image, modal.SandboxOptions{
		Command: []string{"cat"},
	})
	if err != nil {
		fmt.Errorf("Failed to create sandbox: %w", err)
	}
	log.Printf("sandbox: %s\n", sb.SandboxId)

	_, err = sb.Stdin.Write([]byte("this is input that should be mirrored by cat"))
	if err != nil {
		fmt.Errorf("Failed to write to sandbox stdin: %w", err)
	}
	err = sb.Stdin.Close()
	if err != nil {
		fmt.Errorf("Failed to close sandbox stdin: %w", err)
	}

	output, err := io.ReadAll(sb.Stdout)
	if err != nil {
		fmt.Errorf("Failed to read from sandbox stdout: %w", err)
	}

	log.Printf("output: %s\n", string(output))
}
