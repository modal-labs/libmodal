package main

import (
	"context"
	"io"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()

	app, err := modal.AppLookup(ctx, "libmodal-example", &modal.LookupOptions{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to lookup or create app: %v", err)
	}

	image, err := app.ImageFromRegistry("alpine/curl:8.14.1", nil)
	if err != nil {
		log.Fatalf("Failed to create image from registry: %v", err)
	}

	proxy, err := modal.ProxyFromName(ctx, "libmodal-test-proxy", nil)
	if err != nil {
		log.Fatalf("Failed to get proxy: %v", err)
	}
	log.Printf("Using proxy: %s", proxy.ProxyId)

	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Proxy: proxy,
	})
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	log.Printf("Created sandbox with proxy: %s", sb.SandboxId)

	p, err := sb.Exec([]string{"curl", "-s", "ifconfig.me"}, modal.ExecOptions{})
	if err != nil {
		log.Fatalf("Failed to start IP fetch command: %v", err)
	}

	ip, err := io.ReadAll(p.Stdout)
	if err != nil {
		log.Fatalf("Failed to read IP output: %v", err)
	}

	log.Printf("External IP: %s", string(ip))

	err = sb.Terminate()
	if err != nil {
		log.Fatalf("Failed to terminate sandbox: %v", err)
	}
}
