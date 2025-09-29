// Function calls and invocations, to be used with Modal Functions.

import { createHash } from "node:crypto";

import {
  DataFormat,
  FunctionCallInvocationType,
  FunctionInput,
} from "../proto/modal_proto/api";
import type { LookupOptions } from "./app";
import { getDefaultClient, ModalGrpcClient, type ModalClient } from "./client";
import { FunctionCall } from "./function_call";
import { InternalFailure, NotFoundError } from "./errors";
import { dumps } from "./pickle";
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
  async lookup(
    appName: string,
    name: string,
    options: LookupOptions = {},
  ): Promise<Function_> {
    try {
      const resp = await this.#client.cpClient.functionGet({
        appName,
        objectTag: name,
        environmentName: this.#client.environmentName(options.environment),
      });
      return new Function_(
        this.#client,
        resp.functionId,
        undefined,
        resp.handleMetadata?.inputPlaneUrl,
        resp.handleMetadata?.webUrl,
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

/** Options for overriding a Function's autoscaler behavior. */
export interface UpdateAutoscalerOptions {
  minContainers?: number;
  maxContainers?: number;
  bufferContainers?: number;
  scaledownWindow?: number;
}

/** Represents a deployed Modal Function, which can be invoked remotely. */
export class Function_ {
  readonly functionId: string;
  readonly methodName?: string;
  #inputPlaneUrl?: string;
  #webUrl?: string;
  #client: ModalClient;

  /** @ignore */
  constructor(
    client: ModalClient,
    functionId: string,
    methodName?: string,
    inputPlaneUrl?: string,
    webUrl?: string,
  ) {
    this.functionId = functionId;
    this.methodName = methodName;
    this.#inputPlaneUrl = inputPlaneUrl;
    this.#webUrl = webUrl;
    this.#client = client;
  }

  /**
   * @deprecated Use `client.functions.lookup()` instead.
   */
  static async lookup(
    appName: string,
    name: string,
    options: LookupOptions = {},
  ): Promise<Function_> {
    return await getDefaultClient().functions.lookup(appName, name, options);
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
    if (this.#inputPlaneUrl) {
      return await InputPlaneInvocation.create(
        this.#client,
        this.#inputPlaneUrl,
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
  async updateAutoscaler(options: UpdateAutoscalerOptions): Promise<void> {
    await this.#client.cpClient.functionUpdateSchedulingParams({
      functionId: this.functionId,
      warmPoolSizeOverride: 0, // Deprecated field, always set to 0
      settings: {
        minContainers: options.minContainers,
        maxContainers: options.maxContainers,
        bufferContainers: options.bufferContainers,
        scaledownWindow: options.scaledownWindow,
      },
    });
  }

  /**
   * URL of a Function running as a web endpoint.
   * @returns The web URL if this Function is a web endpoint, otherwise undefined
   */
  async getWebUrl(): Promise<string | undefined> {
    return this.#webUrl || undefined;
  }

  async #createInput(
    args: any[] = [],
    kwargs: Record<string, any> = {},
  ): Promise<FunctionInput> {
    const payload = dumps([args, kwargs]);

    let argsBlobId: string | undefined = undefined;
    if (payload.length > maxObjectSizeBytes) {
      argsBlobId = await blobUpload(this.#client.cpClient, payload);
    }

    // Single input sync invocation
    return {
      args: argsBlobId ? undefined : payload,
      argsBlobId,
      dataFormat: DataFormat.DATA_FORMAT_PICKLE,
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
