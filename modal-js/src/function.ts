// Function calls and invocations, to be used with Modal Functions.

import { createHash } from "node:crypto";

import type {
  FunctionGetOutputsItem,
  FunctionPutInputsItem,
  GenericResult,
  ModalClientClient,
} from "../proto/modal_proto/api";
import {
  DataFormat,
  DeploymentNamespace,
  FunctionCallInvocationType,
  FunctionCallType,
  GeneratorDone,
  GenericResult_GenericStatus,
} from "../proto/modal_proto/api";
import type { LookupOptions } from "./app";
import { client, getOrCreateClient } from "./client";

import { FunctionCall } from "./function_call";
import { environmentName } from "./config";
import {
  InternalFailure,
  NotFoundError,
  RemoteError,
  FunctionTimeoutError,
} from "./errors";
import { dumps, loads } from "./pickle";
import { ClientError, Status } from "nice-grpc";

// From: modal/_utils/blob_utils.py
const maxObjectSizeBytes = 2 * 1024 * 1024; // 2 MiB

// From: modal-client/modal/_utils/function_utils.py
const outputsTimeout = 55 * 1000;

function timeNowSeconds() {
  return Date.now() / 1e3;
}

/** Represents a deployed Modal Function, which can be invoked remotely. */
export class Function_ {
  readonly functionId: string;
  readonly methodName: string | undefined;
  private readonly inputPlaneUrl: string | undefined;

  /** @ignore */
  constructor(functionId: string, methodName?: string, inputPlaneUrl?: string) {
    this.functionId = functionId;
    this.methodName = methodName;
    this.inputPlaneUrl = inputPlaneUrl;
  }

  static async lookup(
    appName: string,
    name: string,
    options: LookupOptions = {},
  ): Promise<Function_> {
    try {
      const resp = await client.functionGet({
        appName,
        objectTag: name,
        namespace: DeploymentNamespace.DEPLOYMENT_NAMESPACE_WORKSPACE,
        environmentName: environmentName(options.environment),
      });
      return new Function_(
        resp.functionId,
        undefined,
        resp.handleMetadata?.inputPlaneUrl,
      );
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.NOT_FOUND)
        throw new NotFoundError(`Function '${appName}/${name}' not found`);
      throw err;
    }
  }

  // Execute a single input into a remote Function.
  async remote(
    args: any[] = [],
    kwargs: Record<string, any> = {},
  ): Promise<any> {
    const functionOutputPoller = await this.#execFunctionCall(
      args,
      kwargs,
      FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_SYNC,
    );
    return await functionOutputPoller.poll();
  }

  // Spawn a single input into a remote function.
  async spawn(
    args: any[] = [],
    kwargs: Record<string, any> = {},
  ): Promise<FunctionCall> {
    const functionOutputPoller = await this.#execFunctionCall(
      args,
      kwargs,
      FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_SYNC,
    );
    return FunctionCall.fromPoller(functionOutputPoller);
  }

  async #execFunctionCall(
    args: any[] = [],
    kwargs: Record<string, any> = {},
    invocationType: FunctionCallInvocationType = FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_SYNC,
  ): Promise<FunctionOutputPoller> {
    const payload = dumps([args, kwargs]);

    let argsBlobId: string | undefined = undefined;
    if (payload.length > maxObjectSizeBytes) {
      argsBlobId = await blobUpload(payload);
    }

    // Single input sync invocation
    const functionInputs = [
      {
        idx: 0,
        input: {
          args: argsBlobId ? undefined : payload,
          argsBlobId,
          dataFormat: DataFormat.DATA_FORMAT_PICKLE,
          methodName: this.methodName,
          finalInput: false, // This field isn't specified in the Python client, so it defaults to false.
        },
      },
    ];

    if (this.inputPlaneUrl !== undefined) {
      return this.remoteInputPlane(functionInputs);
    }
    return this.remoteControlPlane(functionInputs, invocationType);
  }

  private async remoteInputPlane(
    functionInputs: FunctionPutInputsItem[],
  ): Promise<any> {
    if (!this.inputPlaneUrl) {
      throw new Error("Input plane URL is not set");
    }
    const client = getOrCreateClient(this.inputPlaneUrl);

    const attemptStartResponse = await client.attemptStart({
      functionId: this.functionId,
      input: functionInputs[0],
    });
    return FunctionOutputPoller.fromAttemptToken(
      client,
      attemptStartResponse.attemptToken,
    );
  }

  private async remoteControlPlane(
    functionInputs: FunctionPutInputsItem[],
    invocationType: FunctionCallInvocationType = FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_SYNC,
  ): Promise<any> {
    const functionMapResponse = await client.functionMap({
      functionId: this.functionId,
      functionCallType: FunctionCallType.FUNCTION_CALL_TYPE_UNARY,
      functionCallInvocationType: invocationType,
      pipelinedInputs: functionInputs,
    });

    return FunctionOutputPoller.fromFunctionCallId(
      client,
      functionMapResponse.functionCallId,
    );
  }
}

/**
 * The `FunctionOutputPoller` class is responsible for polling the outputs of a remote function call.
 * When an instance is created using one of the static factory methods, it is configured to poll either
 * the input plane or the control plane.
 */
export class FunctionOutputPoller {
  functionCallId?: string;
  attemptToken?: string;
  client: ModalClientClient;

  static fromFunctionCallId(
    client: ModalClientClient,
    functionCallId: string,
  ): FunctionOutputPoller {
    return new FunctionOutputPoller(client, functionCallId, undefined);
  }

