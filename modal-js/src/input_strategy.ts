import {
  FunctionCallInvocationType,
  FunctionCallType,
  FunctionGetOutputsItem,
  FunctionPutInputsItem,
  FunctionRetryInputsItem,
  ModalClientClient,
} from "../proto/modal_proto/api";
import { client } from "./client";

/**
 * This abstraction exists so that we can easily send inputs to either the control plane or the input plane.
 * Function definitions that include an option like this `experimental_options={"input_plane_region": "us-west"}`
 * are sent to the input plane, otherwise they are sent to the control plane.
 *
 * Once the input plane is no longer experimental, and all user traffic has been moved to it, we can remove
 * this abstraction.
 */
export interface InputStrategy {
  attemptStart(): Promise<void>;
  attemptRetry(): Promise<void>;
  attemptAwait(timeout_seconds: number): Promise<FunctionGetOutputsItem[]>;
}

/**
 * Implementation of InputStrategy which sends inputs to the control plane.
 */
export class ControlPlaneStrategy implements InputStrategy {
  private readonly client: ModalClientClient;
  private readonly functionId: string;
  private readonly input: FunctionPutInputsItem;
  private readonly invocationType: FunctionCallInvocationType;
  private functionCallJwt: string | undefined;
  private inputJwt: string | undefined;
  functionCallId: string | undefined;

  constructor(
    client: ModalClientClient,
    functionId: string,
    input: FunctionPutInputsItem,
    invocationType: FunctionCallInvocationType,
  ) {
    this.client = client;
    this.functionId = functionId;
    this.input = input;
    this.invocationType = invocationType;
  }

  async attemptStart(): Promise<void> {
    const functionMapResponse = await this.client.functionMap({
      functionId: this.functionId,
      functionCallType: FunctionCallType.FUNCTION_CALL_TYPE_UNARY,
      functionCallInvocationType: this.invocationType,
      pipelinedInputs: [this.input],
    });

    this.functionCallId = functionMapResponse.functionCallId;
    this.functionCallJwt = functionMapResponse.functionCallJwt;
    this.inputJwt = functionMapResponse.pipelinedInputs[0].inputJwt;
  }

  async attemptRetry(): Promise<void> {
    const retryItem: FunctionRetryInputsItem = {
      inputJwt: this.inputJwt!,
      input: this.input.input!,
      retryCount: 0,
    };
    const functionRetryResponse = await client.functionRetryInputs({
      functionCallJwt: this.functionCallJwt,
      inputs: [retryItem],
    });
    this.inputJwt = functionRetryResponse.inputJwts[0];
  }

  async attemptAwait(timeoutMillis: number): Promise<FunctionGetOutputsItem[]> {
    try {
      const response = await this.client.functionGetOutputs({
        functionCallId: this.functionCallId,
        maxValues: 1,
        timeout: timeoutMillis / 1000,
        lastEntryId: "0-0",
        clearOnSuccess: true,
        requestedAt: timeNowSeconds(),
        inputJwts: [this.inputJwt!],
      });
      return response.outputs;
    } catch (err) {
      throw new Error(`FunctionGetOutputs failed: ${err}`);
    }
  }
}

/**
 * Implementation of InputStrategy which sends inputs to the input plane.
 */
export class InputPlaneStrategy implements InputStrategy {
  private readonly client: ModalClientClient;
  private readonly functionId: string;
  private readonly input: FunctionPutInputsItem;
  private attemptToken: string | undefined;

  constructor(
    client: ModalClientClient,
    functionId: string,
    input: FunctionPutInputsItem,
  ) {
    this.client = client;
    this.functionId = functionId;
    this.input = input;
  }

  async attemptStart(): Promise<void> {
    const attemptStartResponse = await this.client.attemptStart({
      functionId: this.functionId,
      input: this.input,
    });
    this.attemptToken = attemptStartResponse.attemptToken;
  }

  async attemptRetry(): Promise<void> {
    const attemptRetryResponse = await this.client.attemptRetry({
      functionId: this.functionId,
      input: this.input,
      attemptToken: this.attemptToken,
    });
    this.attemptToken = attemptRetryResponse.attemptToken;
  }

  async attemptAwait(timeoutMillis: number): Promise<FunctionGetOutputsItem[]> {
    try {
      const response = await this.client.attemptAwait({
        attemptToken: this.attemptToken,
        requestedAt: timeNowSeconds(),
        timeoutSecs: timeoutMillis / 1000,
      });
      return response.output ? [response.output] : [];
    } catch (err) {
      throw new Error(`AttemptAwait failed: ${err}`);
    }
  }
}

function timeNowSeconds() {
  return Date.now() / 1e3;
}
