package main

import (
	"testing"

	"github.com/modal-labs/libmodal/modal-go"
)

func TestVersion(t *testing.T) {
	client, err := modal.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	want := "v0.0.99"
	got := client.Version()
	if want != got {
		t.Fatalf("got version %s != expected version %s", got, want)
	}
}
