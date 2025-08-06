package test

import (
	"context"
	"strings"
	"testing"

	"github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
)

func TestCreateSandboxWithProxy(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	app, err := modal.AppLookup(context.Background(), "libmodal-test", &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image, err := app.ImageFromRegistry("alpine:3.21", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	proxy, err := modal.ProxyFromName(context.Background(), "libmodal-test-proxy", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(proxy.ProxyId).ShouldNot(gomega.BeEmpty())
	g.Expect(strings.HasPrefix(proxy.ProxyId, "pr-")).To(gomega.BeTrue())

	sb, err := app.CreateSandbox(image, &modal.SandboxOptions{
		Proxy:   proxy,
		Command: []string{"echo", "hello, sandbox with proxy"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(sb.SandboxId).ShouldNot(gomega.BeEmpty())

	err = sb.Terminate()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	exitcode, err := sb.Wait()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(exitcode).To(gomega.Equal(137))
}

func TestProxyNotFound(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	_, err := modal.ProxyFromName(context.Background(), "non-existent-proxy-name", nil)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("Proxy 'non-existent-proxy-name' not found"))
}
