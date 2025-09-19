import { App, Image, Proxy } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await Image.fromRegistry("alpine/curl:8.14.1");

const proxy = await Proxy.fromName("libmodal-test-proxy");
console.log("Using Proxy with ID:", proxy.proxyId);

const sb = await app.createSandbox(image, {
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
