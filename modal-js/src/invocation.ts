import {
  DataFormat,
  FunctionCallInvocationType,
  FunctionCallType,
  FunctionGetOutputsResponse,
  FunctionInput,
  FunctionRetryInputsItem,
  GeneratorDone,
  GenericResult,
  GenericResult_GenericStatus,
} from "../proto/modal_proto/api";
import { client } from "./client";
import { FunctionTimeoutError, InternalFailure, RemoteError } from "./errors";
import { loads } from "./pickle";

// From: modal-client/modal/_utils/function_utils.py
const outputsTimeout = 55 * 1000;

/**
 * This abstraction exists so that we can easily send inputs to either the control plane or the input plane.
 * For the control plane, we call the FunctionMap, FunctionRetryInputs, and FunctionGetOutputs RPCs.
 * For the input plane, we call the AttemptStart, AttemptRetry, and AttemptAwait RPCs.
 * For now, we support just the control plane, and will add support for the input plane soon.
 */
export interface Invocation {
  awaitOutput(timeout?: number): Promise<any>;
  retry(retryCount: number): Promise<void>;
}

/**
 * Implementation of Invocation which sends inputs to the control plane.
 */
export class ControlPlaneInvocation implements Invocation {
  readonly functionCallId: string;
  private readonly input?: FunctionInput;
  private readonly functionCallJwt?: string;
  private inputJwt?: string;

  private constructor(
    functionCallId: string,
    input?: FunctionInput,
    functionCallJwt?: string,
    inputJwt?: string,
  ) {
    this.functionCallId = functionCallId;
    this.input = input;
    this.functionCallJwt = functionCallJwt;
    this.inputJwt = inputJwt;
  }

  static async create(
    functionId: string,
    input: FunctionInput,
    invocationType: FunctionCallInvocationType,
  ) {
    const functionPutInputsItem = {
      idx: 0,
      input,
    };

    const functionMapResponse = await client.functionMap({
      functionId,
      functionCallType: FunctionCallType.FUNCTION_CALL_TYPE_UNARY,
      functionCallInvocationType: invocationType,
      pipelinedInputs: [functionPutInputsItem],
    });

    return new ControlPlaneInvocation(
      functionMapResponse.functionCallId,
      input,
      functionMapResponse.functionCallJwt,
      functionMapResponse.pipelinedInputs[0].inputJwt,
    );
  }

  static fromFunctionCallId(functionCallId: string) {
    return new ControlPlaneInvocation(functionCallId);
  }

  async awaitOutput(timeout?: number): Promise<any> {
    return await pollFunctionOutput(this.functionCallId, timeout);
  }

  async retry(retryCount: number): Promise<void> {
    // we do not expect this to happen
    if (!this.input) {
      throw new Error("Cannot retry function invocation - input missing");
    }

    const retryItem: FunctionRetryInputsItem = {
      inputJwt: this.inputJwt!,
      input: this.input,
      retryCount,
    };

    const functionRetryResponse = await client.functionRetryInputs({
      functionCallJwt: this.functionCallJwt,
      inputs: [retryItem],
    });
    this.inputJwt = functionRetryResponse.inputJwts[0];
  }
}

function timeNowSeconds() {
  return Date.now() / 1e3;
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
        functionCallId,
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
