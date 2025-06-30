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

	pickle "github.com/kisielk/og-rek"
	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// From: modal/_utils/blob_utils.py
const maxObjectSizeBytes int = 2 * 1024 * 1024 // 2 MiB

// From: modal-client/modal/_utils/function_utils.py
const outputsTimeout time.Duration = time.Second * 55

// From: client/modal/_functions.py
const maxSystemRetries = 8

func timeNowSeconds() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}

// Function references a deployed Modal Function.
type Function struct {
	FunctionId string
	MethodName *string // used for class methods
	ctx        context.Context
}

// FunctionLookup looks up an existing Function.
func FunctionLookup(ctx context.Context, appName string, name string, options *LookupOptions) (*Function, error) {
	if options == nil {
		options = &LookupOptions{}
	}
	var err error
	ctx, err = clientContext(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.FunctionGet(ctx, pb.FunctionGetRequest_builder{
		AppName:         appName,
		ObjectTag:       name,
		Namespace:       pb.DeploymentNamespace_DEPLOYMENT_NAMESPACE_WORKSPACE,
		EnvironmentName: environmentName(options.Environment),
	}.Build())

	if status, ok := status.FromError(err); ok && status.Code() == codes.NotFound {
		return nil, NotFoundError{fmt.Sprintf("function '%s/%s' not found", appName, name)}
	}
	if err != nil {
		return nil, err
	}

	return &Function{FunctionId: resp.GetFunctionId(), ctx: ctx}, nil
}

// Serialize Go data types to the Python pickle format.
func pickleSerialize(v any) (bytes.Buffer, error) {
	var inputBuffer bytes.Buffer

	e := pickle.NewEncoder(&inputBuffer)
	err := e.Encode(v)

	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("error pickling data: %w", err)
	}
	return inputBuffer, nil
}

// Deserialize from Python pickle into Go basic types.
func pickleDeserialize(buffer []byte) (any, error) {
	decoder := pickle.NewDecoder(bytes.NewReader(buffer))
	result, err := decoder.Decode()
	if err != nil {
		return nil, fmt.Errorf("error unpickling data: %w", err)
	}
	return result, nil
}

// Serializes inputs, make a function call and return its ID
func (f *Function) createInput(args []any, kwargs map[string]any) (*pb.FunctionInput, error) {
	payload, err := pickleSerialize(pickle.Tuple{args, kwargs})
	if err != nil {
		return nil, err
	}

	argsBytes := payload.Bytes()
	var argsBlobId *string
	if payload.Len() > maxObjectSizeBytes {
		blobId, err := blobUpload(f.ctx, argsBytes)
		if err != nil {
			return nil, err
		}
		argsBytes = nil
		argsBlobId = &blobId
	}

	return pb.FunctionInput_builder{
		Args:       argsBytes,
		ArgsBlobId: argsBlobId,
		DataFormat: pb.DataFormat_DATA_FORMAT_PICKLE,
		MethodName: f.MethodName,
	}.Build(), nil
}

// Remote executes a single input on a remote Function.
func (f *Function) Remote(args []any, kwargs map[string]any) (any, error) {
	input, err := f.createInput(args, kwargs)
	if err != nil {
		return nil, err
	}
	invocation, err := createControlPlaneInvocation(f.ctx, f.FunctionId, input, pb.FunctionCallInvocationType_FUNCTION_CALL_INVOCATION_TYPE_SYNC)
	if err != nil {
		return nil, err
	}
	// TODO(ryan): Add tests for retries.
	retryCount := uint32(0)
	for {
		output, err := invocation.awaitOutput(nil)
		if err == nil {
			return output, nil
		}
		if errors.As(err, &InternalFailure{}) && retryCount <= maxSystemRetries {
			if retryErr := invocation.retry(retryCount); retryErr != nil {
				return nil, retryErr
			}
			retryCount++
			continue
		}
		return nil, err
	}
}

// Spawn starts running a single input on a remote function.
func (f *Function) Spawn(args []any, kwargs map[string]any) (*FunctionCall, error) {
	input, err := f.createInput(args, kwargs)
	if err != nil {
		return nil, err
	}
	invocation, err := createControlPlaneInvocation(f.ctx, f.FunctionId, input, pb.FunctionCallInvocationType_FUNCTION_CALL_INVOCATION_TYPE_SYNC)
	if err != nil {
		return nil, err
	}
	functionCall := FunctionCall{
		FunctionCallId: invocation.FunctionCallId,
		ctx:            f.ctx,
	}
	return &functionCall, nil
}

// blobUpload uploads a blob to storage and returns its ID.
func blobUpload(ctx context.Context, data []byte) (string, error) {
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
		return "", fmt.Errorf("function input size exceeds multipart upload threshold, unsupported by this SDK version")

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
