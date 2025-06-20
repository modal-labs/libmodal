import { App } from "modal";

/**
 * Example demonstrating filesystem operations in a Modal sandbox.
 *
 * This example shows how to:
 * - Open files for reading and writing
 * - Read file contents as text and binary data
 * - Write data to files
 * - Seek to specific positions
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
  await writeHandle.write("Hello, Modal filesystem!\n");
  await writeHandle.write("This is line 2.\n");
  await writeHandle.write("And this is line 3.\n");
  await writeHandle.close();

  // Read the entire file as text
  const readHandle = await sb.open("/tmp/example.txt", "r");
  const content = await readHandle.read({ encoding: "utf8" });
  console.log("File content:", content);
  await readHandle.close();

  // Read specific number of bytes as text
  const partialHandle = await sb.open("/tmp/example.txt", "r");
  const firstLine = await partialHandle.read({ length: 25, encoding: "utf8" });
  console.log("First line:", firstLine);
  await partialHandle.close();

  // Append to the file
  const appendHandle = await sb.open("/tmp/example.txt", "a");
  await appendHandle.write("This line was appended.\n");
  await appendHandle.close();

  // Read from a specific position
  const seekHandle = await sb.open("/tmp/example.txt", "r");
  await seekHandle.seek(10); // Skip first 10 bytes
  const fromPosition = await seekHandle.read({ length: 15, encoding: "utf8" });
  console.log("From position 10:", fromPosition);
  await seekHandle.close();

  // Binary file operations
  const binaryHandle = await sb.open("/tmp/data.bin", "w");
  const binaryData = new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);
  await binaryHandle.write(binaryData);
  await binaryHandle.close();

  // Read binary data (default encoding is 'binary')
  const readBinaryHandle = await sb.open("/tmp/data.bin", "r");
  const readData = await readBinaryHandle.read({ encoding: "binary" });
  console.log("Binary data:", Array.from(readData));
  await readBinaryHandle.close();

  // Demonstrate different encoding options
  const textHandle1 = await sb.open("/tmp/example.txt", "r");
  const textHandle2 = await sb.open("/tmp/example.txt", "r");

  // Read as binary (Uint8Array)
  const binaryContent = await textHandle1.read({ encoding: "binary" });
  console.log("Binary content length:", binaryContent.length);

  // Read as text (string)
  const textContent = await textHandle2.read({ encoding: "utf8" });
  console.log("Text content length:", textContent.length);

  await textHandle1.close();
  await textHandle2.close();
} catch (error) {
  console.error("Filesystem operation failed:", error);
} finally {
  // Clean up the sandbox
  await sb.terminate();
}
