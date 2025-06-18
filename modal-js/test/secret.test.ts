import { Secret } from "modal";
import { expect, test } from "vitest";

test("SecretFromName", async () => {
  const secret = await Secret.fromName("libmodal-test-secret");
  expect(secret).toBeDefined();
  expect(secret.secretId).toBeDefined();
  expect(secret.secretId).toMatch(/^st-/);

  const promise = Secret.fromName("missing-secret");
  await expect(promise).rejects.toThrowError(
    /Secret 'missing-secret' not found/,
  );
});

test("SecretFromNameWithRequiredKeys", async () => {
  const secret = await Secret.fromName("libmodal-test-secret", {
    requiredKeys: ["a", "b", "c"],
  });
  expect(secret).toBeDefined();

  const promise = Secret.fromName("libmodal-test-secret", {
    requiredKeys: ["a", "b", "c", "missing-key"],
  });
  await expect(promise).rejects.toThrowError(
    /Secret is missing key\(s\): missing-key/,
  );
});
