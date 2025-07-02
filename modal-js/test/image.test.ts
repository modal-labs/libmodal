import { App, Secret } from "modal";
import { expect, test } from "vitest";

test("ImageFromRegistry", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  expect(app.appId).toBeTruthy();

  const image = await app.imageFromRegistry("alpine:3.21");
  expect(image.imageId).toBeTruthy();
  expect(image.imageId).toMatch(/^im-/);
});

test("ImageFromRegistryWithSecret", async () => {
  // GCP Artifact Registry also supports auth using username and password, if the username is "_json_key"
  // and the password is the service account JSON blob. See:
  // https://cloud.google.com/artifact-registry/docs/docker/authentication#json-key
  // So we use GCP Artifact Registry to test this too.

  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  expect(app.appId).toBeTruthy();

  const image = await app.imageFromRegistry(
    "us-east1-docker.pkg.dev/modal-prod-367916/private-repo-test/my-image",
    await Secret.fromName("libmodal-gcp-artifact-registry-test", {
      requiredKeys: ["REGISTRY_USERNAME", "REGISTRY_PASSWORD"],
    }),
  );
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

test("ImageFromGcpArtifactRegistry", { timeout: 30_000 }, async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  expect(app.appId).toBeTruthy();

  const image = await app.imageFromGcpArtifactRegistry(
    "us-east1-docker.pkg.dev/modal-prod-367916/private-repo-test/my-image",
    await Secret.fromName("libmodal-gcp-artifact-registry-test", {
      requiredKeys: ["SERVICE_ACCOUNT_JSON"],
    }),
  );
  expect(image.imageId).toBeTruthy();
  expect(image.imageId).toMatch(/^im-/);
});
