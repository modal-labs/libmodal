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

export class FunctionCall {
  readonly functionCallId: string;

  constructor(functionCallId: string) {
    this.functionCallId = functionCallId;
  }

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
