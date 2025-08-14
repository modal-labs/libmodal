import { App, Image, Secret } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await Image.fromRegistry("alpine:3.21");
const secret = await Secret.fromName("libmodal-test-secret", {
  requiredKeys: ["c"],
});

const sandbox = await app.createSandbox(image, {
  command: ["printenv", "c"],
  secrets: [secret],
});

console.log("Sandbox created:", sandbox.sandboxId);

const result = await sandbox.stdout.readText();
console.log("Environment variable c:", result);
