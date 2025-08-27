// We use `Image.Build` to create an Image object on Modal
// that eagerly pulls from the registry. The first Sandbox created with this Image
// will ues this "pre-warmed" Image and will start faster.
package main

import (
	"context"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()

	app, err := modal.AppLookup(ctx, "libmodal-example", &modal.LookupOptions{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to lookup or create App: %v", err)
	}

	// With `.Build(app)`, we create an Image object on Modal that eagerly pulls
	// from the registry.
	image, err := modal.NewImageFromRegistry("alpine:3.21", nil).Build(app)
	if err != nil {
		log.Fatalf("Unable to build Image: %v", err)
	}
	log.Printf("Image has ID: %v", image.ImageId)

	// You can save the ImageId and create a new Image object that referes to it.
	imageId := image.ImageId
	image2 := modal.NewImageFromId(imageId)

	sb, err := app.CreateSandbox(image2, &modal.SandboxOptions{
		Command: []string{"cat"},
	})
	defer sb.Terminate()
	if err != nil {
		log.Fatalf("Failed to create Sandbox: %v", err)
	}
	log.Printf("Sandbox: %s\n", sb.SandboxId)
}
