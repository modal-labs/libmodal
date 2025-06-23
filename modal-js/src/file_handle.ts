import {
  SeekWhence,
  ContainerFilesystemExecRequest,
  DeepPartial,
  ContainerFilesystemExecResponse,
} from "../proto/modal_proto/api";
import { client, isRetryableGrpc } from "./client";
import { InvalidError, RemoteError } from "./errors";

/** File open modes supported by the filesystem API. */
export type FileMode = "r" | "w" | "a" | "r+" | "w+" | "a+";

export type ReadOptions = {
  /** Number of bytes to read. If not specified, reads until end of file. */
  length?: number;
  /** Position to seek to before reading. */
  position?: number;
  /** Encoding for text operations. Defaults to 'binary' for Uint8Array. */
  encoding?: "utf8" | "binary";
};

export type WriteOptions = {
  /** Position to seek to before writing. */
  position?: number;
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
    // Handle position seeking if specified
    const is_utf8 = options?.encoding === "utf8";
    if (options?.position !== undefined) {
      if (is_utf8)
        throw new InvalidError(
          "position can only be set if encoding is 'utf8'",
        );
      await this.#seek(options.position);
    }

    if (options?.length !== undefined && is_utf8) {
      throw new InvalidError("length can only be set if encoding is 'utf8'");
    }

    const resp = await runFilesystemExec({
      fileReadRequest: {
        fileDescriptor: this.#fileDescriptor,
        n: options?.length ?? undefined,
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
   * @param options - Options for the write operation
   */
  async write(
    data: string | Uint8Array,
    options?: WriteOptions,
  ): Promise<void> {
    // Handle position seeking if specified
    const is_utf8 = typeof data === "string";
    if (options?.position !== undefined) {
      if (is_utf8)
        throw new InvalidError("Position can only be set if with binary data");
      await this.#seek(options.position);
    }

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
   * Seek to a specific position in the file.
   * @param offset - Offset to seek to
   */
  async #seek(offset: number): Promise<void> {
    await runFilesystemExec({
      fileSeekRequest: {
        fileDescriptor: this.#fileDescriptor,
        offset: offset,
        whence: SeekWhence.SEEK_SET,
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
