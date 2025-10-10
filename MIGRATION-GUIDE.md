# Migration Guide for the beta Modal SDKs for JS and Go, v0.5.0

The Modal JS and Go SDKs went into beta with the version 0.5 release in October 2025. This release brings us closer to feature parity with the Python SDK (with notable exceptions like defining functions, Volume filesystem API, some Image building APIs, and Dicts not yet supported). It's a big step towards bringing JavaScript/TypeScript and Go to the same high level of developer experience and stability as the Python SDK.

The beta release includes breaking changes to improve SDK ergonomics and align with general SDK best practices. While adapting requires some code changes, we believe these improvements make Modal easier to use going forward.

The main changes are:

- The SDKs now expose a central Modal Client object as the main entry point for interacting with Modal resources.
- The interface for working with Modal object instances (Functions, Sandboxes, Images, etc.) is largely the same as before, with some naming changes.
- Calling deployed Functions and classes now uses a new protocol for payload serialization which requires the deployed apps to use the Modal Python SDK 1.2 or newer.
- Internally removed the global client (and config/profile data in global scope), moving all that to the Client type.
- Consistent parameter naming across both SDKs: all `Options` structs/interfaces renamed to `Params`.
- Go-specific changes:
  - Changed how we do context passing, so contexts now only affect the current operation and are not used for lifecycle management of the created resources.
  - All `Params` structs are now passed as pointers for consistency and to support optional parameters.
  - Field names follow Go casing conventions (e.g., `Id` → `ID`, `Url` → `URL`, `TokenId` → `TokenID`).

## Calling deployed Modal Functions and classes

