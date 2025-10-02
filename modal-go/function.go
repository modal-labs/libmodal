package modal

// Function calls and invocations, to be used with Modal Functions.

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/fxamacker/cbor/v2"
	pickle "github.com/kisielk/og-rek"
	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FunctionService provides Function related operations.
type FunctionService interface {
	FromName(ctx context.Context, appName string, name string, params *FunctionFromNameParams) (*Function, error)
}

type functionServiceImpl struct{ client *Client }

// From: modal/_utils/blob_utils.py
const maxObjectSizeBytes int = 2 * 1024 * 1024 // 2 MiB

// From: modal-client/modal/_utils/function_utils.py
const outputsTimeout time.Duration = time.Second * 55

// From: client/modal/_functions.py
const maxSystemRetries = 8

func timeNowSeconds() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}

// FunctionStats represents statistics for a running Function.
type FunctionStats struct {
	Backlog         int
	NumTotalRunners int
}

// FunctionUpdateAutoscalerParams contains options for overriding a Function's autoscaler behavior.
type FunctionUpdateAutoscalerParams struct {
	MinContainers    *uint32
	MaxContainers    *uint32
	BufferContainers *uint32
	ScaledownWindow  *uint32
}

// Function references a deployed Modal Function.
type Function struct {
	FunctionID     string
	MethodName     *string                    // used for class methods
	HandleMetadata *pb.FunctionHandleMetadata // TODO: make private?

	client *Client
}

// FunctionFromNameParams are options for client.Functions.FromName.
type FunctionFromNameParams struct {
	Environment     string
	CreateIfMissing bool
}

// FromName references a Function from a deployed App by its name.
func (s *functionServiceImpl) FromName(ctx context.Context, appName string, name string, params *FunctionFromNameParams) (*Function, error) {
	if params == nil {
		params = &FunctionFromNameParams{}
	}

	resp, err := s.client.cpClient.FunctionGet(ctx, pb.FunctionGetRequest_builder{
		AppName:         appName,
		ObjectTag:       name,
		EnvironmentName: environmentName(params.Environment, s.client.profile),
	}.Build())

	if status, ok := status.FromError(err); ok && status.Code() == codes.NotFound {
		return nil, NotFoundError{fmt.Sprintf("Function '%s/%s' not found", appName, name)}
	}
	if err != nil {
		return nil, err
	}

	handleMetadata := resp.GetHandleMetadata()
	return &Function{FunctionID: resp.GetFunctionId(), HandleMetadata: handleMetadata, client: s.client}, nil
}

// pickleSerialize serializes Go data types to the Python pickle format.
// NOTE: This is only used by Queue operations. Function calls use CBOR only.
func pickleSerialize(v any) (bytes.Buffer, error) {
	var inputBuffer bytes.Buffer

	e := pickle.NewEncoder(&inputBuffer)
	err := e.Encode(v)

	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("error pickling data: %w", err)
	}
	return inputBuffer, nil
}

// pickleDeserialize deserializes from Python pickle into Go basic types.
// NOTE: This is only used by Queue operations. Function calls use CBOR only.
func pickleDeserialize(buffer []byte) (any, error) {
	decoder := pickle.NewDecoder(bytes.NewReader(buffer))
	result, err := decoder.Decode()
	if err != nil {
		return nil, fmt.Errorf("error unpickling data: %w", err)
	}
	return result, nil
}

// cborEncoder is configured with time tags enabled so that time.Time values
// are represented as datetime objects in Python. Uses TimeRFC3339Nano to preserve
// nanosecond precision (Python datetime has microsecond precision).
//
// Both options are required:
//   - Time: TimeRFC3339Nano - specifies the format (RFC3339 with nanosecond precision)
//   - TimeTag: EncTagRequired - wraps the time in CBOR tag 0, signaling it's a datetime
//     Without the tag, Python would receive it as a plain string, not a datetime object.
var cborEncoder, _ = cbor.EncOptions{
	Time:    cbor.TimeRFC3339Nano,
	TimeTag: cbor.EncTagRequired,
}.EncMode()

// cborSerialize serializes Go data types to the CBOR format.
// Uses CBOR time tags so that time.Time values are represented as
// datetime objects in Python.
func cborSerialize(v any) ([]byte, error) {
	data, err := cborEncoder.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("error encoding CBOR data: %w", err)
	}
	return data, nil
}

