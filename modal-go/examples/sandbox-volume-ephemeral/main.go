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

	app, err := modal.AppLookup(ctx, "libmodal-example", &modal.LookupOptions{
		CreateIfMissing: true,
	})
	if err != nil {
		log.Fatalf("Failed to lookup app: %v", err)
	}

	image := modal.NewImageFromRegistry("alpine:3.21", nil)

	volume, err := modal.VolumeEphemeral(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to create ephemeral volume: %v", err)
	}
	defer volume.CloseEphemeral()

	writerSandbox, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command: []string{
			"sh",
			"-c",
			"echo 'Hello from writer sandbox!' > /mnt/volume/message.txt",
		},
		Volumes: map[string]*modal.Volume{
			"/mnt/volume": volume,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create writer sandbox: %v", err)
	}
	fmt.Printf("Writer sandbox: %s\n", writerSandbox.SandboxId)

	exitCode, err := writerSandbox.Wait()
	if err != nil {
		log.Fatalf("Failed to wait for writer sandbox: %v", err)
	}
	fmt.Printf("Writer finished with exit code: %d\n", exitCode)

	if err := writerSandbox.Terminate(); err != nil {
		log.Printf("Failed to terminate writer sandbox: %v", err)
	}

	readerSandbox, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command: []string{"cat", "/mnt/volume/message.txt"},
		Volumes: map[string]*modal.Volume{
			"/mnt/volume": volume.ReadOnly(),
		},
	})
	if err != nil {
		log.Fatalf("Failed to create reader sandbox: %v", err)
	}
	fmt.Printf("Reader sandbox: %s\n", readerSandbox.SandboxId)

	readerOutput, err := io.ReadAll(readerSandbox.Stdout)
	if err != nil {
		log.Fatalf("Failed to read reader output: %v", err)
	}
	fmt.Printf("Reader output: %s", string(readerOutput))

	if err := readerSandbox.Terminate(); err != nil {
		log.Printf("Failed to terminate reader sandbox: %v", err)
	}
}
