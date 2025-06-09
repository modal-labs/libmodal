package modal

// Queue object, to be used with Modal Queues.

import (
	"context"
	"fmt"
	"iter"
	"time"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// From: modal/_object.py
const ephemeralObjectHeartbeatSleep = 300 * time.Second

const queueInitialPutBackoff = 100 * time.Millisecond
const queueDefaultPartitionTtl = 24 * time.Hour

func validatePartitionKey(partition string) ([]byte, error) {
	if partition == "" {
		return nil, nil // default partition
	}
	b := []byte(partition)
	if len(b) == 0 || len(b) > 64 {
		return nil, InvalidError{"queue partition key must be 1–64 bytes long"}
	}
	return b, nil
}

type QueueClearOptions struct {
	Partition string // partition to clear (default "")
	All       bool   // clear *all* partitions (mutually exclusive with Partition)
}

type QueueGetOptions struct {
	Timeout   *time.Duration // wait max (nil = indefinitely)
	Partition string
}

type QueuePutOptions struct {
	Timeout      *time.Duration // max wait for space (nil = indefinitely)
	Partition    string
	PartitionTtl time.Duration // ttl for the *partition* (default 24h)
}

type QueueLenOptions struct {
	Partition string
	Total     bool // total across all partitions (mutually exclusive with Partition)
}

type QueueIterateOptions struct {
	ItemPollTimeout time.Duration // exit if no new items within this period
	Partition       string
}

// Queue is a distributed, FIFO queue for data flow in Modal apps.
type Queue struct {
	QueueId   string
	cancel    context.CancelFunc // only for ephemeral queues
	ephemeral bool
	ctx       context.Context
}

// QueueEphemeral creates a nameless, temporary queue. Caller must CloseEphemeral.
func QueueEphemeral(ctx context.Context, options *EphemeralOptions) (*Queue, error) {
	if options == nil {
		options = &EphemeralOptions{}
	}
	ctx = clientContext(ctx)

	resp, err := client.QueueGetOrCreate(ctx, pb.QueueGetOrCreateRequest_builder{
		ObjectCreationType: pb.ObjectCreationType_OBJECT_CREATION_TYPE_EPHEMERAL,
		EnvironmentName:    environmentName(options.Environment),
	}.Build())
	if err != nil {
		return nil, err
	}

	heartbeatCtx, cancel := context.WithCancel(ctx)
	q := &Queue{QueueId: resp.GetQueueId(), cancel: cancel, ephemeral: true, ctx: ctx}

	// backgroundheart‑beat goroutine
	go func() {
		t := time.NewTicker(ephemeralObjectHeartbeatSleep)
		defer t.Stop()
		for {
			select {
			case <-heartbeatCtx.Done():
				return
			case <-t.C:
				_, _ = client.QueueHeartbeat(heartbeatCtx, pb.QueueHeartbeatRequest_builder{
					QueueId: q.QueueId,
				}.Build()) // ignore errors – next call will retry or context will cancel
			}
		}
	}()

	return q, nil
}

// CloseEphemeral deletes an ephemeral queue, only used with QueueEphemeral.
func (q *Queue) CloseEphemeral() {
	if q.ephemeral {
		q.cancel() // will stop heartbeat
	} else {
		// We panic in this case because of invalid usage. In general, methods
		// used with `defer` like CloseEphemeral should not return errors.
		panic(fmt.Sprintf("queue %s is not ephemeral", q.QueueId))
	}
}

// QueueLookup returns a handle to a (possibly new) queue by deployment name.
func QueueLookup(ctx context.Context, name string, options *LookupOptions) (*Queue, error) {
	if options == nil {
		options = &LookupOptions{}
	}
	ctx = clientContext(ctx)

	creationType := pb.ObjectCreationType_OBJECT_CREATION_TYPE_UNSPECIFIED
	if options.CreateIfMissing {
		creationType = pb.ObjectCreationType_OBJECT_CREATION_TYPE_CREATE_IF_MISSING
	}

	resp, err := client.QueueGetOrCreate(ctx, pb.QueueGetOrCreateRequest_builder{
		DeploymentName:     name,
		Namespace:          pb.DeploymentNamespace_DEPLOYMENT_NAMESPACE_WORKSPACE,
		EnvironmentName:    environmentName(options.Environment),
		ObjectCreationType: creationType,
	}.Build())
	if err != nil {
		return nil, err
	}
	return &Queue{ctx: ctx, QueueId: resp.GetQueueId()}, nil
}

// QueueDelete removes a queue by name.
func QueueDelete(ctx context.Context, name string, options *DeleteOptions) error {
	q, err := QueueLookup(ctx, name, &LookupOptions{Environment: options.Environment})
	if err != nil {
		return err
	}
	_, err = client.QueueDelete(ctx, pb.QueueDeleteRequest_builder{QueueId: q.QueueId}.Build())
	return err
}

// Clear removes all objects from a queue partition.
func (q *Queue) Clear(options *QueueClearOptions) error {
	if options == nil {
		options = &QueueClearOptions{}
	}
	if options.Partition != "" && options.All {
		return InvalidError{"options.Partition must be \"\" when clearing all partitions"}
	}
	key, err := validatePartitionKey(options.Partition)
	if err != nil {
		return err
	}
	_, err = client.QueueClear(q.ctx, pb.QueueClearRequest_builder{
		QueueId:       q.QueueId,
		PartitionKey:  key,
		AllPartitions: options.All,
	}.Build())
	return err
}

// internal helper for both Get and GetMany.
func (q *Queue) get(n int, options *QueueGetOptions) ([]any, error) {
	if options == nil {
		options = &QueueGetOptions{}
	}
	partitionKey, err := validatePartitionKey(options.Partition)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	pollTimeout := 50 * time.Second
	if options.Timeout != nil && pollTimeout > *options.Timeout {
		pollTimeout = *options.Timeout
	}

	for {
		resp, err := client.QueueGet(q.ctx, pb.QueueGetRequest_builder{
			QueueId:      q.QueueId,
			PartitionKey: partitionKey,
			Timeout:      float32(pollTimeout.Seconds()),
			NValues:      int32(n),
		}.Build())
		if err != nil {
			return nil, err
		}
		if len(resp.GetValues()) > 0 {
			out := make([]any, len(resp.GetValues()))
			for i, raw := range resp.GetValues() {
				v, err := pickleDeserialize(raw)
				if err != nil {
					return nil, err
				}
				out[i] = v
			}
			return out, nil
		}
		if options.Timeout != nil {
			remaining := *options.Timeout - time.Since(startTime)
			if remaining <= 0 {
				message := fmt.Sprintf("queue %s did not return values within %s", q.QueueId, *options.Timeout)
				return nil, QueueEmptyError{message}
			}
			pollTimeout = min(pollTimeout, remaining)
		}
	}
}

// Get removes and returns one item (blocking by default).
//
// By default, this will wait until at least one item is present in the queue.
// If `timeout` is set, returns `QueueEmptyError` if no items are available
// within that timeout in milliseconds.
func (q *Queue) Get(options *QueueGetOptions) (any, error) {
	vals, err := q.get(1, options)
	if err != nil {
		return nil, err
	}
	return vals[0], nil // guaranteed len>=1
}

// GetMany removes up to n items.
//
// By default, this will wait until at least one item is present in the queue.
// If `timeout` is set, returns `QueueEmptyError` if no items are available
// within that timeout in milliseconds.
func (q *Queue) GetMany(n int, options *QueueGetOptions) ([]any, error) {
	return q.get(n, options)
}

// internal put helper (single/many).
func (q *Queue) put(values []any, options *QueuePutOptions) error {
	if options == nil {
		options = &QueuePutOptions{}
	}
	key, err := validatePartitionKey(options.Partition)
	if err != nil {
		return err
	}

	valuesEncoded := make([][]byte, len(values))
	for i, v := range values {
		b, err := pickleSerialize(v)
		if err != nil {
			return err
		}
		valuesEncoded[i] = b.Bytes()
	}

	deadline := time.Time{}
	if options.Timeout != nil {
		deadline = time.Now().Add(*options.Timeout)
	}

	delay := queueInitialPutBackoff
	ttl := options.PartitionTtl
	if ttl == 0 {
		ttl = queueDefaultPartitionTtl
	}

	for {
		_, err := client.QueuePut(q.ctx, pb.QueuePutRequest_builder{
			QueueId:             q.QueueId,
			Values:              valuesEncoded,
			PartitionKey:        key,
			PartitionTtlSeconds: int32(ttl.Seconds()),
		}.Build())
		if err == nil {
			return nil // success
		}

		if status.Code(err) != codes.ResourceExhausted {
			return err
		}

		// Queue is full, retry with exponential backoff up to the deadline.
		delay = min(delay*2, 30*time.Second)
		if !deadline.IsZero() {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return QueueFullError{fmt.Sprintf("Put failed on %s", q.QueueId)}
			}
			delay = min(delay, remaining)
		}
		select {
		case <-q.ctx.Done():
			return q.ctx.Err()
		case <-time.After(delay):
		}
	}
}

