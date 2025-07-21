import { App, Secret } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await app.imageFromRegistry("python:3.13-slim");

const sb = await app.createSandbox(image);
console.log("Started sandbox:", sb.sandboxId);

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

  const secret = await Secret.fromName("libmodal-test-secret", {
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
