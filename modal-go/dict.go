package modal

// Dict object, to be used with Modal Dicts.

import (
	"context"
	"fmt"
	"iter"
	"time"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DictPutOptions struct {
	SkipIfExists bool // skip adding the key if it already exists
}

// Dict is a distributed dictionary for key-value storage in Modal apps.
type Dict struct {
	DictId    string
	cancel    context.CancelFunc // only for ephemeral Dicts
	ephemeral bool
	ctx       context.Context
}

// DictEphemeral creates a nameless, temporary Dict. Caller must CloseEphemeral.
func DictEphemeral(ctx context.Context, options *EphemeralOptions) (*Dict, error) {
	if options == nil {
		options = &EphemeralOptions{}
	}
	var err error
	ctx, err = clientContext(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.DictGetOrCreate(ctx, pb.DictGetOrCreateRequest_builder{
		ObjectCreationType: pb.ObjectCreationType_OBJECT_CREATION_TYPE_EPHEMERAL,
		EnvironmentName:    environmentName(options.Environment),
	}.Build())
	if err != nil {
		return nil, err
	}

	heartbeatCtx, cancel := context.WithCancel(ctx)
	d := &Dict{DictId: resp.GetDictId(), cancel: cancel, ephemeral: true, ctx: ctx}

	// background heartbeat goroutine
	go func() {
		t := time.NewTicker(ephemeralObjectHeartbeatSleep)
		defer t.Stop()
		for {
			select {
			case <-heartbeatCtx.Done():
				return
			case <-t.C:
				_, _ = client.DictHeartbeat(heartbeatCtx, pb.DictHeartbeatRequest_builder{
					DictId: d.DictId,
				}.Build()) // ignore errors – next call will retry or context will cancel
			}
		}
	}()

	return d, nil
}

// CloseEphemeral deletes an ephemeral Dict, only used with DictEphemeral.
func (d *Dict) CloseEphemeral() {
	if d.ephemeral {
		d.cancel() // will stop heartbeat
	} else {
		// We panic in this case because of invalid usage. In general, methods
		// used with `defer` like CloseEphemeral should not return errors.
		panic(fmt.Sprintf("Dict %s is not ephemeral", d.DictId))
	}
}

// DictLookup returns a handle to a (possibly new) Dict by deployment name.
func DictLookup(ctx context.Context, name string, options *LookupOptions) (*Dict, error) {
	if options == nil {
		options = &LookupOptions{}
	}
	var err error
	ctx, err = clientContext(ctx)
	if err != nil {
		return nil, err
	}

	creationType := pb.ObjectCreationType_OBJECT_CREATION_TYPE_UNSPECIFIED
	if options.CreateIfMissing {
		creationType = pb.ObjectCreationType_OBJECT_CREATION_TYPE_CREATE_IF_MISSING
	}

	resp, err := client.DictGetOrCreate(ctx, pb.DictGetOrCreateRequest_builder{
		DeploymentName:     name,
		EnvironmentName:    environmentName(options.Environment),
		ObjectCreationType: creationType,
	}.Build())
	if err != nil {
		return nil, err
	}
	return &Dict{ctx: ctx, DictId: resp.GetDictId()}, nil
}

// DictDelete removes a Dict by name.
func DictDelete(ctx context.Context, name string, options *DeleteOptions) error {
	if options == nil {
		options = &DeleteOptions{}
	}
	d, err := DictLookup(ctx, name, &LookupOptions{Environment: options.Environment})
	if err != nil {
		return err
	}
	_, err = client.DictDelete(ctx, pb.DictDeleteRequest_builder{DictId: d.DictId}.Build())
	return err
}

// Get retrieves a value by key, with optional default value.
func (d *Dict) Get(key any, defaultValue *any) (any, error) {
	keyBytes, err := pickleSerialize(key)
	if err != nil {
		return nil, err
	}

	resp, err := client.DictGet(d.ctx, pb.DictGetRequest_builder{
		DictId: d.DictId,
		Key:    keyBytes.Bytes(),
	}.Build())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			if defaultValue != nil {
				return *defaultValue, nil
			}
			return nil, KeyError{"Key not found"}
		}
		return nil, err
	}

	if !resp.GetFound() {
		if defaultValue != nil {
			return *defaultValue, nil
		}
		return nil, KeyError{"Key not found"}
	}

	return pickleDeserialize(resp.GetValue())
}

// Put adds or updates a key-value pair. Returns true if the key was created (not updated).
func (d *Dict) Put(key any, value any, options *DictPutOptions) (bool, error) {
	if options == nil {
		options = &DictPutOptions{}
	}

	keyBytes, err := pickleSerialize(key)
	if err != nil {
		return false, err
	}

	valueBytes, err := pickleSerialize(value)
	if err != nil {
		return false, err
	}

	resp, err := client.DictUpdate(d.ctx, pb.DictUpdateRequest_builder{
		DictId: d.DictId,
		Updates: []*pb.DictEntry{pb.DictEntry_builder{
			Key:   keyBytes.Bytes(),
			Value: valueBytes.Bytes(),
		}.Build()},
		IfNotExists: options.SkipIfExists,
	}.Build())
	if err != nil {
		return false, err
	}

	return resp.GetCreated(), nil
}