// Put adds a single item to the end of the queue.
//
// If the queue is full, this will retry with exponential backoff until the
// provided `timeout` is reached, or indefinitely if `timeout` is not set.
// Raises `QueueFullError` if the queue is still full after the timeout.
func (q *Queue) Put(v any, options *QueuePutOptions) error {
	return q.put([]any{v}, options)
}

// PutMany adds multiple items to the end of the queue.
//
// If the queue is full, this will retry with exponential backoff until the
// provided `timeout` is reached, or indefinitely if `timeout` is not set.
// Raises `QueueFullError` if the queue is still full after the timeout.
func (q *Queue) PutMany(values []any, options *QueuePutOptions) error {
	return q.put(values, options)
}

// Len returns the number of objects in the queue.
func (q *Queue) Len(options *QueueLenOptions) (int, error) {
	if options == nil {
		options = &QueueLenOptions{}
	}
	if options.Partition != "" && options.Total {
		return 0, InvalidError{"partition must be empty when requesting total length"}
	}
	key, err := validatePartitionKey(options.Partition)
	if err != nil {
		return 0, err
	}
	resp, err := client.QueueLen(q.ctx, pb.QueueLenRequest_builder{
		QueueId:      q.QueueId,
		PartitionKey: key,
		Total:        options.Total,
	}.Build())
	if err != nil {
		return 0, err
	}
	return int(resp.GetLen()), nil
}

