package test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
)

func TestQueueInvalidName(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	for _, name := range []string{"has space", "has/slash", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"} {
		_, err := modal.QueueLookup(context.Background(), name, nil)
		g.Expect(err).Should(gomega.HaveOccurred(), "Queue lookup should fail for invalid name: %s", name)
	}
}

func TestQueueEphemeral(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	queue, err := modal.QueueEphemeral(context.Background(), nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer queue.CloseEphemeral()

	err = queue.Put(123, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	len, err := queue.Len(nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(len).To(gomega.Equal(1))

	result, err := queue.Get(nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).To(gomega.Equal(int64(123)))
}

func TestQueueSuite1(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	queue, err := modal.QueueEphemeral(ctx, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer queue.CloseEphemeral()

	// queue.len() == 0
	n, err := queue.Len(nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(n).To(gomega.Equal(0))

	// put / len / get
	g.Expect(queue.Put(123, nil)).ToNot(gomega.HaveOccurred())

	n, err = queue.Len(nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(n).To(gomega.Equal(1))

	item, err := queue.Get(nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(item).To(gomega.Equal(int64(123)))

	// put, then non-blocking get
	g.Expect(queue.Put(432, nil)).ToNot(gomega.HaveOccurred())

	var timeout time.Duration
	item, err = queue.Get(&modal.QueueGetOptions{Timeout: &timeout})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(item).To(gomega.Equal(int64(432)))

	// queue is now empty – non-blocking get should error
	_, err = queue.Get(&modal.QueueGetOptions{Timeout: &timeout})
	g.Expect(errors.As(err, &modal.QueueEmptyError{})).To(gomega.BeTrue())

	n, err = queue.Len(nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(n).To(gomega.Equal(0))

	// putMany + iterator
	g.Expect(queue.PutMany([]any{1, 2, 3}, nil)).ToNot(gomega.HaveOccurred())

	results := make([]int64, 0, 3)
	for v, err := range queue.Iterate(nil) {
		g.Expect(err).ToNot(gomega.HaveOccurred())
		results = append(results, v.(int64))
	}
	g.Expect(results).To(gomega.Equal([]int64{1, 2, 3}))
}

func TestQueueSuite2(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	queue, err := modal.QueueEphemeral(ctx, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer queue.CloseEphemeral()

	var wg sync.WaitGroup
	results := make([]int64, 0, 10)

	// producer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range 10 {
			_ = queue.Put(i, nil) // ignore error for brevity
		}
	}()

	// consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for v, err := range queue.Iterate(&modal.QueueIterateOptions{ItemPollTimeout: time.Second}) {
			g.Expect(err).ToNot(gomega.HaveOccurred())
			results = append(results, v.(int64))
		}
	}()

	wg.Wait()
	g.Expect(results).To(gomega.Equal([]int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}))
}

func TestQueuePutAndGetMany(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	queue, err := modal.QueueEphemeral(ctx, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer queue.CloseEphemeral()

	g.Expect(queue.PutMany([]any{1, 2, 3}, nil)).ToNot(gomega.HaveOccurred())

	n, err := queue.Len(nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(n).To(gomega.Equal(3))

	items, err := queue.GetMany(3, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(items).To(gomega.Equal([]any{int64(1), int64(2), int64(3)}))
}

func TestQueueNonBlocking(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	queue, err := modal.QueueEphemeral(ctx, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer queue.CloseEphemeral()

	var timeout time.Duration
	err = queue.Put(123, &modal.QueuePutOptions{Timeout: &timeout})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	n, err := queue.Len(nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(n).To(gomega.Equal(1))

	item, err := queue.Get(&modal.QueueGetOptions{Timeout: &timeout})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(item).To(gomega.Equal(int64(123)))
}
