package test

import (
	"context"
	"io"
	"testing"

	"github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
)

func TestCreateOneSandbox(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromRegistry("alpine:3.21", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	sb, err := app.CreateSandbox(image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(sb.SandboxId).ShouldNot(gomega.BeEmpty())

	err = sb.Terminate()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	exitcode, err := sb.Wait()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(exitcode).To(gomega.Equal(int32(0)))
}

func TestPassCatToStdin(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromRegistry("alpine:3.21", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Spawn a sandbox running the "cat" command.
	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{Command: []string{"cat"}})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Write to the sandbox's stdin and read from its stdout.
	_, err = sb.Stdin.Write([]byte("this is input that should be mirrored by cat"))
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = sb.Stdin.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	output, err := io.ReadAll(sb.Stdout)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(string(output)).To(gomega.Equal("this is input that should be mirrored by cat"))

	// Terminate the sandbox.
	err = sb.Terminate()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestIgnoreLargeStdout(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromRegistry("python:3.13-alpine", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	sb, err := app.CreateSandbox(image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer sb.Terminate()

	p, err := sb.Exec([]string{"python", "-c", `print("a" * 1_000_000)`}, modal.ExecOptions{Stdout: modal.Ignore})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	buf, err := io.ReadAll(p.Stdout)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(len(buf)).To(gomega.Equal(0)) // Stdout is ignored

	// Stdout should be consumed after cancel, without blocking the process.
	exitCode, err := p.Wait()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(exitCode).To(gomega.Equal(int32(0)))
}

func TestSandboxExecOptions(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromRegistry("alpine:3.21", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	sb, err := app.CreateSandbox(image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer sb.Terminate()

	// Test with a custom working directory and timeout.
	p, err := sb.Exec([]string{"pwd"}, modal.ExecOptions{
		Workdir: "/tmp",
		Timeout: 5,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	output, err := io.ReadAll(p.Stdout)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(string(output)).To(gomega.Equal("/tmp\n"))

	exitCode, err := p.Wait()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(exitCode).To(gomega.Equal(int32(0)))
}

func TestSandboxWithVolume(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := modal.AppLookup(ctx, "libmodal-test", &modal.LookupOptions{
		CreateIfMissing: true,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromRegistry("alpine:3.21", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	volume, err := modal.VolumeFromName(ctx, "libmodal-test-sandbox-volume", &modal.VolumeFromNameOptions{
		CreateIfMissing: true,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	sandbox, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command: []string{"echo", "volume test"},
		Volumes: map[string]*modal.Volume{
			"/mnt/test": volume,
		},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(sandbox).ShouldNot(gomega.BeNil())
	g.Expect(sandbox.SandboxId).Should(gomega.HavePrefix("sb-"))

	exitCode, err := sandbox.Wait()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(exitCode).Should(gomega.Equal(int32(0)))
}
