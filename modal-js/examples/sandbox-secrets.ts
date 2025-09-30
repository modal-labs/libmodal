import { ModalClient } from "modal";

const mc = new ModalClient();

const app = await mc.apps.fromName("libmodal-example", {
  createIfMissing: true,
});
const image = mc.images.fromRegistry("alpine:3.21");

const secret = await mc.secrets.fromName("libmodal-test-secret", {
  requiredKeys: ["c"],
});

const ephemeralSecret = await mc.secrets.fromObject({
  d: "123",
});

const sandbox = await mc.sandboxes.create(app, image, {
  command: ["sh", "-lc", "printenv | grep -E '^c|d='"],
  secrets: [secret, ephemeralSecret],
});

console.log("Sandbox created:", sandbox.sandboxId);

console.log("Sandbox environment variables from Secrets:");
console.log(await sandbox.stdout.readText());
