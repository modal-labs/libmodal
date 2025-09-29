import { ModalClient } from "modal";

const mc = new ModalClient();

const app = await mc.apps.lookup("libmodal-example", {
  createIfMissing: true,
});
const image = mc.images.fromRegistry("alpine:3.21");

// Create a Sandbox that waits for input, then exits with code 42
const sandbox = await mc.sandboxes.create(app, image, {
  command: ["sh", "-c", "read line; exit 42"],
});

console.log("Started Sandbox:", sandbox.sandboxId);

console.log("Poll result while running:", await sandbox.poll());

console.log("\nSending input to trigger completion...");
await sandbox.stdin.writeText("hello, goodbye");
await sandbox.stdin.close();

const exitCode = await sandbox.wait();
console.log("\nSandbox completed with exit code:", exitCode);
console.log("Poll result after completion:", await sandbox.poll());
