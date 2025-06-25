import {
  ContainerFilesystemExecRequest,
  DeepPartial,
  ContainerFilesystemExecResponse,
} from "../proto/modal_proto/api";
import { client, isRetryableGrpc } from "./client";
import { RemoteError } from "./errors";

/** File open modes supported by the filesystem API. */
export type FileMode = "r" | "w" | "a" | "r+" | "w+" | "a+";

export type ReadOptions = {
  /** Encoding for text operations. Defaults to 'utf8' for utf8 encoded data. */
  encoding?: "utf8" | "binary";
};

/**
 * FileHandle represents an open file in the sandbox filesystem.
 * Provides read/write operations similar to Node.js `fsPromises.FileHandle`.
 */
export class FileHandle {
  readonly #fileDescriptor: string;
  readonly #taskId: string;

  /** @ignore */
  constructor(fileDescriptor: string, taskId: string) {
    this.#fileDescriptor = fileDescriptor;
    this.#taskId = taskId;
  }

  /**
   * Read data from the file.
   * @returns Promise that resolves to the read data as Uint8Array or string
   */
  async read(): Promise<Uint8Array>;
  async read(options: ReadOptions & { encoding: "utf8" }): Promise<string>;
  async read(
    options: ReadOptions & { encoding: "binary" },
  ): Promise<Uint8Array>;
  async read(options?: ReadOptions): Promise<Uint8Array | string> {
    // Default encoding to 'utf8' if not specified
    const encoding = options?.encoding ?? "utf8";
    const is_utf8 = encoding === "utf8";

    const resp = await runFilesystemExec({
      fileReadRequest: {
        fileDescriptor: this.#fileDescriptor,
      },
      taskId: this.#taskId,
    });
    const chunks = resp.chunks;

    // Concatenate all chunks into a single Uint8Array
    const totalLength = chunks.reduce((acc, chunk) => acc + chunk.length, 0);
    const result = new Uint8Array(totalLength);
    let offset = 0;
    for (const chunk of chunks) {
      result.set(chunk, offset);
      offset += chunk.length;
    }

    if (is_utf8) {
      // At this point, we can assume that `position` and `length` is not set.
      return new TextDecoder().decode(result);
    }
    return result;
  }

  /**
   * Write data to the file.
   * @param data - Data to write (string or Uint8Array)
   */
  async write(data: string | Uint8Array): Promise<void> {
    // Handle position seeking if specified
    const is_utf8 = typeof data === "string";

    const bytes = is_utf8 ? new TextEncoder().encode(data) : data;

    await runFilesystemExec({
      fileWriteRequest: {
        fileDescriptor: this.#fileDescriptor,
        data: bytes,
      },
      taskId: this.#taskId,
    });
  }

  /**
   * Flush any buffered data to the file.
   */
  async flush(): Promise<void> {
    await runFilesystemExec({
      fileFlushRequest: {
        fileDescriptor: this.#fileDescriptor,
      },
      taskId: this.#taskId,
    });
  }

  /**
   * Close the file handle.
   */
  async close(): Promise<void> {
    await runFilesystemExec({
      fileCloseRequest: {
        fileDescriptor: this.#fileDescriptor,
      },
      taskId: this.#taskId,
    });
  }
}

export async function runFilesystemExec(
  request: DeepPartial<ContainerFilesystemExecRequest>,
): Promise<{
  chunks: Uint8Array[];
  response: ContainerFilesystemExecResponse;
}> {
  const response = await client.containerFilesystemExec(request);

  const chunks: Uint8Array[] = [];
  let retries = 10;
  let completed = false;
  while (!completed) {
    chunks.length = 0;
    try {
      const outputIterator = client.containerFilesystemExecGetOutput({
        execId: response.execId,
        timeout: 55,
      });
      for await (const batch of outputIterator) {
        chunks.push(...batch.output);
        if (batch.eof) {
          completed = true;
          break;
        }
        if (batch.error !== undefined) {
          if (retries > 0) {
            retries--;
            break;
          }
          throw new RemoteError(batch.error.errorMessage);
        }
      }
    } catch (err) {
      if (isRetryableGrpc(err) && retries > 0) {
        retries--;
      } else throw err;
    }
  }
  return { chunks, response };
}
