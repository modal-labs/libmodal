import { tc } from "../test-support/test-client";
import { expect, test } from "vitest";
import { mergeEnvIntoSecrets } from "../src/secret";

test("SecretFromName", async () => {
  const secret = await tc.secrets.fromName("libmodal-test-secret");
  expect(secret).toBeDefined();
  expect(secret.secretId).toBeDefined();
  expect(secret.secretId).toMatch(/^st-/);
  expect(secret.name).toBe("libmodal-test-secret");

  const promise = tc.secrets.fromName("missing-secret");
  await expect(promise).rejects.toThrowError(
    /Secret 'missing-secret' not found/,
  );
});

test("SecretFromNameWithRequiredKeys", async () => {
  const secret = await tc.secrets.fromName("libmodal-test-secret", {
    requiredKeys: ["a", "b", "c"],
  });
  expect(secret).toBeDefined();

  const promise = tc.secrets.fromName("libmodal-test-secret", {
    requiredKeys: ["a", "b", "c", "missing-key"],
  });
  await expect(promise).rejects.toThrowError(
    /Secret is missing key\(s\): missing-key/,
  );
});

test("SecretFromObject", async () => {
  const secret = await tc.secrets.fromObject({ key: "value" });
  expect(secret).toBeDefined();

  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sandbox = await tc.sandboxes.create(app, image, {
    command: ["printenv", "key"],
    secrets: [secret],
  });

  const output = await sandbox.stdout.readText();
  expect(output).toBe("value\n");
});

test("SecretFromObjectInvalid", async () => {
  // @ts-expect-error testing runtime validation
  await expect(tc.secrets.fromObject({ key: 123 })).rejects.toThrowError(
    /entries must be an object mapping string keys to string values/,
  );
});

test("mergeEnvIntoSecrets merges env with existing secrets", async () => {
  const existingSecret = await tc.secrets.fromObject({ A: "1" });
  const env = { B: "2", C: "3" };

  const result = await mergeEnvIntoSecrets(tc, env, [existingSecret]);

  expect(result).toHaveLength(2);
  expect(result[0]).toBe(existingSecret);
  expect(result[1].secretId).toMatch(/^st-/);
});

test("mergeEnvIntoSecrets with only env parameter", async () => {
  const env = { B: "2", C: "3" };

  const result = await mergeEnvIntoSecrets(tc, env);

  expect(result).toHaveLength(1);
  expect(result[0].secretId).toMatch(/^st-/);
});

test("mergeEnvIntoSecrets with empty env object returns existing secrets", async () => {
  const existingSecret = await tc.secrets.fromObject({ A: "1" });
  const env = {};

  const result = await mergeEnvIntoSecrets(tc, env, [existingSecret]);

  expect(result).toHaveLength(1);
  expect(result[0]).toBe(existingSecret);
});

test("mergeEnvIntoSecrets with undefined env returns existing secrets", async () => {
  const existingSecret = await tc.secrets.fromObject({ A: "1" });

  const result = await mergeEnvIntoSecrets(tc, undefined, [existingSecret]);

  expect(result).toHaveLength(1);
  expect(result[0]).toBe(existingSecret);
});

test("mergeEnvIntoSecrets with only existing secrets", async () => {
  const secret1 = await tc.secrets.fromObject({ A: "1" });
  const secret2 = await tc.secrets.fromObject({ B: "2" });

  const result = await mergeEnvIntoSecrets(tc, undefined, [secret1, secret2]);

  expect(result).toHaveLength(2);
  expect(result[0]).toBe(secret1);
  expect(result[1]).toBe(secret2);
});

test("mergeEnvIntoSecrets with no env and no secrets returns empty array", async () => {
  const result = await mergeEnvIntoSecrets(tc);

  expect(result).toHaveLength(0);
  expect(result).toEqual([]);
});
