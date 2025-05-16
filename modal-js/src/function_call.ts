// Manage existing Function Calls (look-ups, polling for output, cancellation).

import { client } from "./client";
import { pollFunctionOutput } from "./function";
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
    const timeout = options.timeout;

    if (!timeout) return await pollFunctionOutput(this.functionCallId);

    return new Promise(async (resolve, reject) => {
      const timer = setTimeout(
        () => reject(new TimeoutError(`timeout after ${timeout}s`)),
        timeout * 1_000,
      );

      await pollFunctionOutput(this.functionCallId)
        .then((result) => {
          clearTimeout(timer);
          resolve(result);
        })
        .catch((err) => {
          clearTimeout(timer);
          reject(err);
        });
    });
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
