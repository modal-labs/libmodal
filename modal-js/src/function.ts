// Function calls and invocations, to be used with Modal Functions.

import { createHash } from "node:crypto";

import type { GenericResult } from "../proto/modal_proto/api";
import {
  DataFormat,
  DeploymentNamespace,
  FunctionCallInvocationType,
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
import {
  ControlPlaneStrategy,
  InputPlaneStrategy,
  InputStrategy,
} from "./input_strategy";

// From: modal/_utils/blob_utils.py
const maxObjectSizeBytes = 2 * 1024 * 1024; // 2 MiB

// From: modal-client/modal/_utils/function_utils.py
export const outputsTimeoutMillis = 55 * 1000;

const maxSystemRetries = 8;

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
    // InputStrategy sends inputs to either the control plane or the input plane,
    // depending on how the function is configured.
    const inputStrategy = await this.#createInputStrategy(
      args,
      kwargs,
      FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_SYNC,
    );
    await inputStrategy.attemptStart();
    // TODO(ryan): Write tests for retry logic
    let retryCount = 0;
    while (true) {
      try {
        return await pollFunctionOutput(inputStrategy);
      } catch (err) {
        if (err instanceof InternalFailure && retryCount <= maxSystemRetries) {
          await inputStrategy.attemptRetry();
          retryCount++;
        } else {
          throw err;
        }
      }
    }
  }

  // Spawn a single input into a remote function.
  async spawn(
    args: any[] = [],
    kwargs: Record<string, any> = {},
  ): Promise<FunctionCall> {
    const inputStrategy = await this.#createInputStrategy(
      args,
      kwargs,
      FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_SYNC,
    );
    await inputStrategy.attemptStart();
    return FunctionCall.fromInputStrategy(
      inputStrategy as ControlPlaneStrategy,
    );
  }

  async #createInputStrategy(
    args: any[] = [],
    kwargs: Record<string, any> = {},
    invocationType: FunctionCallInvocationType = FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_SYNC,
  ): Promise<InputStrategy> {
    const payload = dumps([args, kwargs]);

    let argsBlobId: string | undefined = undefined;
    if (payload.length > maxObjectSizeBytes) {
      argsBlobId = await blobUpload(payload);
    }

    // Single input sync invocation
    const functionInput = {
      idx: 0,
      input: {
        args: argsBlobId ? undefined : payload,
        argsBlobId,
        dataFormat: DataFormat.DATA_FORMAT_PICKLE,
        methodName: this.methodName,
        finalInput: false, // This field isn't specified in the Python client, so it defaults to false.
      },
    };

    if (this.inputPlaneUrl === undefined) {
      return new ControlPlaneStrategy(
        client,
        this.functionId,
        functionInput,
        invocationType,
      );
    }

    // Input plane does not support ASYNC inputs
    if (
      invocationType !==
      FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_SYNC
    ) {
      throw new Error("Only SYNC invocations types are supported");
    }

    return new InputPlaneStrategy(
      getOrCreateClient(this.inputPlaneUrl),
      this.functionId,
      functionInput,
    );
  }
}

export async function pollFunctionOutput(
  inputStrategy: InputStrategy,
  timeoutMillis?: number,
): Promise<any> {
  const startTime = Date.now();
  let pollTimeout = outputsTimeoutMillis;
  if (timeoutMillis !== undefined) {
    pollTimeout = Math.min(timeoutMillis, outputsTimeoutMillis);
  }

  while (true) {
    const outputs = await inputStrategy.attemptAwait(pollTimeout);
    if (outputs.length > 0) {
      return await processResult(outputs[0].result, outputs[0].dataFormat);
    }

    if (timeoutMillis !== undefined) {
      const remainingTime = timeoutMillis - (Date.now() - startTime);
      if (remainingTime <= 0) {
        const message = `Timeout exceeded: ${(timeoutMillis / 1000).toFixed(1)}s`;
        throw new FunctionTimeoutError(message);
      }
      pollTimeout = Math.min(outputsTimeoutMillis, remainingTime);
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
