package test

import (
	"context"
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

	reader, err := sb.Open("/tmp/text.txt", "r")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	output := make([]byte, 12)
	n, err = reader.Read(output)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(n).Should(gomega.Equal(12))
}

func TestAppendToFile(t *testing.T) {}
func TestFlush(t *testing.T)        {}

// func _TestSandboxFilesystem(t *testing.T) {
// 	t.Parallel()
// 	g := gomega.NewWithT(t)
// 	ctx := context.Background()

// 	app, err := modal.AppLookup(ctx, "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
// 	g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 	image, err := app.ImageFromRegistry("alpine:3.21")
// 	g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{
// 		Command: []string{"sleep", "3600"}, // Keep sandbox alive
// 	})
// 	g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 	defer func() {
// 		if err := sb.Terminate(); err != nil {
// 			t.Logf("Failed to terminate sandbox: %v", err)
// 		}
// 	}()

// 	t.Run("WriteAndReadTextFile", func(t *testing.T) {
// 		g := gomega.NewWithT(t)

// 		// Write a file
// 		writeFile, err := sb.Open("/tmp/test.txt", "w")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		_, err = writeFile.WriteString("Hello, Modal filesystem!\n")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		_, err = writeFile.WriteString("This is a test file.\n")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		err = writeFile.Close()
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		// Read the file
// 		readFile, err := sb.Open("/tmp/test.txt", "r")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		content, err := readFile.ReadAll()
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		expected := "Hello, Modal filesystem!\nThis is a test file.\n"
// 		g.Expect(string(content)).To(gomega.Equal(expected))

// 		err = readFile.Close()
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())
// 	})

// 	t.Run("WriteAndReadBinaryFile", func(t *testing.T) {
// 		g := gomega.NewWithT(t)
// 		testData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

// 		// Write binary data
// 		writeFile, err := sb.Open("/tmp/test.bin", "w")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		_, err = writeFile.Write(testData)
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		err = writeFile.Close()
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		// Read binary data
// 		readFile, err := sb.Open("/tmp/test.bin", "r")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		readData, err := readFile.ReadAll()
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		g.Expect(len(readData)).To(gomega.Equal(len(testData)))

// 		for i, b := range readData {
// 			g.Expect(b).To(gomega.Equal(testData[i]))
// 		}

// 		err = readFile.Close()
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())
// 	})

// 	t.Run("AppendToFile", func(t *testing.T) {
// 		g := gomega.NewWithT(t)

// 		// Write initial content
// 		writeFile, err := sb.Open("/tmp/append.txt", "w")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		_, err = writeFile.WriteString("Initial content\n")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		err = writeFile.Close()
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		// Append more content
// 		appendFile, err := sb.Open("/tmp/append.txt", "a")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		_, err = appendFile.WriteString("Appended content\n")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		err = appendFile.Close()
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		// Read the entire file
// 		readFile, err := sb.Open("/tmp/append.txt", "r")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		content, err := readFile.ReadAll()
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		expected := "Initial content\nAppended content\n"
// 		g.Expect(string(content)).To(gomega.Equal(expected))

// 		err = readFile.Close()
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())
// 	})

// 	t.Run("SeekAndRead", func(t *testing.T) {
// 		g := gomega.NewWithT(t)

// 		// Write a file
// 		writeFile, err := sb.Open("/tmp/seek.txt", "w")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		_, err = writeFile.WriteString("Hello, world! This is a test.")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		err = writeFile.Close()
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		// Seek to position 7 and read
// 		readFile, err := sb.Open("/tmp/seek.txt", "r")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		_, err = readFile.Seek(7, 0) // Seek from beginning
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		buffer := make([]byte, 5)
// 		n, err := readFile.Read(buffer)
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		expected := "world"
// 		g.Expect(string(buffer[:n])).To(gomega.Equal(expected))

// 		err = readFile.Close()
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())
// 	})

// 	t.Run("FileFlush", func(t *testing.T) {
// 		g := gomega.NewWithT(t)

// 		file, err := sb.Open("/tmp/flush.txt", "w")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		_, err = file.WriteString("Test data")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		err = file.Flush() // Ensure data is written to disk
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		err = file.Close()
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		// Verify the data was written
// 		readFile, err := sb.Open("/tmp/flush.txt", "r")
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		content, err := readFile.ReadAll()
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())

// 		expected := "Test data"
// 		g.Expect(string(content)).To(gomega.Equal(expected))

// 		err = readFile.Close()
// 		g.Expect(err).ShouldNot(gomega.HaveOccurred())
// 	})
// }
