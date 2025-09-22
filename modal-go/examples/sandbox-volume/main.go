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

	app, err := mc.Apps.Lookup(ctx, "libmodal-example", &modal.LookupOptions{
		CreateIfMissing: true,
	})
	if err != nil {
		log.Fatalf("Failed to lookup App: %v", err)
	}

	image := mc.Images.FromRegistry("alpine:3.21", nil)

	volume, err := mc.Volumes.FromName(ctx, "libmodal-example-volume", &modal.VolumeFromNameOptions{
		CreateIfMissing: true,
	})
	if err != nil {
		log.Fatalf("Failed to create Volume: %v", err)
	}

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

	readerSandbox, err := mc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateOptions{
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

	rp, err := readerSandbox.Exec(ctx, []string{"cat", "/mnt/volume/message.txt"}, modal.ExecOptions{
		Stdout: modal.Pipe,
	})
	if err != nil {
		log.Fatalf("Failed to exec read command: %v", err)
	}
	readOutput, err := io.ReadAll(rp.Stdout)
	if err != nil {
		log.Fatalf("Failed to read output: %v", err)
	}
	fmt.Printf("Reader output: %s", string(readOutput))

	wp, err := readerSandbox.Exec(ctx, []string{"sh", "-c", "echo 'This should fail' >> /mnt/volume/message.txt"}, modal.ExecOptions{
		Stdout: modal.Pipe,
		Stderr: modal.Pipe,
	})
	if err != nil {
		log.Fatalf("Failed to exec write command: %v", err)
	}

	writeExitCode, err := wp.Wait(ctx)
	if err != nil {
		log.Fatalf("Failed to wait for write process: %v", err)
	}
	writeStderr, err := io.ReadAll(wp.Stderr)
	if err != nil {
		log.Fatalf("Failed to read stderr: %v", err)
	}

	fmt.Printf("Write attempt exit code: %d\n", writeExitCode)
	fmt.Printf("Write attempt stderr: %s", string(writeStderr))
}