// Iterate yields items from the queue until it is empty.
func (q *Queue) Iterate(options *QueueIterateOptions) iter.Seq2[any, error] {
	if options == nil {
		options = &QueueIterateOptions{}
	}

	itemPoll := options.ItemPollTimeout
	lastEntryID := ""
	maxPoll := 30 * time.Second

	return func(yield func(any, error) bool) {
		key, err := validatePartitionKey(options.Partition)
		if err != nil {
			yield(nil, err)
			return
		}

		fetchDeadline := time.Now().Add(itemPoll)
		for {
			pollDuration := max(0, min(maxPoll, time.Until(fetchDeadline)))
			resp, err := client.QueueNextItems(q.ctx, pb.QueueNextItemsRequest_builder{
				QueueId:         q.QueueId,
				PartitionKey:    key,
				ItemPollTimeout: float32(pollDuration.Seconds()),
				LastEntryId:     lastEntryID,
			}.Build())
			if err != nil {
				yield(nil, err)
				return
			}
			if len(resp.GetItems()) > 0 {
				for _, item := range resp.GetItems() {
					v, err := pickleDeserialize(item.GetValue())
					if err != nil {
						return
					}
					if !yield(v, nil) {
						return
					}
					lastEntryID = item.GetEntryId()
				}
				fetchDeadline = time.Now().Add(itemPoll)
			} else if time.Now().After(fetchDeadline) {
				return // exit on idle
			}
		}
	}
}
