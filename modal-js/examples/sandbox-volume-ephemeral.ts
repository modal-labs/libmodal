import { App, Image, Volume } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await Image.fromRegistry("alpine:3.21");

const volume = await Volume.ephemeral();

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
await writerSandbox.terminate();

const readerSandbox = await app.createSandbox(image, {
  command: ["cat", "/mnt/volume/message.txt"],
  volumes: { "/mnt/volume": volume.readOnly() },
});
console.log("Reader sandbox:", readerSandbox.sandboxId);
console.log("Reader output:", await readerSandbox.stdout.readText());

await readerSandbox.terminate();
volume.closeEphemeral();
