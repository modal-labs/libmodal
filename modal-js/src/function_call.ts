// Manage existing Function Calls (look-ups, polling for output, cancellation).

import { client } from "./client";
import { pollFunctionOutput, outputsTimeout } from "./function";
import { TimeoutError } from "./errors";

export type FunctionCallGetOptions = {
  timeout?: number; // in seconds
};

export type FunctionCallCancelOptions = {
  terminateContainers?: boolean;
};

/** Represents a Modal FunctionCall, Function Calls are
Function invocations with a given input. They can be consumed
asynchronously (see get()) or cancelled (see cancel()).
*/
export class FunctionCall {
  readonly functionCallId: string;

  constructor(functionCallId: string) {
    this.functionCallId = functionCallId;
  }

  // Get output for a FunctionCall ID.
  async get(options: FunctionCallGetOptions = {}): Promise<any> {
    const timeout = options.timeout || outputsTimeout;
    return await pollFunctionOutput(this.functionCallId, timeout);
  }

  // Cancel ongoing FunctionCall.
  async cancel(options: FunctionCallCancelOptions = {}) {
    await client.functionCallCancel({
      functionCallId: this.functionCallId,
      terminateContainers: options.terminateContainers,
    });
  }
}

async function functionCallFromId(
  functionCallId: string,
): Promise<FunctionCall> {
  return new FunctionCall(functionCallId);
}
