import { App, Volume } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await app.imageFromRegistry("alpine:3.21");

const volume = await Volume.fromName("libmodal-example-volume", {
  createIfMissing: true,
});

const writerSandbox = await app.createSandbox(image, {
  command: [
    "sh",
    "-c",
    "echo 'Hello from writer sandbox!' > /mnt/volume/message.txt",
  ],
  volumes: { "/mnt/volume": volume },
});
console.log("Writer sandbox:", writerSandbox.sandboxId);

await writerSandbox.wait();
console.log("Writer finished");

const readerSandbox = await app.createSandbox(image, {
  command: ["sh", "-c", "cat /mnt/volume/message.txt"],
  volumes: { "/mnt/volume": volume },
});
console.log("Reader sandbox:", readerSandbox.sandboxId);

console.log("Reader output:", await readerSandbox.stdout.readText());

await writerSandbox.terminate();
await readerSandbox.terminate();
