import { App, Secret } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await app.imageFromAwsEcr(
  "459781239556.dkr.ecr.us-east-1.amazonaws.com/ecr-private-registry-test-7522615:python",
  await Secret.fromName("aws-ecr-private-registry-test-secret", {
    requiredKeys: ["AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"],
  }),
);

// Spawn a sandbox running a simple Python version of the "cat" command.
const sb = await app.createSandbox(image, {
  command: ["python", "-c", `import sys; sys.stdout.write(sys.stdin.read())`],
});
console.log("sandbox:", sb.sandboxId);

// Write to the sandbox's stdin and read from its stdout.
await sb.stdin.writeText(
  "this is input that should be mirrored by the Python one-liner",
);
await sb.stdin.close();
console.log("output:", await sb.stdout.readText());

// Terminate the sandbox.
await sb.terminate();
