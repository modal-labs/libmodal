import { ClientError, Status } from "nice-grpc";
import {
  ClassParameterInfo_ParameterSerializationFormat,
  ClassParameterSet,
  ClassParameterSpec,
  ClassParameterValue,
  FunctionOptions,
  FunctionRetryPolicy,
  ParameterType,
  VolumeMount,
} from "../proto/modal_proto/api";
import type { LookupOptions } from "./app";
import { NotFoundError } from "./errors";
import { client } from "./client";
import { environmentName } from "./config";
import { Function_ } from "./function";
import { parseGpuConfig } from "./app";
import type { Secret } from "./secret";
import { mergeEnvAndSecrets } from "./secret";
import { Retries, parseRetries } from "./retries";
import type { Volume } from "./volume";

export type ClsOptions = {
  cpu?: number;
  memory?: number;
  gpu?: string;
  env?: Record<string, string>;
  secrets?: Secret[];
  volumes?: Record<string, Volume>;
  retries?: number | Retries;
  maxContainers?: number;
  bufferContainers?: number;
  scaledownWindow?: number; // in milliseconds
  timeout?: number; // in milliseconds
};

export type ClsConcurrencyOptions = {
  maxInputs: number;
  targetInputs?: number;
};

export type ClsBatchingOptions = {
  maxBatchSize: number;
  waitMs: number;
};

type ServiceOptions = ClsOptions & {
  maxConcurrentInputs?: number;
  targetConcurrentInputs?: number;
  batchMaxSize?: number;
  batchWaitMs?: number;
};

/** Represents a deployed Modal Cls. */
export class Cls {
  #serviceFunctionId: string;
  #schema: ClassParameterSpec[];
  #methodNames: string[];
  #inputPlaneUrl?: string;
  #options?: ServiceOptions;

  /** @ignore */
  constructor(
    serviceFunctionId: string,
    schema: ClassParameterSpec[],
    methodNames: string[],
    inputPlaneUrl?: string,
    options?: ServiceOptions,
  ) {
    this.#serviceFunctionId = serviceFunctionId;
    this.#schema = schema;
    this.#methodNames = methodNames;
    this.#inputPlaneUrl = inputPlaneUrl;
    this.#options = options;
  }

