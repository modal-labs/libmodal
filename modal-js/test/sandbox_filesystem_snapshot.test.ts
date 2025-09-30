import { tc } from "../test-support/test-client";
import { expect, test, onTestFinished } from "vitest";

test("snapshotFilesystem", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sb = await tc.sandboxes.create(app, image);
  onTestFinished(async () => {
    await sb.terminate();
  });

  await sb.exec(["sh", "-c", "echo -n 'test content' > /tmp/test.txt"]);
  await sb.exec(["mkdir", "-p", "/tmp/testdir"]);

  const snapshotImage = await sb.snapshotFilesystem();
  expect(snapshotImage).toBeDefined();
  expect(snapshotImage.imageId).toMatch(/^im-/);

  await sb.terminate();

  // Create new sandbox from snapshot
  const sb2 = await tc.sandboxes.create(app, snapshotImage);
  onTestFinished(async () => {
    await sb2.terminate();
  });

  // Verify file exists in snapshot
  const proc = await sb2.exec(["cat", "/tmp/test.txt"]);
  const output = await proc.stdout.readText();
  expect(output).toBe("test content");

  // Verify directory exists in snapshot
  const dirCheck = await sb2.exec(["test", "-d", "/tmp/testdir"]);
  expect(await dirCheck.wait()).toBe(0);
});
