import { ModalClient } from "modal";

const mc = new ModalClient();

const app = await mc.apps.lookup("libmodal-example", {
  createIfMissing: true,
});
const image = mc.images.fromRegistry("alpine:3.21");

const sb = await mc.sandboxes.create(app, image, { command: ["cat"] });
console.log("Sandbox:", sb.sandboxId);

const sbFromId = await mc.sandboxes.fromId(sb.sandboxId);
console.log("Queried Sandbox from ID:", sbFromId.sandboxId);

await sb.stdin.writeText("this is input that should be mirrored by cat");
await sb.stdin.close();
console.log("output:", await sb.stdout.readText());

await sb.terminate();
