package modal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega"
)

func TestGetConfigPath_WithEnvVar(t *testing.T) {
	g := gomega.NewWithT(t)

	originalPath := os.Getenv("MODAL_CONFIG_PATH")
	defer func() {
		if originalPath != "" {
			err := os.Setenv("MODAL_CONFIG_PATH", originalPath)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
		} else {
			err := os.Unsetenv("MODAL_CONFIG_PATH")
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
		}
	}()

	customPath := "/custom/path/to/config.toml"
	err := os.Setenv("MODAL_CONFIG_PATH", customPath)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	path, err := configFilePath()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(path).Should(gomega.Equal(customPath))
}

func TestGetConfigPath_WithoutEnvVar(t *testing.T) {
	g := gomega.NewWithT(t)

	originalPath := os.Getenv("MODAL_CONFIG_PATH")
	defer func() {
		if originalPath != "" {
			err := os.Setenv("MODAL_CONFIG_PATH", originalPath)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
		} else {
			err := os.Unsetenv("MODAL_CONFIG_PATH")
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
		}
	}()

	err := os.Unsetenv("MODAL_CONFIG_PATH")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	path, err := configFilePath()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	home, _ := os.UserHomeDir()
	expectedPath := filepath.Join(home, ".modal.toml")
	g.Expect(path).Should(gomega.Equal(expectedPath))
}
