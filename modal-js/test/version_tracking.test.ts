import { expect, test } from "vitest";

declare const __MODAL_SDK_VERSION__: string;

test("VersionConstantFormat", () => {
  expect(__MODAL_SDK_VERSION__).toBeDefined();
  expect(__MODAL_SDK_VERSION__).toMatch(/^modal-js\/\d+\.\d+\.\d+$/);
});
