import { ModalClient } from "modal";

const mc = new ModalClient();

const app = await mc.apps.lookup("libmodal-example", {
  createIfMissing: true,
});
const image = mc.images.fromRegistry("nvidia/cuda:12.4.0-devel-ubuntu22.04");

const sb = await mc.sandboxes.create(app, image, { gpu: "A10G" });
console.log("Started Sandbox with A10G GPU:", sb.sandboxId);

try {
  console.log("Running `nvidia-smi` in Sandbox:");

  const gpuCheck = await sb.exec(["nvidia-smi"]);

  console.log(await gpuCheck.stdout.readText());
} finally {
  await sb.terminate();
}
