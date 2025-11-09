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

// checkLeaksPerTest enables per-test goroutine leak detection using goleak.VerifyNone.
// To enable leak detection for a test, call `parallelOrCheckLeaks(t)` in the beginning of the test function, where you would normally call `t.Parallel()`.
var checkLeaksPerTest = flag.Bool("check-leaks-per-test", false, "enable per-test goroutine leak detection (disables t.Parallel() execution)")

var goleakOptions []goleak.Option

func TestMain(m *testing.M) {
	flag.Parse()

	if *checkLeaks && *checkLeaksPerTest {
		panic("Error: -check-leaks and -check-leaks-per-test are mutually exclusive")
	}

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

// parallelOrCheckLeaks configures the test for parallel execution or leak detection.
// When -check-leaks-per-test is passed to go test, runs sequentially and with goleak.VerifyNone.
// Otherwise it enables t.Parallel() for fast concurrent tests.
// See: https://github.com/uber-go/goleak?tab=readme-ov-file#note
func parallelOrCheckLeaks(t *testing.T, opts ...goleak.Option) {
	t.Helper()
	if !*checkLeaksPerTest {
		t.Parallel()
	} else {
		t.Cleanup(func() {
			if !t.Failed() {
				goleak.VerifyNone(t, append(goleakOptions, opts...)...)
			}
		})
	}
}

func terminateSandbox(g *gomega.WithT, sb *modal.Sandbox) {
	err := sb.Terminate(context.Background())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}
