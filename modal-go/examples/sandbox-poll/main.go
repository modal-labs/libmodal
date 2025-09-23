package main

import (
	"context"
	"fmt"
	"log"

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

	image := mc.Images.FromRegistry("alpine:3.21", nil)

	// Create a sandbox that waits for input, then exits with code 42
	sandbox, err := mc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateOptions{
		Command: []string{"sh", "-c", "read line; exit 42"},
	})
	if err != nil {
		log.Fatalf("Failed to create Sandbox: %v", err)
	}
	fmt.Printf("Started Sandbox: %s\n", sandbox.SandboxId)
	defer func() {
		if err := sandbox.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", sandbox.SandboxId, err)
		}
	}()

	initialPoll, err := sandbox.Poll(ctx)
	if err != nil {
		log.Fatalf("Failed to poll Sandbox: %v", err)
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

	exitCode, err := sandbox.Wait(ctx)
	if err != nil {
		log.Fatalf("Failed to wait for Sandbox: %v", err)
	}
	fmt.Printf("\nSandbox completed with exit code: %d\n", exitCode)

	finalPoll, err := sandbox.Poll(ctx)
	if err != nil {
		log.Fatalf("Failed to poll Sandbox after completion: %v", err)
	}
	fmt.Printf("Poll result after completion: %d\n", *finalPoll)
}
