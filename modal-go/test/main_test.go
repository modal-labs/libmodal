package test

import (
	"context"
	"flag"
	"net/http"
	"os"
	"testing"

	modal "github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
	"go.uber.org/goleak"
)

// tc is the test Client, used for running tests against Modal infra.
var tc *modal.Client

// testMainWrapper wraps testing.M, so we can do cleanup before goleak looks for leaks.
// We need to do cleanup (closing the test client etc.) after m.Run() and before
// goleak.Find(), but goleak.VerifyTestMain doesn't offer a way to do this (goleak.Cleanup()
// is deferred until after goleak.Find() returns).
type testMainWrapper struct {
	m *testing.M
}

func (w *testMainWrapper) Run() int {
	c, err := modal.NewClient()
	if err != nil {
		panic(err)
	}
	tc = c
	defer c.Close()
	defer http.DefaultClient.CloseIdleConnections()
	return w.m.Run()
}

// checkLeaks enables package-level goroutine leak detection using goleak.VerifyTestMain.
// This approach works with t.Parallel() tests.
var checkLeaks = flag.Bool("check-leaks", false, "enable package-level goroutine leak detection (works with t.Parallel() tests)")

func TestMain(m *testing.M) {
	flag.Parse()

	wrapper := &testMainWrapper{m: m}
	if *checkLeaks {
		var goleakOptions []goleak.Option
		// Capture baseline goroutines before any test infrastructure is created.
		// The wrapper will create/close the client within its Run() call.
		goleakOptions = append(goleakOptions, goleak.IgnoreCurrent())

		goleak.VerifyTestMain(wrapper, goleakOptions...)
	} else {
		code := wrapper.Run()
		os.Exit(code)
	}
}

func terminateSandbox(g *gomega.WithT, sb *modal.Sandbox) {
	err := sb.Terminate(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}
