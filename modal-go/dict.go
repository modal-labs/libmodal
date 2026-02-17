package modal

// Dict implements the Modal Dict distributed key-value store.
//
// Dict keys are matched on the server by byte-equality of their serialized
// pickle representation. The Go SDK serializes keys using og-rek (pickle
// protocol 4, StrictUnicode) with post-processing to inject FRAME and MEMOIZE
// opcodes, producing bytes byte-identical to Python's cloudpickle for all
// supported primitive types. This enables cross-language interop: keys written
// by Go can be read by Python and vice versa.
//
// Per the Modal docs: "cloudpickle serialization is not guaranteed to be
// deterministic, so it is generally recommended to use primitive types for keys."
// See https://modal.com/docs/reference/modal.Dict
//
// Values are serialized using og-rek (protocol 4) without post-processing.
// Values only need to be valid pickle for deserialization — they do not require
// byte-equality with cloudpickle. This supports complex types (maps, slices,
// nested structures) as values.
//
// Supported key types (Go → Python):
//
//	nil            → None
//	bool           → bool
//	int, int8-64   → int
//	uint, uint8-64 → int
//	float32, float64 → float
//	string         → str
//	[]byte         → bytes

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"iter"
	"math/big"
	"strings"

	pickle "github.com/kisielk/og-rek"
	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Pickle protocol 4 opcodes used by ogrekToCloudpickle / cloudpickleToOgRek.
const (
	pickleShortBinBytes   = 'C'  // bytes len < 256
	pickleBinBytes        = 'B'  // bytes len < 2^32
	pickleShortBinUnicode = 0x8c // string len < 256
	pickleBinUnicode      = 'X'  // string len >= 256
	pickleLong1           = 0x8a // variable-length two's complement int
	pickleASCIIInt        = 'I'  // og-rek uses this for ints outside int32 range
	pickleMemoize         = 0x94
	pickleFrame           = 0x95
	pickleStop            = '.'
)

// dictOgRekP4Serialize uses og-rek with Protocol 4 + StrictUnicode.
// This is the base encoder used by both dictSerializeKey (with post-processing)
// and dictSerializeValue (without post-processing).
func dictOgRekP4Serialize(v any) ([]byte, error) {
	var buf bytes.Buffer
	e := pickle.NewEncoderWithConfig(&buf, &pickle.EncoderConfig{Protocol: 4, StrictUnicode: true})
	if err := e.Encode(v); err != nil {
		return nil, fmt.Errorf("og-rek pickle error: %w", err)
	}
	return buf.Bytes(), nil
}

// dictSerializeKey serializes a Dict key to pickle bytes that are byte-identical
// to Python's cloudpickle (protocol 4) for all supported primitive types.
// This enables cross-language Dict key lookups: keys written in Go can be
// read by Python and vice versa.
//
// Internally: runs og-rek (protocol 4, StrictUnicode) then post-processes the
// output via ogrekToCloudpickle to inject FRAME and MEMOIZE opcodes.
//
// Per the Modal docs: "cloudpickle serialization is not guaranteed to be
// deterministic, so it is generally recommended to use primitive types for keys."
func dictSerializeKey(v any) ([]byte, error) {
	raw, err := dictOgRekP4Serialize(v)
	if err != nil {
		return nil, err
	}
	return ogrekToCloudpickle(raw), nil
}

// og-rek encodes []byte as builtins.bytearray(SHORT_BINBYTES(data)), but
// cloudpickle uses bare SHORT_BINBYTES/BINBYTES. This prefix is the og-rek
// bytearray constructor pattern that we convert.
var ogrekBytearrayPrefix = []byte{
	0x8c, 0x08, 'b', 'u', 'i', 'l', 't', 'i', 'n', 's',
	0x8c, 0x09, 'b', 'y', 't', 'e', 'a', 'r', 'r', 'a', 'y',
	0x93, // STACK_GLOBAL
}

