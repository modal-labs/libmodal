package test

import (
	"context"
	"io"
	"testing"
	"time"

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
	g.Expect(exitcode).To(gomega.Equal(137))
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
	g.Expect(exitCode).To(gomega.Equal(0))
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
	g.Expect(exitCode).To(gomega.Equal(0))
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
	g.Expect(exitCode).Should(gomega.Equal(0))
}

func TestSandboxWithTunnels(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := modal.AppLookup(ctx, "libmodal-test", &modal.LookupOptions{
		CreateIfMissing: true,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromRegistry("alpine:3.21", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Create a sandbox with port forwarding
	sandbox, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command:          []string{"cat"},
		EncryptedPorts:   []int{8443},
		UnencryptedPorts: []int{8080},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(sandbox).ShouldNot(gomega.BeNil())
	g.Expect(sandbox.SandboxId).Should(gomega.HavePrefix("sb-"))

	defer sandbox.Terminate()

	tunnels, err := sandbox.Tunnels(30 * time.Second)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(tunnels).Should(gomega.HaveLen(2))

	// Test encrypted tunnel (port 8443)
	encryptedTunnel := tunnels[8443]
	g.Expect(encryptedTunnel.Host).Should(gomega.MatchRegexp(`\.modal\.host$`))
	g.Expect(encryptedTunnel.Port).Should(gomega.Equal(443))
	g.Expect(encryptedTunnel.URL()).Should(gomega.HavePrefix("https://"))

	host, port := encryptedTunnel.TLSSocket()
	g.Expect(host).Should(gomega.Equal(encryptedTunnel.Host))
	g.Expect(port).Should(gomega.Equal(encryptedTunnel.Port))

	// Test unencrypted tunnel (port 8080)
	unencryptedTunnel := tunnels[8080]
	g.Expect(unencryptedTunnel.UnencryptedHost).Should(gomega.MatchRegexp(`\.modal\.host$`))
	g.Expect(unencryptedTunnel.UnencryptedPort).Should(gomega.BeNumerically(">", 0))

	tcpHost, tcpPort, err := unencryptedTunnel.TCPSocket()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(tcpHost).Should(gomega.Equal(unencryptedTunnel.UnencryptedHost))
	g.Expect(tcpPort).Should(gomega.Equal(unencryptedTunnel.UnencryptedPort))
}

func TestSandboxPollAndReturnCode(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromRegistry("alpine:3.21", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	sandbox, err := app.CreateSandbox(image, &modal.SandboxOptions{Command: []string{"cat"}})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	pollResult, err := sandbox.Poll()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(pollResult).Should(gomega.BeNil())

	// Send input to make the cat command complete
	_, err = sandbox.Stdin.Write([]byte("hello, sandbox"))
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = sandbox.Stdin.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	waitResult, err := sandbox.Wait()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(waitResult).To(gomega.Equal(0))

	pollResult, err = sandbox.Poll()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(pollResult).ShouldNot(gomega.BeNil())
	g.Expect(*pollResult).To(gomega.Equal(0))
}

func TestSandboxPollAfterFailure(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromRegistry("alpine:3.21", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	sandbox, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Command: []string{"sh", "-c", "exit 42"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	waitResult, err := sandbox.Wait()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(waitResult).To(gomega.Equal(42))

	pollResult, err := sandbox.Poll()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(pollResult).ShouldNot(gomega.BeNil())
	g.Expect(*pollResult).To(gomega.Equal(42))
}
