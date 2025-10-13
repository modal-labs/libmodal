package main

import (
	"fmt"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	client, err := modal.NewClient()
	if err != nil {
		panic(fmt.Sprintf("ERROR: Failed to create client: %v", err))
	}
	defer client.Close()

	fmt.Println(client.Version())
}
