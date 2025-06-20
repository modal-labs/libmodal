import { SeekWhence } from "../proto/modal_proto/api";
import { client, isRetryableGrpc } from "./client";

/** File open modes supported by the filesystem API. */
export type FileMode = "r" | "w" | "a" | "r+" | "w+" | "a+";

/** Options for file operations. */
export type FileOptions = {
  /** Number of bytes to read. If not specified, reads until end of file. */
  length?: number;
  /** Position to seek to before reading/writing. */
  position?: number;
  /** Encoding for text operations. Defaults to 'binary' for Uint8Array. */
  encoding?: "utf8" | "binary";
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
  async read(options: FileOptions & { encoding: "utf8" }): Promise<string>;
  async read(
    options: FileOptions & { encoding?: "binary" },
  ): Promise<Uint8Array>;
  async read(options?: FileOptions): Promise<Uint8Array | string> {
    // Handle position seeking if specified
    if (options?.position !== undefined) {
      await this.seek(options.position);
    }

    const resp = await client.containerFilesystemExec({
      fileReadRequest: {
        fileDescriptor: this.#fileDescriptor,
        n: options?.length ?? undefined,
      },
      taskId: this.#taskId,
    });
    await this._wait(resp.execId);

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
    if (options?.encoding === "utf8") {
      return new TextDecoder().decode(result);
    }
    return result;
  }

  /**
   * Write data to the file.
   * @param data - Data to write (string or Uint8Array)
   * @param options - Options for the write operation
   */
  async write(data: string | Uint8Array, options?: FileOptions): Promise<void> {
    // Handle position seeking if specified
    if (options?.position !== undefined) {
      await this.seek(options.position);
    }

    const bytes =
      typeof data === "string" ? new TextEncoder().encode(data) : data;

    const req = await client.containerFilesystemExec({
      fileWriteRequest: {
        fileDescriptor: this.#fileDescriptor,
        data: bytes,
      },
      taskId: this.#taskId,
    });
    await this._wait(req.execId);
  }

  /**
   * Seek to a specific position in the file.
   * @param offset - Offset to seek to
   */
  async seek(offset: number): Promise<void> {
    const req = await client.containerFilesystemExec({
      fileSeekRequest: {
        fileDescriptor: this.#fileDescriptor,
        offset: offset,
        whence: SeekWhence.SEEK_SET,
      },
      taskId: this.#taskId,
    });
    await this._wait(req.execId);
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

    await this._wait(req.execId);
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
    await this._wait(req.execId);
  }

  async _wait(execId: string): Promise<void> {
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
        }
      } catch (err) {
        if (isRetryableGrpc(err) && retries > 0) retries--;
        else throw err;
      }
    }
  }
}