// ogrekToCloudpickle post-processes og-rek protocol 4 bytes into cloudpickle-
// compatible bytes by injecting FRAME and MEMOIZE opcodes, converting the
// ASCII 'I' opcode to LONG1, and unwrapping bytearray constructors.
//
// og-rek protocol 4 produces:   PROTO(4) + <opcodes> + STOP
// cloudpickle protocol 4 needs: PROTO(4) + [FRAME(len)] + <opcodes> [+ MEMOIZE] + STOP
//
// The post-processing:
//  1. Strips the PROTO header (0x80 0x04) and trailing STOP (0x2e)
//  2. Converts ASCII 'I' opcode (og-rek uses for ints > int32) to LONG1 (cloudpickle)
//  3. Unwraps builtins.bytearray(...) to bare SHORT_BINBYTES/BINBYTES
//  4. If the first opcode is a string or bytes opcode, appends MEMOIZE (0x94)
//  5. Re-appends STOP
//  6. If the content (after PROTO) is >= 4 bytes, wraps it in a FRAME opcode
//  7. Reassembles: PROTO(4) + [FRAME] + content + STOP
func ogrekToCloudpickle(raw []byte) []byte {
	// raw layout: [0x80 0x04] <body> [0x2e]
	// Minimum valid pickle is 3 bytes: PROTO(2) + STOP(1).
	if len(raw) < 3 {
		return raw
	}

	body := raw[2 : len(raw)-1] // strip PROTO header and STOP

	// og-rek uses the ASCII 'I' opcode (protocol 0) for ints outside int32 range.
	// cloudpickle uses LONG1 (0x8a) with binary two's complement encoding.
	// Convert: I<digits>\n → LONG1 <len> <bytes>
	if len(body) > 0 && body[0] == pickleASCIIInt {
		s := strings.TrimSuffix(string(body[1:]), "\n")
		n, ok := new(big.Int).SetString(s, 10)
		if ok {
			body = encodeLong1(n)
		}
	}

	// og-rek encodes []byte as builtins.bytearray(SHORT_BINBYTES(data))
	// but cloudpickle uses bare SHORT_BINBYTES/BINBYTES. Unwrap.
	if bytes.HasPrefix(body, ogrekBytearrayPrefix) {
		inner := body[len(ogrekBytearrayPrefix):]
		if extracted := extractBytesOpcode(inner); extracted != nil {
			body = extracted
		}
	}

	// Inject MEMOIZE after string/bytes opcodes.
	// cloudpickle always memoizes strings and bytes; og-rek never does.
	needsMemoize := len(body) > 0 &&
		(body[0] == pickleShortBinUnicode || body[0] == pickleBinUnicode ||
			body[0] == pickleShortBinBytes || body[0] == pickleBinBytes)

	var content bytes.Buffer
	content.Write(body)
	if needsMemoize {
		content.WriteByte(pickleMemoize)
	}
	content.WriteByte(pickleStop)

	// Assemble: PROTO(4) + optional FRAME + content.
	// cloudpickle emits a FRAME when content >= 4 bytes.
	var buf bytes.Buffer
	buf.Write([]byte{0x80, 0x04})

	data := content.Bytes()
	if len(data) >= 4 {
		buf.WriteByte(pickleFrame)
		binary.Write(&buf, binary.LittleEndian, uint64(len(data)))
	}
	buf.Write(data)

	return buf.Bytes()
}

// extractBytesOpcode extracts the bare SHORT_BINBYTES/BINBYTES from an og-rek
// bytearray constructor body: <bytes_opcode> <data> TUPLE1(0x85) REDUCE(0x52).
// Returns nil if the pattern doesn't match.
func extractBytesOpcode(inner []byte) []byte {
	if len(inner) < 4 {
		return nil
	}
	switch inner[0] {
	case pickleShortBinBytes:
		dataLen := int(inner[1])
		end := 2 + dataLen
		if end+2 <= len(inner) && inner[end] == 0x85 && inner[end+1] == 0x52 {
			return inner[:end]
		}
	case pickleBinBytes:
		if len(inner) < 5 {
			return nil
		}
		dataLen := int(binary.LittleEndian.Uint32(inner[1:5]))
		end := 5 + dataLen
		if end+2 <= len(inner) && inner[end] == 0x85 && inner[end+1] == 0x52 {
			return inner[:end]
		}
	}
	return nil
}

