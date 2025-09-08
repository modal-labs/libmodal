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

	app, err := mc.Apps.Lookup(ctx, "libmodal-example", &modal.LookupOptions{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to lookup or create App: %v", err)
	}

	image := mc.Images.FromRegistry("alpine/curl:8.14.1", nil)

	proxy, err := mc.Proxies.FromName(ctx, "libmodal-test-proxy", nil)
	if err != nil {
		log.Fatalf("Failed to get Proxy: %v", err)
	}
	log.Printf("Using Proxy: %s", proxy.ProxyId)

	sb, err := mc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateOptions{
		Proxy: proxy,
	})
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	log.Printf("Created sandbox with proxy: %s", sb.SandboxId)

	p, err := sb.Exec(ctx, []string{"curl", "-s", "ifconfig.me"}, modal.ExecOptions{})
	if err != nil {
		log.Fatalf("Failed to start IP fetch command: %v", err)
	}

	ip, err := io.ReadAll(p.Stdout)
	if err != nil {
		log.Fatalf("Failed to read IP output: %v", err)
	}

	log.Printf("External IP: %s", string(ip))

	err = sb.Terminate(ctx)
	if err != nil {
		log.Fatalf("Failed to terminate Sandbox: %v", err)
	}
}