  static fromAttemptToken(
    client: ModalClientClient,
    attemptToken: string,
  ): FunctionOutputPoller {
    return new FunctionOutputPoller(client, undefined, attemptToken);
  }

  private constructor(
    client: ModalClientClient,
    functionCallId?: string,
    attemptToken?: string,
  ) {
    if (!functionCallId && !attemptToken) {
      throw new Error("Either functionCallId or attemptToken must be provided");
    }
    this.client = client;
    this.functionCallId = functionCallId;
    this.attemptToken = attemptToken;
  }

  async poll(
    timeout?: number, // in milliseconds
  ): Promise<any> {
    const startTime = Date.now();
    let pollTimeout = outputsTimeout;
    if (timeout !== undefined) {
      pollTimeout = Math.min(timeout, outputsTimeout);
    }

    while (true) {
      const outputs = this.attemptToken
        ? await this.#pollInputPlane(pollTimeout)
        : await this.#pollControlPlane(pollTimeout);
      if (outputs.length > 0) {
        return await processResult(outputs[0].result, outputs[0].dataFormat);
      }

      if (timeout !== undefined) {
        const remainingTime = timeout - (Date.now() - startTime);
        if (remainingTime <= 0) {
          const message = `Timeout exceeded: ${(timeout / 1000).toFixed(1)}s`;
          throw new FunctionTimeoutError(message);
        }
        pollTimeout = Math.min(outputsTimeout, remainingTime);
      }
    }
  }

  async #pollControlPlane(
    timeout: number, // in milliseconds
  ): Promise<FunctionGetOutputsItem[]> {
    try {
      const response = await this.client.functionGetOutputs({
        functionCallId: this.functionCallId,
        maxValues: 1,
        timeout: timeout / 1000, // Backend needs seconds
        lastEntryId: "0-0",
        clearOnSuccess: true,
        requestedAt: timeNowSeconds(),
      });
      return response.outputs;
    } catch (err) {
      throw new Error(`FunctionGetOutputs failed: ${err}`);
    }
  }

  async #pollInputPlane(
    timeout: number, // in milliseconds
  ): Promise<FunctionGetOutputsItem[]> {
    try {
      const response = await this.client.attemptAwait({
        attemptToken: this.attemptToken,
        requestedAt: timeNowSeconds(),
        timeoutSecs: timeout / 1000, // Convert to seconds
      });
      return response.output ? [response.output] : [];
    } catch (err) {
      throw new Error(`AttemptAwait failed: ${err}`);
    }
  }
}

async function processResult(
  result: GenericResult | undefined,
  dataFormat: DataFormat,
): Promise<unknown> {
  if (!result) {
    throw new Error("Received null result from invocation");
  }

  let data = new Uint8Array();
  if (result.data !== undefined) {
    data = result.data;
  } else if (result.dataBlobId) {
    data = await blobDownload(result.dataBlobId);
  }

  switch (result.status) {
    case GenericResult_GenericStatus.GENERIC_STATUS_TIMEOUT:
      throw new FunctionTimeoutError(`Timeout: ${result.exception}`);
    case GenericResult_GenericStatus.GENERIC_STATUS_INTERNAL_FAILURE:
      throw new InternalFailure(`Internal failure: ${result.exception}`);
    case GenericResult_GenericStatus.GENERIC_STATUS_SUCCESS:
      // Proceed to deserialize the data.
      break;
    default:
      // Handle other statuses, e.g., remote error.
      throw new RemoteError(`Remote error: ${result.exception}`);
  }

  return deserializeDataFormat(data, dataFormat);
}

async function blobUpload(data: Uint8Array): Promise<string> {
  const contentMd5 = createHash("md5").update(data).digest("base64");
  const contentSha256 = createHash("sha256").update(data).digest("base64");
  const resp = await client.blobCreate({
    contentMd5,
    contentSha256Base64: contentSha256,
    contentLength: data.length,
  });
  if (resp.multipart) {
    throw new Error(
      "Function input size exceeds multipart upload threshold, unsupported by this SDK version",
    );
  } else if (resp.uploadUrl) {
    const uploadResp = await fetch(resp.uploadUrl, {
      method: "PUT",
      headers: {
        "Content-Type": "application/octet-stream",
        "Content-MD5": contentMd5,
      },
      body: data,
    });
    if (uploadResp.status < 200 || uploadResp.status >= 300) {
      throw new Error(`Failed blob upload: ${uploadResp.statusText}`);
    }
    // Skip client-side ETag header validation for now (MD5 checksum).
    return resp.blobId;
  } else {
    throw new Error("Missing upload URL in BlobCreate response");
  }
}

async function blobDownload(blobId: string): Promise<Uint8Array> {
  const resp = await client.blobGet({ blobId });
  const s3resp = await fetch(resp.downloadUrl);
  if (!s3resp.ok) {
    throw new Error(`Failed to download blob: ${s3resp.statusText}`);
  }
  const buf = await s3resp.arrayBuffer();
  return new Uint8Array(buf);
}

function deserializeDataFormat(
  data: Uint8Array | undefined,
  dataFormat: DataFormat,
): unknown {
  if (!data) {
    return null; // No data to deserialize.
  }

  switch (dataFormat) {
    case DataFormat.DATA_FORMAT_PICKLE:
      return loads(data);
    case DataFormat.DATA_FORMAT_ASGI:
      throw new Error("ASGI data format is not supported in Go");
    case DataFormat.DATA_FORMAT_GENERATOR_DONE:
      return GeneratorDone.decode(data);
    default:
      throw new Error(`Unsupported data format: ${dataFormat}`);
  }
}
