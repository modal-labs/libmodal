package test

import (
	"context"
	"testing"

	"github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
)

func TestVolumeFromName(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	volume, err := modal.VolumeFromName(context.Background(), "libmodal-test-volume", &modal.VolumeFromNameOptions{
		CreateIfMissing: true,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(volume).ShouldNot(gomega.BeNil())
	g.Expect(volume.VolumeId).Should(gomega.HavePrefix("vo-"))
	g.Expect(volume.Name).To(gomega.Equal("libmodal-test-volume"))

	// Test that missing volume returns an error
	_, err = modal.VolumeFromName(context.Background(), "missing-volume", nil)
	g.Expect(err).Should(gomega.MatchError(gomega.ContainSubstring("Volume 'missing-volume' not found")))
}

func TestVolumeReadOnly(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	volume, err := modal.VolumeFromName(context.Background(), "libmodal-test-volume", &modal.VolumeFromNameOptions{
		CreateIfMissing: true,
	})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(volume.IsReadOnly()).To(gomega.BeFalse())

	readOnlyVolume := volume.ReadOnly()
	g.Expect(readOnlyVolume.IsReadOnly()).To(gomega.BeTrue())
	g.Expect(readOnlyVolume.VolumeId).To(gomega.Equal(volume.VolumeId))
	g.Expect(readOnlyVolume.Name).To(gomega.Equal(volume.Name))

	g.Expect(volume.IsReadOnly()).To(gomega.BeFalse())
}
