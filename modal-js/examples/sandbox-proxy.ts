import { ModalClient } from "modal";

const mc = new ModalClient();

const app = await mc.apps.fromName("libmodal-example", {
  createIfMissing: true,
});
const image = mc.images.fromRegistry("alpine/curl:8.14.1");

const proxy = await mc.proxies.fromName("libmodal-test-proxy", {
  environment: "libmodal",
});
console.log("Using Proxy with ID:", proxy.proxyId);

const sb = await mc.sandboxes.create(app, image, {
  proxy,
});
console.log("Created Sandbox with Proxy:", sb.sandboxId);

try {
  const p = await sb.exec(["curl", "-s", "ifconfig.me"]);
  const ip = await p.stdout.readText();

  console.log("External IP:", ip.trim());
} finally {
  await sb.terminate();
}
