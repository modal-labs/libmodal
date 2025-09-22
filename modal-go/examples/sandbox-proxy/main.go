package main

import (
	"context"
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

	app, err := mc.Apps.FromName(ctx, "libmodal-example", &modal.AppFromNameParams{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to get or create App: %v", err)
	}

	image := mc.Images.FromRegistry("alpine/curl:8.14.1", nil)

	proxy, err := mc.Proxies.FromName(ctx, "libmodal-test-proxy", nil)
	if err != nil {
		log.Fatalf("Failed to get Proxy: %v", err)
	}
	log.Printf("Using Proxy: %s", proxy.ProxyID)

	sb, err := mc.Sandboxes.Create(ctx, *app, *image, &modal.SandboxCreateParams{
		Proxy: proxy,
	})
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	log.Printf("Created sandbox with proxy: %s", sb.SandboxID)
	defer func() {
		if err := sb.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", sb.SandboxID, err)
		}
	}()

	p, err := sb.Exec(ctx, []string{"curl", "-s", "ifconfig.me"}, nil)
	if err != nil {
		log.Fatalf("Failed to start IP fetch command: %v", err)
	}

	ip, err := io.ReadAll(p.Stdout)
	if err != nil {
		log.Fatalf("Failed to read IP output: %v", err)
	}

	log.Printf("External IP: %s", string(ip))
}