// Pop removes and returns a value by key.
func (d *Dict) Pop(key any) (any, error) {
	keyBytes, err := pickleSerialize(key)
	if err != nil {
		return nil, err
	}

	resp, err := client.DictPop(d.ctx, pb.DictPopRequest_builder{
		DictId: d.DictId,
		Key:    keyBytes.Bytes(),
	}.Build())
	if err != nil {
		return nil, err
	}

	if !resp.GetFound() {
		return nil, KeyError{"Key not found"}
	}

	return pickleDeserialize(resp.GetValue())
}

// Contains checks if a key exists in the Dict.
func (d *Dict) Contains(key any) (bool, error) {
	keyBytes, err := pickleSerialize(key)
	if err != nil {
		return false, err
	}

	resp, err := client.DictContains(d.ctx, pb.DictContainsRequest_builder{
		DictId: d.DictId,
		Key:    keyBytes.Bytes(),
	}.Build())
	if err != nil {
		return false, err
	}

	return resp.GetFound(), nil
}

// Update adds multiple key-value pairs to the Dict.
func (d *Dict) Update(items map[any]any) error {
	updates := make([]*pb.DictEntry, 0, len(items))
	for key, value := range items {
		keyBytes, err := pickleSerialize(key)
		if err != nil {
			return err
		}
		valueBytes, err := pickleSerialize(value)
		if err != nil {
			return err
		}
		updates = append(updates, pb.DictEntry_builder{
			Key:   keyBytes.Bytes(),
			Value: valueBytes.Bytes(),
		}.Build())
	}

	_, err := client.DictUpdate(d.ctx, pb.DictUpdateRequest_builder{
		DictId:  d.DictId,
		Updates: updates,
	}.Build())
	return err
}

// Clear removes all entries from the Dict.
func (d *Dict) Clear() error {
	_, err := client.DictClear(d.ctx, pb.DictClearRequest_builder{
		DictId: d.DictId,
	}.Build())
	return err
}

// Len returns the number of key-value pairs in the Dict.
func (d *Dict) Len() (int, error) {
	resp, err := client.DictLen(d.ctx, pb.DictLenRequest_builder{
		DictId: d.DictId,
	}.Build())
	if err != nil {
		return 0, err
	}
	return int(resp.GetLen()), nil
}

// Keys returns an iterator over the Dict's keys.
func (d *Dict) Keys() iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		stream, err := client.DictContents(d.ctx, pb.DictContentsRequest_builder{
			DictId: d.DictId,
			Keys:   true,
		}.Build())
		if err != nil {
			yield(nil, err)
			return
		}

		for {
			entry, err := stream.Recv()
			if err != nil {
				if err.Error() != "EOF" {
					yield(nil, err)
				}
				return
			}

			key, err := pickleDeserialize(entry.GetKey())
			if err != nil {
				yield(nil, err)
				return
			}

			if !yield(key, nil) {
				return
			}
		}
	}
}

// Values returns an iterator over the Dict's values.
func (d *Dict) Values() iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		stream, err := client.DictContents(d.ctx, pb.DictContentsRequest_builder{
			DictId: d.DictId,
			Values: true,
		}.Build())
		if err != nil {
			yield(nil, err)
			return
		}

		for {
			entry, err := stream.Recv()
			if err != nil {
				if err.Error() != "EOF" {
					yield(nil, err)
				}
				return
			}

			value, err := pickleDeserialize(entry.GetValue())
			if err != nil {
				yield(nil, err)
				return
			}

			if !yield(value, nil) {
				return
			}
		}
	}
}

// Items returns an iterator over the Dict's key-value pairs.
func (d *Dict) Items() iter.Seq2[[2]any, error] {
	return func(yield func([2]any, error) bool) {
		stream, err := client.DictContents(d.ctx, pb.DictContentsRequest_builder{
			DictId: d.DictId,
			Keys:   true,
			Values: true,
		}.Build())
		if err != nil {
			yield([2]any{}, err)
			return
		}

		for {
			entry, err := stream.Recv()
			if err != nil {
				if err.Error() != "EOF" {
					yield([2]any{}, err)
				}
				return
			}

			key, err := pickleDeserialize(entry.GetKey())
			if err != nil {
				yield([2]any{}, err)
				return
			}

			value, err := pickleDeserialize(entry.GetValue())
			if err != nil {
				yield([2]any{}, err)
				return
			}

			if !yield([2]any{key, value}, nil) {
				return
			}
		}
	}
}
