package modal

import (
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/onsi/gomega"
)

func TestClientWithLogger(t *testing.T) {
	g := gomega.NewWithT(t)

	// Use a buffer to capture log output
	r, w, err := os.Pipe()
	g.Expect(err).To(gomega.BeNil())

	logger := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	g.Expect(logger).NotTo(gomega.BeNil())

	client, err := NewClientWithOptions(&ClientParams{Logger: logger})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(client).NotTo(gomega.BeNil())

	w.Close()

	output, err := io.ReadAll(r)
	g.Expect(err).To(gomega.BeNil())

	g.Expect(output).To(gomega.ContainSubstring("Initializing Modal client"))
	g.Expect(output).To(gomega.ContainSubstring("Modal client initialized successfully"))
}
