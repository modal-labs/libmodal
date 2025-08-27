package main

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()

	app, err := modal.AppLookup(ctx, "libmodal-example", &modal.LookupOptions{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to lookup App: %v", err)
	}

	baseImage := modal.NewImageFromRegistry("alpine:3.21", nil)

	sb, err := app.CreateSandbox(baseImage, &modal.SandboxOptions{})
	if err != nil {
		log.Fatalf("Failed to create Sandbox: %v", err)
	}
	log.Printf("Started Sandbox: %s", sb.SandboxId)

	defer sb.Terminate()

	_, err = sb.Exec([]string{"mkdir", "-p", "/app/data"}, modal.ExecOptions{})
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	_, err = sb.Exec([]string{"sh", "-c", "echo 'This file was created in the first Sandbox' > /app/data/info.txt"}, modal.ExecOptions{})
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	log.Printf("Created file in first Sandbox")

	snapshotImage, err := sb.SnapshotFilesystem(55 * time.Second)
	if err != nil {
		log.Fatalf("Failed to snapshot filesystem: %v", err)
	}
	log.Printf("Filesystem snapshot created with Image ID: %s", snapshotImage.ImageId)

	sb.Terminate()
	log.Printf("Terminated first Sandbox")

	// Create new Sandbox from snapshot Image
	sb2, err := app.CreateSandbox(snapshotImage, nil)
	if err != nil {
		log.Fatalf("Failed to create Sandbox from snapshot: %v", err)
	}
	log.Printf("Started new Sandbox from snapshot: %s", sb2.SandboxId)

	defer sb2.Terminate()

	proc, err := sb2.Exec([]string{"cat", "/app/data/info.txt"}, modal.ExecOptions{})
	if err != nil {
		log.Fatalf("Failed to exec cat command: %v", err)
	}

	content, err := io.ReadAll(proc.Stdout)
	if err != nil {
		log.Fatalf("Failed to read output: %v", err)
	}
	log.Printf("File data read in second Sandbox: %s", string(content))

	sb2.Terminate()
}
