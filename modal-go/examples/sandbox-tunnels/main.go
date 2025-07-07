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

	app, err := modal.AppLookup(ctx, "libmodal-example", &modal.LookupOptions{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to lookup or create app: %v", err)
	}

	// Create a sandbox with Python's built-in HTTP server
	image, err := app.ImageFromRegistry("python:3.12-alpine", nil)
	if err != nil {
		log.Fatalf("Failed to create image from registry: %v", err)
	}

	sandbox, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command:        []string{"python3", "-m", "http.server", "8000"},
		EncryptedPorts: []int{8000},
		Timeout:        1 * time.Minute,
	})
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	log.Printf("Sandbox created: %s", sandbox.SandboxId)

	log.Printf("Waiting for server to start...")
	time.Sleep(3 * time.Second)

	log.Printf("Getting tunnel information...")
	tunnels, err := sandbox.Tunnels(30 * time.Second)
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

	err = sandbox.Terminate()
	if err != nil {
		log.Fatalf("Failed to terminate sandbox: %v", err)
	}
}
