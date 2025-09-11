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

	app, err := modal.AppLookup(ctx, "libmodal-example", &modal.LookupOptions{CreateIfMissing: true})
	if err != nil {
		log.Fatal(err)
	}

	secret, err := modal.SecretFromMap(ctx, map[string]string{
		"CURL_VERSION": "8.12.1-r1",
	}, nil)
	if err != nil {
		log.Fatal(err)
	}

	image := modal.NewImageFromRegistry("alpine:3.21", nil).
		DockerfileCommands([]string{"RUN apk add --no-cache curl=$CURL_VERSION"}, &modal.ImageDockerfileCommandsOptions{
			Secrets: []*modal.Secret{secret},
		}).
		DockerfileCommands([]string{"ENV SERVER=ipconfig.me"}, nil)

	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command: []string{"sh", "-c", "curl -Ls $SERVER"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created Sandbox with ID:", sb.SandboxId)

	output, err := io.ReadAll(sb.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sandbox output:", string(output))

	err = sb.Terminate()
	if err != nil {
		log.Fatal(err)
	}
}
