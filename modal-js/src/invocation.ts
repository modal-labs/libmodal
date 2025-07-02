import {
  DataFormat,
  FunctionCallInvocationType,
  FunctionCallType,
  FunctionGetOutputsItem,
  FunctionInput,
  FunctionPutInputsItem,
  FunctionRetryInputsItem,
  GeneratorDone,
  GenericResult,
  GenericResult_GenericStatus,
  ModalClientClient,
} from "../proto/modal_proto/api";
import { client, getOrCreateInputPlaneClient } from "./client";
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
    return await pollFunctionOutput(
      (timeoutMillis: number) => this.#getOutput(timeoutMillis),
      timeout,
    );
  }

  async #getOutput(
    timeoutMillis: number,
  ): Promise<FunctionGetOutputsItem | undefined> {
    const response = await client.functionGetOutputs({
      functionCallId: this.functionCallId,
      maxValues: 1,
      timeout: timeoutMillis / 1000, // Backend needs seconds
      lastEntryId: "0-0",
      clearOnSuccess: true,
      requestedAt: timeNowSeconds(),
    });
    return response.outputs ? response.outputs[0] : undefined;
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

/**
 * Implementation of Invocation which sends inputs to the input plane.
 */
export class InputPlaneInvocation implements Invocation {
  private readonly client: ModalClientClient;
  private readonly functionId: string;
  private readonly input: FunctionPutInputsItem;
  private attemptToken: string;

  constructor(
    client: ModalClientClient,
    functionId: string,
    input: FunctionPutInputsItem,
    attemptToken: string,
  ) {
    this.client = client;
    this.functionId = functionId;
    this.input = input;
    this.attemptToken = attemptToken;
  }

  static async create(
    inputPlaneUrl: string,
    functionId: string,
    input: FunctionInput,
  ) {
    const functionPutInputsItem = {
      idx: 0,
      input,
    };
    const client = getOrCreateInputPlaneClient(inputPlaneUrl);
    // Single input sync invocation
    const attemptStartResponse = await client.attemptStart({
      functionId,
      input: functionPutInputsItem,
    });
    return new InputPlaneInvocation(
      client,
      functionId,
      functionPutInputsItem,
      attemptStartResponse.attemptToken,
    );
  }

  async awaitOutput(timeout?: number): Promise<any> {
    return await pollFunctionOutput(
      (timeoutMillis: number) => this.#getOutput(timeoutMillis),
      timeout,
    );
  }

  async #getOutput(
    timeoutMillis: number,
  ): Promise<FunctionGetOutputsItem | undefined> {
    const response = await this.client.attemptAwait({
      attemptToken: this.attemptToken,
      requestedAt: timeNowSeconds(),
      timeoutSecs: timeoutMillis / 1000,
    });
    return response.output;
  }

  async retry(_retryCount: number): Promise<void> {
    const attemptRetryResponse = await this.client.attemptRetry({
      functionId: this.functionId,
      input: this.input,
      attemptToken: this.attemptToken,
    });
    this.attemptToken = attemptRetryResponse.attemptToken;
  }
}

function timeNowSeconds() {
  return Date.now() / 1e3;
}

/**
 * Signature of a function that fetches a single output using the given timeout. Used by `pollForOutputs` to fetch
 * from either the control plane or the input plane, depending on the implementation.
 */
type GetOutput = (
  timeoutMillis: number,
) => Promise<FunctionGetOutputsItem | undefined>;

/***
 * Repeatedly tries to fetch an output using the provided `getOutput` function, and the specified timeout value.
 * We use a timeout value of 55 seconds if the caller does not specify a timeout value, or if the specified timeout
 * value is greater than 55 seconds.
 */
async function pollFunctionOutput(
  getOutput: GetOutput,
  timeout?: number, // in milliseconds
): Promise<any> {
  const startTime = Date.now();
  let pollTimeout = outputsTimeout;
  if (timeout !== undefined) {
    pollTimeout = Math.min(timeout, outputsTimeout);
  }

  while (true) {
    const output = await getOutput(pollTimeout);
    if (output) {
      return await processResult(output.result, output.dataFormat);
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
