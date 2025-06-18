import { App, Secret } from "modal";
import { expect, test } from "vitest";

test("ImageFromRegistry", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  expect(app.appId).toBeTruthy();

  const image = await app.imageFromRegistry("alpine:3.21");
  expect(image.imageId).toBeTruthy();
  expect(image.imageId).toMatch(/^im-/);
});

test("ImageFromAwsEcr", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  expect(app.appId).toBeTruthy();

  const image = await app.imageFromAwsEcr(
    "459781239556.dkr.ecr.us-east-1.amazonaws.com/ecr-private-registry-test-7522615:python",
    await Secret.fromName("libmodal-aws-ecr-test", {
      requiredKeys: ["AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"],
    }),
  );
  expect(image.imageId).toBeTruthy();
  expect(image.imageId).toMatch(/^im-/);
});
