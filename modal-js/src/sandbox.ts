import { ClientError, Status } from "nice-grpc";
import {
  FileDescriptor,
  GenericResult,
  GenericResult_GenericStatus,
} from "../proto/modal_proto/api";
import { client, isRetryableGrpc } from "./client";
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
import { type Secret } from "./secret";
import { InvalidError, NotFoundError, SandboxTimeoutError } from "./errors";
import { Image } from "./image";

/**
 * Stdin is always present, but this option allow you to drop stdout or stderr
 * if you don't need them. The default is "pipe", matching Node.js behavior.
 *
 * If behavior is set to "ignore", the output streams will be empty.
 */
export type StdioBehavior = "pipe" | "ignore";

/**
 * Specifies the type of data that will be read from the sandbox or container
 * process. "text" means the data will be read as UTF-8 text, while "binary"
 * means the data will be read as raw bytes (Uint8Array).
 */
export type StreamMode = "text" | "binary";

/** Options to configure a `Sandbox.exec()` operation. */
export type ExecOptions = {
  /** Specifies text or binary encoding for input and output streams. */
  mode?: StreamMode;
  /** Whether to pipe or ignore standard output. */
  stdout?: StdioBehavior;
  /** Whether to pipe or ignore standard error. */
  stderr?: StdioBehavior;
  /** Working directory to run the command in. */
  workdir?: string;
  /** Timeout for the process in milliseconds. Defaults to 0 (no timeout). */
  timeout?: number;
  /** Secrets with environment variables for the command. */
  secrets?: [Secret];
};

/** A port forwarded from within a running Modal sandbox. */
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

/** Sandboxes are secure, isolated containers in Modal that boot in seconds. */
export class Sandbox {
  readonly sandboxId: string;
  stdin: ModalWriteStream<string>;
  stdout: ModalReadStream<string>;
  stderr: ModalReadStream<string>;

  #taskId: string | undefined;
  #tunnels: Record<number, Tunnel> | undefined;

  /** @ignore */
  constructor(sandboxId: string) {
    this.sandboxId = sandboxId;

    this.stdin = toModalWriteStream(inputStreamSb(sandboxId));
    this.stdout = toModalReadStream(
      streamConsumingIter(
        outputStreamSb(sandboxId, FileDescriptor.FILE_DESCRIPTOR_STDOUT),
      ).pipeThrough(new TextDecoderStream()),
    );
    this.stderr = toModalReadStream(
      streamConsumingIter(
        outputStreamSb(sandboxId, FileDescriptor.FILE_DESCRIPTOR_STDERR),
      ).pipeThrough(new TextDecoderStream()),
    );
  }

  /** Returns a running Sandbox object from an ID.
   *
   * @returns Sandbox with ID
   */
  static async fromId(sandboxId: string): Promise<Sandbox> {
    try {
      await client.sandboxWait({
        sandboxId,
        timeout: 0,
      });
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.NOT_FOUND)
        throw new NotFoundError(`Sandbox with id: '${sandboxId}' not found`);
      throw err;
    }

