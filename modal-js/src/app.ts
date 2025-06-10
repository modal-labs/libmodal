import { ClientError, Status } from "nice-grpc";
import {
  NetworkAccess_NetworkAccessType,
  ObjectCreationType,
  RegistryAuthType,
} from "../proto/modal_proto/api";
import { client } from "./client";
import { environmentName } from "./config";
import { fromRegistryInternal, type Image } from "./image";
import { Sandbox } from "./sandbox";
import { NotFoundError } from "./errors";
import { Secret } from "./secret";

/** Options for functions that find deployed Modal objects. */
export type LookupOptions = {
  environment?: string;
  createIfMissing?: boolean;
};

/** Options for deleting a named object. */
export type DeleteOptions = {
  environment?: string;
};

/** Options for constructors that create a temporary, nameless object. */
export type EphemeralOptions = {
  environment?: string;
};

/** Options for `App.createSandbox()`. */
export type SandboxCreateOptions = {
  /** Reservation of physical CPU cores for the sandbox, can be fractional. */
  cpu?: number;

  /** Reservation of memory in MiB. */
  memory?: number;

  /** Timeout of the sandbox container, defaults to 10 minutes. */
  timeout?: number;

  /**
   * Sequence of program arguments for the main process.
   * Default behavior is to sleep indefinitely until timeout or termination.
   */
  command?: string[]; // default is ["sleep", "48h"]
};

/** Represents a deployed Modal App. */
export class App {
  readonly appId: string;

  /** @ignore */
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

  async imageFromRegistry(tag: string, secret?: Secret): Promise<Image> {
    let imageRegistryConfig;
    if (secret) {
      if (!(secret instanceof Secret)) {
        throw new TypeError(
          "secret must be a reference to an existing Secret, e.g. `await Secret.fromName('my_secret')`",
        );
      }
      imageRegistryConfig = {
        registryAuthType: RegistryAuthType.REGISTRY_AUTH_TYPE_STATIC_CREDS,
        secretId: secret.secretId,
      };
    }
    return await fromRegistryInternal(this.appId, tag, imageRegistryConfig);
  }

  async imageFromAwsEcr(tag: string, secret: Secret): Promise<Image> {
    if (!(secret instanceof Secret)) {
      throw new TypeError(
        "secret must be a reference to an existing Secret, e.g. `await Secret.fromName('my_secret')`",
      );
    }

    const imageRegistryConfig = {
      registryAuthType: RegistryAuthType.REGISTRY_AUTH_TYPE_AWS,
      secretId: secret.secretId,
    };

    return await fromRegistryInternal(this.appId, tag, imageRegistryConfig);
  }

  async imageFromGcpArtifactRegistry(
    tag: string,
    secret: Secret,
  ): Promise<Image> {
    if (!(secret instanceof Secret)) {
      throw new TypeError(
        "secret must be a reference to an existing Secret, e.g. `await Secret.fromName('my_secret')`",
      );
    }

    const imageRegistryConfig = {
      registryAuthType: RegistryAuthType.REGISTRY_AUTH_TYPE_GCP,
      secretId: secret.secretId,
    };

    return await fromRegistryInternal(this.appId, tag, imageRegistryConfig);
  }
}
