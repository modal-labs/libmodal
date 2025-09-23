package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()
	mc, err := modal.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	app, err := mc.Apps.FromName(ctx, "libmodal-example", &modal.AppFromNameParams{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to get or create App: %v", err)
	}

	// Create a Sandbox with Python's built-in HTTP server
	image := mc.Images.FromRegistry("python:3.12-alpine", nil)

	sandbox, err := mc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Command:        []string{"python3", "-m", "http.server", "8000"},
		EncryptedPorts: []int{8000},
		Timeout:        1 * time.Minute,
		IdleTimeout:    30 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create Sandbox: %v", err)
	}
	defer func() {
		if err := sandbox.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", sandbox.SandboxId, err)
		}
	}()

	log.Printf("Sandbox created: %s", sandbox.SandboxId)

	log.Printf("Waiting for server to start...")
	time.Sleep(3 * time.Second)

	log.Printf("Getting tunnel information...")
	tunnels, err := sandbox.Tunnels(ctx, 30*time.Second)
	if err != nil {
		log.Fatalf("Failed to get tunnels: %v", err)
	}

	tunnel := tunnels[8000]
	if tunnel == nil {
		log.Fatalf("No tunnel found for port 8000")
	}

	log.Printf("Tunnel information:")
	log.Printf("  URL: %s", tunnel.URL())
	log.Printf("  Port: %d", tunnel.Port)

	log.Printf("\nMaking GET request to the tunneled server at %s", tunnel.URL())

	// Make a GET request to the tunneled server
	resp, err := http.Get(tunnel.URL())
	if err != nil {
		log.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("HTTP error! status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	// Display first 500 characters of the response
	bodyStr := string(body)
	if len(bodyStr) > 500 {
		bodyStr = bodyStr[:500]
	}

	fmt.Printf("\nDirectory listing from server (first 500 chars):\n%s\n", bodyStr)

	log.Printf("\n✅ Successfully connected to the tunneled server!")
}
