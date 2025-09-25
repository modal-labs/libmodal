import { ModalClient, AlreadyExistsError } from "modal";

const mc = new ModalClient();

const app = await mc.apps.fromName("libmodal-example", {
  createIfMissing: true,
});
const image = mc.images.fromRegistry("alpine:3.21");

const sandboxName = `libmodal-example-named-sandbox`;

const sb = await mc.sandboxes.create(app, image, {
  name: sandboxName,
  command: ["cat"],
});

console.log(`Created Sandbox with name: ${sandboxName}`);
console.log(`Sandbox ID: ${sb.sandboxId}`);

try {
  await mc.sandboxes.create(app, image, {
    name: sandboxName,
    command: ["cat"],
  });
} catch (e) {
  if (e instanceof AlreadyExistsError) {
    console.log(
      "Trying to create one more Sandbox with the same name throws:",
      e.message,
    );
  } else {
    throw e;
  }
}

const sbFromName = await mc.sandboxes.fromName("libmodal-example", sandboxName);
console.log(`Retrieved the same Sandbox from name: ${sbFromName.sandboxId}`);

await sbFromName.stdin.writeText("hello, named Sandbox");
await sbFromName.stdin.close();

console.log("Reading output:");
console.log(await sbFromName.stdout.readText());

await sb.terminate();
console.log("Sandbox terminated");