// encodeLong1 encodes a big.Int as a pickle LONG1 opcode: 0x8a <len> <bytes>.
// The bytes are minimal-length little-endian two's complement, matching Python's
// pickle LONG1 format.
func encodeLong1(n *big.Int) []byte {
	if n.Sign() == 0 {
		return []byte{pickleLong1, 0x00}
	}

	// big.Int.Bytes() returns unsigned big-endian bytes. We need signed
	// little-endian (two's complement). For positive numbers, Bytes() is
	// already the unsigned representation. For negative, compute two's
	// complement manually.
	var data []byte
	if n.Sign() > 0 {
		be := n.Bytes() // big-endian unsigned
		data = make([]byte, len(be))
		for i, b := range be {
			data[len(be)-1-i] = b // reverse to little-endian
		}
		// If high bit is set, append 0x00 to keep it positive.
		if data[len(data)-1] >= 0x80 {
			data = append(data, 0x00)
		}
	} else {
		// Two's complement for negative: subtract 1 from abs, then invert bits.
		abs := new(big.Int).Abs(n)
		abs.Sub(abs, big.NewInt(1))
		be := abs.Bytes()
		if len(be) == 0 {
			// n == -1: abs-1 == 0, Bytes() is empty. Two's complement is 0xff.
			data = []byte{0xff}
		} else {
			data = make([]byte, len(be))
			for i, b := range be {
				data[len(be)-1-i] = ^b // reverse and invert
			}
			// If high bit is not set, append 0xff to keep it negative.
			if data[len(data)-1] < 0x80 {
				data = append(data, 0xff)
			}
		}
	}

	result := make([]byte, 0, 2+len(data))
	result = append(result, pickleLong1, byte(len(data)))
	result = append(result, data...)
	return result
}

// decodeLong1 decodes the body portion of a LONG1 opcode (after the 0x8a byte)
// back to an ASCII 'I' opcode string for cloudpickleToOgRek. The input is:
// <len_byte> <data_bytes>.
func decodeLong1(body []byte) (string, int) {
	if len(body) < 1 {
		return "", 0
	}
	length := int(body[0])
	if len(body) < 1+length {
		return "", 0
	}
	data := body[1 : 1+length]

	if length == 0 {
		return "I0\n", 1 + length
	}

	// Determine sign from the high bit of the last byte (most significant in LE).
	negative := data[length-1] >= 0x80

	n := new(big.Int)
	if !negative {
		// Convert LE to BE.
		be := make([]byte, length)
		for i, b := range data {
			be[length-1-i] = b
		}
		n.SetBytes(be)
	} else {
		// Two's complement negative: invert bits, convert, then negate and subtract 1.
		be := make([]byte, length)
		for i, b := range data {
			be[length-1-i] = ^b
		}
		n.SetBytes(be)
		n.Add(n, big.NewInt(1))
		n.Neg(n)
	}

	return "I" + n.String() + "\n", 1 + length
}

// cloudpickleToOgRek strips FRAME and MEMOIZE opcodes from cloudpickle protocol 4
// bytes, producing plain og-rek protocol 4 bytes. This is the inverse of
// ogrekToCloudpickle.
func cloudpickleToOgRek(raw []byte) []byte {
	if len(raw) < 3 {
		return raw
	}

	// Start after PROTO header (0x80 0x04).
	pos := 2

	// Skip FRAME opcode + 8-byte length if present.
	if pos < len(raw) && raw[pos] == pickleFrame {
		pos += 1 + 8 // opcode + uint64 frame length
	}

	body := raw[pos:]

	// Strip trailing MEMOIZE before STOP: ...body... [0x94] [0x2e]
	if len(body) >= 2 && body[len(body)-2] == pickleMemoize && body[len(body)-1] == pickleStop {
		body = append(body[:len(body)-2], pickleStop)
	}

	// Convert LONG1 back to ASCII 'I' opcode (inverse of ogrekToCloudpickle).
	if len(body) >= 2 && body[0] == pickleLong1 {
		asciiInt, consumed := decodeLong1(body[1:])
		if consumed > 0 {
			var newBody bytes.Buffer
			newBody.WriteString(asciiInt)
			newBody.Write(body[1+consumed:])
			body = newBody.Bytes()
		}
	}

	// Wrap bare SHORT_BINBYTES/BINBYTES in builtins.bytearray() constructor
	// (inverse of ogrekToCloudpickle's bytearray unwrapping).
	if len(body) >= 2 && (body[0] == pickleShortBinBytes || body[0] == pickleBinBytes) {
		// body = <bytes_opcode> <data> STOP
		// Wrap: prefix + <bytes_opcode> <data> + TUPLE1 + REDUCE + STOP
		stopIdx := len(body) - 1 // last byte should be STOP
		if body[stopIdx] == pickleStop {
			var newBody bytes.Buffer
			newBody.Write(ogrekBytearrayPrefix)
			newBody.Write(body[:stopIdx]) // bytes opcode + data (without STOP)
			newBody.WriteByte(0x85)       // TUPLE1
			newBody.WriteByte(0x52)       // REDUCE
			newBody.WriteByte(pickleStop)
			body = newBody.Bytes()
		}
	}

	var buf bytes.Buffer
	buf.Write([]byte{0x80, 0x04})
	buf.Write(body)
	return buf.Bytes()
}

