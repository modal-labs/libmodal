import { ModalClient } from "modal";

const mc = new ModalClient();

const app = await mc.apps.lookup("libmodal-example", {
  createIfMissing: true,
});
const image = mc.images.fromAwsEcr(
  "459781239556.dkr.ecr.us-east-1.amazonaws.com/ecr-private-registry-test-7522615:python",
  await mc.secrets.fromName("libmodal-aws-ecr-test", {
    requiredKeys: ["AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"],
  }),
);

// Spawn a Sandbox running a simple Python version of the "cat" command.
const sb = await mc.sandboxes.create(app, image, {
  command: ["python", "-c", `import sys; sys.stdout.write(sys.stdin.read())`],
});
console.log("Sandbox:", sb.sandboxId);

await sb.stdin.writeText(
  "this is input that should be mirrored by the Python one-liner",
);
await sb.stdin.close();
console.log("output:", await sb.stdout.readText());

await sb.terminate();
