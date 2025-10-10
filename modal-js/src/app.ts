import { ClientError, Status } from "nice-grpc";
import { ObjectCreationType } from "../proto/modal_proto/api";
import { getDefaultClient, type ModalClient } from "./client";
import { Image } from "./image";
import { Sandbox, SandboxCreateParams } from "./sandbox";
import { NotFoundError } from "./errors";
import { Secret } from "./secret";
import { GPUConfig } from "../proto/modal_proto/api";

/**
 * Service for managing Apps.
 */
export class AppService {
  readonly #client: ModalClient;
  constructor(client: ModalClient) {
    this.#client = client;
  }

  /**
   * Referencea deployed App by name, or create if it does not exist.
   */
  async fromName(name: string, params: AppFromNameParams = {}): Promise<App> {
    try {
      const resp = await this.#client.cpClient.appGetOrCreate({
        appName: name,
        environmentName: this.#client.environmentName(params.environment),
        objectCreationType: params.createIfMissing
          ? ObjectCreationType.OBJECT_CREATION_TYPE_CREATE_IF_MISSING
          : ObjectCreationType.OBJECT_CREATION_TYPE_UNSPECIFIED,
      });
      return new App(resp.appId, name);
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.NOT_FOUND)
        throw new NotFoundError(`App '${name}' not found`);
      throw err;
    }
  }
}

/** Optional parameters for `client.apps.fromName()`. */
export type AppFromNameParams = {
  environment?: string;
  createIfMissing?: boolean;
};

/** @deprecated Use specific Params types instead. */
export type LookupOptions = {
  environment?: string;
  createIfMissing?: boolean;
};

/** @deprecated Use specific Params types instead. */
export type DeleteOptions = {
  environment?: string;
};

/** @deprecated Use specific Params types instead. */
export type EphemeralOptions = {
  environment?: string;
};

/**
 * Parse a GPU configuration string into a GPUConfig object.
 * @param gpu - GPU string in format "type" or "type:count" (e.g. "T4", "A100:2")
 * @returns GPUConfig object or undefined if no GPU specified
 */
export function parseGpuConfig(gpu: string | undefined): GPUConfig | undefined {
  if (!gpu) {
    return undefined;
  }

  let gpuType = gpu;
  let count = 1;

  if (gpu.includes(":")) {
    const [type, countStr] = gpu.split(":", 2);
    gpuType = type;
    count = parseInt(countStr, 10);
    if (isNaN(count) || count < 1) {
      throw new Error(
        `Invalid GPU count: ${countStr}. Value must be a positive integer.`,
      );
    }
  }

  return {
    type: 0, // Deprecated field, but required by proto
    count,
    gpuType: gpuType.toUpperCase(),
  };
}

/** Represents a deployed Modal App. */
export class App {
  readonly appId: string;
  readonly name?: string;

  /** @ignore */
  constructor(appId: string, name?: string) {
    this.appId = appId;
    this.name = name;
  }

  /**
   * @deprecated Use `client.apps.fromName()` instead.
   */
  static async lookup(name: string, options: LookupOptions = {}): Promise<App> {
    return getDefaultClient().apps.fromName(name, options);
  }

  /**
   * @deprecated Use `client.sandboxes.create()` instead.
   */
  async createSandbox(
    image: Image,
    options: SandboxCreateParams = {},
  ): Promise<Sandbox> {
    return getDefaultClient().sandboxes.create(this, image, options);
  }

  /**
   * @deprecated Use `client.images.fromRegistry()` instead.
   */
  async imageFromRegistry(tag: string, secret?: Secret): Promise<Image> {
    return getDefaultClient().images.fromRegistry(tag, secret).build(this);
  }

  /**
   * @deprecated Use `client.images.fromAwsEcr()` instead.
   */
  async imageFromAwsEcr(tag: string, secret: Secret): Promise<Image> {
    return getDefaultClient().images.fromAwsEcr(tag, secret).build(this);
  }

  /**
   * @deprecated Use `client.images.fromGcpArtifactRegistry()` instead.
   */
  async imageFromGcpArtifactRegistry(
    tag: string,
    secret: Secret,
  ): Promise<Image> {
    return getDefaultClient()
      .images.fromGcpArtifactRegistry(tag, secret)
      .build(this);
  }
}
