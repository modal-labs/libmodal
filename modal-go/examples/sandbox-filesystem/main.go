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
		log.Fatalf("Failed to lookup or create app: %v", err)
	}

	image, err := app.ImageFromRegistry("alpine:3.21")
	if err != nil {
		log.Fatalf("Failed to create image from registry: %v", err)
	}

	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{})
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	log.Printf("Started sandbox: %s", sb.SandboxId)

	defer func() {
		if err := sb.Terminate(); err != nil {
			log.Printf("Failed to terminate sandbox: %v", err)
		}
	}()

	// Write a file
	writeFile, err := sb.Open("/tmp/example.txt", "w")
	if err != nil {
		log.Fatalf("Failed to open file for writing: %v", err)
	}

	_, err = writeFile.Write([]byte("Hello, Modal filesystem!\n"))
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	if err := writeFile.Close(); err != nil {
		log.Fatalf("Failed to close file: %v", err)
	}

	// Read the file
	reader, err := sb.Open("/tmp/example.txt", "r")
	if err != nil {
		log.Fatalf("Failed to open file for reading: %v", err)
	}

	content, err := io.ReadAll(reader)
	if err != nil && err != io.EOF {
		log.Fatalf("Failed to read file: %v", err)
	}

	fmt.Printf("File content:\n%s", string(content))
	if err := reader.Close(); err != nil {
		log.Fatalf("Failed to close file: %v", err)
	}
}
