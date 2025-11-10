package test

import (
	"context"
	"flag"
	"fmt"
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
var checkLeaks = flag.Bool("check-leaks", true, "enable package-level goroutine leak detection")

func TestMain(m *testing.M) {
	flag.Parse()

	wrapper := &testMainWrapper{m: m}

	if *checkLeaks {
		var goleakOptions []goleak.Option
		goleakOptions = append(goleakOptions, goleak.Cleanup(func(exitCode int) {
			if exitCode == 0 {
				fmt.Println("goleak: no goroutine leaks detected.")
			}
			os.Exit(exitCode)
		}))

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
