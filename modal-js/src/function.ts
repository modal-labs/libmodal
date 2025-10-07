// Function calls and invocations, to be used with Modal Functions.

import { createHash } from "node:crypto";

import {
  DataFormat,
  FunctionCallInvocationType,
  FunctionHandleMetadata,
  FunctionInput,
} from "../proto/modal_proto/api";
import { getDefaultClient, ModalGrpcClient, type ModalClient } from "./client";
import { FunctionCall } from "./function_call";
import { InternalFailure, InvalidError, NotFoundError } from "./errors";
import { cborEncode } from "./serialization";
import { ClientError, Status } from "nice-grpc";
import {
  ControlPlaneInvocation,
  InputPlaneInvocation,
  Invocation,
} from "./invocation";

// From: modal/_utils/blob_utils.py
const maxObjectSizeBytes = 2 * 1024 * 1024; // 2 MiB

// From: client/modal/_functions.py
const maxSystemRetries = 8;

/** Optional parameters for `client.functions.fromName()`. */
export type FunctionFromNameParams = {
  environment?: string;
  createIfMissing?: boolean;
};

/**
 * Service for managing Functions.
 */
export class FunctionService {
  readonly #client: ModalClient;
  constructor(client: ModalClient) {
    this.#client = client;
  }

  /**
   * Reference a Function by its name in an App.
   */
  async fromName(
    appName: string,
    name: string,
    params: FunctionFromNameParams = {},
  ): Promise<Function_> {
    try {
      const resp = await this.#client.cpClient.functionGet({
        appName,
        objectTag: name,
        environmentName: this.#client.environmentName(params.environment),
      });
      return new Function_(
        this.#client,
        resp.functionId,
        undefined,
        resp.handleMetadata,
      );
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.NOT_FOUND)
        throw new NotFoundError(`Function '${appName}/${name}' not found`);
      throw err;
    }
  }
}

/** Simple data structure storing stats for a running Function. */
export interface FunctionStats {
  backlog: number;
  numTotalRunners: number;
}

/** Optional parameters for `Function_.updateAutoscaler()`. */
export interface FunctionUpdateAutoscalerParams {
  minContainers?: number;
  maxContainers?: number;
  bufferContainers?: number;
  scaledownWindow?: number;
}

/** Represents a deployed Modal Function, which can be invoked remotely. */
export class Function_ {
  readonly functionId: string;
  readonly methodName?: string;
  #client: ModalClient;
  #handleMetadata?: FunctionHandleMetadata;

  /** @ignore */
  constructor(
    client: ModalClient,
    functionId: string,
    methodName?: string,
    functionHandleMetadata?: FunctionHandleMetadata,
  ) {
    this.functionId = functionId;
    this.methodName = methodName;

    this.#client = client;
    this.#handleMetadata = functionHandleMetadata;
  }

  /**
   * @deprecated Use `client.functions.fromName()` instead.
   */
  static async lookup(
    appName: string,
    name: string,
    params: FunctionFromNameParams = {},
  ): Promise<Function_> {
    return await getDefaultClient().functions.fromName(appName, name, params);
  }

  // Execute a single input into a remote Function.
  async remote(
    args: any[] = [],
    kwargs: Record<string, any> = {},
  ): Promise<any> {
    const input = await this.#createInput(args, kwargs);
    const invocation = await this.#createRemoteInvocation(input);
    // TODO(ryan): Add tests for retries.
    let retryCount = 0;
    while (true) {
      try {
        return await invocation.awaitOutput();
      } catch (err) {
        if (err instanceof InternalFailure && retryCount <= maxSystemRetries) {
          await invocation.retry(retryCount);
          retryCount++;
        } else {
          throw err;
        }
      }
    }
  }

  async #createRemoteInvocation(input: FunctionInput): Promise<Invocation> {
    if (this.#handleMetadata?.inputPlaneUrl) {
      return await InputPlaneInvocation.create(
        this.#client,
        this.#handleMetadata.inputPlaneUrl,
        this.functionId,
        input,
      );
    }

    return await ControlPlaneInvocation.create(
      this.#client,
      this.functionId,
      input,
      FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_SYNC,
    );
  }

  // Spawn a single input into a remote Function.
  async spawn(
    args: any[] = [],
    kwargs: Record<string, any> = {},
  ): Promise<FunctionCall> {
    const input = await this.#createInput(args, kwargs);
    const invocation = await ControlPlaneInvocation.create(
      this.#client,
      this.functionId,
      input,
      FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_ASYNC,
    );
    return new FunctionCall(this.#client, invocation.functionCallId);
  }

  // Returns statistics about the Function.
  async getCurrentStats(): Promise<FunctionStats> {
    const resp = await this.#client.cpClient.functionGetCurrentStats(
      { functionId: this.functionId },
      { timeout: 10000 },
    );
    return {
      backlog: resp.backlog,
      numTotalRunners: resp.numTotalTasks,
    };
  }

  // Overrides the current autoscaler behavior for this Function.
  async updateAutoscaler(
    params: FunctionUpdateAutoscalerParams,
  ): Promise<void> {
    await this.#client.cpClient.functionUpdateSchedulingParams({
      functionId: this.functionId,
      warmPoolSizeOverride: 0, // Deprecated field, always set to 0
      settings: {
        minContainers: params.minContainers,
        maxContainers: params.maxContainers,
        bufferContainers: params.bufferContainers,
        scaledownWindow: params.scaledownWindow,
      },
    });
  }

  /**
   * URL of a Function running as a web endpoint.
   * @returns The web URL if this Function is a web endpoint, otherwise undefined
   */
  async getWebUrl(): Promise<string | undefined> {
    return this.#handleMetadata?.webUrl || undefined;
  }

  async #createInput(
    args: any[] = [],
    kwargs: Record<string, any> = {},
  ): Promise<FunctionInput> {
    const supported_input_formats = this.#handleMetadata?.supportedInputFormats
      ?.length
      ? this.#handleMetadata.supportedInputFormats
      : [DataFormat.DATA_FORMAT_PICKLE];
    if (!supported_input_formats.includes(DataFormat.DATA_FORMAT_CBOR)) {
      // the remote function isn't cbor compatible for inputs
      // so we can error early
      throw new InvalidError(
        "the deployed Function does not support libmodal - please redeploy it using Modal Python SDK version >= 1.2",
      );
    }
    const payload = cborEncode([args, kwargs]);

    let argsBlobId: string | undefined = undefined;
    if (payload.length > maxObjectSizeBytes) {
      argsBlobId = await blobUpload(this.#client.cpClient, payload);
    }

    // Single input sync invocation
    return {
      args: argsBlobId ? undefined : payload,
      argsBlobId,
      dataFormat: DataFormat.DATA_FORMAT_CBOR,
      methodName: this.methodName,
      finalInput: false, // This field isn't specified in the Python client, so it defaults to false.
    };
  }
}

async function blobUpload(
  cpClient: ModalGrpcClient,
  data: Uint8Array,
): Promise<string> {
  const contentMd5 = createHash("md5").update(data).digest("base64");
  const contentSha256 = createHash("sha256").update(data).digest("base64");
  const resp = await cpClient.blobCreate({
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
