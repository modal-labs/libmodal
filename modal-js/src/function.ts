// Function calls and invocations, to be used with Modal Functions.

import { createHash } from "node:crypto";

import {
  DataFormat,
  DeploymentNamespace,
  FunctionCallInvocationType,
  FunctionCallType,
  FunctionGetOutputsResponse,
  GeneratorDone,
  GenericResult,
  GenericResult_GenericStatus,
} from "../proto/modal_proto/api";
import type { LookupOptions } from "./app";
import { client } from "./client";
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

  /** @ignore */
  constructor(functionId: string, methodName?: string) {
    this.functionId = functionId;
    this.methodName = methodName;
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
      return new Function_(resp.functionId);
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
    const functionCallId = await this.#execFunctionCall(
      args,
      kwargs,
      FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_SYNC,
    );
    return await pollFunctionOutput(functionCallId);
  }

  // Spawn a single input into a remote function.
  async spawn(
    args: any[] = [],
    kwargs: Record<string, any> = {},
  ): Promise<FunctionCall> {
    const functionCallId = await this.#execFunctionCall(
      args,
      kwargs,
      FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_SYNC,
    );
    return new FunctionCall(functionCallId);
  }

  async #execFunctionCall(
    args: any[] = [],
    kwargs: Record<string, any> = {},
    invocationType: FunctionCallInvocationType = FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_SYNC,
  ): Promise<string> {
    const payload = dumps([args, kwargs]);

    let argsBlobId: string | undefined = undefined;
    if (payload.length > maxObjectSizeBytes) {
      argsBlobId = await blobUpload(payload);
    }

    // Single input sync invocation
    const functionMapResponse = await client.functionMap({
      functionId: this.functionId,
      functionCallType: FunctionCallType.FUNCTION_CALL_TYPE_UNARY,
      functionCallInvocationType: invocationType,
      pipelinedInputs: [
        {
          idx: 0,
          input: {
            args: argsBlobId ? undefined : payload,
            argsBlobId,
            dataFormat: DataFormat.DATA_FORMAT_PICKLE,
            methodName: this.methodName,
          },
        },
      ],
    });

    return functionMapResponse.functionCallId;
  }
}

export async function pollFunctionOutput(
  functionCallId: string,
  timeout?: number, // in milliseconds
): Promise<any> {
  const startTime = Date.now();
  let pollTimeout = outputsTimeout;
  if (timeout !== undefined) {
    pollTimeout = Math.min(timeout, outputsTimeout);
  }

  while (true) {
    let response: FunctionGetOutputsResponse;
    try {
      response = await client.functionGetOutputs({
        functionCallId: functionCallId,
        maxValues: 1,
        timeout: pollTimeout / 1000, // Backend needs seconds
        lastEntryId: "0-0",
        clearOnSuccess: true,
        requestedAt: timeNowSeconds(),
      });
    } catch (err) {
      throw new Error(`FunctionGetOutputs failed: ${err}`);
    }

    const outputs = response.outputs;
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
