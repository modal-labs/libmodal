import { App, Image, Sandbox } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await Image.fromRegistry("alpine:3.21");

const sb = await app.createSandbox(image, { command: ["cat"] });
console.log("Sandbox:", sb.sandboxId);

const sbFromId = await Sandbox.fromId(sb.sandboxId);
console.log("Queried Sandbox from ID:", sbFromId.sandboxId);

await sb.stdin.writeText("this is input that should be mirrored by cat");
await sb.stdin.close();
console.log("output:", await sb.stdout.readText());

await sb.terminate();