// DictService provides Dict related operations.
type DictService interface {
	Ephemeral(ctx context.Context, params *DictEphemeralParams) (*Dict, error)
	FromName(ctx context.Context, name string, params *DictFromNameParams) (*Dict, error)
	Delete(ctx context.Context, name string, params *DictDeleteParams) error
}

type dictServiceImpl struct{ client *Client }

// Dict is a distributed dictionary for storage in Modal Apps.
//
// Keys should be primitive types (see package doc for the full list).
// cloudpickle serialization is not guaranteed to be deterministic, so
// complex types as keys may produce inconsistent lookups across languages.
type Dict struct {
	DictID          string
	Name            string
	cancelEphemeral context.CancelFunc

	client *Client
}

// DictEphemeralParams are options for client.Dicts.Ephemeral.
type DictEphemeralParams struct {
	Environment string
}

// Ephemeral creates a nameless, temporary Dict that persists until CloseEphemeral is called, or the process exits.
func (s *dictServiceImpl) Ephemeral(ctx context.Context, params *DictEphemeralParams) (*Dict, error) {
	if params == nil {
		params = &DictEphemeralParams{}
	}

	resp, err := s.client.cpClient.DictGetOrCreate(ctx, pb.DictGetOrCreateRequest_builder{
		ObjectCreationType: pb.ObjectCreationType_OBJECT_CREATION_TYPE_EPHEMERAL,
		EnvironmentName:    environmentName(params.Environment, s.client.profile),
	}.Build())
	if err != nil {
		return nil, err
	}

	s.client.logger.DebugContext(ctx, "Created ephemeral Dict", "dict_id", resp.GetDictId())

	ephemeralCtx, cancel := context.WithCancel(context.Background())
	startEphemeralHeartbeat(ephemeralCtx, func() error {
		_, err := s.client.cpClient.DictHeartbeat(ephemeralCtx, pb.DictHeartbeatRequest_builder{
			DictId: resp.GetDictId(),
		}.Build())
		return err
	})

	return &Dict{
		DictID:          resp.GetDictId(),
		cancelEphemeral: cancel,
		client:          s.client,
	}, nil
}

// CloseEphemeral deletes an ephemeral Dict, only used with DictEphemeral.
func (d *Dict) CloseEphemeral() {
	if d.cancelEphemeral != nil {
		d.cancelEphemeral()
	} else {
		panic(fmt.Sprintf("Dict %s is not ephemeral", d.DictID))
	}
}

// DictFromNameParams are options for client.Dicts.FromName.
type DictFromNameParams struct {
	Environment     string
	CreateIfMissing bool
}

// FromName references a named Dict, creating if necessary.
func (s *dictServiceImpl) FromName(ctx context.Context, name string, params *DictFromNameParams) (*Dict, error) {
	if params == nil {
		params = &DictFromNameParams{}
	}

	creationType := pb.ObjectCreationType_OBJECT_CREATION_TYPE_UNSPECIFIED
	if params.CreateIfMissing {
		creationType = pb.ObjectCreationType_OBJECT_CREATION_TYPE_CREATE_IF_MISSING
	}

	resp, err := s.client.cpClient.DictGetOrCreate(ctx, pb.DictGetOrCreateRequest_builder{
		DeploymentName:     name,
		EnvironmentName:    environmentName(params.Environment, s.client.profile),
		ObjectCreationType: creationType,
	}.Build())

	if status, ok := status.FromError(err); ok && status.Code() == codes.NotFound {
		return nil, NotFoundError{fmt.Sprintf("Dict '%s' not found", name)}
	}
	if err != nil {
		return nil, err
	}

	s.client.logger.DebugContext(ctx, "Retrieved Dict", "dict_id", resp.GetDictId(), "dict_name", name)
	return &Dict{
		DictID:          resp.GetDictId(),
		Name:            name,
		cancelEphemeral: nil,
		client:          s.client,
	}, nil
}

