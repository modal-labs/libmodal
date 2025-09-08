package test

import (
	"os"
	"testing"

	modal "github.com/modal-labs/libmodal/modal-go"
)

// tc is the test Client, used for running tests against Modal infra.
var tc *modal.Client

func TestMain(m *testing.M) {
	c, err := modal.NewClient()
	if err != nil {
		panic(err)
	}
	tc = c

	code := m.Run()
	os.Exit(code)
}
