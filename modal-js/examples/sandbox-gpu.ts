import { App } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await app.imageFromRegistry(
  "nvidia/cuda:12.4.0-devel-ubuntu22.04",
);

const sb = await app.createSandbox(image, { gpu: "A10G" });
console.log("Started sandbox with A10G GPU:", sb.sandboxId);

try {
  console.log("Running `nvidia-smi` in sandbox:");

  const gpuCheck = await sb.exec(["nvidia-smi"]);

  console.log(await gpuCheck.stdout.readText());
} finally {
  await sb.terminate();
}
