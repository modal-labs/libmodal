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
	mc, err := modal.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	app, err := mc.Apps.FromName(ctx, "libmodal-example", &modal.LookupOptions{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to get or create App: %v", err)
	}

	baseImage := mc.Images.FromRegistry("alpine:3.21", nil)

	sb, err := mc.Sandboxes.Create(ctx, app, baseImage, &modal.SandboxCreateOptions{})
	if err != nil {
		log.Fatalf("Failed to create Sandbox: %v", err)
	}
	log.Printf("Started Sandbox: %s", sb.SandboxId)
	defer func() {
		if err := sb.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", sb.SandboxId, err)
		}
	}()

	_, err = sb.Exec(ctx, []string{"mkdir", "-p", "/app/data"}, modal.ExecOptions{})
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	_, err = sb.Exec(ctx, []string{"sh", "-c", "echo 'This file was created in the first Sandbox' > /app/data/info.txt"}, modal.ExecOptions{})
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	log.Printf("Created file in first Sandbox")

	snapshotImage, err := sb.SnapshotFilesystem(ctx, 55*time.Second)
	if err != nil {
		log.Fatalf("Failed to snapshot filesystem: %v", err)
	}
	log.Printf("Filesystem snapshot created with Image ID: %s", snapshotImage.ImageId)

	err = sb.Terminate(ctx)
	if err != nil {
		log.Fatalf("Failed to terminate Sandbox %s: %v", sb.SandboxId, err)
	}
	log.Printf("Terminated first Sandbox")

	// Create new Sandbox from snapshot Image
	sb2, err := mc.Sandboxes.Create(ctx, app, snapshotImage, nil)
	if err != nil {
		log.Fatalf("Failed to create Sandbox from snapshot: %v", err)
	}
	log.Printf("Started new Sandbox from snapshot: %s", sb2.SandboxId)

	defer func() {
		if err := sb2.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", sb2.SandboxId, err)
		}
	}()

	proc, err := sb2.Exec(ctx, []string{"cat", "/app/data/info.txt"}, modal.ExecOptions{})
	if err != nil {
		log.Fatalf("Failed to exec cat command: %v", err)
	}

	content, err := io.ReadAll(proc.Stdout)
	if err != nil {
		log.Fatalf("Failed to read output: %v", err)
	}
	log.Printf("File data read in second Sandbox: %s", string(content))
}
