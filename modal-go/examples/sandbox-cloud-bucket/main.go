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

	image, err := app.ImageFromRegistry("alpine:3.21", nil)
	if err != nil {
		log.Fatalf("Failed to create image from registry: %v", err)
	}

	secret, err := modal.SecretFromName(ctx, "libmodal-aws-bucket-secret", nil)
	if err != nil {
		log.Fatalf("Failed to lookup secret: %v", err)
	}

	keyPrefix := "data/"
	cloudBucketMount, err := modal.NewCloudBucketMount("my-s3-bucket", &modal.CloudBucketMountOptions{
		Secret:    secret,
		KeyPrefix: &keyPrefix,
		ReadOnly:  true,
	})
	if err != nil {
		log.Fatalf("Failed to create cloud bucket mount: %v", err)
	}

	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command: []string{"sh", "-c", "ls -la /mnt/s3-bucket"},
		CloudBucketMounts: map[string]*modal.CloudBucketMount{
			"/mnt/s3-bucket": cloudBucketMount,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	log.Printf("S3 sandbox: %s", sb.SandboxId)

	output, err := io.ReadAll(sb.Stdout)
	if err != nil {
		log.Fatalf("Failed to read from sandbox stdout: %v", err)
	}

	log.Printf("Sandbox directory listing of /mnt/s3-bucket:\n%s", string(output))

	if err := sb.Terminate(); err != nil {
		log.Printf("Failed to terminate sandbox: %v", err)
	}
}
