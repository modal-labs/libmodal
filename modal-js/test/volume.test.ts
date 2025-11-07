import { tc } from "../test-support/test-client";
import { expect, test } from "vitest";
import { createMockModalClients } from "../test-support/grpc_mock";
import { NotFoundError } from "../src/errors";

test("Volume.fromName", async () => {
  const volume = await tc.volumes.fromName("libmodal-test-volume", {
    createIfMissing: true,
  });
  expect(volume).toBeDefined();
  expect(volume.volumeId).toBeDefined();
  expect(volume.volumeId).toMatch(/^vo-/);
  expect(volume.name).toBe("libmodal-test-volume");

  const promise = tc.volumes.fromName("missing-volume");
  await expect(promise).rejects.toThrowError(
    /Volume 'missing-volume' not found/,
  );
});

test("Volume.readOnly", async () => {
  const volume = await tc.volumes.fromName("libmodal-test-volume", {
    createIfMissing: true,
  });

  expect(volume.isReadOnly).toBe(false);

  const readOnlyVolume = volume.readOnly();
  expect(readOnlyVolume.isReadOnly).toBe(true);
  expect(readOnlyVolume.volumeId).toBe(volume.volumeId);
  expect(readOnlyVolume.name).toBe(volume.name);

  expect(volume.isReadOnly).toBe(false);
});

test("VolumeEphemeral", async () => {
  const volume = await tc.volumes.ephemeral();
  expect(volume.name).toBeUndefined();
  expect(volume.volumeId).toMatch(/^vo-/);
  expect(volume.isReadOnly).toBe(false);
  expect(volume.readOnly().isReadOnly).toBe(true);
  volume.closeEphemeral();
});

test("VolumeDelete success", async () => {
  const { mockClient: mc, mockCpClient: mock } = createMockModalClients();

  mock.handleUnary("/VolumeGetOrCreate", () => ({
    volumeId: "vo-test-123",
    metadata: { name: "test-volume" },
  }));

  mock.handleUnary("/VolumeDelete", (req: any) => {
    expect(req.volumeId).toBe("vo-test-123");
    return {};
  });

  await mc.volumes.delete("test-volume");
  mock.assertExhausted();
});

test("VolumeDelete with allowMissing=true", async () => {
  const { mockClient: mc, mockCpClient: mock } = createMockModalClients();

  mock.handleUnary("/VolumeGetOrCreate", () => {
    throw new NotFoundError("Volume 'missing' not found");
  });

  await mc.volumes.delete("missing", { allowMissing: true });
  mock.assertExhausted();
});

test("VolumeDelete with allowMissing=false throws", async () => {
  const { mockClient: mc, mockCpClient: mock } = createMockModalClients();

  mock.handleUnary("/VolumeGetOrCreate", () => {
    throw new NotFoundError("Volume 'missing' not found");
  });

  await expect(
    mc.volumes.delete("missing", { allowMissing: false }),
  ).rejects.toThrow(NotFoundError);
});
