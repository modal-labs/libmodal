package test

import (
	"context"
	"os"
	"testing"

	modal "github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
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

func terminateSandbox(g *gomega.WithT, sb *modal.Sandbox) {
	err := sb.Terminate(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}
