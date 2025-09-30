import { ModalClient } from "modal";

const mc = new ModalClient();

const app = await mc.apps.fromName("libmodal-example", {
  createIfMissing: true,
});
const image = mc.images.fromRegistry("python:3.13-slim");

const sb = await mc.sandboxes.create(app, image);
console.log("Started Sandbox:", sb.sandboxId);

try {
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

  // Read both the stdout and stderr streams.
  const [contentStdout, contentStderr] = await Promise.all([
    p.stdout.readText(),
    p.stderr.readText(),
  ]);
  console.log(
    `Got ${contentStdout.length} bytes stdout and ${contentStderr.length} bytes stderr`,
  );
  console.log("Return code:", await p.wait());

  const secret = await mc.secrets.fromName("libmodal-test-secret", {
    requiredKeys: ["c"],
  });
  const printSecret = await sb.exec(["printenv", "c"], {
    stdout: "pipe",
    secrets: [secret],
  });
  const secretText = await printSecret.stdout.readText();
  console.log(`Got environment variable c=${secretText}`);
} finally {
  await sb.terminate();
}
