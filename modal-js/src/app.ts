import { ClientError, Status } from "nice-grpc";
import {
  NetworkAccess_NetworkAccessType,
  ObjectCreationType,
  PortSpec,
  TunnelType,
  NetworkAccess,
  GPUConfig,
  SchedulerPlacement,
  VolumeMount,
  CloudBucketMount as CloudBucketMountProto,
} from "../proto/modal_proto/api";
import { client } from "./client";
import { environmentName } from "./config";
import { Image } from "./image";
import { Sandbox } from "./sandbox";
import { NotFoundError } from "./errors";
import { Secret } from "./secret";
import { Volume } from "./volume";
import { Proxy } from "./proxy";
import {
  CloudBucketMount,
  cloudBucketMountToProto,
} from "./cloud_bucket_mount";

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

  /** GPU reservation for the sandbox (e.g. "A100", "T4:2", "A100-80GB:4"). */
  gpu?: string;

  /** Timeout of the sandbox container, defaults to 10 minutes. */
  timeout?: number;

  /** Working directory of the sandbox. */
  workdir?: string;

  /**
   * Sequence of program arguments for the main process.
   * Default behavior is to sleep indefinitely until timeout or termination.
   */
  command?: string[]; // default is ["sleep", "48h"]

  /** Secrets to inject into the sandbox. */
  secrets?: Secret[];

  /** Mount points for Modal Volumes. */
  volumes?: Record<string, Volume>;

  /** Mount points for cloud buckets. */
  cloudBucketMounts?: Record<string, CloudBucketMount>;

  /** List of ports to tunnel into the sandbox. Encrypted ports are tunneled with TLS. */
  encryptedPorts?: number[];

  /** List of encrypted ports to tunnel into the sandbox, using HTTP/2. */
  h2Ports?: number[];

  /** List of ports to tunnel into the sandbox without encryption. */
  unencryptedPorts?: number[];

  /** Whether to block all network access from the sandbox. */
  blockNetwork?: boolean;

  /** List of CIDRs the sandbox is allowed to access. If None, all CIDRs are allowed. Cannot be used with blockNetwork. */
  cidrAllowlist?: string[];

  /** Cloud provider to run the sandbox on. */
  cloud?: string;

  /** Region(s) to run the sandbox on. */
  regions?: string[];

  /** Enable verbose logging. */
  verbose?: boolean;

  /** Reference to a Modal Proxy to use in front of this Sandbox. */
  proxy?: Proxy;
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
      return new App(resp.appId, name);
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
    const gpuConfig = parseGpuConfig(options.gpu);

    if (options.timeout && options.timeout % 1000 !== 0) {
      // The gRPC API only accepts a whole number of seconds.
      throw new Error(
        `Timeout must be a multiple of 1000ms, got ${options.timeout}`,
      );
    }
    await image._build(this.appId);

    if (options.workdir && !options.workdir.startsWith("/")) {
      throw new Error(
        `workdir must be an absolute path, got: ${options.workdir}`,
      );
    }

    const volumeMounts: VolumeMount[] = options.volumes
      ? Object.entries(options.volumes).map(([mountPath, volume]) => ({
          volumeId: volume.volumeId,
          mountPath,
          allowBackgroundCommits: true,
          readOnly: volume.isReadOnly,
        }))
      : [];

    const cloudBucketMounts: CloudBucketMountProto[] = options.cloudBucketMounts
      ? Object.entries(options.cloudBucketMounts).map(([mountPath, mount]) =>
          cloudBucketMountToProto(mount, mountPath),
        )
      : [];

    const openPorts: PortSpec[] = [];
    if (options.encryptedPorts) {
      openPorts.push(
        ...options.encryptedPorts.map((port) => ({
          port,
          unencrypted: false,
        })),
      );
    }
    if (options.h2Ports) {
      openPorts.push(
        ...options.h2Ports.map((port) => ({
          port,
          unencrypted: false,
          tunnelType: TunnelType.TUNNEL_TYPE_H2,
        })),
      );
    }
    if (options.unencryptedPorts) {
      openPorts.push(
        ...options.unencryptedPorts.map((port) => ({
          port,
          unencrypted: true,
        })),
      );
    }

    const secretIds = options.secrets
      ? options.secrets.map((secret) => secret.secretId)
      : [];

    let networkAccess: NetworkAccess;
    if (options.blockNetwork) {
      if (options.cidrAllowlist) {
        throw new Error(
          "cidrAllowlist cannot be used when blockNetwork is enabled",
        );
      }
      networkAccess = {
        networkAccessType: NetworkAccess_NetworkAccessType.BLOCKED,
        allowedCidrs: [],
      };
    } else if (options.cidrAllowlist) {
      networkAccess = {
        networkAccessType: NetworkAccess_NetworkAccessType.ALLOWLIST,
        allowedCidrs: options.cidrAllowlist,
      };
    } else {
      networkAccess = {
        networkAccessType: NetworkAccess_NetworkAccessType.OPEN,
        allowedCidrs: [],
      };
    }

    const schedulerPlacement = SchedulerPlacement.create({
      regions: options.regions ?? [],
    });

    const createResp = await client.sandboxCreate({
      appId: this.appId,
      definition: {
        // Sleep default is implicit in image builder version <=2024.10
        entrypointArgs: options.command ?? ["sleep", "48h"],
        imageId: image.imageId,
        timeoutSecs:
          options.timeout != undefined ? options.timeout / 1000 : 600,
        workdir: options.workdir ?? undefined,
        networkAccess,
        resources: {
          // https://modal.com/docs/guide/resources
          milliCpu: Math.round(1000 * (options.cpu ?? 0.125)),
          memoryMb: options.memory ?? 128,
          gpuConfig,
        },
        volumeMounts,
        cloudBucketMounts,
        secretIds,
        openPorts: openPorts.length > 0 ? { ports: openPorts } : undefined,
        cloudProviderStr: options.cloud ?? "",
        schedulerPlacement,
        verbose: options.verbose ?? false,
        proxyId: options.proxy?.proxyId,
      },
    });

    return new Sandbox(createResp.sandboxId);
  }

  /**
   * @deprecated Use `Image.fromRegistry` instead.
   */
  async imageFromRegistry(tag: string, secret?: Secret): Promise<Image> {
    return await Image.fromRegistry(tag, secret)._build(this.appId);
  }

  /**
   * @deprecated Use `Image.fromAwsEcr` instead.
   */
  async imageFromAwsEcr(tag: string, secret: Secret): Promise<Image> {
    return await Image.fromAwsEcr(tag, secret)._build(this.appId);
  }

  /**
   * @deprecated Use `Image.fromGcpArtifactRegistry` instead.
   */
  async imageFromGcpArtifactRegistry(
    tag: string,
    secret: Secret,
  ): Promise<Image> {
    return await Image.fromGcpArtifactRegistry(tag, secret)._build(this.appId);
  }
}
