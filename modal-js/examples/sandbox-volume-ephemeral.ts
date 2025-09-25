import { ModalClient } from "modal";

const mc = new ModalClient();

const app = await mc.apps.fromName("libmodal-example", {
  createIfMissing: true,
});
const image = mc.images.fromRegistry("alpine:3.21");

const volume = await mc.volumes.ephemeral();

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
await writerSandbox.terminate();

const readerSandbox = await mc.sandboxes.create(app, image, {
  command: ["cat", "/mnt/volume/message.txt"],
  volumes: { "/mnt/volume": volume.readOnly() },
});
console.log("Reader Sandbox:", readerSandbox.sandboxId);
console.log("Reader output:", await readerSandbox.stdout.readText());

await readerSandbox.terminate();
volume.closeEphemeral();
