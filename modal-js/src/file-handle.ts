import { SeekWhence } from "../proto/modal_proto/api";
import { client } from "./client";

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
   * @param options - Options for the read operation
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

    const request: any = {
      fileReadRequest: {
        fileDescriptor: this.#fileDescriptor,
      },
      taskId: this.#taskId,
    };

    if (options?.length !== undefined) {
      request.fileReadRequest.n = options.length;
    }

    const resp = await client.containerFilesystemExec(request);

    // Get the output stream to read the actual data
    const outputIterator = client.containerFilesystemExecGetOutput({
      execId: resp.execId,
      timeout: 55,
    });

    const chunks: Uint8Array[] = [];
    for await (const batch of outputIterator) {
      chunks.push(...batch.output);
      if (batch.eof) break;
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

    const request: any = {
      fileWriteRequest: {
        fileDescriptor: this.#fileDescriptor,
        data: bytes,
      },
      taskId: this.#taskId,
    };

    await client.containerFilesystemExec(request);
  }

  /**
   * Seek to a specific position in the file.
   * @param offset - Offset to seek to
   * @param whence - Where to seek from (SEEK_SET, SEEK_CUR, SEEK_END)
   */
  async seek(
    offset: number,
    whence: SeekWhence = SeekWhence.SEEK_SET,
  ): Promise<void> {
    await client.containerFilesystemExec({
      fileSeekRequest: {
        fileDescriptor: this.#fileDescriptor,
        offset,
        whence,
      },
      taskId: this.#taskId,
    });
  }

  /**
   * Flush any buffered data to the file.
   */
  async flush(): Promise<void> {
    await client.containerFilesystemExec({
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
    await client.containerFilesystemExec({
      fileCloseRequest: {
        fileDescriptor: this.#fileDescriptor,
      },
      taskId: this.#taskId,
    });
  }
}