// DictDeleteParams are options for client.Dicts.Delete.
type DictDeleteParams struct {
	Environment  string
	AllowMissing bool
}

// Delete removes a Dict by name.
//
// Warning: Deletion is irreversible and will affect any Apps currently using the Dict.
func (s *dictServiceImpl) Delete(ctx context.Context, name string, params *DictDeleteParams) error {
	if params == nil {
		params = &DictDeleteParams{}
	}

	d, err := s.FromName(ctx, name, &DictFromNameParams{
		Environment:     params.Environment,
		CreateIfMissing: false,
	})

	if err != nil {
		if _, ok := err.(NotFoundError); ok && params.AllowMissing {
			return nil
		}
		return err
	}

	_, err = s.client.cpClient.DictDelete(ctx, pb.DictDeleteRequest_builder{DictId: d.DictID}.Build())
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound && params.AllowMissing {
			return nil
		}
		return err
	}

	s.client.logger.DebugContext(ctx, "Deleted Dict", "dict_name", name, "dict_id", d.DictID)
	return nil
}

// Clear removes all items from the Dict.
func (d *Dict) Clear(ctx context.Context) error {
	_, err := d.client.cpClient.DictClear(ctx, pb.DictClearRequest_builder{
		DictId: d.DictID,
	}.Build())
	return err
}

// Get returns the value for a key. The second return value indicates whether the key was found.
func (d *Dict) Get(ctx context.Context, key any) (any, bool, error) {
	keyBytes, err := dictSerializeKey(key)
	if err != nil {
		return nil, false, err
	}

	resp, err := d.client.cpClient.DictGet(ctx, pb.DictGetRequest_builder{
		DictId: d.DictID,
		Key:    keyBytes,
	}.Build())
	if err != nil {
		return nil, false, err
	}

	if !resp.GetFound() {
		return nil, false, nil
	}

	val, err := pickleDeserialize(resp.GetValue())
	if err != nil {
		return nil, false, err
	}
	return val, true, nil
}

// Contains returns whether a key is present in the Dict.
func (d *Dict) Contains(ctx context.Context, key any) (bool, error) {
	keyBytes, err := dictSerializeKey(key)
	if err != nil {
		return false, err
	}

	resp, err := d.client.cpClient.DictContains(ctx, pb.DictContainsRequest_builder{
		DictId: d.DictID,
		Key:    keyBytes,
	}.Build())
	if err != nil {
		return false, err
	}
	return resp.GetFound(), nil
}

// Len returns the number of items in the Dict.
func (d *Dict) Len(ctx context.Context) (int, error) {
	resp, err := d.client.cpClient.DictLen(ctx, pb.DictLenRequest_builder{
		DictId: d.DictID,
	}.Build())
	if err != nil {
		return 0, err
	}
	return int(resp.GetLen()), nil
}

// DictPutParams are options for Dict.Put.
type DictPutParams struct {
	SkipIfExists bool
}

// Put adds a key-value pair to the Dict.
// Returns true if the entry was created, false if the key already existed and SkipIfExists was set.
func (d *Dict) Put(ctx context.Context, key any, value any, params *DictPutParams) (bool, error) {
	if params == nil {
		params = &DictPutParams{}
	}

	entries, err := serializeDictEntries(key, value)
	if err != nil {
		return false, err
	}

	resp, err := d.client.cpClient.DictUpdate(ctx, pb.DictUpdateRequest_builder{
		DictId:      d.DictID,
		Updates:     entries,
		IfNotExists: params.SkipIfExists,
	}.Build())
	if err != nil {
		return false, err
	}
	return resp.GetCreated(), nil
}

