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
	ctx := context.Background()

	app, err := tc.Apps.FromName(ctx, "libmodal-test", &modal.AppFromNameOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	image := tc.Images.FromRegistry("alpine:3.21", nil)

	proxy, err := tc.Proxies.FromName(ctx, "libmodal-test-proxy", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(proxy.ProxyId).ShouldNot(gomega.BeEmpty())
	g.Expect(strings.HasPrefix(proxy.ProxyId, "pr-")).To(gomega.BeTrue())

	sb, err := tc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateOptions{
		Proxy:   proxy,
		Command: []string{"echo", "hello, Sandbox with Proxy"},
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(sb.SandboxId).ShouldNot(gomega.BeEmpty())

	err = sb.Terminate(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	exitcode, err := sb.Wait(ctx)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(exitcode).To(gomega.Equal(137))
}

func TestProxyNotFound(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	_, err := tc.Proxies.FromName(ctx, "non-existent-proxy-name", nil)
	g.Expect(err).Should(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("Proxy 'non-existent-proxy-name' not found"))
}