// cborDeserialize deserializes from CBOR into Go basic types.
func cborDeserialize(buffer []byte) (any, error) {
	var result any
	err := cbor.Unmarshal(buffer, &result)
	if err != nil {
		return nil, fmt.Errorf("error decoding CBOR data: %w", err)
	}
	return result, nil
}

// createInput serializes inputs, makes a function call and returns its ID
func (f *Function) createInput(ctx context.Context, args []any, kwargs map[string]any) (*pb.FunctionInput, error) {
	// Check supported input formats and require CBOR
	supportedInputFormats := f.getSupportedInputFormats()
	cborSupported := false
	for _, format := range supportedInputFormats {
		if format == pb.DataFormat_DATA_FORMAT_CBOR {
			cborSupported = true
			break
		}
	}

	// Error if CBOR is not supported
	if !cborSupported {
		return nil, fmt.Errorf("function does not support DATA_FORMAT_CBOR - please redeploy the function using modal client >= 1.2")
	}

	// Use CBOR encoding
	// Ensure args and kwargs are not nil to match expected behavior
	cborArgs := args
	if cborArgs == nil {
		cborArgs = []any{}
	}
	cborKwargs := kwargs
	if cborKwargs == nil {
		cborKwargs = map[string]any{}
	}
	argsBytes, err := cborSerialize([]any{cborArgs, cborKwargs})
	if err != nil {
		return nil, err
	}
	dataFormat := pb.DataFormat_DATA_FORMAT_CBOR
	var argsBlobID *string
	if len(argsBytes) > maxObjectSizeBytes {
		blobID, err := blobUpload(ctx, f.client.cpClient, argsBytes)
		if err != nil {
			return nil, err
		}
		argsBytes = nil
		argsBlobID = &blobID
	}

	return pb.FunctionInput_builder{
		Args:       argsBytes,
		ArgsBlobId: argsBlobID,
		DataFormat: dataFormat,
		MethodName: f.MethodName,
	}.Build(), nil
}

// getSupportedInputFormats returns the supported input formats for this function.
// If no metadata is available, it returns an empty slice.
func (f *Function) getSupportedInputFormats() []pb.DataFormat {
	if f.HandleMetadata != nil && len(f.HandleMetadata.GetSupportedInputFormats()) > 0 {
		return f.HandleMetadata.GetSupportedInputFormats()
	}
	// Return empty slice if no metadata is available - this will cause CBOR validation to fail
	return []pb.DataFormat{}
}

// getInputPlaneUrl returns the input plane URL for this function, if available.
func (f *Function) getInputPlaneUrl() string {
	if f.HandleMetadata != nil {
		return f.HandleMetadata.GetInputPlaneUrl()
	}
	return ""
}

// getWebURL returns the web URL for this function, if it's a web endpoint.
func (f *Function) getWebURL() string {
	if f.HandleMetadata != nil {
		return f.HandleMetadata.GetWebUrl()
	}
	return ""
}

// Remote executes a single input on a remote Function.
func (f *Function) Remote(ctx context.Context, args []any, kwargs map[string]any) (any, error) {
	input, err := f.createInput(ctx, args, kwargs)
	if err != nil {
		return nil, err
	}
	invocation, err := f.createRemoteInvocation(ctx, input)
	if err != nil {
		return nil, err
	}
	// TODO(ryan): Add tests for retries.
	retryCount := uint32(0)
	for {
		output, err := invocation.awaitOutput(ctx, nil)
		if err == nil {
			return output, nil
		}
		if errors.As(err, &InternalFailure{}) && retryCount <= maxSystemRetries {
			if retryErr := invocation.retry(ctx, retryCount); retryErr != nil {
				return nil, retryErr
			}
			retryCount++
			continue
		}
		return nil, err
	}
}

// createRemoteInvocation creates an Invocation using either the input plane or control plane.
func (f *Function) createRemoteInvocation(ctx context.Context, input *pb.FunctionInput) (invocation, error) {
	inputPlaneURL := f.getInputPlaneUrl()
	if inputPlaneURL != "" {
		ipClient, err := f.client.ipClient(inputPlaneURL)
		if err != nil {
			return nil, err
		}
		return createInputPlaneInvocation(ctx, ipClient, f.FunctionID, input)
	}
	return createControlPlaneInvocation(ctx, f.client.cpClient, f.FunctionID, input, pb.FunctionCallInvocationType_FUNCTION_CALL_INVOCATION_TYPE_SYNC)
}

