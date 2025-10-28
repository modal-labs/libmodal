import { ClientError, Status } from "nice-grpc";
import {
  FileDescriptor,
  GenericResult,
  GenericResult_GenericStatus,
  PTYInfo,
  PTYInfo_PTYType,
  ContainerExecRequest,
  SandboxTagsGetResponse,
  SandboxCreateRequest,
  NetworkAccess,
  NetworkAccess_NetworkAccessType,
  VolumeMount,
  CloudBucketMount as CloudBucketMountProto,
  SchedulerPlacement,
  TunnelType,
  PortSpec,
} from "../proto/modal_proto/api";
import {
  getDefaultClient,
  type ModalClient,
  isRetryableGrpc,
  ModalGrpcClient,
} from "./client";
import {
  runFilesystemExec,
  SandboxFile,
  SandboxFileMode,
} from "./sandbox_filesystem";
import {
  type ModalReadStream,
  type ModalWriteStream,
  streamConsumingIter,
  toModalReadStream,
  toModalWriteStream,
} from "./streams";
import { type Secret, mergeEnvIntoSecrets } from "./secret";
import {
  InvalidError,
  NotFoundError,
  SandboxTimeoutError,
  AlreadyExistsError,
} from "./errors";
import { Image } from "./image";
import type { Volume } from "./volume";
import type { Proxy } from "./proxy";
import type { CloudBucketMount } from "./cloud_bucket_mount";
import { cloudBucketMountToProto } from "./cloud_bucket_mount";
import type { App } from "./app";
import { parseGpuConfig } from "./app";
import { checkForRenamedParams } from "./validation";

/**
 * Stdin is always present, but this option allow you to drop stdout or stderr
 * if you don't need them. The default is "pipe", matching Node.js behavior.
 *
 * If behavior is set to "ignore", the output streams will be empty.
 */
export type StdioBehavior = "pipe" | "ignore";

/**
 * Specifies the type of data that will be read from the Sandbox or container
 * process. "text" means the data will be read as UTF-8 text, while "binary"
 * means the data will be read as raw bytes (Uint8Array).
 */
export type StreamMode = "text" | "binary";

/** Optional parameters for {@link SandboxService#create client.sandboxes.create()}. */
export type SandboxCreateParams = {
  /** Reservation of physical CPU cores for the Sandbox, can be fractional. */
  cpu?: number;

  /** Hard limit of physical CPU cores for the Sandbox, can be fractional. */
  cpuLimit?: number;

  /** Reservation of memory in MiB. */
  memoryMiB?: number;

  /** Hard limit of memory in MiB. */
  memoryLimitMiB?: number;

  /** GPU reservation for the Sandbox (e.g. "A100", "T4:2", "A100-80GB:4"). */
  gpu?: string;

  /** Timeout of the Sandbox container in milliseconds, defaults to 10 minutes. */
  timeoutMs?: number;

  /** The amount of time in milliseconds that a sandbox can be idle before being terminated. */
  idleTimeoutMs?: number;

  /** Working directory of the Sandbox. */
  workdir?: string;

  /**
   * Sequence of program arguments for the main process.
   * Default behavior is to sleep indefinitely until timeout or termination.
   */
  command?: string[]; // default is ["sleep", "48h"]

  /** Environment variables to set in the Sandbox. */
  env?: Record<string, string>;

  /** {@link Secret}s to inject into the Sandbox as environment variables. */
  secrets?: Secret[];

  /** Mount points for Modal {@link Volume}s. */
  volumes?: Record<string, Volume>;

  /** Mount points for {@link CloudBucketMount}s. */
  cloudBucketMounts?: Record<string, CloudBucketMount>;

  /** Enable a PTY for the Sandbox. */
  pty?: boolean;

  /** List of ports to tunnel into the Sandbox. Encrypted ports are tunneled with TLS. */
  encryptedPorts?: number[];

  /** List of encrypted ports to tunnel into the Sandbox, using HTTP/2. */
  h2Ports?: number[];

  /** List of ports to tunnel into the Sandbox without encryption. */
  unencryptedPorts?: number[];

  /** Whether to block all network access from the Sandbox. */
  blockNetwork?: boolean;

  /** List of CIDRs the Sandbox is allowed to access. If None, all CIDRs are allowed. Cannot be used with blockNetwork. */
  cidrAllowlist?: string[];

  /** Cloud provider to run the Sandbox on. */
  cloud?: string;

  /** Region(s) to run the Sandbox on. */
  regions?: string[];

  /** Enable verbose logging. */
  verbose?: boolean;

  /** Reference to a Modal {@link Proxy} to use in front of this Sandbox. */
  proxy?: Proxy;

  /** Optional name for the Sandbox. Unique within an App. */
  name?: string;
};

