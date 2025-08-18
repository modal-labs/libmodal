import { App, Image, Volume } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await Image.fromRegistry("alpine:3.21");

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
  volumes: { "/mnt/volume": volume.readOnly() },
});
console.log("Reader sandbox:", readerSandbox.sandboxId);

const rp = await readerSandbox.exec(["cat", "/mnt/volume/message.txt"]);
console.log("Reader output:", await rp.stdout.readText());

const wp = await readerSandbox.exec([
  "sh",
  "-c",
  "echo 'This should fail' >> /mnt/volume/message.txt",
]);
const wpExitCode = await wp.wait();

console.log("Write attempt exit code:", wpExitCode);
console.log("Write attempt stderr:", await wp.stderr.readText());

await writerSandbox.terminate();
await readerSandbox.terminate();
