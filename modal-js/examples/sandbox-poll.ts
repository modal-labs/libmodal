import { App } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await app.imageFromRegistry("alpine:3.21");

// Create a sandbox that waits for input, then exits with code 42
const sandbox = await app.createSandbox(image, {
  command: ["sh", "-c", "read line; exit 42"],
});

console.log("Started sandbox:", sandbox.sandboxId);

const initialPoll = await sandbox.poll();
console.log("Poll result while running:", initialPoll);
console.log("Return code while running:", sandbox.returncode);

console.log("\nSending input to trigger completion...");
await sandbox.stdin.writeText("hello, goodbye");
await sandbox.stdin.close();

const exitCode = await sandbox.wait();
console.log("\nSandbox completed with exit code:", exitCode);

const finalPoll = await sandbox.poll();
console.log("Poll result after completion:", finalPoll);
console.log("Return code after completion:", sandbox.returncode);
