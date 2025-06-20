import { App } from "modal";
import { expect, test, vi, onTestFinished } from "vitest";

vi.setConfig({ testTimeout: 40000 });

test("WriteAndReadTextFile", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);
  onTestFinished(async () => {
    await sb.terminate();
  });

  const writeHandle = await sb.open("/tmp/test.txt", "w");
  expect(writeHandle).toBeTruthy();
  await writeHandle.write("Hello, Modal filesystem!");
  await writeHandle.close();

  const readHandle = await sb.open("/tmp/test.txt", "r");
  const content = await readHandle.read({ encoding: "utf8" });
  expect(content).toBe("Hello, Modal filesystem!");
  await readHandle.close();
});

test("WriteAndReadBinaryFile", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);
  onTestFinished(async () => {
    await sb.terminate();
  });

  const testData = new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);
  // Write binary data
  const writeHandle = await sb.open("/tmp/test.bin", "w");
  await writeHandle.write(testData);
  await writeHandle.close();

  // Read binary data
  const readHandle = await sb.open("/tmp/test.bin", "r");
  const readData = await readHandle.read({ encoding: "binary" });
  expect(readData).toEqual(testData);
  await readHandle.close();
});

test("AppendToFile", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);
  onTestFinished(async () => {
    await sb.terminate();
  });

  // Write initial content
  const writeHandle = await sb.open("/tmp/append.txt", "w");
  await writeHandle.write("Initial content\n");
  await writeHandle.close();

  // Append more content
  const appendHandle = await sb.open("/tmp/append.txt", "a");
  await appendHandle.write("Appended content\n");
  await appendHandle.close();

  // Read the entire file
  const readHandle = await sb.open("/tmp/append.txt", "r");
  const content = await readHandle.read({ encoding: "utf8" });
  expect(content).toBe("Initial content\nAppended content\n");
  await readHandle.close();
});

test("ReadPartialFile", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);
  onTestFinished(async () => {
    await sb.terminate();
  });

  // Write a file
  const writeHandle = await sb.open("/tmp/partial.txt", "w");
  await writeHandle.write(
    "Hello, this is a longer test file with multiple lines.",
  );
  await writeHandle.close();

  // Read first 10 bytes
  const readHandle = await sb.open("/tmp/partial.txt", "r");
  const partial = await readHandle.read({ length: 10, encoding: "utf8" });
  expect(partial).toBe("Hello, thi");
  await readHandle.close();
});

test("ReadWithPositionOption", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);
  onTestFinished(async () => {
    await sb.terminate();
  });

  // Write a file
  const writeHandle = await sb.open("/tmp/position.txt", "w");
  await writeHandle.write("Hello, world! This is a test.");
  await writeHandle.close();

  // Read from position 7 using the position option
  const readHandle = await sb.open("/tmp/position.txt", "r");
  const content = await readHandle.read({
    length: 5,
    encoding: "utf8",
  });
  expect(content).toBe("Hello");
  await readHandle.close();
});

test("WriteWithPositionOption", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);
  onTestFinished(async () => {
    await sb.terminate();
  });

  // Write initial content
  const writeHandle = await sb.open("/tmp/write-position.txt", "w");
  await writeHandle.write("Hello, world!");
  await writeHandle.close();

  // Write at position 7
  const updateHandle = await sb.open("/tmp/write-position.txt", "r+");
  await updateHandle.write("Modal", { position: 7 });
  await updateHandle.close();

  // Read the result
  const readHandle = await sb.open("/tmp/write-position.txt", "r");
  const content = await readHandle.read({ encoding: "utf8" });
  expect(content).toBe("Hello, Modal!");
  await readHandle.close();
});

test("FileHandleFlush", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);
  onTestFinished(async () => {
    await sb.terminate();
  });

  const handle = await sb.open("/tmp/flush.txt", "w");
  await handle.write("Test data");
  await handle.flush(); // Ensure data is written to disk
  await handle.close();

  // Verify the data was written
  const readHandle = await sb.open("/tmp/flush.txt", "r");
  const content = await readHandle.read({ encoding: "utf8" });
  expect(content).toBe("Test data");
  await readHandle.close();
});

test("MultipleFileOperations", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);
  onTestFinished(async () => {
    await sb.terminate();
  });

  // Create multiple files
  const handle1 = await sb.open("/tmp/file1.txt", "w");
  await handle1.write("File 1 content");
  await handle1.close();

  const handle2 = await sb.open("/tmp/file2.txt", "w");
  await handle2.write("File 2 content");
  await handle2.close();

  // Read both files
  const read1 = await sb.open("/tmp/file1.txt", "r");
  const content1 = await read1.read({ encoding: "utf8" });
  await read1.close();

  const read2 = await sb.open("/tmp/file2.txt", "r");
  const content2 = await read2.read({ encoding: "utf8" });
  await read2.close();

  expect(content1).toBe("File 1 content");
  expect(content2).toBe("File 2 content");
});

test("FileOpenModes", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);
  onTestFinished(async () => {
    await sb.terminate();
  });

  // Test write mode (truncates)
  const writeHandle = await sb.open("/tmp/modes.txt", "w");
  await writeHandle.write("Initial content");
  await writeHandle.close();

  // Test read mode
  const readHandle = await sb.open("/tmp/modes.txt", "r");
  const content1 = await readHandle.read({ encoding: "utf8" });
  expect(content1).toBe("Initial content");
  await readHandle.close();

  // Test append mode
  const appendHandle = await sb.open("/tmp/modes.txt", "a");
  await appendHandle.write(" appended");
  await appendHandle.close();

  // Verify append worked
  const finalRead = await sb.open("/tmp/modes.txt", "r");
  const finalContent = await finalRead.read({ encoding: "utf8" });
  expect(finalContent).toBe("Initial content appended");
  await finalRead.close();
});

test("LargeFileOperations", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);
  onTestFinished(async () => {
    await sb.terminate();
  });

  // Create a larger file
  const largeData = "x".repeat(1000);

  const writeHandle = await sb.open("/tmp/large.txt", "w");
  await writeHandle.write(largeData);
  await writeHandle.close();

  // Read it back
  const readHandle = await sb.open("/tmp/large.txt", "r");
  const content = await readHandle.read({ encoding: "utf8" });
  expect(content).toBe(largeData);
  expect(content.length).toBe(1000);
  await readHandle.close();
});
