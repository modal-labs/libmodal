// Run a bunch of container exec commands, alerting of any output issues.

import PQueue from "p-queue";
import { ModalClient } from "modal";

const mc = new ModalClient();

const app = await mc.apps.lookup("libmodal-example", {
  createIfMissing: true,
});
const image = mc.images.fromRegistry("python:3.13-slim");

const sandboxes = [
  await mc.sandboxes.create(app, image),
  await mc.sandboxes.create(app, image),
  await mc.sandboxes.create(app, image),
  await mc.sandboxes.create(app, image),
  await mc.sandboxes.create(app, image),
  await mc.sandboxes.create(app, image),
  await mc.sandboxes.create(app, image),
  await mc.sandboxes.create(app, image),
  await mc.sandboxes.create(app, image),
  await mc.sandboxes.create(app, image),
];

try {
  const expectedContent = Array.from(
    { length: 50000 },
    (_, i) => `${i}\n`,
  ).join("");

  const queue = new PQueue({ concurrency: 50 });

  let success = 0;
  let failure = 0;

  for (let i = 0; i < 10000; i++) {
    await queue.onEmpty();

    queue.add(async () => {
      const sb = sandboxes[i % sandboxes.length];
      const p = await sb.exec(
        [
          "python",
          "-c",
          `
import time
import sys
for i in range(50000):
  if i % 1000 == 0:
    time.sleep(0.01)
  print(i)
  print(i, file=sys.stderr)`,
        ],
        {
          stdout: "pipe",
          stderr: "pipe",
        },
      );
      const [contentStdout, contentStderr] = await Promise.all([
        p.stdout.readText(),
        p.stderr.readText(),
      ]);
      if (
        contentStdout === expectedContent &&
        contentStderr === expectedContent
      ) {
        success++;
        console.log("Output matches expected content.", i);
      } else {
        failure++;
        console.error("MISMATCH", i);
      }
    });
  }

  await queue.onIdle();
  console.log("Success:", success);
  console.log("Failure:", failure);
} finally {
  for (const sb of sandboxes) {
    await sb.terminate();
  }
}
