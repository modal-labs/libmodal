package modal

// Function calls and invocations, to be used with Modal Functions.

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	pickle "github.com/kisielk/og-rek"
	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/protobuf/proto"
)

// From: modal/_utils/blob_utils.py
const maxObjectSizeBytes = 2 * 1024 * 1024 // 2 MiB

func timeNow() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}

// Function references a deployed Modal Function.
type Function struct {
	FunctionId string
	ctx        context.Context
}

// FunctionLookup looks up an existing Function.
func FunctionLookup(ctx context.Context, appName string, name string, options LookupOptions) (*Function, error) {
	ctx = clientContext(ctx)

	resp, err := client.FunctionGet(ctx, pb.FunctionGetRequest_builder{
		AppName:         appName,
		ObjectTag:       name,
		Namespace:       pb.DeploymentNamespace_DEPLOYMENT_NAMESPACE_WORKSPACE,
		EnvironmentName: environmentName(options.Environment),
	}.Build())

	if err != nil {
		return nil, err
	}

	return &Function{FunctionId: resp.GetFunctionId(), ctx: ctx}, nil
}

// Serialize function inputs to the Python pickle format.
func pickleSerialize(args []any, kwargs map[string]any) (bytes.Buffer, error) {
	var inputBuffer bytes.Buffer

	e := pickle.NewEncoder(&inputBuffer)
	err := e.Encode(pickle.Tuple{
		args,
		kwargs,
	})

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

// Execute a single input into a remote Function.
func (function *Function) Remote(ctx context.Context, args []any, kwargs map[string]any) (any, error) {
	ctx = clientContext(ctx)

	payload, err := pickleSerialize(args, kwargs)
	if err != nil {
		return nil, err
	}

	argsBytes := payload.Bytes()
	var argsBlobId *string
	if payload.Len() > maxObjectSizeBytes {
		blobId, err := blobUpload(ctx, argsBytes)
		if err != nil {
			return nil, err
		}
		argsBytes = nil
		argsBlobId = &blobId
	}

	// Single input sync invocation
	var functionInputs []*pb.FunctionPutInputsItem
	functionInputItem := pb.FunctionPutInputsItem_builder{
		Idx: 0,
		Input: pb.FunctionInput_builder{
			Args:       argsBytes,
			ArgsBlobId: argsBlobId,
			DataFormat: pb.DataFormat_DATA_FORMAT_PICKLE,
		}.Build(),
	}.Build()
	functionInputs = append(functionInputs, functionInputItem)

	functionMapResponse, err := client.FunctionMap(ctx, pb.FunctionMapRequest_builder{
		FunctionId:                 function.FunctionId,
		FunctionCallType:           pb.FunctionCallType_FUNCTION_CALL_TYPE_UNARY,
		FunctionCallInvocationType: pb.FunctionCallInvocationType_FUNCTION_CALL_INVOCATION_TYPE_SYNC,
		PipelinedInputs:            functionInputs,
	}.Build())
	if err != nil {
		return nil, fmt.Errorf("FunctionMap error: %v", err)
	}

	for {
		response, err := client.FunctionGetOutputs(ctx, pb.FunctionGetOutputsRequest_builder{
			FunctionCallId: functionMapResponse.GetFunctionCallId(),
			MaxValues:      1,
			Timeout:        55,
			LastEntryId:    "0-0",
			ClearOnSuccess: true,
			RequestedAt:    timeNow(),
		}.Build())
		if err != nil {
			return nil, fmt.Errorf("FunctionGetOutputs failed: %v", err)
		}

		// Output serialization may fail if any of the output items can't be deserialized
		// into a supported Go type. Users are expected to serialize outputs correctly.
		outputs := response.GetOutputs()
		if len(outputs) > 0 {
			return processResult(ctx, outputs[0].GetResult(), outputs[0].GetDataFormat())
		}
	}
}

// processResult processes the result from an invocation.
func processResult(ctx context.Context, result *pb.GenericResult, dataFormat pb.DataFormat) (any, error) {
	if result == nil {
		return nil, RemoteError{"Received null result from invocation"}
	}

	var data []byte
	var err error
	switch result.WhichDataOneof() {
	case pb.GenericResult_Data_case:
		data = result.GetData()
	case pb.GenericResult_DataBlobId_case:
		data, err = blobDownload(ctx, result.GetDataBlobId())
		if err != nil {
			return nil, err
		}
	case pb.GenericResult_DataOneof_not_set_case:
		data = nil
	}

	switch result.GetStatus() {
	case pb.GenericResult_GENERIC_STATUS_TIMEOUT:
		return nil, TimeoutError{result.GetException()}
	case pb.GenericResult_GENERIC_STATUS_INTERNAL_FAILURE:
		return nil, InternalFailure{result.GetException()}
	case pb.GenericResult_GENERIC_STATUS_SUCCESS:
		// Proceed to the block below this switch statement.
	default:
		// In this case, `result.GetData()` may have a pickled user code exception with traceback
		// from Python. We ignore this and only take the string representation.
		return nil, RemoteError{result.GetException()}
	}

	return deserializeDataFormat(data, dataFormat)
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

// blobDownload downloads a blob by its ID.
func blobDownload(ctx context.Context, blobId string) ([]byte, error) {
	resp, err := client.BlobGet(ctx, pb.BlobGetRequest_builder{
		BlobId: blobId,
	}.Build())
	if err != nil {
		return nil, err
	}
	s3resp, err := http.Get(resp.GetDownloadUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to download blob: %w", err)
	}
	defer s3resp.Body.Close()
	buf, err := io.ReadAll(s3resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob data: %w", err)
	}
	return buf, nil
}

func deserializeDataFormat(data []byte, dataFormat pb.DataFormat) (any, error) {
	switch dataFormat {
	case pb.DataFormat_DATA_FORMAT_PICKLE:
		return pickleDeserialize(data)
	case pb.DataFormat_DATA_FORMAT_ASGI:
		return nil, fmt.Errorf("ASGI data format is not supported in Go")
	case pb.DataFormat_DATA_FORMAT_GENERATOR_DONE:
		var done pb.GeneratorDone
		if err := proto.Unmarshal(data, &done); err != nil {
			return nil, fmt.Errorf("failed to unmarshal GeneratorDone: %w", err)
		}
		return &done, nil
	default:
		return nil, fmt.Errorf("unsupported data format: %s", dataFormat.String())
	}
}
