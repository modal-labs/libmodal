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

	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command: []string{"sleep", "3600"}, // Keep sandbox alive
	})
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

	_, err = writeFile.WriteString("Hello, Modal filesystem!\n")
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	_, err = writeFile.WriteString("This is a test file.\n")
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	if err := writeFile.Close(); err != nil {
		log.Fatalf("Failed to close file: %v", err)
	}

	// Read the file
	readFile, err := sb.Open("/tmp/example.txt", "r")
	if err != nil {
		log.Fatalf("Failed to open file for reading: %v", err)
	}

	content, err := readFile.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	fmt.Printf("File content:\n%s", string(content))
	if err := readFile.Close(); err != nil {
		log.Fatalf("Failed to close file: %v", err)
	}

	// Demonstrate seeking
	seekFile, err := sb.Open("/tmp/example.txt", "r")
	if err != nil {
		log.Fatalf("Failed to open file for seeking: %v", err)
	}

	_, err = seekFile.Seek(7, 0) // Seek to position 7 from beginning
	if err != nil {
		log.Fatalf("Failed to seek in file: %v", err)
	}

	buffer := make([]byte, 5)
	n, err := seekFile.Read(buffer)
	if err != nil && err != io.EOF {
		log.Fatalf("Failed to read from position: %v", err)
	}

	fmt.Printf("From position 7: %s\n", string(buffer[:n]))
	if err := seekFile.Close(); err != nil {
		log.Fatalf("Failed to close file: %v", err)
	}
}