export async function buildSandboxCreateRequestProto(
  appId: string,
  imageId: string,
  params: SandboxCreateParams = {},
): Promise<SandboxCreateRequest> {
  checkForRenamedParams(params, {
    memory: "memoryMiB",
    memoryLimit: "memoryLimitMiB",
    timeout: "timeoutMs",
    idleTimeout: "idleTimeoutMs",
  });

  const gpuConfig = parseGpuConfig(params.gpu);

  // The gRPC API only accepts a whole number of seconds.
  if (params.timeoutMs && params.timeoutMs % 1000 !== 0) {
    throw new Error(
      `timeoutMs must be a multiple of 1000ms, got ${params.timeoutMs}`,
    );
  }
  if (params.idleTimeoutMs && params.idleTimeoutMs % 1000 !== 0) {
    throw new Error(
      `idleTimeoutMs must be a multiple of 1000ms, got ${params.idleTimeoutMs}`,
    );
  }

  if (params.workdir && !params.workdir.startsWith("/")) {
    throw new Error(`workdir must be an absolute path, got: ${params.workdir}`);
  }

  const volumeMounts: VolumeMount[] = params.volumes
    ? Object.entries(params.volumes).map(([mountPath, volume]) => ({
        volumeId: volume.volumeId,
        mountPath,
        allowBackgroundCommits: true,
        readOnly: volume.isReadOnly,
      }))
    : [];

  const cloudBucketMounts: CloudBucketMountProto[] = params.cloudBucketMounts
    ? Object.entries(params.cloudBucketMounts).map(([mountPath, mount]) =>
        cloudBucketMountToProto(mount, mountPath),
      )
    : [];

  const openPorts: PortSpec[] = [];
  if (params.encryptedPorts) {
    openPorts.push(
      ...params.encryptedPorts.map((port) => ({
        port,
        unencrypted: false,
      })),
    );
  }
  if (params.h2Ports) {
    openPorts.push(
      ...params.h2Ports.map((port) => ({
        port,
        unencrypted: false,
        tunnelType: TunnelType.TUNNEL_TYPE_H2,
      })),
    );
  }
  if (params.unencryptedPorts) {
    openPorts.push(
      ...params.unencryptedPorts.map((port) => ({
        port,
        unencrypted: true,
      })),
    );
  }

  const secretIds = (params.secrets || []).map((secret) => secret.secretId);

  let networkAccess: NetworkAccess;
  if (params.blockNetwork) {
    if (params.cidrAllowlist) {
      throw new Error(
        "cidrAllowlist cannot be used when blockNetwork is enabled",
      );
    }
    networkAccess = {
      networkAccessType: NetworkAccess_NetworkAccessType.BLOCKED,
      allowedCidrs: [],
    };
  } else if (params.cidrAllowlist) {
    networkAccess = {
      networkAccessType: NetworkAccess_NetworkAccessType.ALLOWLIST,
      allowedCidrs: params.cidrAllowlist,
    };
  } else {
    networkAccess = {
      networkAccessType: NetworkAccess_NetworkAccessType.OPEN,
      allowedCidrs: [],
    };
  }

  const schedulerPlacement = SchedulerPlacement.create({
    regions: params.regions ?? [],
  });

  let ptyInfo: PTYInfo | undefined;
  if (params.pty) {
    ptyInfo = defaultSandboxPTYInfo();
  }

  let milliCpu: number | undefined = undefined;
  let milliCpuMax: number | undefined = undefined;
  if (params.cpu === undefined && params.cpuLimit !== undefined) {
    throw new Error("must also specify cpu when cpuLimit is specified");
  }
  if (params.cpu !== undefined) {
    if (params.cpu <= 0) {
      throw new Error(`cpu (${params.cpu}) must be a positive number`);
    }
    milliCpu = Math.trunc(1000 * params.cpu);
    if (params.cpuLimit !== undefined) {
      if (params.cpuLimit < params.cpu) {
        throw new Error(
          `cpu (${params.cpu}) cannot be higher than cpuLimit (${params.cpuLimit})`,
        );
      }
      milliCpuMax = Math.trunc(1000 * params.cpuLimit);
    }
  }

  let memoryMb: number | undefined = undefined;
  let memoryMbMax: number | undefined = undefined;
  if (params.memoryMiB === undefined && params.memoryLimitMiB !== undefined) {
    throw new Error(
      "must also specify memoryMiB when memoryLimitMiB is specified",
    );
  }
  if (params.memoryMiB !== undefined) {
    if (params.memoryMiB <= 0) {
      throw new Error(
        `the memoryMiB request (${params.memoryMiB}) must be a positive number`,
      );
    }
    memoryMb = params.memoryMiB;
    if (params.memoryLimitMiB !== undefined) {
      if (params.memoryLimitMiB < params.memoryMiB) {
        throw new Error(
          `the memoryMiB request (${params.memoryMiB}) cannot be higher than memoryLimitMiB (${params.memoryLimitMiB})`,
        );
      }
      memoryMbMax = params.memoryLimitMiB;
    }
  }

  return SandboxCreateRequest.create({
    appId,
    definition: {
      // Sleep default is implicit in image builder version <=2024.10
      entrypointArgs: params.command ?? ["sleep", "48h"],
      imageId,
      timeoutSecs:
        params.timeoutMs != undefined ? params.timeoutMs / 1000 : 600,
      idleTimeoutSecs:
        params.idleTimeoutMs != undefined
          ? params.idleTimeoutMs / 1000
          : undefined,
      workdir: params.workdir ?? undefined,
      networkAccess,
      resources: {
        milliCpu,
        milliCpuMax,
        memoryMb,
        memoryMbMax,
        gpuConfig,
      },
      volumeMounts,
      cloudBucketMounts,
      ptyInfo,
      secretIds,
      openPorts: openPorts.length > 0 ? { ports: openPorts } : undefined,
      cloudProviderStr: params.cloud ?? "",
      schedulerPlacement,
      verbose: params.verbose ?? false,
      proxyId: params.proxy?.proxyId,
      name: params.name,
    },
  });
}

