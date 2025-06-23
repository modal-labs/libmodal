import { seekWhenceFromJSON } from "../proto/modal_proto/api";
import { client, isRetryableGrpc } from "./client";
import { InvalidError } from "./errors";

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
 * Provides read/write operations similar to Node.js fs.promises.FileHandle.
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

    const resp = await client.containerFilesystemExec({
      fileReadRequest: {
        fileDescriptor: this.#fileDescriptor,
        n: options?.length ?? undefined,
      },
      taskId: this.#taskId,
    });

    let retries = 10;
    let completed = false;
    const chunks: Uint8Array[] = [];

    while (!completed) {
      chunks.length = 0;
      try {
        const outputIterator = client.containerFilesystemExecGetOutput({
          execId: resp.execId,
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
            } else {
              throw new Error(batch.error.errorMessage);
            }
          }
        }
      } catch (err) {
        if (isRetryableGrpc(err) && retries > 0) {
          retries--;
        } else throw err;
      }
    }

    // Concatenate all chunks into a single Uint8Array
    const totalLength = chunks.reduce((acc, chunk) => acc + chunk.length, 0);
    const result = new Uint8Array(totalLength);
    let offset = 0;
    for (const chunk of chunks) {
      result.set(chunk, offset);
      offset += chunk.length;
    }

    // Return text or binary based on encoding option
    if (is_utf8) {
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

    const req = await client.containerFilesystemExec({
      fileWriteRequest: {
        fileDescriptor: this.#fileDescriptor,
        data: bytes,
      },
      taskId: this.#taskId,
    });
    await waitContainerFilesystemExec(req.execId);
  }

  /**
   * Seek to a specific position in the file.
   * @param offset - Offset to seek to
   * @param whence - 0 for aboslute file positioning, 1 for relative to the current position and 2 for relative to the file's end.
   */
  async #seek(offset: number, whence: number = 0): Promise<void> {
    const req = await client.containerFilesystemExec({
      fileSeekRequest: {
        fileDescriptor: this.#fileDescriptor,
        offset: offset,
        whence: seekWhenceFromJSON(whence),
      },
      taskId: this.#taskId,
    });
    await waitContainerFilesystemExec(req.execId);
  }

  /**
   * Flush any buffered data to the file.
   */
  async flush(): Promise<void> {
    const req = await client.containerFilesystemExec({
      fileFlushRequest: {
        fileDescriptor: this.#fileDescriptor,
      },
      taskId: this.#taskId,
    });
    await waitContainerFilesystemExec(req.execId);
  }

  /**
   * Close the file handle.
   */
  async close(): Promise<void> {
    const req = await client.containerFilesystemExec({
      fileCloseRequest: {
        fileDescriptor: this.#fileDescriptor,
      },
      taskId: this.#taskId,
    });
    await waitContainerFilesystemExec(req.execId);
  }
}

export async function waitContainerFilesystemExec(
  execId: string,
): Promise<void> {
  let retries = 10;
  let completed = false;
  while (!completed) {
    try {
      const outputIterator = client.containerFilesystemExecGetOutput({
        execId,
        timeout: 55,
      });
      for await (const batch of outputIterator) {
        if (batch.eof) {
          completed = true;
          break;
        }
        if (batch.error !== undefined) {
          if (retries > 0) {
            retries--;
            break;
          } else {
            throw new Error(batch.error.errorMessage);
          }
        }
      }
    } catch (err) {
      if (isRetryableGrpc(err) && retries > 0) {
        retries--;
      } else throw err;
    }
  }
}
