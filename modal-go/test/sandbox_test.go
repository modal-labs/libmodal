package test

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
)

func TestCreateOneSandbox(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(app.Name).To(gomega.Equal("libmodal-test"))

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sb, err := tc.Sandboxes.Create(ctx, app, image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(sb.SandboxId).ShouldNot(gomega.BeEmpty())
	defer terminateSandbox(g, sb)

	err = sb.Terminate(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	exitcode, err := sb.Wait(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(exitcode).To(gomega.Equal(137))
}

func TestPassCatToStdin(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sb, err := tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{Command: []string{"cat"}})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer terminateSandbox(g, sb)

	_, err = sb.Stdin.Write([]byte("this is input that should be mirrored by cat"))
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = sb.Stdin.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	output, err := io.ReadAll(sb.Stdout)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(string(output)).To(gomega.Equal("this is input that should be mirrored by cat"))
}

func TestIgnoreLargeStdout(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("python:3.13-alpine", nil)

	sb, err := tc.Sandboxes.Create(ctx, app, image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer terminateSandbox(g, sb)

	p, err := sb.Exec(ctx, []string{"python", "-c", `print("a" * 1_000_000)`}, &modal.SandboxExecParams{Stdout: modal.Ignore})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	buf, err := io.ReadAll(p.Stdout)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(len(buf)).To(gomega.Equal(0)) // Stdout is ignored

	// Stdout should be consumed after cancel, without blocking the process.
	exitCode, err := p.Wait(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(exitCode).To(gomega.Equal(0))
}

func TestSandboxCreateOptions(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{
		CreateIfMissing: true,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sb, err := tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Command: []string{"echo", "hello, params"},
		Cloud:   "aws",
		Regions: []string{"us-east-1", "us-west-2"},
		Verbose: true,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(sb).ShouldNot(gomega.BeNil())
	g.Expect(sb.SandboxId).Should(gomega.HavePrefix("sb-"))

	defer terminateSandbox(g, sb)

	exitCode, err := sb.Wait(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(exitCode).Should(gomega.Equal(0))

	_, err = tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Cloud: "invalid-cloud",
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("InvalidArgument"))

	_, err = tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Regions: []string{"invalid-region"},
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("InvalidArgument"))
}

func TestSandboxExecOptions(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sb, err := tc.Sandboxes.Create(ctx, app, image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer terminateSandbox(g, sb)

	// Test with a custom working directory and timeout.
	p, err := sb.Exec(ctx, []string{"pwd"}, &modal.SandboxExecParams{
		Workdir: "/tmp",
		Timeout: 5,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	output, err := io.ReadAll(p.Stdout)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(string(output)).To(gomega.Equal("/tmp\n"))

	exitCode, err := p.Wait(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(exitCode).To(gomega.Equal(0))
}

func TestSandboxWithVolume(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{
		CreateIfMissing: true,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	volume, err := tc.Volumes.FromName(ctx, "libmodal-test-sandbox-volume", &modal.VolumeFromNameParams{
		CreateIfMissing: true,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	sandbox, err := tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Command: []string{"echo", "volume test"},
		Volumes: map[string]*modal.Volume{
			"/mnt/test": volume,
		},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(sandbox).ShouldNot(gomega.BeNil())
	g.Expect(sandbox.SandboxId).Should(gomega.HavePrefix("sb-"))

	exitCode, err := sandbox.Wait(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(exitCode).Should(gomega.Equal(0))
}

func TestSandboxWithReadOnlyVolume(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{
		CreateIfMissing: true,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	volume, err := tc.Volumes.FromName(ctx, "libmodal-test-sandbox-volume", &modal.VolumeFromNameParams{
		CreateIfMissing: true,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	readOnlyVolume := volume.ReadOnly()
	g.Expect(readOnlyVolume.IsReadOnly()).To(gomega.BeTrue())

	sb, err := tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Command: []string{"sh", "-c", "echo 'test' > /mnt/test/test.txt"},
		Volumes: map[string]*modal.Volume{
			"/mnt/test": readOnlyVolume,
		},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer terminateSandbox(g, sb)

	exitCode, err := sb.Wait(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(exitCode).Should(gomega.Equal(1))

	stderr, err := io.ReadAll(sb.Stderr)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(string(stderr)).Should(gomega.ContainSubstring("Read-only file system"))
}

func TestSandboxWithTunnels(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{
		CreateIfMissing: true,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sandbox, err := tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Command:          []string{"cat"},
		EncryptedPorts:   []int{8443},
		UnencryptedPorts: []int{8080},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer terminateSandbox(g, sandbox)

	g.Expect(sandbox.SandboxId).Should(gomega.HavePrefix("sb-"))

	tunnels, err := sandbox.Tunnels(ctx, 30*time.Second)
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

func TestCreateSandboxWithSecrets(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	secret, err := tc.Secrets.FromName(ctx, "libmodal-test-secret", &modal.SecretFromNameParams{RequiredKeys: []string{"c"}})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sb, err := tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{Secrets: []*modal.Secret{secret}, Command: []string{"printenv", "c"}})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	output, err := io.ReadAll(sb.Stdout)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(string(output)).To(gomega.Equal("hello world\n"))
}
func TestSandboxPollAndReturnCode(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sandbox, err := tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{Command: []string{"cat"}})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	pollResult, err := sandbox.Poll(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(pollResult).Should(gomega.BeNil())

	// Send input to make the cat command complete
	_, err = sandbox.Stdin.Write([]byte("hello, sandbox"))
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = sandbox.Stdin.Close()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	waitResult, err := sandbox.Wait(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(waitResult).To(gomega.Equal(0))

	pollResult, err = sandbox.Poll(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(pollResult).ShouldNot(gomega.BeNil())
	g.Expect(*pollResult).To(gomega.Equal(0))
}

func TestSandboxPollAfterFailure(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sandbox, err := tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Command: []string{"sh", "-c", "exit 42"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	waitResult, err := sandbox.Wait(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(waitResult).To(gomega.Equal(42))

	pollResult, err := sandbox.Poll(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(pollResult).ShouldNot(gomega.BeNil())
	g.Expect(*pollResult).To(gomega.Equal(42))
}

func TestCreateSandboxWithNetworkAccessParams(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{
		CreateIfMissing: true,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sb, err := tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Command:       []string{"echo", "hello, network access"},
		BlockNetwork:  false,
		CIDRAllowlist: []string{"10.0.0.0/8", "192.168.0.0/16"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer terminateSandbox(g, sb)

	g.Expect(sb).ShouldNot(gomega.BeNil())
	g.Expect(sb.SandboxId).Should(gomega.HavePrefix("sb-"))

	exitCode, err := sb.Wait(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(exitCode).Should(gomega.Equal(0))

	_, err = tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		BlockNetwork:  false,
		CIDRAllowlist: []string{"not-an-ip/8"},
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("Invalid CIDR: not-an-ip/8"))

	_, err = tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		BlockNetwork:  true,
		CIDRAllowlist: []string{"10.0.0.0/8"},
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).Should(gomega.ContainSubstring("CIDRAllowlist cannot be used when BlockNetwork is enabled"))
}

func TestSandboxExecSecret(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sb, err := tc.Sandboxes.Create(ctx, app, image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer terminateSandbox(g, sb)

	secret, err := tc.Secrets.FromName(ctx, "libmodal-test-secret", &modal.SecretFromNameParams{RequiredKeys: []string{"c"}})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	secret2, err := tc.Secrets.FromMap(ctx, map[string]string{"d": "3"}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	p, err := sb.Exec(ctx, []string{"printenv", "c", "d"}, &modal.SandboxExecParams{Secrets: []*modal.Secret{secret, secret2}})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	buf, err := io.ReadAll(p.Stdout)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(string(buf)).Should(gomega.Equal("hello world\n3\n"))
}

func TestSandboxFromId(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sb, err := tc.Sandboxes.Create(ctx, app, image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer terminateSandbox(g, sb)

	g.Expect(sb.SandboxId).ShouldNot(gomega.BeEmpty())

	sbFromId, err := tc.Sandboxes.FromId(ctx, sb.SandboxId)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(sbFromId.SandboxId).Should(gomega.Equal(sb.SandboxId))
}

func TestSandboxWithWorkdir(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sb, err := tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Command: []string{"pwd"},
		Workdir: "/tmp",
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer terminateSandbox(g, sb)

	output, err := io.ReadAll(sb.Stdout)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(string(output)).To(gomega.Equal("/tmp\n"))

	exitCode, err := sb.Wait(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(exitCode).To(gomega.Equal(0))

	_, err = tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Workdir: "relative/path",
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("the Workdir value must be an absolute path"))
}

func TestSandboxSetTagsAndList(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sb, err := tc.Sandboxes.Create(ctx, app, image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer terminateSandbox(g, sb)

	unique := fmt.Sprintf("%d", rand.Int())

	var before []string
	it, err := tc.Sandboxes.List(ctx, &modal.SandboxListParams{Tags: map[string]string{"test-key": unique}})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	for s, err := range it {
		g.Expect(err).ShouldNot(gomega.HaveOccurred())
		before = append(before, s.SandboxId)
	}
	g.Expect(before).To(gomega.HaveLen(0))

	err = sb.SetTags(ctx, map[string]string{"test-key": unique})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	var after []string
	it, err = tc.Sandboxes.List(ctx, &modal.SandboxListParams{Tags: map[string]string{"test-key": unique}})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	for s, err := range it {
		g.Expect(err).ShouldNot(gomega.HaveOccurred())
		after = append(after, s.SandboxId)
	}
	g.Expect(after).To(gomega.Equal([]string{sb.SandboxId}))
}

func TestSandboxTags(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sb, err := tc.Sandboxes.Create(ctx, app, image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer terminateSandbox(g, sb)

	retrievedTagsBefore, err := sb.GetTags(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(retrievedTagsBefore).To(gomega.Equal(map[string]string{}))

	tagA := fmt.Sprintf("%d", rand.Int())
	tagB := fmt.Sprintf("%d", rand.Int())
	tagC := fmt.Sprintf("%d", rand.Int())

	err = sb.SetTags(ctx, map[string]string{"key-a": tagA, "key-b": tagB, "key-c": tagC})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	retrievedTags, err := sb.GetTags(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(retrievedTags).To(gomega.Equal(map[string]string{"key-a": tagA, "key-b": tagB, "key-c": tagC}))

	var ids []string
	it, err := tc.Sandboxes.List(ctx, &modal.SandboxListParams{Tags: map[string]string{"key-a": tagA}})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	for s, err := range it {
		g.Expect(err).ShouldNot(gomega.HaveOccurred())
		ids = append(ids, s.SandboxId)
	}
	g.Expect(ids).To(gomega.Equal([]string{sb.SandboxId}))

	ids = nil
	it, err = tc.Sandboxes.List(ctx, &modal.SandboxListParams{Tags: map[string]string{"key-a": tagA, "key-b": tagB}})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	for s, err := range it {
		g.Expect(err).ShouldNot(gomega.HaveOccurred())
		ids = append(ids, s.SandboxId)
	}
	g.Expect(ids).To(gomega.Equal([]string{sb.SandboxId}))

	ids = nil
	it, err = tc.Sandboxes.List(ctx, &modal.SandboxListParams{Tags: map[string]string{"key-a": tagA, "key-b": tagB, "key-d": "not-set"}})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	for s, err := range it {
		g.Expect(err).ShouldNot(gomega.HaveOccurred())
		ids = append(ids, s.SandboxId)
	}
	g.Expect(ids).To(gomega.HaveLen(0))
}

func TestSandboxListByAppId(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sb, err := tc.Sandboxes.Create(ctx, app, image, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer terminateSandbox(g, sb)

	count := 0
	it, err := tc.Sandboxes.List(ctx, &modal.SandboxListParams{AppId: app.AppId})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	for s, err := range it {
		g.Expect(err).ShouldNot(gomega.HaveOccurred())
		g.Expect(s.SandboxId).Should(gomega.HavePrefix("sb-"))
		count++
		if count >= 1 {
			break
		}
	}
	g.Expect(count).ToNot(gomega.Equal(0))
}

func TestNamedSandbox(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameParams{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	sandboxName := fmt.Sprintf("test-sandbox-%d", rand.Int())

	sb, err := tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Name:    sandboxName,
		Command: []string{"sleep", "60"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(sb.SandboxId).ShouldNot(gomega.BeEmpty())

	defer terminateSandbox(g, sb)

	sb1FromName, err := tc.Sandboxes.FromName(ctx, "libmodal-test", sandboxName, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(sb1FromName.SandboxId).To(gomega.Equal(sb.SandboxId))

	sb2FromName, err := tc.Sandboxes.FromName(ctx, "libmodal-test", sandboxName, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(sb2FromName.SandboxId).To(gomega.Equal(sb1FromName.SandboxId))

	_, err = tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Name:    sandboxName,
		Command: []string{"sleep", "60"},
	})
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("already exists"))
}

func TestNamedSandboxNotFound(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	_, err := tc.Sandboxes.FromName(ctx, "libmodal-test", "non-existent-sandbox", nil)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("not found"))
}