    return new Sandbox(sandboxId);
  }

  /**
   * Open a file in the sandbox filesystem.
   * @param path - Path to the file to open
   * @param mode - File open mode (r, w, a, r+, w+, a+)
   * @returns Promise that resolves to a SandboxFile
   */
  async open(path: string, mode: SandboxFileMode = "r"): Promise<SandboxFile> {
    const taskId = await this.#getTaskId();
    const resp = await runFilesystemExec({
      fileOpenRequest: {
        path,
        mode,
      },
      taskId,
    });
    // For Open request, the file descriptor is always set
    const fileDescriptor = resp.response.fileDescriptor as string;
    return new SandboxFile(fileDescriptor, taskId);
  }

  async exec(
    command: string[],
    options?: ExecOptions & { mode?: "text" },
  ): Promise<ContainerProcess<string>>;

  async exec(
    command: string[],
    options: ExecOptions & { mode: "binary" },
  ): Promise<ContainerProcess<Uint8Array>>;

  async exec(
    command: string[],
    options?: {
      mode?: StreamMode;
      stdout?: StdioBehavior;
      stderr?: StdioBehavior;
      workdir?: string;
      timeout?: number;
      secrets?: [Secret];
    },
  ): Promise<ContainerProcess> {
    const taskId = await this.#getTaskId();

    const secretIds = options?.secrets
      ? options.secrets.map((secret) => secret.secretId)
      : [];

    const resp = await client.containerExec({
      taskId,
      command,
      workdir: options?.workdir,
      timeoutSecs: options?.timeout ? options.timeout / 1000 : 0,
      secretIds,
    });

    return new ContainerProcess(resp.execId, options);
  }

  async #getTaskId(): Promise<string> {
    if (this.#taskId === undefined) {
      const resp = await client.sandboxGetTaskId({
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
    await client.sandboxTerminate({ sandboxId: this.sandboxId });
    this.#taskId = undefined; // Reset task ID after termination
  }

  async wait(): Promise<number> {
    while (true) {
      const resp = await client.sandboxWait({
        sandboxId: this.sandboxId,
        timeout: 55,
      });
      if (resp.result) {
        return Sandbox.#getReturnCode(resp.result)!;
      }
    }
  }

  /** Get Tunnel metadata for the sandbox.
   *
   * Raises `SandboxTimeoutError` if the tunnels are not available after the timeout.
   *
   * @returns A dictionary of Tunnel objects which are keyed by the container port.
   */
  async tunnels(timeout = 50000): Promise<Record<number, Tunnel>> {
    if (this.#tunnels) {
      return this.#tunnels;
    }

    const resp = await client.sandboxGetTunnels({
      sandboxId: this.sandboxId,
      timeout: timeout / 1000, // Convert to seconds
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
   * Returns an `Image` object which can be used to spawn a new Sandbox with the same filesystem.
   *
   * @param timeout - Timeout for the snapshot operation in milliseconds
   * @returns Promise that resolves to an Image
   */
  async snapshotFilesystem(timeout = 55000): Promise<Image> {
    const resp = await client.sandboxSnapshotFs({
      sandboxId: this.sandboxId,
      timeout: timeout / 1000,
    });

    if (
      resp.result?.status !== GenericResult_GenericStatus.GENERIC_STATUS_SUCCESS
    ) {
      throw new Error(
        `Sandbox snapshot failed: ${resp.result?.exception || "Unknown error"}`,
      );
    }

    if (!resp.imageId) {
      throw new Error("Sandbox snapshot response missing image ID");
    }

    return new Image(resp.imageId);
  }

  /**
   * Check if the Sandbox has finished running.
   *
   * Returns `null` if the Sandbox is still running, else returns the exit code.
   */
  async poll(): Promise<number | null> {
    const resp = await client.sandboxWait({
      sandboxId: this.sandboxId,
      timeout: 0,
    });

    return Sandbox.#getReturnCode(resp.result);
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

  readonly #execId: string;

  constructor(execId: string, options?: ExecOptions) {
    const mode = options?.mode ?? "text";
    const stdout = options?.stdout ?? "pipe";
    const stderr = options?.stderr ?? "pipe";

    this.#execId = execId;

    this.stdin = toModalWriteStream(inputStreamCp<R>(execId));

    let stdoutStream = streamConsumingIter(
      outputStreamCp(execId, FileDescriptor.FILE_DESCRIPTOR_STDOUT),
    );
    if (stdout === "ignore") {
      stdoutStream.cancel();
      stdoutStream = ReadableStream.from([]);
    }

    let stderrStream = streamConsumingIter(
      outputStreamCp(execId, FileDescriptor.FILE_DESCRIPTOR_STDERR),
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
      const resp = await client.containerExecWait({
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
  sandboxId: string,
  fileDescriptor: FileDescriptor,
): AsyncIterable<Uint8Array> {
  let lastIndex = "0-0";
  let completed = false;
  let retries = 10;
  while (!completed) {
    try {
      const outputIterator = client.sandboxGetLogs({
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
  execId: string,
  fileDescriptor: FileDescriptor,
): AsyncIterable<Uint8Array> {
  let lastIndex = 0;
  let completed = false;
  let retries = 10;
  while (!completed) {
    try {
      const outputIterator = client.containerExecGetOutput({
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

function inputStreamSb(sandboxId: string): WritableStream<string> {
  let index = 1;
  return new WritableStream<string>({
    async write(chunk) {
      await client.sandboxStdinWrite({
        sandboxId,
        input: encodeIfString(chunk),
        index,
      });
      index++;
    },
    async close() {
      await client.sandboxStdinWrite({
        sandboxId,
        index,
        eof: true,
      });
    },
  });
}

function inputStreamCp<R extends string | Uint8Array>(
  execId: string,
): WritableStream<R> {
  let messageIndex = 1;
  return new WritableStream<R>({
    async write(chunk) {
      await client.containerExecPutInput({
        execId,
        input: {
          message: encodeIfString(chunk),
          messageIndex,
        },
      });
      messageIndex++;
    },
    async close() {
      await client.containerExecPutInput({
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
