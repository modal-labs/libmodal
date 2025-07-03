import { ClientError, Status } from "nice-grpc";
import {
  ClassParameterInfo_ParameterSerializationFormat,
  ClassParameterSet,
  ClassParameterSpec,
  ClassParameterValue,
  ParameterType,
} from "../proto/modal_proto/api";
import type { LookupOptions } from "./app";
import { NotFoundError } from "./errors";
import { client } from "./client";
import { environmentName } from "./config";
import { Function_ } from "./function";

/** Represents a deployed Modal Cls. */
export class Cls {
  #serviceFunctionId: string;
  #schema: ClassParameterSpec[];
  #methodNames: string[];
  #inputPlaneUrl?: string;

  /** @ignore */
  constructor(
    serviceFunctionId: string,
    schema: ClassParameterSpec[],
    methodNames: string[],
    inputPlaneUrl?: string,
  ) {
    this.#serviceFunctionId = serviceFunctionId;
    this.#schema = schema;
    this.#methodNames = methodNames;
    this.#inputPlaneUrl = inputPlaneUrl;
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
      );
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.NOT_FOUND)
        throw new NotFoundError(`Class '${appName}/${name}' not found`);
      throw err;
    }
  }

  /** Create a new instance of the Cls with parameters. */
  async instance(params: Record<string, any> = {}): Promise<ClsInstance> {
    let functionId: string;
    if (this.#schema.length === 0) {
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

  /** Bind parameters to the Cls function. */
  async #bindParameters(params: Record<string, any>): Promise<string> {
    const serializedParams = encodeParameterSet(this.#schema, params);
    const bindResp = await client.functionBindParams({
      functionId: this.#serviceFunctionId,
      serializedParams,
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