// Spawn starts running a single input on a remote Function.
func (f *Function) Spawn(ctx context.Context, args []any, kwargs map[string]any) (*FunctionCall, error) {
	input, err := f.createInput(ctx, args, kwargs)
	if err != nil {
		return nil, err
	}
	invocation, err := createControlPlaneInvocation(ctx, f.client.cpClient, f.FunctionID, input, pb.FunctionCallInvocationType_FUNCTION_CALL_INVOCATION_TYPE_SYNC)
	if err != nil {
		return nil, err
	}
	functionCall := FunctionCall{
		FunctionCallID: invocation.FunctionCallID,
		client:         f.client,
	}
	return &functionCall, nil
}

// GetCurrentStats returns a FunctionStats object with statistics about the Function.
func (f *Function) GetCurrentStats(ctx context.Context) (*FunctionStats, error) {
	resp, err := f.client.cpClient.FunctionGetCurrentStats(ctx, pb.FunctionGetCurrentStatsRequest_builder{
		FunctionId: f.FunctionID,
	}.Build())
	if err != nil {
		return nil, err
	}

	return &FunctionStats{
		Backlog:         int(resp.GetBacklog()),
		NumTotalRunners: int(resp.GetNumTotalTasks()),
	}, nil
}

// UpdateAutoscaler overrides the current autoscaler behavior for this Function.
func (f *Function) UpdateAutoscaler(ctx context.Context, params *FunctionUpdateAutoscalerParams) error {
	if params == nil {
		params = &FunctionUpdateAutoscalerParams{}
	}

	settings := pb.AutoscalerSettings_builder{
		MinContainers:    params.MinContainers,
		MaxContainers:    params.MaxContainers,
		BufferContainers: params.BufferContainers,
		ScaledownWindow:  params.ScaledownWindow,
	}.Build()

	_, err := f.client.cpClient.FunctionUpdateSchedulingParams(ctx, pb.FunctionUpdateSchedulingParamsRequest_builder{
		FunctionId:           f.FunctionID,
		WarmPoolSizeOverride: 0, // Deprecated field, always set to 0
		Settings:             settings,
	}.Build())

	return err
}

// GetWebURL returns the URL of a Function running as a web endpoint.
// Returns empty string if this Function is not a web endpoint.
func (f *Function) GetWebURL() string {
	return f.getWebURL()
}

// blobUpload uploads a blob to storage and returns its ID.
func blobUpload(ctx context.Context, client pb.ModalClientClient, data []byte) (string, error) {
	md5sum := md5.Sum(data)
	sha256sum := sha256.Sum256(data)
	contentMd5 := base64.StdEncoding.EncodeToString(md5sum[:])
	contentSha256 := base64.StdEncoding.EncodeToString(sha256sum[:])

	resp, err := client.BlobCreate(ctx, pb.BlobCreateRequest_builder{
		ContentMd5:          contentMd5,
		ContentSha256Base64: contentSha256,
		ContentLength:       int64(len(data)),
	}.Build())
	if err != nil {
		return "", fmt.Errorf("failed to create blob: %w", err)
	}

	switch resp.WhichUploadTypeOneof() {
	case pb.BlobCreateResponse_Multipart_case:
		return "", fmt.Errorf("Function input size exceeds multipart upload threshold, unsupported by this SDK version")

	case pb.BlobCreateResponse_UploadUrl_case:
		req, err := http.NewRequest("PUT", resp.GetUploadUrl(), bytes.NewReader(data))
		if err != nil {
			return "", fmt.Errorf("failed to create upload request: %w", err)
		}
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Content-MD5", contentMd5)
		uploadResp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to upload blob: %w", err)
		}
		defer uploadResp.Body.Close()
		if uploadResp.StatusCode < 200 || uploadResp.StatusCode >= 300 {
			return "", fmt.Errorf("failed blob upload: %s", uploadResp.Status)
		}
		// Skip client-side ETag header validation for now (MD5 checksum).
		return resp.GetBlobId(), nil

	default:
		return "", fmt.Errorf("missing upload URL in BlobCreate response")
	}
}
