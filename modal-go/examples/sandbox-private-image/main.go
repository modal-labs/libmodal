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
		log.Fatalf("Failed to lookup or create App: %v", err)
	}

	secret, err := modal.SecretFromName(ctx, "libmodal-aws-ecr-test", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},
	})
	if err != nil {
		log.Fatalf("Failed to get Secret: %v", err)
	}

	image := modal.NewImageFromAwsEcr("459781239556.dkr.ecr.us-east-1.amazonaws.com/ecr-private-registry-test-7522615:python", secret)

	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command: []string{"python", "-c", `import sys; sys.stdout.write(sys.stdin.read())`},
	})
	if err != nil {
		log.Fatalf("Failed to create Sandbox: %v", err)
	}
	log.Printf("Sandbox: %s\n", sb.SandboxId)

	_, err = sb.Stdin.Write([]byte("this is input that should be mirrored by the Python one-liner"))
	if err != nil {
		log.Fatalf("Failed to write to Sandbox stdin: %v", err)
	}
	err = sb.Stdin.Close()
	if err != nil {
		log.Fatalf("Failed to close Sandbox stdin: %v", err)
	}

	output, err := io.ReadAll(sb.Stdout)
	if err != nil {
		log.Fatalf("Failed to read from Sandbox stdout: %v", err)
	}

	log.Printf("output: %s\n", string(output))
}
