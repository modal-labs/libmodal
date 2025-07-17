package main

import (
	"context"
	"fmt"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()

	app, err := modal.AppLookup(ctx, "libmodal-example", &modal.LookupOptions{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to lookup or create app: %v", err)
	}

	image, err := app.ImageFromRegistry("alpine:3.21", nil)
	if err != nil {
		log.Fatalf("Failed to create image from registry: %v", err)
	}

	// Create a sandbox that waits for input, then exits with code 42
	sandbox, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command: []string{"sh", "-c", "read line; exit 42"},
	})
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	fmt.Printf("Started sandbox: %s\n", sandbox.SandboxId)

	initialPoll, err := sandbox.Poll()
	if err != nil {
		log.Fatalf("Failed to poll sandbox: %v", err)
	}
	fmt.Printf("Poll result while running: %v\n", initialPoll)

	fmt.Println("\nSending input to trigger completion...")
	_, err = sandbox.Stdin.Write([]byte("hello, goodbye\n"))
	if err != nil {
		log.Fatalf("Failed to write to stdin: %v", err)
	}
	err = sandbox.Stdin.Close()
	if err != nil {
		log.Fatalf("Failed to close stdin: %v", err)
	}

	exitCode, err := sandbox.Wait()
	if err != nil {
		log.Fatalf("Failed to wait for sandbox: %v", err)
	}
	fmt.Printf("\nSandbox completed with exit code: %d\n", exitCode)

	finalPoll, err := sandbox.Poll()
	if err != nil {
		log.Fatalf("Failed to poll sandbox after completion: %v", err)
	}
	fmt.Printf("Poll result after completion: %d\n", *finalPoll)
}