/**
 * Service for managing {@link Sandbox}es.
 *
 * Normally only ever accessed via the client as:
 * ```typescript
 * const modal = new ModalClient();
 * const sandbox = await modal.sandboxes.create(app, image);
 * ```
 */
export class SandboxService {
  readonly #client: ModalClient;
  constructor(client: ModalClient) {
    this.#client = client;
  }

  /**
   * Create a new {@link Sandbox} in the {@link App} with the specified {@link Image} and options.
   */
  async create(
    app: App,
    image: Image,
    params: SandboxCreateParams = {},
  ): Promise<Sandbox> {
    await image.build(app);

    const mergedSecrets = await mergeEnvIntoSecrets(
      this.#client,
      params.env,
      params.secrets,
    );
    const mergedParams = {
      ...params,
      secrets: mergedSecrets,
      env: undefined, // setting env to undefined just to clarify it's not needed anymore
    };

    const createReq = await buildSandboxCreateRequestProto(
      app.appId,
      image.imageId,
      mergedParams,
    );
    let createResp;
    try {
      createResp = await this.#client.cpClient.sandboxCreate(createReq);
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.ALREADY_EXISTS) {
        throw new AlreadyExistsError(err.details || err.message);
      }
      throw err;
    }

    return new Sandbox(this.#client, createResp.sandboxId);
  }

