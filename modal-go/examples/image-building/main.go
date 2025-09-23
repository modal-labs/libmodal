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
	mc, err := modal.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	app, err := mc.Apps.FromName(ctx, "libmodal-example", &modal.AppFromNameOptions{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to get or create App: %v", err)
	}

	secret, err := mc.Secrets.FromMap(ctx, map[string]string{
		"CURL_VERSION": "8.12.1-r1",
	}, nil)
	if err != nil {
		log.Fatal(err)
	}

	image := mc.Images.FromRegistry("alpine:3.21", nil).
		DockerfileCommands([]string{"RUN apk add --no-cache curl=$CURL_VERSION"}, &modal.ImageDockerfileCommandsOptions{
			Secrets: []*modal.Secret{secret},
		}).
		DockerfileCommands([]string{"ENV SERVER=ipconfig.me"}, nil)

	sb, err := mc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateOptions{
		Command: []string{"sh", "-c", "curl -Ls $SERVER"},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := sb.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", sb.SandboxId, err)
		}
	}()

	fmt.Println("Created Sandbox with ID:", sb.SandboxId)

	output, err := io.ReadAll(sb.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sandbox output:", string(output))
}
