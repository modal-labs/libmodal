package test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
)

func createSandbox(g *gomega.WithT) *modal.Sandbox {
	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromRegistry("alpine:3.21")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	sb, err := app.CreateSandbox(image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(sb.SandboxId).ShouldNot(gomega.BeEmpty())
	return sb
}

func terminateSandbox(g *gomega.WithT, sb *modal.Sandbox) {
	err := sb.Terminate()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestSandboxWriteAndReadBinaryFile(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	sb := createSandbox(g)
	defer terminateSandbox(g, sb)

	writer, err := sb.Open("/tmp/test.bin", "w")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	text := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n, err := writer.Write(text)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(n).Should(gomega.Equal(len(text)))
	err = writer.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	reader, err := sb.Open("/tmp/test.bin", "r")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	output := make([]byte, 10)
	n, err = reader.Read(output)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(n).Should(gomega.Equal(10))
	g.Expect(output).Should(gomega.Equal(text))

	err = reader.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestSandboxAppendToFileBinary(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	sb := createSandbox(g)
	defer terminateSandbox(g, sb)

	writer, err := sb.Open("/tmp/append.txt", "w")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	text := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n, err := writer.Write(text)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(n).Should(gomega.Equal(len(text)))
	err = writer.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	appender, err := sb.Open("/tmp/append.txt", "a")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	more_text := []byte{7, 8, 9, 10}
	appender.Write(more_text)

	reader, err := sb.Open("/tmp/append.txt", "r")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	expected_text := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 7, 8, 9, 10}
	out, err := io.ReadAll(reader)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(out).Should(gomega.Equal(expected_text))

	err = reader.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}
func TestSandboxFileFlush(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	sb := createSandbox(g)
	defer terminateSandbox(g, sb)

	writer, err := sb.Open("/tmp/flush.txt", "w")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	text := []byte("Test data")
	n, err := writer.Write(text)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(n).Should(gomega.Equal(len(text)))
	err = writer.Flush()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	reader, err := sb.Open("/tmp/flush.txt", "r")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	out, err := io.ReadAll(reader)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(out).Should(gomega.Equal(text))

	err = writer.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	err = reader.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestSandboxMultipleFileOperations(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	sb := createSandbox(g)
	defer terminateSandbox(g, sb)

	// Create multiple files
	content1 := []byte("File 1 content")
	writer, err := sb.Open("/tmp/file1.txt", "w")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	n, err := writer.Write(content1)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(n).Should(gomega.Equal(len(content1)))
	err = writer.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	content2 := []byte("File 2 content")
	writer, err = sb.Open("/tmp/file2.txt", "w")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	n, err = writer.Write(content2)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(n).Should(gomega.Equal(len(content2)))
	err = writer.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Read both files
	reader1, err := sb.Open("/tmp/file1.txt", "r")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	readContent1, err := io.ReadAll(reader1)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = reader1.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	reader2, err := sb.Open("/tmp/file2.txt", "r")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	readContent2, err := io.ReadAll(reader2)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = reader2.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(readContent1).Should(gomega.Equal(content1))
	g.Expect(readContent2).Should(gomega.Equal(content2))

}

func TestSandboxFileOpenModes(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	sb := createSandbox(g)
	defer terminateSandbox(g, sb)

	// Test write mode (truncates)
	content1 := []byte("Initial content")
	writer, err := sb.Open("/tmp/modes.txt", "w")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	n, err := writer.Write(content1)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(n).Should(gomega.Equal(len(content1)))
	err = writer.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Test read mode
	reader1, err := sb.Open("/tmp/modes.txt", "r")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	readContent1, err := io.ReadAll(reader1)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(readContent1).Should(gomega.Equal(content1))
	err = reader1.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Test append mode
	content2 := []byte(" appended")
	appender, err := sb.Open("/tmp/modes.txt", "a")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	n, err = appender.Write(content2)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(n).Should(gomega.Equal(len(content2)))
	err = appender.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Verify append worked
	reader2, err := sb.Open("/tmp/modes.txt", "r")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	readContent2, err := io.ReadAll(reader2)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	expectedContent := []byte("Initial content appended")
	g.Expect(readContent2).Should(gomega.Equal(expectedContent))
	err = reader2.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestSandboxLargeFileOperations(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	sb := createSandbox(g)
	defer terminateSandbox(g, sb)

	xByte := []byte{'x'}
	largeData := bytes.Repeat(xByte, 1000)

	writer, err := sb.Open("/tmp/large.txt", "w")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	n, err := writer.Write(largeData)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(n).Should(gomega.Equal(len(largeData)))
	err = writer.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	reader1, err := sb.Open("/tmp/large.txt", "r")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	readContent1, err := io.ReadAll(reader1)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(readContent1).Should(gomega.Equal(largeData))
	err = reader1.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}
