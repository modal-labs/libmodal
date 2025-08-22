import { App, Image, Secret } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await Image.fromRegistry("alpine:3.21");

const secret = await Secret.fromName("libmodal-test-secret", {
  requiredKeys: ["c"],
});

const ephemeralSecret = await Secret.fromObject({
  d: "123",
});

const sandbox = await app.createSandbox(image, {
  command: ["sh", "-lc", "printenv | grep -E '^c|d='"],
  secrets: [secret, ephemeralSecret],
});

console.log("Sandbox created:", sandbox.sandboxId);

console.log("Sandbox environment variables from secrets:");
console.log(await sandbox.stdout.readText());
