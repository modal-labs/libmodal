import { App } from "modal";

const app = await App.lookup("libmodal-example", {
  createIfMissing: true,
});
const baseImage = await app.imageFromRegistry("alpine:3.21");

const sb = await app.createSandbox(baseImage);
console.log("Started sandbox:", sb.sandboxId);

await sb.exec(["mkdir", "-p", "/app/data"]);
await sb.exec([
  "sh",
  "-c",
  "echo 'This file was created in the first sandbox' > /app/data/info.txt",
]);
console.log("Created file in first sandbox");

const snapshotImage = await sb.snapshotFilesystem();
console.log(
  "Filesystem snapshot created with image ID:",
  snapshotImage.imageId,
);

await sb.terminate();
console.log("Terminated first sandbox");

// Create new sandbox from the snapshot image
const sb2 = await app.createSandbox(snapshotImage);
console.log("\nStarted new sandbox from snapshot:", sb2.sandboxId);

const proc = await sb2.exec(["cat", "/app/data/info.txt"]);
const info = await proc.stdout.readText();
console.log("File data read in second sandbox:", info);

await sb2.terminate();
