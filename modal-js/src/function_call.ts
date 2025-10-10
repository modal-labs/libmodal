// Manage existing Function Calls (look-ups, polling for output, cancellation).

import { getDefaultClient, type ModalClient } from "./client";
import { ControlPlaneInvocation } from "./invocation";

/**
 * Service for managing FunctionCalls.
 */
export class FunctionCallService {
  readonly #client: ModalClient;
  constructor(client: ModalClient) {
    this.#client = client;
  }

  /**
   * Create a new Function call from ID.
   */
  async fromId(functionCallId: string): Promise<FunctionCall> {
    return new FunctionCall(this.#client, functionCallId);
  }
}

/** Optional parameters for `FunctionCall.get()`. */
export type FunctionCallGetParams = {
  timeout?: number; // in milliseconds
};

/** Optional parameters for `FunctionCall.cancel()`. */
export type FunctionCallCancelParams = {
  terminateContainers?: boolean;
};

/**
 * Represents a Modal FunctionCall. Function Calls are Function invocations with
 * a given input. They can be consumed asynchronously (see `get()`) or cancelled
 * (see `cancel()`).
 */
export class FunctionCall {
  readonly functionCallId: string;
  #client?: ModalClient;

  /** @ignore */
  constructor(client: ModalClient | undefined, functionCallId: string) {
    this.#client = client;
    this.functionCallId = functionCallId;
  }

  /**
   * @deprecated Use `client.functionCalls.fromId()` instead.
   */
  static fromId(functionCallId: string): FunctionCall {
    return new FunctionCall(undefined, functionCallId);
  }

  /** Get the result of a Function call, optionally waiting with a timeout. */
  async get(params: FunctionCallGetParams = {}): Promise<any> {
    const timeout = params.timeout;
    const invocation = ControlPlaneInvocation.fromFunctionCallId(
      this.#client || getDefaultClient(),
      this.functionCallId,
    );
    return invocation.awaitOutput(timeout);
  }

  /** Cancel a running Function call. */
  async cancel(params: FunctionCallCancelParams = {}) {
    const cpClient = this.#client?.cpClient || getDefaultClient().cpClient;

    await cpClient.functionCallCancel({
      functionCallId: this.functionCallId,
      terminateContainers: params.terminateContainers,
    });
  }
}
