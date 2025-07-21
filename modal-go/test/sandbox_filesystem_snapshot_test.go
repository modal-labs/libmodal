package test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
)

func TestSnapshotFilesystem(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromRegistry("alpine:3.21", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	sb, err := app.CreateSandbox(image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer sb.Terminate()

	_, err = sb.Exec([]string{"sh", "-c", "echo -n 'test content' > /tmp/test.txt"}, modal.ExecOptions{})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	_, err = sb.Exec([]string{"mkdir", "-p", "/tmp/testdir"}, modal.ExecOptions{})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	snapshotImage, err := sb.SnapshotFilesystem(55 * time.Second)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(snapshotImage).ShouldNot(gomega.BeNil())
	g.Expect(snapshotImage.ImageId).To(gomega.HavePrefix("im-"))

	err = sb.Terminate()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Create new sandbox from snapshot
	sb2, err := app.CreateSandbox(snapshotImage, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer sb2.Terminate()

	// Verify file exists in snapshot
	proc, err := sb2.Exec([]string{"cat", "/tmp/test.txt"}, modal.ExecOptions{})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	output, err := io.ReadAll(proc.Stdout)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(string(output)).To(gomega.Equal("test content"))

	// Verify directory exists in snapshot
	dirCheck, err := sb2.Exec([]string{"test", "-d", "/tmp/testdir"}, modal.ExecOptions{})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	exitCode, err := dirCheck.Wait()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(exitCode).To(gomega.Equal(0))
}
