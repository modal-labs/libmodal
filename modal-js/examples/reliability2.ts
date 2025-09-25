// Quick script for making sure Sandboxes can be created and wait() without stalling.

import PQueue from "p-queue";
import { ModalClient } from "modal";

const mc = new ModalClient();

const app = await mc.apps.fromName("libmodal-example", {
  createIfMissing: true,
});
const image = mc.images.fromRegistry("python:3.13-slim");

async function createAndWaitOne() {
  const sb = await mc.sandboxes.create(app, image);
  if (!sb.sandboxId) throw new Error("Sandbox ID is missing");
  await sb.terminate();
  const exitCode = await Promise.race([
    sb.wait(),
    new Promise<number>((_, reject) => {
      setTimeout(() => reject(new Error("wait() timed out")), 10000).unref();
    }),
  ]);
  console.log("Sandbox wait completed with exit code:", exitCode);
  if (exitCode !== 0) throw new Error(`Sandbox exited with code ${exitCode}`);
}

const queue = new PQueue({ concurrency: 50 });

let success = 0;
let failure = 0;

for (let i = 0; i < 150; i++) {
  await queue.onEmpty();

  queue.add(async () => {
    try {
      await createAndWaitOne();
      success++;
      console.log("Sandbox created and waited successfully.", i);
    } catch (error) {
      failure++;
      console.error("Error in Sandbox creation/waiting:", error, i);
    }
  });
}

await queue.onIdle();
console.log("Success:", success);
console.log("Failure:", failure);