  /** Returns a running {@link Sandbox} object from an ID.
   *
   * @returns Sandbox with ID
   */
  async fromId(sandboxId: string): Promise<Sandbox> {
    try {
      await this.#client.cpClient.sandboxWait({
        sandboxId,
        timeout: 0,
      });
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.NOT_FOUND)
        throw new NotFoundError(`Sandbox with id: '${sandboxId}' not found`);
      throw err;
    }

    return new Sandbox(this.#client, sandboxId);
  }

  /** Get a running {@link Sandbox} by name from a deployed {@link App}.
   *
   * Raises a {@link NotFoundError} if no running Sandbox is found with the given name.
   * A Sandbox's name is the `name` argument passed to {@link SandboxService#create sandboxes.create()}.
   *
   * @param appName - Name of the deployed App
   * @param name - Name of the Sandbox
   * @param params - Optional parameters for getting the Sandbox
   * @returns Promise that resolves to a Sandbox
   */
  async fromName(
    appName: string,
    name: string,
    params?: SandboxFromNameParams,
  ): Promise<Sandbox> {
    try {
      const resp = await this.#client.cpClient.sandboxGetFromName({
        sandboxName: name,
        appName,
        environmentName: this.#client.environmentName(params?.environment),
      });
      return new Sandbox(this.#client, resp.sandboxId);
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.NOT_FOUND)
        throw new NotFoundError(
          `Sandbox with name '${name}' not found in App '${appName}'`,
        );
      throw err;
    }
  }

  /**
   * List all {@link Sandbox}es for the current Environment or App ID (if specified).
   * If tags are specified, only Sandboxes that have at least those tags are returned.
   */
  async *list(
    params: SandboxListParams = {},
  ): AsyncGenerator<Sandbox, void, unknown> {
    const env = this.#client.environmentName(params.environment);
    const tagsList = params.tags
      ? Object.entries(params.tags).map(([tagName, tagValue]) => ({
          tagName,
          tagValue,
        }))
      : [];

    let beforeTimestamp: number | undefined = undefined;
    while (true) {
      try {
        const resp = await this.#client.cpClient.sandboxList({
          appId: params.appId,
          beforeTimestamp,
          environmentName: env,
          includeFinished: false,
          tags: tagsList,
        });
        if (!resp.sandboxes || resp.sandboxes.length === 0) {
          return;
        }
        for (const info of resp.sandboxes) {
          yield new Sandbox(this.#client, info.id);
        }
        beforeTimestamp = resp.sandboxes[resp.sandboxes.length - 1].createdAt;
      } catch (err) {
        if (
          err instanceof ClientError &&
          err.code === Status.INVALID_ARGUMENT
        ) {
          throw new InvalidError(err.details || err.message);
        }
        throw err;
      }
    }
  }
}

/** Optional parameters for {@link SandboxService#list client.sandboxes.list()}. */
export type SandboxListParams = {
  /** Filter Sandboxes for a specific {@link App}. */
  appId?: string;
  /** Only return Sandboxes that include all specified tags. */
  tags?: Record<string, string>;
  /** Override environment for the request; defaults to current profile. */
  environment?: string;
};

/** Optional parameters for {@link SandboxService#fromName client.sandboxes.fromName()}. */
export type SandboxFromNameParams = {
  environment?: string;
};

/** Optional parameters for {@link Sandbox#exec Sandbox.exec()}. */
export type SandboxExecParams = {
  /** Specifies text or binary encoding for input and output streams. */
  mode?: StreamMode;
  /** Whether to pipe or ignore standard output. */
  stdout?: StdioBehavior;
  /** Whether to pipe or ignore standard error. */
  stderr?: StdioBehavior;
  /** Working directory to run the command in. */
  workdir?: string;
  /** Timeout for the process in milliseconds. Defaults to 0 (no timeout). */
  timeoutMs?: number;
  /** Environment variables to set for the command. */
  env?: Record<string, string>;
  /** {@link Secret}s to inject as environment variables for the commmand.*/
  secrets?: Secret[];
  /** Enable a PTY for the command. */
  pty?: boolean;
};

