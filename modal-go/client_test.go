package modal

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestGetLibModalVersionDev(t *testing.T) {
	// Version tracking is properly tested in test/integration/version-detection/
	g := gomega.NewWithT(t)

	g.Expect(getLibModalVersion()).To(gomega.Equal("modal-go/dev"))
}
