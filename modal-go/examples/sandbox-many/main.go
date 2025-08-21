package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()
	n_sandboxes := 30

	app, err := modal.AppLookup(ctx, "libmodal-example", &modal.LookupOptions{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to lookup or create app: %v", err)
	}

	image := modal.NewImageFromRegistry("alpine:3.21", nil)

	type SandboxWithTunnel struct {
		sandbox *modal.Sandbox
		tunnels map[int]*modal.Tunnel
	}

	sandboxes := []SandboxWithTunnel{}

	for i := 0; i < n_sandboxes; i++ {
		sb, err := app.CreateSandbox(image, nil)
		if err != nil {
			log.Fatalf("Failed to create sandbox: %v", err)
		}
		sb.Exec([]string{"echo", fmt.Sprintf("%d", i)}, modal.ExecOptions{})
		tunnels, err := sb.Tunnels(30 * time.Second)
		if err != nil {
			log.Fatalf("Failed to create sandbox: %v", err)
		}

		sandboxes = append(sandboxes, SandboxWithTunnel{sandbox: sb, tunnels: tunnels})
	}

	for _, sandbox := range sandboxes {
		sandbox.sandbox.Terminate()
	}

}
