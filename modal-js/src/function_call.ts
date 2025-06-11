// Manage existing Function Calls (look-ups, polling for output, cancellation).

import { client } from "./client";
import type { FunctionOutputPoller } from "./function";

/** Options for `FunctionCall.get()`. */
export type FunctionCallGetOptions = {
  timeout?: number; // in milliseconds
};

/** Options for `FunctionCall.cancel()`. */
export type FunctionCallCancelOptions = {
  terminateContainers?: boolean;
};

/**
 * Represents a Modal FunctionCall. Function Calls are Function invocations with
 * a given input. They can be consumed asynchronously (see `get()`) or cancelled
 * (see `cancel()`).
 */
export class FunctionCall {
  readonly functionOutputPoller: FunctionOutputPoller;

  /** @ignore */
  constructor(functionOutputPoller: FunctionOutputPoller) {
    this.functionOutputPoller = functionOutputPoller;
  }

  /** Create a new function call from ID. */
  static fromPoller(functionOutputPoller: FunctionOutputPoller): FunctionCall {
    return new FunctionCall(functionOutputPoller);
  }

  /** Get the result of a function call, optionally waiting with a timeout. */
  async get(options: FunctionCallGetOptions = {}): Promise<any> {
    const timeout = options.timeout;
    return await this.functionOutputPoller.poll(timeout);
  }

  /** Cancel a running function call. */
  async cancel(options: FunctionCallCancelOptions = {}) {
    await client.functionCallCancel({
      functionCallId: this.functionOutputPoller.functionCallId,
      terminateContainers: options.terminateContainers,
    });
  }
}
