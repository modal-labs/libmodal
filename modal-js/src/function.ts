// Function calls and invocations, to be used with Modal Functions.

import {
  DataFormat,
  DeploymentNamespace,
  FunctionCallInvocationType,
  FunctionCallType,
  GeneratorDone,
  GenericResult,
  GenericResult_GenericStatus,
} from "../proto/modal_proto/api";
import { LookupOptions } from "./app";
import { client } from "./client";
import { environmentName } from "./config";
import { InternalFailure, RemoteError, TimeoutError } from "./errors";
import { dumps, loads } from "./pickle";

function timeNow() {
  return Date.now() / 1e3;
}

/** Represents a deployed Modal Function, which can be invoked remotely. */
export class Function_ {
  readonly functionId: string;

  constructor(functionId: string) {
    this.functionId = functionId;
  }

  static async lookup(
    appName: string,
    name: string,
    options: LookupOptions = {},
  ): Promise<Function_> {
    const resp = await client.functionGet({
      appName,
      objectTag: name,
      namespace: DeploymentNamespace.DEPLOYMENT_NAMESPACE_WORKSPACE,
      environmentName: environmentName(options.environment),
    });
    return new Function_(resp.functionId);
  }

  // Execute a single input into a remote Function.
  async remote(
    args: any[] = [],
    kwargs: Record<string, any> = {},
  ): Promise<any> {
    const payload = dumps([args, kwargs]);

    // Single input sync invocation
    const functionInputs = [
      {
        idx: 0,
        input: {
          args: payload,
          dataFormat: DataFormat.DATA_FORMAT_PICKLE,
        },
      },
    ];

    const functionMapResponse = await client.functionMap({
      functionId: this.functionId,
      functionCallType: FunctionCallType.FUNCTION_CALL_TYPE_UNARY,
      functionCallInvocationType:
        FunctionCallInvocationType.FUNCTION_CALL_INVOCATION_TYPE_SYNC,
      pipelinedInputs: functionInputs,
    });

    while (true) {
      const response = await client.functionGetOutputs({
        functionCallId: functionMapResponse.functionCallId,
        maxValues: 1,
        timeout: 55,
        lastEntryId: "0-0",
        clearOnSuccess: true,
        requestedAt: timeNow(),
      });

      const outputs = response.outputs;
      if (outputs.length > 0) {
        return await processResult(outputs[0].result, outputs[0].dataFormat);
      }
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
      throw new TimeoutError(`Timeout: ${result.exception}`);
    case GenericResult_GenericStatus.GENERIC_STATUS_INTERNAL_FAILURE:
      throw new InternalFailure(`Internal failure: ${result.exception}`);
    case GenericResult_GenericStatus.GENERIC_STATUS_SUCCESS:
      // Proceed to deserialize the data.
      break;
    default:
      // Handle other statuses, e.g., remote error.
      throw new RemoteError(`Remote error: ${result.exception}`);
  }

  return deserializeDataFormat(result.data, dataFormat);
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
