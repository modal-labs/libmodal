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

	volume, err := modal.VolumeFromName(ctx, "libmodal-example-volume", &modal.VolumeFromNameOptions{
		CreateIfMissing: true,
	})
	if err != nil {
		log.Fatalf("Failed to create volume: %v", err)
	}

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

	readerSandbox, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Volumes: map[string]*modal.Volume{
			"/mnt/volume": volume.ReadOnly(),
		},
	})
	if err != nil {
		log.Fatalf("Failed to create reader sandbox: %v", err)
	}
	fmt.Printf("Reader sandbox: %s\n", readerSandbox.SandboxId)

	rp, err := readerSandbox.Exec([]string{"cat", "/mnt/volume/message.txt"}, modal.ExecOptions{
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

	wp, err := readerSandbox.Exec([]string{"sh", "-c", "echo 'This should fail' >> /mnt/volume/message.txt"}, modal.ExecOptions{
		Stdout: modal.Pipe,
		Stderr: modal.Pipe,
	})
	if err != nil {
		log.Fatalf("Failed to exec write command: %v", err)
	}

	writeExitCode, err := wp.Wait()
	if err != nil {
		log.Fatalf("Failed to wait for write process: %v", err)
	}
	writeStderr, err := io.ReadAll(wp.Stderr)
	if err != nil {
		log.Fatalf("Failed to read stderr: %v", err)
	}

	fmt.Printf("Write attempt exit code: %d\n", writeExitCode)
	fmt.Printf("Write attempt stderr: %s", string(writeStderr))

	if err := writerSandbox.Terminate(); err != nil {
		log.Printf("Failed to terminate writer sandbox: %v", err)
	}
	if err := readerSandbox.Terminate(); err != nil {
		log.Printf("Failed to terminate reader sandbox: %v", err)
	}
}
