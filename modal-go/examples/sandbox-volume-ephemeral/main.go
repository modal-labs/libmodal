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
	mc, err := modal.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	app, err := mc.Apps.FromName(ctx, "libmodal-example", &modal.AppFromNameOptions{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to get or create App: %v", err)
	}

	image := mc.Images.FromRegistry("alpine:3.21", nil)

	volume, err := mc.Volumes.Ephemeral(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to create ephemeral Volume: %v", err)
	}
	defer volume.CloseEphemeral()

	writerSandbox, err := mc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateOptions{
		Command: []string{
			"sh",
			"-c",
			"echo 'Hello from writer Sandbox!' > /mnt/volume/message.txt",
		},
		Volumes: map[string]*modal.Volume{
			"/mnt/volume": volume,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create writer Sandbox: %v", err)
	}
	fmt.Printf("Writer Sandbox: %s\n", writerSandbox.SandboxId)
	defer func() {
		if err := writerSandbox.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", writerSandbox.SandboxId, err)
		}
	}()

	exitCode, err := writerSandbox.Wait(ctx)
	if err != nil {
		log.Fatalf("Failed to wait for writer Sandbox: %v", err)
	}
	fmt.Printf("Writer finished with exit code: %d\n", exitCode)

	if err := writerSandbox.Terminate(ctx); err != nil {
		log.Fatalf("Failed to terminate Sandbox %s: %v", writerSandbox.SandboxId, err)
	}

	readerSandbox, err := mc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateOptions{
		Command: []string{"cat", "/mnt/volume/message.txt"},
		Volumes: map[string]*modal.Volume{
			"/mnt/volume": volume.ReadOnly(),
		},
	})
	if err != nil {
		log.Fatalf("Failed to create reader Sandbox: %v", err)
	}
	fmt.Printf("Reader Sandbox: %s\n", readerSandbox.SandboxId)
	defer func() {
		if err := readerSandbox.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", readerSandbox.SandboxId, err)
		}
	}()

	readerOutput, err := io.ReadAll(readerSandbox.Stdout)
	if err != nil {
		log.Fatalf("Failed to read reader output: %v", err)
	}
	fmt.Printf("Reader output: %s", string(readerOutput))
}
