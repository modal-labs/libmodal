import { App } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });

// Create a sandbox with Python's built-in HTTP server
const image = await app.imageFromRegistry("python:3.12-alpine");
const sandbox = await app.createSandbox(image, {
  command: ["python3", "-m", "http.server", "8000"],
  encryptedPorts: [8000],
  timeout: 60000, // 1 minute
});

console.log("Sandbox created:", sandbox.sandboxId);

console.log("Getting tunnel information...");
const tunnels = await sandbox.tunnels();

console.log("Waiting for server to start...");
await new Promise((resolve) => setTimeout(resolve, 3000));
const tunnel = tunnels[8000];

console.log("Tunnel information:");
console.log("  URL:", tunnel.url);
console.log("  Port:", tunnel.port);

console.log("\nMaking GET request to the tunneled server at " + tunnel.url);

const response = await fetch(tunnel.url);

const html = await response.text();
console.log("\nDirectory listing from server (first 500 chars):");
console.log(html.substring(0, 500));

console.log("\n✅ Successfully connected to the tunneled server!");

await sandbox.terminate();
