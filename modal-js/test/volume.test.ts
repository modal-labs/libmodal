import { Volume } from "modal";
import { expect, test } from "vitest";

test("VolumeFromName", async () => {
  const volume = await Volume.fromName("libmodal-test-volume", {
    createIfMissing: true,
  });
  expect(volume).toBeDefined();
  expect(volume.volumeId).toBeDefined();
  expect(volume.volumeId).toMatch(/^vo-/);

  const promise = Volume.fromName("missing-volume");
  await expect(promise).rejects.toThrowError(
    /Volume 'missing-volume' not found/,
  );
});

test("VolumeFromNameWithVersion", async () => {
  const volume = await Volume.fromName("libmodal-test-volume-v2", {
    createIfMissing: true,
    version: 2,
  });
  expect(volume).toBeDefined();
  expect(volume.volumeId).toBeDefined();
  expect(volume.volumeId).toMatch(/^vo-/);
});