/** A port forwarded from within a running Modal {@link Sandbox}. */
export class Tunnel {
  /** @ignore */
  constructor(
    public host: string,
    public port: number,
    public unencryptedHost?: string,
    public unencryptedPort?: number,
  ) {}

  /** Get the public HTTPS URL of the forwarded port. */
  get url(): string {
    let value = `https://${this.host}`;
    if (this.port !== 443) {
      value += `:${this.port}`;
    }
    return value;
  }

  /** Get the public TLS socket as a [host, port] tuple. */
  get tlsSocket(): [string, number] {
    return [this.host, this.port];
  }

  /** Get the public TCP socket as a [host, port] tuple. */
  get tcpSocket(): [string, number] {
    if (!this.unencryptedHost || this.unencryptedPort === undefined) {
      throw new InvalidError(
        "This tunnel is not configured for unencrypted TCP.",
      );
    }
    return [this.unencryptedHost, this.unencryptedPort];
  }
}

export function defaultSandboxPTYInfo(): PTYInfo {
  return PTYInfo.create({
    enabled: true,
    winszRows: 24,
    winszCols: 80,
    envTerm: "xterm-256color",
    envColorterm: "truecolor",
    envTermProgram: "",
    ptyType: PTYInfo_PTYType.PTY_TYPE_SHELL,
    noTerminateOnIdleStdin: true,
  });
}

export async function buildContainerExecRequestProto(
  taskId: string,
  command: string[],
  params?: SandboxExecParams,
): Promise<ContainerExecRequest> {
  checkForRenamedParams(params, { timeout: "timeoutMs" });

  const secretIds = (params?.secrets || []).map((secret) => secret.secretId);

  let ptyInfo: PTYInfo | undefined;
  if (params?.pty) {
    ptyInfo = defaultSandboxPTYInfo();
  }

  return ContainerExecRequest.create({
    taskId,
    command,
    workdir: params?.workdir,
    timeoutSecs: params?.timeoutMs ? params.timeoutMs / 1000 : 0,
    secretIds,
    ptyInfo,
  });
}

/** Sandboxes are secure, isolated containers in Modal that boot in seconds. */
export class Sandbox {
  readonly #client: ModalClient;
  readonly sandboxId: string;
  stdin: ModalWriteStream<string>;
  stdout: ModalReadStream<string>;
  stderr: ModalReadStream<string>;

  #taskId: string | undefined;
  #tunnels: Record<number, Tunnel> | undefined;

  /** @ignore */
  constructor(client: ModalClient, sandboxId: string) {
    this.#client = client;
    this.sandboxId = sandboxId;

    this.stdin = toModalWriteStream(inputStreamSb(client.cpClient, sandboxId));
    this.stdout = toModalReadStream(
      streamConsumingIter(
        outputStreamSb(
          client.cpClient,
          sandboxId,
          FileDescriptor.FILE_DESCRIPTOR_STDOUT,
        ),
      ).pipeThrough(new TextDecoderStream()),
    );
    this.stderr = toModalReadStream(
      streamConsumingIter(
        outputStreamSb(
          client.cpClient,
          sandboxId,
          FileDescriptor.FILE_DESCRIPTOR_STDERR,
        ),
      ).pipeThrough(new TextDecoderStream()),
    );
  }

