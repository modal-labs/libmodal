import { App, Sandbox } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await app.imageFromRegistry("alpine:3.21");

// Spawn a sandbox running the "cat" command.
const sb = await app.createSandbox(image, { command: ["cat"] });
console.log("sandbox:", sb.sandboxId);

// Get running sandbox from ID
const sbFromId = await Sandbox.fromId(sb.sandboxId);
console.log("Queried sandbox from ID:", sbFromId.sandboxId);

// Write to the sandbox's stdin and read from its stdout.
await sb.stdin.writeText("this is input that should be mirrored by cat");
await sb.stdin.close();
console.log("output:", await sb.stdout.readText());

// Terminate the sandbox.
await sb.terminate();
