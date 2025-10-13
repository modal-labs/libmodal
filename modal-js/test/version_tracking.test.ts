import { expect, test } from "vitest";
import { ModalClient } from "modal";

declare const __MODAL_SDK_VERSION__: string;

test("VersionConstantFormat", () => {
  expect(__MODAL_SDK_VERSION__).toBeDefined();
  expect(__MODAL_SDK_VERSION__).toMatch(/^modal-js\/\d+\.\d+\.\d+$/);
});

test("ClientVersion", () => {
  const client = new ModalClient();
  expect(client.version()).toBeDefined();
  expect(client.version()).toMatch(/^modal-js\/\d+\.\d+\.\d+$/);
  expect(client.version()).toBe(__MODAL_SDK_VERSION__);
  client.close();
});