  /** Set tags (key-value pairs) on the Sandbox. Tags can be used to filter results in {@link SandboxService#list Sandbox.list}. */
  async setTags(tags: Record<string, string>): Promise<void> {
    const tagsList = Object.entries(tags).map(([tagName, tagValue]) => ({
      tagName,
      tagValue,
    }));
    try {
      await this.#client.cpClient.sandboxTagsSet({
        environmentName: this.#client.environmentName(),
        sandboxId: this.sandboxId,
        tags: tagsList,
      });
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.INVALID_ARGUMENT) {
        throw new InvalidError(err.details || err.message);
      }
      throw err;
    }
  }

  /** Get tags (key-value pairs) currently attached to this Sandbox from the server. */
  async getTags(): Promise<Record<string, string>> {
    let resp: SandboxTagsGetResponse;
    try {
      resp = await this.#client.cpClient.sandboxTagsGet({
        sandboxId: this.sandboxId,
      });
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.INVALID_ARGUMENT) {
        throw new InvalidError(err.details || err.message);
      }
      throw err;
    }

    const tags: Record<string, string> = {};
    for (const tag of resp.tags) {
      tags[tag.tagName] = tag.tagValue;
    }
    return tags;
  }

  /**
   * @deprecated Use {@link SandboxService#fromId client.sandboxes.fromId()} instead.
   */
  static async fromId(sandboxId: string): Promise<Sandbox> {
    return getDefaultClient().sandboxes.fromId(sandboxId);
  }

  /**
   * @deprecated Use {@link SandboxService#fromName client.sandboxes.fromName()} instead.
   */
  static async fromName(
    appName: string,
    name: string,
    environment?: string,
  ): Promise<Sandbox> {
    return getDefaultClient().sandboxes.fromName(appName, name, {
      environment,
    });
  }

  /**
   * Open a file in the Sandbox filesystem.
   * @param path - Path to the file to open
   * @param mode - File open mode (r, w, a, r+, w+, a+)
   * @returns Promise that resolves to a {@link SandboxFile}
   */
  async open(path: string, mode: SandboxFileMode = "r"): Promise<SandboxFile> {
    const taskId = await this.#getTaskId();
    const resp = await runFilesystemExec(this.#client.cpClient, {
      fileOpenRequest: {
        path,
        mode,
      },
      taskId,
    });
    // For Open request, the file descriptor is always set
    const fileDescriptor = resp.response.fileDescriptor as string;
    return new SandboxFile(this.#client, fileDescriptor, taskId);
  }

  async exec(
    command: string[],
    params?: SandboxExecParams & { mode?: "text" },
  ): Promise<ContainerProcess<string>>;

  async exec(
    command: string[],
    params: SandboxExecParams & { mode: "binary" },
  ): Promise<ContainerProcess<Uint8Array>>;

  async exec(
    command: string[],
    params?: SandboxExecParams,
  ): Promise<ContainerProcess> {
    const taskId = await this.#getTaskId();

    const mergedSecrets = await mergeEnvIntoSecrets(
      this.#client,
      params?.env,
      params?.secrets,
    );
    const mergedParams = {
      ...params,
      secrets: mergedSecrets,
      env: undefined, // setting env to undefined just to clarify it's not needed anymore
    };

    const req = await buildContainerExecRequestProto(
      taskId,
      command,
      mergedParams,
    );
    const resp = await this.#client.cpClient.containerExec(req);

    return new ContainerProcess(this.#client, resp.execId, params);
  }

  async #getTaskId(): Promise<string> {
    if (this.#taskId === undefined) {
      const resp = await this.#client.cpClient.sandboxGetTaskId({
        sandboxId: this.sandboxId,
      });
      if (!resp.taskId) {
        throw new Error(
          `Sandbox ${this.sandboxId} does not have a task ID. It may not be running.`,
        );
      }
      if (resp.taskResult) {
        throw new Error(
          `Sandbox ${this.sandboxId} has already completed with result: ${resp.taskResult}`,
        );
      }
      this.#taskId = resp.taskId;
    }
    return this.#taskId;
  }

  async terminate(): Promise<void> {
    await this.#client.cpClient.sandboxTerminate({ sandboxId: this.sandboxId });
    this.#taskId = undefined; // Reset task ID after termination
  }

  async wait(): Promise<number> {
    while (true) {
      const resp = await this.#client.cpClient.sandboxWait({
        sandboxId: this.sandboxId,
        timeout: 10,
      });
      if (resp.result) {
        return Sandbox.#getReturnCode(resp.result)!;
      }
    }
  }

  /** Get {@link Tunnel} metadata for the Sandbox.
   *
   * Raises {@link SandboxTimeoutError} if the tunnels are not available after the timeout.
   *
   * @returns A dictionary of {@link Tunnel} objects which are keyed by the container port.
   */
  async tunnels(timeoutMs = 50000): Promise<Record<number, Tunnel>> {
    if (this.#tunnels) {
      return this.#tunnels;
    }

    const resp = await this.#client.cpClient.sandboxGetTunnels({
      sandboxId: this.sandboxId,
      timeout: timeoutMs / 1000,
    });

    if (
      resp.result?.status === GenericResult_GenericStatus.GENERIC_STATUS_TIMEOUT
    ) {
      throw new SandboxTimeoutError();
    }

    this.#tunnels = {};
    for (const t of resp.tunnels) {
      this.#tunnels[t.containerPort] = new Tunnel(
        t.host,
        t.port,
        t.unencryptedHost,
        t.unencryptedPort,
      );
    }

    return this.#tunnels;
  }

  /**
   * Snapshot the filesystem of the Sandbox.
   *
   * Returns an {@link Image} object which can be used to spawn a new Sandbox with the same filesystem.
   *
   * @param timeoutMs - Timeout for the snapshot operation in milliseconds
   * @returns Promise that resolves to an {@link Image}
   */
  async snapshotFilesystem(timeoutMs = 55000): Promise<Image> {
    const resp = await this.#client.cpClient.sandboxSnapshotFs({
      sandboxId: this.sandboxId,
      timeout: timeoutMs / 1000,
    });

    if (
      resp.result?.status !== GenericResult_GenericStatus.GENERIC_STATUS_SUCCESS
    ) {
      throw new Error(
        `Sandbox snapshot failed: ${resp.result?.exception || "Unknown error"}`,
      );
    }

    if (!resp.imageId) {
      throw new Error("Sandbox snapshot response missing `imageId`");
    }

    return new Image(this.#client, resp.imageId, "");
  }

  /**
   * Check if the Sandbox has finished running.
   *
   * Returns `null` if the Sandbox is still running, else returns the exit code.
   */
  async poll(): Promise<number | null> {
    const resp = await this.#client.cpClient.sandboxWait({
      sandboxId: this.sandboxId,
      timeout: 0,
    });

    return Sandbox.#getReturnCode(resp.result);
  }

  /**
   * @deprecated Use {@link SandboxService#list client.sandboxes.list()} instead.
   */
  static async *list(
    params: SandboxListParams = {},
  ): AsyncGenerator<Sandbox, void, unknown> {
    yield* getDefaultClient().sandboxes.list(params);
  }

  static #getReturnCode(result: GenericResult | undefined): number | null {
    if (
      result === undefined ||
      result.status === GenericResult_GenericStatus.GENERIC_STATUS_UNSPECIFIED
    ) {
      return null;
    }

    // Statuses are converted to exitcodes so we can conform to subprocess API.
    if (result.status === GenericResult_GenericStatus.GENERIC_STATUS_TIMEOUT) {
      return 124;
    } else if (
      result.status === GenericResult_GenericStatus.GENERIC_STATUS_TERMINATED
    ) {
      return 137;
    } else {
      return result.exitcode;
    }
  }
}

