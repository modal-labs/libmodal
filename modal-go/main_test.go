package modal

import (
	"flag"
	"os"
	"testing"
)

// The checkLeaks and checkLeaksPerTest are defined and used in the test package in ./test.
// However when running tests across both packages (go test . ./test -check-leaks),
// Go passes it to both packages, and we get a "flag provided but not defined" warning
// if it's not also defined here.

// checkLeaks is ignored in this package, but defined in the test package.
var checkLeaks = flag.Bool("check-leaks", false, "ignored in this package, but defined in the test package")

// checkLeaksPerTest is ignored in this package, but defined in the test package.
var checkLeaksPerTest = flag.Bool("check-leaks-per-test", false, "ignored in this package, but defined in the test package")

func TestMain(m *testing.M) {
	flag.Parse()

	if *checkLeaks && *checkLeaksPerTest {
		panic("Error: -check-leaks and -check-leaks-per-test are mutually exclusive")
	}

	code := m.Run()
	os.Exit(code)
}