  static async lookup(
    appName: string,
    name: string,
    options: LookupOptions = {},
  ): Promise<Cls> {
    try {
      const serviceFunctionName = `${name}.*`;
      const serviceFunction = await client.functionGet({
        appName,
        objectTag: serviceFunctionName,
        environmentName: environmentName(options.environment),
      });

      const parameterInfo = serviceFunction.handleMetadata?.classParameterInfo;
      const schema = parameterInfo?.schema ?? [];
      if (
        schema.length > 0 &&
        parameterInfo?.format !==
          ClassParameterInfo_ParameterSerializationFormat.PARAM_SERIALIZATION_FORMAT_PROTO
      ) {
        throw new Error(
          `Unsupported parameter format: ${parameterInfo?.format}`,
        );
      }

      let methodNames: string[];
      if (serviceFunction.handleMetadata?.methodHandleMetadata) {
        methodNames = Object.keys(
          serviceFunction.handleMetadata.methodHandleMetadata,
        );
      } else {
        // Legacy approach not supported
        throw new Error(
          "Cls requires Modal deployments using client v0.67 or later.",
        );
      }
      return new Cls(
        serviceFunction.functionId,
        schema,
        methodNames,
        serviceFunction.handleMetadata?.inputPlaneUrl,
        undefined,
      );
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.NOT_FOUND)
        throw new NotFoundError(`Class '${appName}/${name}' not found`);
      throw err;
    }
  }

  /** Create a new instance of the Cls with parameters and/or runtime options. */
  async instance(params: Record<string, any> = {}): Promise<ClsInstance> {
    let functionId: string;
    if (this.#schema.length === 0 && this.#options === undefined) {
      functionId = this.#serviceFunctionId;
    } else {
      functionId = await this.#bindParameters(params);
    }
    const methods = new Map<string, Function_>();
    for (const name of this.#methodNames) {
      methods.set(name, new Function_(functionId, name, this.#inputPlaneUrl));
    }
    return new ClsInstance(methods);
  }

  /** Override the static Function configuration at runtime. */
  withOptions(options: ClsOptions): Cls {
    const merged = mergeServiceOptions(this.#options, options);
    return new Cls(
      this.#serviceFunctionId,
      this.#schema,
      this.#methodNames,
      this.#inputPlaneUrl,
      merged,
    );
  }

  /** Create an instance of the Cls with input concurrency enabled or overridden with new values. */
  withConcurrency(options: ClsConcurrencyOptions): Cls {
    const merged = mergeServiceOptions(this.#options, {
      maxConcurrentInputs: options.maxInputs,
      targetConcurrentInputs: options.targetInputs,
    });
    return new Cls(
      this.#serviceFunctionId,
      this.#schema,
      this.#methodNames,
      this.#inputPlaneUrl,
      merged,
    );
  }

  /** Create an instance of the Cls with dynamic batching enabled or overridden with new values. */
  withBatching(options: ClsBatchingOptions): Cls {
    const merged = mergeServiceOptions(this.#options, {
      batchMaxSize: options.maxBatchSize,
      batchWaitMs: options.waitMs,
    });
    return new Cls(
      this.#serviceFunctionId,
      this.#schema,
      this.#methodNames,
      this.#inputPlaneUrl,
      merged,
    );
  }

  /** Bind parameters to the Cls function. */
  async #bindParameters(params: Record<string, any>): Promise<string> {
    const serializedParams = encodeParameterSet(this.#schema, params);
    const functionOptions = await buildFunctionOptionsProto(this.#options);
    const bindResp = await client.functionBindParams({
      functionId: this.#serviceFunctionId,
      serializedParams,
      functionOptions,
    });
    return bindResp.boundFunctionId;
  }
}

export function encodeParameterSet(
  schema: ClassParameterSpec[],
  params: Record<string, any>,
): Uint8Array {
  const encoded: ClassParameterValue[] = [];
  for (const paramSpec of schema) {
    const paramValue = encodeParameter(paramSpec, params[paramSpec.name]);
    encoded.push(paramValue);
  }
  // Sort keys, identical to Python `SerializeToString(deterministic=True)`.
  encoded.sort((a, b) => a.name.localeCompare(b.name));
  return ClassParameterSet.encode({ parameters: encoded }).finish();
}

function mergeServiceOptions(
  base: ServiceOptions | undefined,
  diff: Partial<ServiceOptions>,
): ServiceOptions | undefined {
  const filteredDiff = Object.fromEntries(
    Object.entries(diff).filter(([, value]) => value !== undefined),
  ) as Partial<ServiceOptions>;
  const merged = { ...(base ?? {}), ...filteredDiff } as ServiceOptions;
  return Object.keys(merged).length === 0 ? undefined : merged;
}

async function buildFunctionOptionsProto(
  options?: ServiceOptions,
): Promise<FunctionOptions | undefined> {
  if (!options) return undefined;
  const o = options ?? {};

  const gpuConfig = parseGpuConfig(o.gpu);
  const resources =
    o.cpu !== undefined || o.memory !== undefined || gpuConfig
      ? {
          milliCpu: o.cpu !== undefined ? Math.round(1000 * o.cpu) : undefined,
          memoryMb: o.memory,
          gpuConfig,
        }
      : undefined;

  const mergedSecrets = await mergeEnvAndSecrets(o.env, o.secrets);
  const secretIds = mergedSecrets.map((s) => s.secretId);

  const volumeMounts: VolumeMount[] = o.volumes
    ? Object.entries(o.volumes).map(([mountPath, volume]) => ({
        volumeId: volume.volumeId,
        mountPath,
        allowBackgroundCommits: true,
        readOnly: volume.isReadOnly,
      }))
    : [];

  const parsedRetries = parseRetries(o.retries);
  const retryPolicy: FunctionRetryPolicy | undefined = parsedRetries
    ? {
        retries: parsedRetries.maxRetries,
        backoffCoefficient: parsedRetries.backoffCoefficient,
        initialDelayMs: parsedRetries.initialDelayMs,
        maxDelayMs: parsedRetries.maxDelayMs,
      }
    : undefined;

  if (o.scaledownWindow !== undefined && o.scaledownWindow % 1000 !== 0) {
    throw new Error(
      `scaledownWindow must be a multiple of 1000ms, got ${o.scaledownWindow}`,
    );
  }
  if (o.timeout !== undefined && o.timeout % 1000 !== 0) {
    throw new Error(`timeout must be a multiple of 1000ms, got ${o.timeout}`);
  }

  const functionOptions = FunctionOptions.create({
    secretIds,
    replaceSecretIds: secretIds.length > 0,
    replaceVolumeMounts: volumeMounts.length > 0,
    volumeMounts,
    resources,
    retryPolicy,
    concurrencyLimit: o.maxContainers,
    bufferContainers: o.bufferContainers,
    taskIdleTimeoutSecs:
      o.scaledownWindow !== undefined ? o.scaledownWindow / 1000 : undefined,
    timeoutSecs: o.timeout !== undefined ? o.timeout / 1000 : undefined,
    maxConcurrentInputs: o.maxConcurrentInputs,
    targetConcurrentInputs: o.targetConcurrentInputs,
    batchMaxSize: o.batchMaxSize,
    batchLingerMs: o.batchWaitMs,
  });

  return functionOptions;
}

function encodeParameter(
  paramSpec: ClassParameterSpec,
  value: any,
): ClassParameterValue {
  const name = paramSpec.name;
  const paramType = paramSpec.type;
  const paramValue: ClassParameterValue = { name, type: paramType };

  switch (paramType) {
    case ParameterType.PARAM_TYPE_STRING:
      if (value == null && paramSpec.hasDefault) {
        value = paramSpec.stringDefault ?? "";
      }
      if (typeof value !== "string") {
        throw new Error(`Parameter '${name}' must be a string`);
      }
      paramValue.stringValue = value;
      break;

    case ParameterType.PARAM_TYPE_INT:
      if (value == null && paramSpec.hasDefault) {
        value = paramSpec.intDefault ?? 0;
      }
      if (typeof value !== "number") {
        throw new Error(`Parameter '${name}' must be an integer`);
      }
      paramValue.intValue = value;
      break;

    case ParameterType.PARAM_TYPE_BOOL:
      if (value == null && paramSpec.hasDefault) {
        value = paramSpec.boolDefault ?? false;
      }
      if (typeof value !== "boolean") {
        throw new Error(`Parameter '${name}' must be a boolean`);
      }
      paramValue.boolValue = value;
      break;

    case ParameterType.PARAM_TYPE_BYTES:
      if (value == null && paramSpec.hasDefault) {
        value = paramSpec.bytesDefault ?? new Uint8Array();
      }
      if (!(value instanceof Uint8Array)) {
        throw new Error(`Parameter '${name}' must be a byte array`);
      }
      paramValue.bytesValue = value;
      break;

    default:
      throw new Error(`Unsupported parameter type: ${paramType}`);
  }

  return paramValue;
}

/** Represents an instance of a deployed Modal Cls, optionally with parameters. */
export class ClsInstance {
  #methods: Map<string, Function_>;

  constructor(methods: Map<string, Function_>) {
    this.#methods = methods;
  }

  method(name: string): Function_ {
    const method = this.#methods.get(name);
    if (!method) {
      throw new NotFoundError(`Method '${name}' not found on class`);
    }
    return method;
  }
}