export class ContainerProcess<R extends string | Uint8Array = any> {
  stdin: ModalWriteStream<R>;
  stdout: ModalReadStream<R>;
  stderr: ModalReadStream<R>;
  returncode: number | null = null;

  readonly #client: ModalClient;
  readonly #execId: string;

  constructor(client: ModalClient, execId: string, params?: SandboxExecParams) {
    this.#client = client;
    const mode = params?.mode ?? "text";
    const stdout = params?.stdout ?? "pipe";
    const stderr = params?.stderr ?? "pipe";

    this.#execId = execId;

    this.stdin = toModalWriteStream(inputStreamCp<R>(client.cpClient, execId));

    let stdoutStream = streamConsumingIter(
      outputStreamCp(
        client.cpClient,
        execId,
        FileDescriptor.FILE_DESCRIPTOR_STDOUT,
      ),
    );
    if (stdout === "ignore") {
      stdoutStream.cancel();
      stdoutStream = ReadableStream.from([]);
    }

    let stderrStream = streamConsumingIter(
      outputStreamCp(
        client.cpClient,
        execId,
        FileDescriptor.FILE_DESCRIPTOR_STDERR,
      ),
    );
    if (stderr === "ignore") {
      stderrStream.cancel();
      stderrStream = ReadableStream.from([]);
    }

    if (mode === "text") {
      this.stdout = toModalReadStream(
        stdoutStream.pipeThrough(new TextDecoderStream()),
      ) as ModalReadStream<R>;
      this.stderr = toModalReadStream(
        stderrStream.pipeThrough(new TextDecoderStream()),
      ) as ModalReadStream<R>;
    } else {
      this.stdout = toModalReadStream(stdoutStream) as ModalReadStream<R>;
      this.stderr = toModalReadStream(stderrStream) as ModalReadStream<R>;
    }
  }

