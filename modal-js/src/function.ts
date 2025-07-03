// Function calls and invocations, to be used with Modal Functions.

import { createHash } from "node:crypto";

import {
  DataFormat,
  FunctionCallInvocationType,
  FunctionInput,
} from "../proto/modal_proto/api";
import type { LookupOptions } from "./app";
import { client } from "./client";
import { FunctionCall } from "./function_call";
import { environmentName } from "./config";
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

/** Represents a deployed Modal Function, which can be invoked remotely. */
export class Function_ {
  readonly functionId: string;
  readonly methodName: string | undefined;
  readonly inputPlaneUrl: string | undefined;

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
    if (this.inputPlaneUrl) {
      return await InputPlaneInvocation.create(
        this.inputPlaneUrl,
        this.functionId,
        input,
      );
    }

    return await ControlPlaneInvocation.create(
      this.functionId,
      input,
      FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_SYNC,
    );
  }

  // Spawn a single input into a remote function.
  async spawn(
    args: any[] = [],
    kwargs: Record<string, any> = {},
  ): Promise<FunctionCall> {
    const input = await this.#createInput(args, kwargs);
    const invocation = await ControlPlaneInvocation.create(
      this.functionId,
      input,
      FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_ASYNC,
    );
    return new FunctionCall(invocation.functionCallId);
  }

  async #createInput(
    args: any[] = [],
    kwargs: Record<string, any> = {},
  ): Promise<FunctionInput> {
    const payload = dumps([args, kwargs]);

    let argsBlobId: string | undefined = undefined;
    if (payload.length > maxObjectSizeBytes) {
      argsBlobId = await blobUpload(payload);
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
