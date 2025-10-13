package main

import (
	"fmt"
	"runtime/debug"

	_ "github.com/modal-labs/libmodal/modal-go"
)

func main() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		panic("ERROR: BuildInfo not available")
	}

	for _, dep := range info.Deps {
		if dep.Path == "github.com/modal-labs/libmodal/modal-go" {
			fmt.Printf("modal-go/%s\n", dep.Version)
			return
		}
	}

	panic("ERROR: modal-go not found in dependencies")
}
