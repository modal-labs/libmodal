import { ModalClient } from "modal";

const modal = new ModalClient();

const app = await modal.apps.fromName("libmodal-example", {
  createIfMissing: true,
});

// Create a Sandbox with Python's built-in HTTP server
const image = modal.images.fromRegistry("python:3.12-alpine");
const sb = await modal.sandboxes.create(app, image, {
  command: ["python3", "-m", "http.server", "8000"],
});

const creds = await sb.createConnectToken({userMetadata: "abc"});
console.log("Get url: " + creds.url + " , credentials: " + creds.token);
await sb.terminate();
