// We use `Image.Build` to create an Image object on Modal
// that eagerly pulls from the registry. The first sandbox created with this image
// will ues this "pre-warmed" image and will start faster.
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
		log.Fatalf("Failed to lookup or create app: %v", err)
	}

	// With `.Build(app)`, we create an Image object on Modal that eagerly pulls
	// from the registry.
	image, err := modal.NewImageFromRegistry("alpine:3.21", nil).Build(app)
	if err != nil {
		log.Fatalf("Unable to build image: %v", err)
	}
	log.Printf("Image has id: %v", image.ImageId)

	// You can save the ImageId and create a new image object that referes to it.
	imageId := image.ImageId
	image2 := modal.NewImageFromId(imageId)

	sb, err := app.CreateSandbox(image2, &modal.SandboxOptions{
		Command: []string{"cat"},
	})
	defer sb.Terminate()
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	log.Printf("sandbox: %s\n", sb.SandboxId)
}
