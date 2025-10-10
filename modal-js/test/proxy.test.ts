import { tc } from "../test-support/test-client";
import { expect, test } from "vitest";

test("CreateSandboxWithProxy", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const proxy = await tc.proxies.fromName("libmodal-test-proxy", {
    environment: "libmodal",
  });
  expect(proxy.proxyId).toBeTruthy();
  expect(proxy.proxyId).toMatch(/^pr-/);

  const sb = await tc.sandboxes.create(app, image, {
    proxy,
    command: ["echo", "hello, sandbox with proxy"],
  });
  expect(sb.sandboxId).toBeTruthy();

  await sb.terminate();
  expect(await sb.wait()).toBe(137);
});

test("ProxyNotFound", async () => {
  await expect(tc.proxies.fromName("non-existent-proxy-name")).rejects.toThrow(
    "Proxy 'non-existent-proxy-name' not found",
  );
});
