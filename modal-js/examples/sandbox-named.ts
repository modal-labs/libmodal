import { App, Image, Sandbox, AlreadyExistsError } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await Image.fromRegistry("alpine:3.21");

const sandboxName = `libmodal-example-named-sandbox`;

const sb = await app.createSandbox(image, {
  name: sandboxName,
  command: ["cat"],
});

console.log(`Created sandbox with name: ${sandboxName}`);
console.log(`Sandbox ID: ${sb.sandboxId}`);

try {
  await app.createSandbox(image, {
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

const sbFromName = await Sandbox.fromName("libmodal-example", sandboxName);
console.log(`Retrieved the same sandbox from name: ${sbFromName.sandboxId}`);

await sbFromName.stdin.writeText("hello, named Sandbox");
await sbFromName.stdin.close();

console.log("Reading output:");
console.log(await sbFromName.stdout.readText());

await sb.terminate();
console.log("Sandbox terminated");
