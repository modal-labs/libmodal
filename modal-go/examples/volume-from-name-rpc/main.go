package main

import (
	"context"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()
	mc, err := modal.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	v, err := mc.Volumes.FromName(ctx, "bad-name", nil)
	if err != nil {
		log.Fatalf("Failed to find volume: %v", err)
	}
	log.Panicln(v)

}
