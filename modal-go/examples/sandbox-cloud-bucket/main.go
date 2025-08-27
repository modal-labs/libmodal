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

	image := modal.NewImageFromRegistry("alpine:3.21", nil)

	secret, err := modal.SecretFromName(ctx, "libmodal-aws-bucket-secret", nil)
	if err != nil {
		log.Fatalf("Failed to lookup Secret: %v", err)
	}

	keyPrefix := "data/"
	cloudBucketMount, err := modal.NewCloudBucketMount("my-s3-bucket", &modal.CloudBucketMountOptions{
		Secret:    secret,
		KeyPrefix: &keyPrefix,
		ReadOnly:  true,
	})
	if err != nil {
		log.Fatalf("Failed to create Cloud Bucket Mount: %v", err)
	}

	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command: []string{"sh", "-c", "ls -la /mnt/s3-bucket"},
		CloudBucketMounts: map[string]*modal.CloudBucketMount{
			"/mnt/s3-bucket": cloudBucketMount,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create Sandbox: %v", err)
	}

	log.Printf("S3 Sandbox: %s", sb.SandboxId)

	output, err := io.ReadAll(sb.Stdout)
	if err != nil {
		log.Fatalf("Failed to read from Sandbox stdout: %v", err)
	}

	log.Printf("Sandbox directory listing of /mnt/s3-bucket:\n%s", string(output))

	if err := sb.Terminate(); err != nil {
		log.Printf("Failed to terminate Sandbox: %v", err)
	}
}
