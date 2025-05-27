import { App, Secret } from "modal";
import { expect, test } from "vitest";

test("ImageFromRegistry", { timeout: 30_000 }, async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  expect(app.appId).toBeTruthy();

  const image = await app.imageFromRegistry("alpine:3.21");
  expect(image.imageId).toBeTruthy();
  expect(image.imageId).toMatch(/^im-/);
});

test("ImageFromAwsEcr", { timeout: 30_000 }, async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  expect(app.appId).toBeTruthy();

  const image = await app.imageFromAwsEcr(
    "459781239556.dkr.ecr.us-east-1.amazonaws.com/ecr-private-registry-test-7522615:python",
    await Secret.fromName("aws-ecr-private-registry-test-secret", {
      environment: "libmodal",
      requiredKeys: ["AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"],
    }),
  );
  expect(image.imageId).toBeTruthy();
  expect(image.imageId).toMatch(/^im-/);
});
