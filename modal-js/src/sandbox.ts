import { FileDescriptor } from "../proto/modal_proto/api";
import { client, isRetryableGrpc } from "./client";
import type { ModalReadStream, ModalWriteStream } from "./streams";
import {
  streamConsumingIter,
  toModalReadStream,
  toModalWriteStream,
} from "./streams";

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
  mode?: StreamMode;
  stdout?: StdioBehavior;
  stderr?: StdioBehavior;
};

/** Sandboxes are secure, isolated containers in Modal that boot in seconds. */
export class Sandbox {
  readonly sandboxId: string;
  stdin: ModalWriteStream<string>;
  stdout: ModalReadStream<string>;
  stderr: ModalReadStream<string>;

  #taskId: string | undefined;

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
    },
  ): Promise<ContainerProcess> {
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

    const resp = await client.containerExec({
      taskId: this.#taskId,
      command,
    });

    return new ContainerProcess(resp.execId, options);
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
        return resp.result.exitcode;
      }
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
