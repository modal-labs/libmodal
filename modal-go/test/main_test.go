package test

import (
	"context"
	"flag"
	"os"
	"testing"

	modal "github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
	"go.uber.org/goleak"
)

// tc is the test Client, used for running tests against Modal infra.
var tc *modal.Client

// checkLeaks enables package-level goroutine leak detection using goleak.VerifyTestMain.
// This approach works with t.Parallel() tests.
var checkLeaks = flag.Bool("check-leaks", false, "enable package-level goroutine leak detection (works with t.Parallel() tests)")

var goleakOptions []goleak.Option

func TestMain(m *testing.M) {
	flag.Parse()

	c, err := modal.NewClient()
	if err != nil {
		panic(err)
	}
	tc = c

	// Capture baseline goroutines after client creation
	goleakOptions = append(goleakOptions, goleak.IgnoreCurrent())

	if *checkLeaks {
		goleak.VerifyTestMain(m, goleakOptions...)
	} else {
		code := m.Run()
		os.Exit(code)
	}
}

func terminateSandbox(g *gomega.WithT, sb *modal.Sandbox) {
	err := sb.Terminate(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}
