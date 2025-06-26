package test

import (
	"context"
	"io"
	"testing"

	"github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
)

func create_sandbox(g *gomega.WithT) *modal.Sandbox {
	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromRegistry("alpine:3.21")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	sb, err := app.CreateSandbox(image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(sb.SandboxId).ShouldNot(gomega.BeEmpty())
	return sb
}

func terminate_sandbox(g *gomega.WithT, sb *modal.Sandbox) {
	err := sb.Terminate()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestWriteAndReadTextFile(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	sb := create_sandbox(g)
	defer terminate_sandbox(g, sb)

	writer, err := sb.Open("/tmp/test.txt", "w")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	text := []byte("Hello, Modal filesystem!")
	n, err := writer.Write(text)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(n).Should(gomega.Equal(len(text)))
	err = writer.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	reader, err := sb.Open("/tmp/test.txt", "r")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	output := make([]byte, 12)
	n, err = reader.Read(output)
	g.Expect(err).Should(gomega.Equal(io.EOF))
	g.Expect(n).Should(gomega.Equal(12))
	g.Expect(output).Should(gomega.Equal(text[0:12]))

	err = reader.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestAppendToFile(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	sb := create_sandbox(g)
	defer terminate_sandbox(g, sb)

	writer, err := sb.Open("/tmp/test.txt", "w")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	text := []byte("Hello, Modal filesystem!")
	n, err := writer.Write(text)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(n).Should(gomega.Equal(len(text)))
	err = writer.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	appender, err := sb.Open("/tmp/test.txt", "a")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	more_text := []byte("This is more text")
	appender.Write(more_text)

	reader, err := sb.Open("/tmp/test.txt", "r")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	expected_text := append(text, more_text...)
	out, err := io.ReadAll(reader)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(out).Should(gomega.Equal(expected_text))

	err = reader.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}
func TestFlush(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	sb := create_sandbox(g)
	defer terminate_sandbox(g, sb)

	writer, err := sb.Open("/tmp/test.txt", "w")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	text := []byte("hello world")
	n, err := writer.Write(text)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(n).Should(gomega.Equal(len(text)))
	err = writer.Flush()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	reader, err := sb.Open("/tmp/test.txt", "r")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	out, err := io.ReadAll(reader)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(out).Should(gomega.Equal(text))

}
