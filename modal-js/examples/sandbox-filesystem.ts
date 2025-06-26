import { App } from "modal";

/**
 * Example demonstrating filesystem operations in a Modal sandbox.
 *
 * This example shows how to:
 * - Open files for reading and writing
 * - Read file contents as binary data
 * - Write data to files
 * - Close file handles
 */

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await app.imageFromRegistry("alpine:3.21");

// Create a sandbox
const sb = await app.createSandbox(image);
console.log("Started sandbox:", sb.sandboxId);

try {
  // Write a file
  const writeHandle = await sb.open("/tmp/example.txt", "w");
  const encoder = new TextEncoder();
  const deocder = new TextDecoder();

  await writeHandle.write(encoder.encode("Hello, Modal filesystem!\n"));
  await writeHandle.write(encoder.encode("This is line 2.\n"));
  await writeHandle.write(encoder.encode("And this is line 3.\n"));
  await writeHandle.close();

  // Read the entire file as binary
  const readHandle = await sb.open("/tmp/example.txt", "r");
  const content = await readHandle.read();
  console.log("File content:", deocder.decode(content));
  await readHandle.close();

  // Append to the file
  const appendHandle = await sb.open("/tmp/example.txt", "a");
  await appendHandle.write(encoder.encode("This line was appended.\n"));
  await appendHandle.close();

  // Read with binary
  const seekHandle = await sb.open("/tmp/example.txt", "r");
  const appendedContent = await seekHandle.read();
  console.log("File with appended:", deocder.decode(appendedContent));
  await seekHandle.close();

  // Binary file operations
  const binaryHandle = await sb.open("/tmp/data.bin", "w");
  const binaryData = new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);
  await binaryHandle.write(binaryData);
  await binaryHandle.close();

  // Read binary data
  const readBinaryHandle = await sb.open("/tmp/data.bin", "r");
  const readData = await readBinaryHandle.read();
  console.log("Binary data:", readData);
  await readBinaryHandle.close();
} catch (error) {
  console.error("Filesystem operation failed:", error);
} finally {
  // Clean up the sandbox
  await sb.terminate();
}
