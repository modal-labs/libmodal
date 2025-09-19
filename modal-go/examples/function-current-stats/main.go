// Demonstrates how to get current statistics for a Modal Function.

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	function, err := modal.FunctionLookup(context.Background(), "libmodal-test-support", "echo_string", nil)
	if err != nil {
		log.Fatalf("Failed to lookup Function: %v", err)
	}

	stats, err := function.GetCurrentStats()
	if err != nil {
		log.Fatalf("Failed to get Function stats: %v", err)
	}

	fmt.Println("Function Statistics:")
	fmt.Printf("  Backlog: %d inputs\n", stats.Backlog)
	fmt.Printf("  Total Runners: %d containers\n", stats.NumTotalRunners)
}