Starting with this version, invoking remote Functions and class methods through `.remote()` and similar uses a new serialization protocol that requires the referenced modal Apps to be deployed using the Modal Python SDK 1.2 or newer. In addition, your deployed Apps need to be on the 2025.06 image builder version or newer (see https://modal.com/settings/image-config for more information) or have the `cbor2` Python package installed in their image.

## API changes

See below for a list of all changes in [JavaScript/TypeScript](#javascripttypescript) and [Go](#go). See also the updated examples in [JS](./modal-js/examples) and [Go](./modal-go/examples) for a sense of how the API has changed.

## JavaScript/TypeScript

Brief example of using the new API:

```ts
import { ModalClient } from "modal";

const client = new ModalClient();

const app = await client.apps.fromName("libmodal-example", {
  createIfMissing: true,
});
const image = client.images.fromRegistry("alpine:3.21");
const volume = await client.volumes.fromName("libmodal-example-volume", {
  createIfMissing: true,
});

const sb = await client.sandboxes.create(app, image, {
  volumes: { "/mnt/volume": volume },
});
const p = await sb.exec(["cat", "/mnt/volume/message.txt"]);
console.log(`Message: ${await p.stdout.readText()}`);
await sb.terminate();

const echo = await client.functions.fromName("libmodal-example", "echo");
console.log(await echo.remote(["Hello world!"]));
```

### Client

```ts
import { ModalClient } from "modal";
const client = new ModalClient();
// or customized:
const client = new ModalClient({ tokenId: "...", tokenSecret: "..." });
```

- `initializeClient(...)` -> `new ModalClient(...)`

### App

- `App.lookup(...)` -> `client.apps.fromName(...)`

### Cls

- `Cls.lookup(...)` -> `client.cls.fromName(...)`

### Function

- `Function_.lookup(...)` -> `client.functions.fromName(...)`

### FunctionCall

- `FunctionCall.fromId(...)` -> `client.functionCalls.fromId(...)`

### Image

- `app.imageFromRegistry(...)` -> `client.images.fromRegistry(...)`
- `app.imageFromAwsEcr(...)` -> `client.images.fromAwsEcr(...)`
- `app.imageFromGcpArtifactRegistry(...)` -> `client.images.fromGcpArtifactRegistry(...)`
- `Image.fromRegistry(...)` -> `client.images.fromRegistry(...)`
- `Image.fromAwsEcr(...)` -> `client.images.fromAwsEcr(...)`
- `Image.fromGcpArtifactRegistry(...)` -> `client.images.fromGcpArtifactRegistry(...)`
- `Image.fromId(...)` -> `client.images.fromId(...)`
- `Image.delete(...)` -> `client.images.delete(...)`

### Proxy

- `Proxy.fromName(...)` -> `client.proxies.fromName(...)`

### Queue

- `Queue.lookup(...)` -> `client.queues.fromName(...)`
- `Queue.fromName(...)` -> `client.queues.fromName(...)`
- `Queue.ephemeral(...)` -> `client.queues.ephemeral(...)`
- `Queue.delete(...)` -> `client.queues.delete(...)`

### Sandbox

- `app.createSandbox(image, { ... })` -> `client.sandboxes.create(app, image, { ... })`
- `Sandbox.fromId(...)` -> `client.sandboxes.fromId(...)`
- `Sandbox.fromName(...)` -> `client.sandboxes.fromName(...)`
- `Sandbox.list(...)` -> `client.sandboxes.list(...)`

### Secret

- `Secret.fromName(...)` -> `client.secrets.fromName(...)`
- `Secret.fromObject(...)` -> `client.secrets.fromObject(...)`

### Volume

- `Volume.fromName(...)` -> `client.volumes.fromName(...)`
- `Volume.ephemeral(...)` -> `client.volumes.ephemeral(...)`

### Parameter Type Renames

- `ClsOptions` -> `ClsWithOptionsParams`
- `ClsConcurrencyOptions` -> `ClsWithConcurrencyParams`
- `ClsBatchingOptions` -> `ClsWithBatchingParams`
- `DeleteOptions` -> specific `*DeleteParams` types: `QueueDeleteParams`
- `EphemeralOptions` -> specific `*EphemeralParams` types: `QueueEphemeralParams`, `VolumeEphemeralParams`
- `ExecOptions` -> `SandboxExecParams`
- `UpdateAutoscalerOptions` -> `FunctionUpdateAutoscalerParams`
- `FunctionCallGetOptions` -> `FunctionCallGetParams`
- `FunctionCallCancelOptions` -> `FunctionCallCancelParams`
- `ImageDockerfileCommandsOptions` -> `ImageDockerfileCommandsParams`
- `ImageDeleteOptions` -> `ImageDeleteParams`
- `LookupOptions` -> specific `*FromNameParams` types: `AppFromNameParams`, `ClsFromNameParams`, `FunctionFromNameParams`, `QueueFromNameParams`
- `ProxyFromNameOptions` -> `ProxyFromNameParams`
- `QueueClearOptions` -> `QueueClearParams`
- `QueueGetOptions` -> `QueueGetParams` and `QueueGetManyParams`
- `QueuePutOptions` -> `QueuePutParams` and `QueuePutManyParams`
- `QueueLenOptions` -> `QueueLenParams`
- `QueueIterateOptions` -> `QueueIterateParams`
- `SandboxCreateOptions` -> `SandboxCreateParams`
- `SandboxFromNameOptions` -> `SandboxFromNameParams`
- `SandboxListOptions` -> `SandboxListParams`
- `SecretFromNameOptions` -> `SecretFromNameParams`
- `SecretFromObjectParams` -> new export (no previous equivalent)
- `VolumeFromNameOptions` -> `VolumeFromNameParams`

## Go

Brief example of using the new API (with `err` handling omitted for brevity):

```go
package main

import (
	"context"
	"fmt"
	"io"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()

	client, _ := modal.NewClient()

	app, _ := client.Apps.FromName(ctx, "libmodal-example", &modal.AppFromNameParams{CreateIfMissing: true})
	image := client.Images.FromRegistry("alpine:3.21", nil)
	volume, _ := client.Volumes.FromName(ctx, "libmodal-example-volume", &modal.VolumeFromNameParams{CreateIfMissing: true})

	sb, _ := client.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Volumes: map[string]*modal.Volume{"/mnt/volume": volume},
	})
	defer sb.Terminate(context.Background())
	p, _ := sb.Exec(ctx, []string{"cat", "/mnt/volume/message.txt"}, nil)
	stdout, _ := io.ReadAll(p.Stdout)
	fmt.Printf("Message: %s\n", stdout)

	echo, _ := client.Functions.FromName(ctx, "libmodal-example", "echo", nil)
	result, _ := echo.Remote(ctx, []any{"Hello world!"}, nil)
	fmt.Println(result)
}
```

### General notes

- Many methods now require `ctx context.Context` as the first parameter
- Field renames in structs:
  - `TokenId` -> `TokenID`
  - `AppId` -> `AppID`
  - `SandboxId` -> `SandboxID`
  - `ImageId` -> `ImageID`
  - `SecretId` -> `SecretID`
  - `VolumeId` -> `VolumeID`
  - `QueueId` -> `QueueID`
  - `FunctionId` -> `FunctionID`
  - `ClsId` -> `ClsID`
  - `FunctionCallId` -> `FunctionCallID`
  - `ProxyId` -> `ProxyID`
  - `ServerUrl` -> `ServerURL`
  - `BucketEndpointUrl` -> `BucketEndpointURL`

### Client

```go
import "github.com/modal-labs/libmodal/modal-go"
client, err := modal.NewClient()
// or customized:
client, err := modal.NewClientWithOptions(&modal.ClientParams{
    TokenID:     "...",
    TokenSecret: "...",
})
```

- `modal.InitializeClient(modal.ClientOptions{...})` -> `modal.NewClient()` or `modal.NewClientWithOptions(&modal.ClientParams{...})`

### App

- `modal.AppLookup(ctx, "my-app", &modal.LookupOptions{...})` -> `client.Apps.FromName(ctx, "my-app", &modal.AppFromNameParams{...})`

### CloudBucketMount

- `modal.NewCloudBucketMount(..., &modal.CloudBucketMountOptions{...})` -> `client.CloudBucketMounts.New(..., &modal.CloudBucketMountParams{...})`

### Cls

- `modal.ClsLookup(ctx, ..., &modal.LookupOptions{...})` -> `client.Cls.FromName(ctx, ..., &modal.ClsFromNameParams{...})`

#### Cls methods

- `cls.Instance(...)` -> `cls.Instance(ctx, ...)`
- `cls.WithOptions(modal.ClsOptions{...})` -> `cls.WithOptions(&modal.ClsWithOptionsParams{...})`
- `cls.WithConcurrency(modal.ClsConcurrencyOptions{...})` -> `cls.WithConcurrency(&modal.ClsWithConcurrencyParams{...})`
- `cls.WithBatching(modal.ClsBatchingOptions{...})` -> `cls.WithBatching(&modal.ClsWithBatchingParams{...})`

### Function

- `modal.FunctionLookup(ctx, ..., &modal.LookupOptions{...})` -> `client.Functions.FromName(ctx, ..., &modal.FunctionFromNameParams{...})`

#### Function methods

- `function.Remote(...)` -> `function.Remote(ctx, ...)`
- `function.Spawn(...)` -> `function.Spawn(ctx, ...)`
- `function.GetCurrentStats()` -> `function.GetCurrentStats(ctx)`
- `function.UpdateAutoscaler(modal.UpdateAutoscalerOptions{...})` -> `function.UpdateAutoscaler(ctx, &modal.FunctionUpdateAutoscalerParams{...})`

### FunctionCall

- `modal.FunctionCallFromId(ctx, "call-id")` -> `client.FunctionCalls.FromID(ctx, "call-id")`

#### FunctionCall methods

- `functionCall.Get(&modal.FunctionCallGetOptions{...})` -> `functionCall.Get(ctx, &modal.FunctionCallGetParams{...})`
- `functionCall.Cancel(&modal.FunctionCallCancelOptions{...})` -> `functionCall.Cancel(ctx, &modal.FunctionCallCancelParams{...})`

### Image

- `app.ImageFromRegistry(..., &modal.ImageFromRegistryOptions{...})` -> `client.Images.FromRegistry(ctx, ..., &modal.ImageFromRegistryParams{...})`
- `modal.NewImageFromRegistry(..., &modal.ImageFromRegistryOptions{...})` -> `client.Images.FromRegistry(ctx, ..., &modal.ImageFromRegistryParams{...})`
- `modal.NewImageFromAwsEcr(..., secret)` -> `client.Images.FromAwsEcr(ctx, ..., secret)`
- `modal.NewImageFromGcpArtifactRegistry(..., secret)` -> `client.Images.FromGcpArtifactRegistry(ctx, ..., secret)`
- `modal.NewImageFromId(ctx, ...)` -> `client.Images.FromID(ctx, ...)`
- `modal.ImageDelete(ctx, ..., &modal.ImageDeleteOptions{...})` -> `client.Images.Delete(ctx, ..., &modal.ImageDeleteParams{...})`

#### Image methods

- `image.DockerfileCommands(..., &modal.ImageDockerfileCommandsOptions{...})` -> `image.DockerfileCommands(..., &modal.ImageDockerfileCommandsParams{...})`
- `image.Build(app)` -> `image.Build(ctx, app)`

### Proxy

- `modal.ProxyFromName(..., &modal.ProxyFromNameOptions{...})` -> `client.Proxies.FromName(..., &modal.ProxyFromNameParams{...})`

### Queue

- `modal.QueueLookup(ctx, ..., &modal.LookupOptions{...})` -> `client.Queues.FromName(ctx, ..., &modal.QueueFromNameParams{...})`
- `modal.QueueEphemeral(ctx, &modal.EphemeralOptions{...})` -> `client.Queues.Ephemeral(ctx, &modal.QueueEphemeralParams{...})`
- `modal.QueueDelete(ctx, ..., &modal.DeleteOptions{...})` -> `client.Queues.Delete(ctx, ..., &modal.QueueDeleteParams{...})`

#### Queue methods

- `queue.Clear(&modal.QueueClearOptions{...})` -> `queue.Clear(ctx, &modal.QueueClearParams{...})`
- `queue.Get(&modal.QueueGetOptions{...})` -> `queue.Get(ctx, &modal.QueueGetParams{...})`
- `queue.GetMany(..., &modal.QueueGetOptions{...})` -> `queue.GetMany(ctx, ..., &modal.QueueGetManyParams{...})`
- `queue.Put(..., &modal.QueuePutOptions{...})` -> `queue.Put(ctx, ..., &modal.QueuePutParams{...})`
- `queue.PutMany(..., &modal.QueuePutOptions{...})` -> `queue.PutMany(ctx, ..., &modal.QueuePutManyParams{...})`
- `queue.Len(&modal.QueueLenOptions{...})` -> `queue.Len(ctx, &modal.QueueLenParams{...})`
- `queue.Iterate(&modal.QueueIterateOptions{...})` -> `queue.Iterate(ctx, &modal.QueueIterateParams{...})`

### Retries

- `modal.NewRetries(..., &modal.RetriesOptions{...})` -> `modal.NewRetries(..., &modal.RetriesParams{...})`

### Sandbox

- `app.CreateSandbox(image, &modal.SandboxOptions{...})` -> `client.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{...})`
- `modal.SandboxFromId(ctx, "sandbox-id")` -> `client.Sandboxes.FromID(ctx, "sandbox-id")`
- `modal.SandboxFromName(ctx, "app-name", "sandbox-name", &modal.SandboxFromNameOptions{...})` -> `client.Sandboxes.FromName(ctx, "app-name", "sandbox-name", &modal.SandboxFromNameParams{...})`
- `modal.SandboxList(ctx, &modal.SandboxListOptions{...})` -> `client.Sandboxes.List(ctx, &modal.SandboxListParams{...})`

#### Sandbox methods

- `sandbox.Exec(..., modal.ExecOptions{...})` -> `sandbox.Exec(ctx, ..., &modal.SandboxExecParams{...})`
- `sandbox.Open(...)` -> `sandbox.Open(ctx, ...)`
- `sandbox.Terminate()` -> `sandbox.Terminate(ctx)`
- `sandbox.Wait()` -> `sandbox.Wait(ctx)`
- `sandbox.Tunnels(...)` -> `sandbox.Tunnels(ctx, ...)`
- `sandbox.SnapshotFilesystem(...)` -> `sandbox.SnapshotFilesystem(ctx, ...)`
- `sandbox.Poll()` -> `sandbox.Poll(ctx)`
- `sandbox.SetTags(...)` -> `sandbox.SetTags(ctx, ...)`
- `sandbox.GetTags()` -> `sandbox.GetTags(ctx)`

### Secret

- `modal.SecretFromName(ctx, ..., &modal.SecretFromNameOptions{...})` -> `client.Secrets.FromName(ctx, ..., &modal.SecretFromNameParams{...})`
- `modal.SecretFromMap(ctx, ..., &modal.SecretFromMapOptions{...})` -> `client.Secrets.FromMap(ctx, ..., &modal.SecretFromMapParams{...})`

### Volume

- `modal.VolumeFromName(ctx, ..., &modal.VolumeFromNameOptions{...})` -> `client.Volumes.FromName(ctx, ..., &modal.VolumeFromNameParams{...})`
- `modal.VolumeEphemeral(ctx, &modal.EphemeralOptions{...})` -> `client.Volumes.Ephemeral(ctx, &modal.VolumeEphemeralParams{...})`

### Parameter Type Renames

- `ClientOptions` -> `ClientParams`
- `CloudBucketMountOptions` -> `CloudBucketMountParams`
- `ClsBatchingOptions` -> `ClsWithBatchingParams`
- `ClsConcurrencyOptions` -> `ClsWithConcurrencyParams`
- `ClsOptions` -> `ClsWithOptionsParams`
- `DeleteOptions` -> specific `*DeleteParams` types: `QueueDeleteParams`
- `EphemeralOptions` -> specific `*EphemeralParams` types: `QueueEphemeralParams`, `VolumeEphemeralParams`
- `ExecOptions` -> `SandboxExecParams`
- `FunctionCallCancelOptions` -> `FunctionCallCancelParams`
- `FunctionCallGetOptions` -> `FunctionCallGetParams`
- `ImageDeleteOptions` -> `ImageDeleteParams`
- `ImageDockerfileCommandsOptions` -> `ImageDockerfileCommandsParams`
- `ImageFromRegistryOptions` -> `ImageFromRegistryParams`
- `LookupOptions` -> specific `*FromNameParams` types: `AppFromNameParams`, `ClsFromNameParams`, `FunctionFromNameParams`, `QueueFromNameParams`
- `ProxyFromNameOptions` -> `ProxyFromNameParams`
- `QueueClearOptions` -> `QueueClearParams`
- `QueueGetOptions` -> `QueueGetParams` and `QueueGetManyParams`
- `QueueIterateOptions` -> `QueueIterateParams`
- `QueueLenOptions` -> `QueueLenParams`
- `QueuePutOptions` -> `QueuePutParams` and `QueuePutManyParams`
- `RetriesOptions` -> `RetriesParams`
- `SandboxFromNameOptions` -> `SandboxFromNameParams`
- `SandboxListOptions` -> `SandboxListParams`
- `SandboxOptions` -> `SandboxCreateParams`
- `SecretFromMapOptions` -> `SecretFromMapParams`
- `SecretFromNameOptions` -> `SecretFromNameParams`
- `UpdateAutoscalerOptions` -> `FunctionUpdateAutoscalerParams`
- `VolumeFromNameOptions` -> `VolumeFromNameParams`
