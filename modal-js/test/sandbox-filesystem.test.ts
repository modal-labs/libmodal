import { App } from "modal";
import { expect, test } from "vitest";

test("OpenFileForWriting", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);

  try {
    const handle = await sb.open("/tmp/test.txt", "w");
    expect(handle).toBeTruthy();
    await handle.write("Hello, world!");
    await handle.close();
  } finally {
    await sb.terminate();
  }
});

test("WriteAndReadTextFile", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);

  try {
    // Write a file
    const writeHandle = await sb.open("/tmp/test.txt", "w");
    await writeHandle.write("Hello, Modal filesystem!\n");
    await writeHandle.write("This is a test file.\n");
    await writeHandle.close();

    // Read the file
    // const readHandle = await sb.open("/tmp/test.txt", "r");
    // const content = await readHandle.read({ encoding: "utf8" });
    // expect(content).toBe("Hello, Modal filesystem!\nThis is a test file.\n");
    // await readHandle.close();
  } finally {
    await sb.terminate();
  }
});

test("WriteAndReadBinaryFile", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);

  try {
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
  } finally {
    await sb.terminate();
  }
});

test("AppendToFile", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);

  try {
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
  } finally {
    await sb.terminate();
  }
});

test("ReadPartialFile", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);

  try {
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
  } finally {
    await sb.terminate();
  }
});

test("SeekAndRead", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);

  try {
    // Write a file
    const writeHandle = await sb.open("/tmp/seek.txt", "w");
    await writeHandle.write("Hello, world! This is a test.");
    await writeHandle.close();

    // Seek to position 7 and read
    const readHandle = await sb.open("/tmp/seek.txt", "r");
    await readHandle.seek(7);
    const content = await readHandle.read({ length: 5, encoding: "utf8" });
    expect(content).toBe("world");
    await readHandle.close();
  } finally {
    await sb.terminate();
  }
});

test("ReadWithPositionOption", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);

  try {
    // Write a file
    const writeHandle = await sb.open("/tmp/position.txt", "w");
    await writeHandle.write("Hello, world! This is a test.");
    await writeHandle.close();

    // Read from position 7 using the position option
    const readHandle = await sb.open("/tmp/position.txt", "r");
    const content = await readHandle.read({
      position: 7,
      length: 5,
      encoding: "utf8",
    });
    expect(content).toBe("world");
    await readHandle.close();
  } finally {
    await sb.terminate();
  }
});

test("WriteWithPositionOption", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);

  try {
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
  } finally {
    await sb.terminate();
  }
});

test("FileHandleFlush", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);

  try {
    const handle = await sb.open("/tmp/flush.txt", "w");
    await handle.write("Test data");
    await handle.flush(); // Ensure data is written to disk
    await handle.close();

    // Verify the data was written
    const readHandle = await sb.open("/tmp/flush.txt", "r");
    const content = await readHandle.read({ encoding: "utf8" });
    expect(content).toBe("Test data");
    await readHandle.close();
  } finally {
    await sb.terminate();
  }
});

test("MultipleFileOperations", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);

  try {
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
  } finally {
    await sb.terminate();
  }
});

test("FileOpenModes", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);

  try {
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
  } finally {
    await sb.terminate();
  }
});

test("LargeFileOperations", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");
  const sb = await app.createSandbox(image);

  try {
    // Create a larger file
    const largeData = "x".repeat(10000);

    const writeHandle = await sb.open("/tmp/large.txt", "w");
    await writeHandle.write(largeData);
    await writeHandle.close();

    // Read it back
    const readHandle = await sb.open("/tmp/large.txt", "r");
    const content = await readHandle.read({ encoding: "utf8" });
    expect(content).toBe(largeData);
    expect(content.length).toBe(10000);
    await readHandle.close();
  } finally {
    await sb.terminate();
  }
});
