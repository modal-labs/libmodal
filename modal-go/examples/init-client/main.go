// This example configures a client using a `CUSTOM_MODAL_ID` and `CUSTOM_MODAL_SECRET` environment variable.

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()

	modal_id := os.Getenv("CUSTOM_MODAL_ID")
	if modal_id == "" {
		log.Fatal("CUSTOM_MODAL_ID environment variable not set")
	}
	modal_secret := os.Getenv("CUSTOM_MODAL_SECRET")
	if modal_secret == "" {
		log.Fatal("CUSTOM_MODAL_SECRET environment variable not set")
	}

	modal.InitializeClient(modal.ClientOptions{
		TokenId:     modal_id,
		TokenSecret: modal_secret,
	})

	echo, err := modal.FunctionLookup(ctx, "libmodal-test-support", "echo_string", nil)
	if err != nil {
		log.Fatalf("Failed to lookup function: %v", err)
	}
	fmt.Printf("%#v\n", echo)
}
