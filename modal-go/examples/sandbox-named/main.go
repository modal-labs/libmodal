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

	image := modal.NewImageFromRegistry("alpine:3.21", nil)

	sandboxName := "libmodal-example-named-sandbox"

	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Name:    sandboxName,
		Command: []string{"cat"},
	})
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	fmt.Printf("Created sandbox with name: %s\n", sandboxName)
	fmt.Printf("Sandbox ID: %s\n", sb.SandboxId)

	_, err = app.CreateSandbox(image, &modal.SandboxOptions{
		Name:    sandboxName,
		Command: []string{"cat"},
	})
	if err != nil {
		if alreadyExistsErr, ok := err.(modal.AlreadyExistsError); ok {
			fmt.Printf("Trying to create one more Sandbox with the same name fails: %s\n", alreadyExistsErr.Exception)
		} else {
			log.Fatalf("Unexpected error: %v", err)
		}
	}

	sbFromName, err := modal.SandboxFromName(ctx, "libmodal-example", sandboxName, nil)
	if err != nil {
		log.Fatalf("Failed to get Sandbox by name: %v", err)
	}
	fmt.Printf("Retrieved the same Sandbox from name: %s\n", sbFromName.SandboxId)

	_, err = sbFromName.Stdin.Write([]byte("hello, named Sandbox"))
	if err != nil {
		log.Fatalf("Failed to write to Sandbox stdin: %v", err)
	}
	err = sbFromName.Stdin.Close()
	if err != nil {
		log.Fatalf("Failed to close Sandbox stdin: %v", err)
	}

	fmt.Println("Reading output:")
	output, err := io.ReadAll(sbFromName.Stdout)
	if err != nil {
		log.Fatalf("Failed to read output from Sandbox stdout: %v", err)
	}
	fmt.Printf("%s\n", output)

	err = sb.Terminate()
	if err != nil {
		log.Fatalf("Failed to terminate Sandbox: %v", err)
	}
	fmt.Println("Sandbox terminated")
}