// Pop removes a key from the Dict and returns its value.
// The second return value indicates whether the key was found.
func (d *Dict) Pop(ctx context.Context, key any) (any, bool, error) {
	keyBytes, err := dictSerializeKey(key)
	if err != nil {
		return nil, false, err
	}

	resp, err := d.client.cpClient.DictPop(ctx, pb.DictPopRequest_builder{
		DictId: d.DictID,
		Key:    keyBytes,
	}.Build())
	if err != nil {
		return nil, false, err
	}

	if !resp.GetFound() {
		return nil, false, nil
	}

	val, err := pickleDeserialize(resp.GetValue())
	if err != nil {
		return nil, false, err
	}
	return val, true, nil
}

// Update adds multiple key-value pairs to the Dict.
func (d *Dict) Update(ctx context.Context, data map[any]any) error {
	entries, err := serializeDictEntriesMap(data)
	if err != nil {
		return err
	}

	_, err = d.client.cpClient.DictUpdate(ctx, pb.DictUpdateRequest_builder{
		DictId:  d.DictID,
		Updates: entries,
	}.Build())
	return err
}

// DictItem holds a key-value pair from a Dict iteration.
type DictItem struct {
	Key   any
	Value any
}

// Keys returns an iterator over the keys in the Dict.
func (d *Dict) Keys(ctx context.Context) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		stream, err := d.client.cpClient.DictContents(ctx, pb.DictContentsRequest_builder{
			DictId: d.DictID,
			Keys:   true,
		}.Build())
		if err != nil {
			yield(nil, err)
			return
		}

		for {
			entry, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				yield(nil, err)
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

// Values returns an iterator over the values in the Dict.
func (d *Dict) Values(ctx context.Context) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		stream, err := d.client.cpClient.DictContents(ctx, pb.DictContentsRequest_builder{
			DictId: d.DictID,
			Values: true,
		}.Build())
		if err != nil {
			yield(nil, err)
			return
		}

		for {
			entry, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				yield(nil, err)
				return
			}
			val, err := pickleDeserialize(entry.GetValue())
			if err != nil {
				yield(nil, err)
				return
			}
			if !yield(val, nil) {
				return
			}
		}
	}
}

// Items returns an iterator over the key-value pairs in the Dict.
func (d *Dict) Items(ctx context.Context) iter.Seq2[DictItem, error] {
	return func(yield func(DictItem, error) bool) {
		stream, err := d.client.cpClient.DictContents(ctx, pb.DictContentsRequest_builder{
			DictId: d.DictID,
			Keys:   true,
			Values: true,
		}.Build())
		if err != nil {
			yield(DictItem{}, err)
			return
		}

		for {
			entry, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				yield(DictItem{}, err)
				return
			}
			key, err := pickleDeserialize(entry.GetKey())
			if err != nil {
				yield(DictItem{}, err)
				return
			}
			val, err := pickleDeserialize(entry.GetValue())
			if err != nil {
				yield(DictItem{}, err)
				return
			}
			if !yield(DictItem{Key: key, Value: val}, nil) {
				return
			}
		}
	}
}

// dictSerializeValue serializes a Dict value using og-rek (protocol 4, StrictUnicode).
// Unlike keys, values don't need byte-equality with cloudpickle — they just need to be
// valid pickle that Python can deserialize. og-rek handles primitives and complex types
// (maps, slices, nested structures).
func dictSerializeValue(v any) ([]byte, error) {
	return dictOgRekP4Serialize(v)
}

// serializeDictEntries serializes a single key-value pair into a DictEntry slice.
func serializeDictEntries(key any, value any) ([]*pb.DictEntry, error) {
	keyBytes, err := dictSerializeKey(key)
	if err != nil {
		return nil, err
	}
	valBytes, err := dictSerializeValue(value)
	if err != nil {
		return nil, err
	}
	return []*pb.DictEntry{
		pb.DictEntry_builder{
			Key:   keyBytes,
			Value: valBytes,
		}.Build(),
	}, nil
}

// serializeDictEntriesMap serializes a map of key-value pairs into a DictEntry slice.
func serializeDictEntriesMap(data map[any]any) ([]*pb.DictEntry, error) {
	entries := make([]*pb.DictEntry, 0, len(data))
	for k, v := range data {
		keyBytes, err := dictSerializeKey(k)
		if err != nil {
			return nil, err
		}
		valBytes, err := dictSerializeValue(v)
		if err != nil {
			return nil, err
		}
		entries = append(entries, pb.DictEntry_builder{
			Key:   keyBytes,
			Value: valBytes,
		}.Build())
	}
	return entries, nil
}
