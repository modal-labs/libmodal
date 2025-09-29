import { ModalClient } from "modal";

const mc = new ModalClient();

const app = await mc.apps.lookup("libmodal-example", {
  createIfMissing: true,
});
const image = mc.images.fromRegistry("alpine:3.21");

const volume = await mc.volumes.fromName("libmodal-example-volume", {
  createIfMissing: true,
});

const writerSandbox = await mc.sandboxes.create(app, image, {
  command: [
    "sh",
    "-c",
    "echo 'Hello from writer Sandbox!' > /mnt/volume/message.txt",
  ],
  volumes: { "/mnt/volume": volume },
});
console.log("Writer Sandbox:", writerSandbox.sandboxId);

await writerSandbox.wait();
console.log("Writer finished");

const readerSandbox = await mc.sandboxes.create(app, image, {
  volumes: { "/mnt/volume": volume.readOnly() },
});
console.log("Reader Sandbox:", readerSandbox.sandboxId);

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
