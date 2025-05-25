import { ClientError, Status } from "nice-grpc";
import {
  NetworkAccess_NetworkAccessType,
  ObjectCreationType,
} from "../proto/modal_proto/api";
import { client } from "./client";
import { environmentName } from "./config";
import { fromRegistryInternal, Image } from "./image";
import { Sandbox } from "./sandbox";
import { NotFoundError } from "./errors";

export type LookupOptions = {
  environment?: string;
  createIfMissing?: boolean;
};

export type SandboxCreateOptions = {
  cpu?: number; // in physical cores
  memory?: number; // in MiB
  timeout?: number; // in milliseconds
  command?: string[]; // default is ["sleep", "48h"]
};

export class App {
  readonly appId: string;

  constructor(appId: string) {
    this.appId = appId;
  }

  /** Lookup a deployed app by name, or create if it does not exist. */
  static async lookup(name: string, options: LookupOptions = {}): Promise<App> {
    try {
      const resp = await client.appGetOrCreate({
        appName: name,
        environmentName: environmentName(options.environment),
        objectCreationType: options.createIfMissing
          ? ObjectCreationType.OBJECT_CREATION_TYPE_CREATE_IF_MISSING
          : ObjectCreationType.OBJECT_CREATION_TYPE_UNSPECIFIED,
      });
      return new App(resp.appId);
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.NOT_FOUND)
        throw new NotFoundError(`App '${name}' not found`);
      throw err;
    }
  }

  async createSandbox(
    image: Image,
    options: SandboxCreateOptions = {},
  ): Promise<Sandbox> {
    if (options.timeout && options.timeout % 1000 !== 0) {
      // The gRPC API only accepts a whole number of seconds.
      throw new Error(
        `Timeout must be a multiple of 1000ms, got ${options.timeout}`,
      );
    }

    const createResp = await client.sandboxCreate({
      appId: this.appId,
      definition: {
        // Sleep default is implicit in image builder version <=2024.10
        entrypointArgs: options.command ?? ["sleep", "48h"],
        imageId: image.imageId,
        timeoutSecs:
          options.timeout != undefined ? options.timeout / 1000 : 600,
        networkAccess: {
          networkAccessType: NetworkAccess_NetworkAccessType.OPEN,
        },
        resources: {
          // https://modal.com/docs/guide/resources
          milliCpu: Math.round(1000 * (options.cpu ?? 0.125)),
          memoryMb: options.memory ?? 128,
        },
      },
    });

    return new Sandbox(createResp.sandboxId);
  }

  async imageFromRegistry(tag: string): Promise<Image> {
    return await fromRegistryInternal(this.appId, tag);
  }
}
