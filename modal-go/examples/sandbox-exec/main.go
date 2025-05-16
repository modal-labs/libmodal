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

	image, err := app.ImageFromRegistry("python:3.13-slim")
	if err != nil {
		fmt.Errorf("Failed to create image from registry: %w", err)
	}

	sb, err := app.CreateSandbox(image, modal.SandboxOptions{})
	if err != nil {
		fmt.Errorf("Failed to create sandbox: %w", err)
	}
	log.Println("Started sandbox:", sb.SandboxId)
	defer sb.Terminate()

	p, err := sb.Exec(
		[]string{
			"python",
			"-c",
			`
import time
import sys
for i in range(50000):
	if i % 1000 == 0:
		time.sleep(0.01)
	print(i)
	print(i, file=sys.stderr)`,
		},
		modal.ExecOptions{
			Stdout: modal.Pipe,
			Stderr: modal.Pipe,
		},
	)
	if err != nil {
		fmt.Errorf("Failed to execute command in sandbox: %w", err)
	}

	contentStdout, err := io.ReadAll(p.Stdout)
	if err != nil {
		fmt.Errorf("Failed to read stdout: %w", err)
	}
	contentStderr, err := io.ReadAll(p.Stderr)
	if err != nil {
		fmt.Errorf("Failed to read stderr: %w", err)
	}

	log.Printf("Got %d bytes stdout and %d bytes stderr\n", len(contentStdout), len(contentStderr))
	returnCode, err := p.Wait()
	if err != nil {
		fmt.Errorf("Failed to wait for process completion: %w", err)
	}
	log.Println("Return code:", returnCode)

}
