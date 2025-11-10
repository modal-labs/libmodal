package modal

import (
	"flag"
	"os"
	"testing"
)

// The checkLeaks flag is defined and used in the test package in ./test.
// However when running tests across both packages (go test . ./test -check-leaks),
// Go passes it to both packages, and we get a "flag provided but not defined" warning
// if it's not also defined here.
var checkLeaks = flag.Bool("check-leaks", false, "ignored in this package, but defined in the test package") //nolint:unused

func TestMain(m *testing.M) {
	flag.Parse()

	code := m.Run()
	os.Exit(code)
}