  /** Wait for process completion and return the exit code. */
  async wait(): Promise<number> {
    while (true) {
      const resp = await this.#client.cpClient.containerExecWait({
        execId: this.#execId,
        timeout: 55,
      });
      if (resp.completed) {
        return resp.exitCode ?? 0;
      }
    }
  }
}

// Like _StreamReader with object_type == "sandbox".
async function* outputStreamSb(
  cpClient: ModalGrpcClient,
  sandboxId: string,
  fileDescriptor: FileDescriptor,
): AsyncIterable<Uint8Array> {
  let lastIndex = "0-0";
  let completed = false;
  let retries = 10;
  while (!completed) {
    try {
      const outputIterator = cpClient.sandboxGetLogs({
        sandboxId,
        fileDescriptor,
        timeout: 55,
        lastEntryId: lastIndex,
      });
      for await (const batch of outputIterator) {
        lastIndex = batch.entryId;
        yield* batch.items.map((item) => new TextEncoder().encode(item.data));
        if (batch.eof) {
          completed = true;
          break;
        }
      }
    } catch (err) {
      if (isRetryableGrpc(err) && retries > 0) retries--;
      else throw err;
    }
  }
}

// Like _StreamReader with object_type == "container_process".
async function* outputStreamCp(
  cpClient: ModalGrpcClient,
  execId: string,
  fileDescriptor: FileDescriptor,
): AsyncIterable<Uint8Array> {
  let lastIndex = 0;
  let completed = false;
  let retries = 10;
  while (!completed) {
    try {
      const outputIterator = cpClient.containerExecGetOutput({
        execId,
        fileDescriptor,
        timeout: 55,
        getRawBytes: true,
        lastBatchIndex: lastIndex,
      });
      for await (const batch of outputIterator) {
        lastIndex = batch.batchIndex;
        yield* batch.items.map((item) => item.messageBytes);
        if (batch.exitCode !== undefined) {
          // The container process exited. Python code also doesn't handle this
          // exit code, so we don't either right now.
          completed = true;
          break;
        }
      }
    } catch (err) {
      if (isRetryableGrpc(err) && retries > 0) retries--;
      else throw err;
    }
  }
}

function inputStreamSb(
  cpClient: ModalGrpcClient,
  sandboxId: string,
): WritableStream<string> {
  let index = 1;
  return new WritableStream<string>({
    async write(chunk) {
      await cpClient.sandboxStdinWrite({
        sandboxId,
        input: encodeIfString(chunk),
        index,
      });
      index++;
    },
    async close() {
      await cpClient.sandboxStdinWrite({
        sandboxId,
        index,
        eof: true,
      });
    },
  });
}

function inputStreamCp<R extends string | Uint8Array>(
  cpClient: ModalGrpcClient,
  execId: string,
): WritableStream<R> {
  let messageIndex = 1;
  return new WritableStream<R>({
    async write(chunk) {
      await cpClient.containerExecPutInput({
        execId,
        input: {
          message: encodeIfString(chunk),
          messageIndex,
        },
      });
      messageIndex++;
    },
    async close() {
      await cpClient.containerExecPutInput({
        execId,
        input: {
          messageIndex,
          eof: true,
        },
      });
    },
  });
}

function encodeIfString(chunk: Uint8Array | string): Uint8Array {
  return typeof chunk === "string" ? new TextEncoder().encode(chunk) : chunk;
}
