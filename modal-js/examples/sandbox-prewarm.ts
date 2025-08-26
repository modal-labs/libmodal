// We use `Image.build` to create an Image object on Modal
// that eagerly pulls from the registry. The first sandbox created with this image
// will ues this "pre-warmed" image and will start faster.
import { App, Image } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });

// With `.build(app)`, we create an Image object on Modal that eagerly pulls
// from the registry.
const image = await Image.fromRegistry("alpine:3.21").build(app);
console.log("image id:", image.imageId);

const imageId = image.imageId;
// You can save the ImageId and create a new image object that referes to it.
const image2 = Image.fromId(imageId);

// Spawn a sandbox running the "cat" command.
const sb = await app.createSandbox(image2, { command: ["cat"] });
console.log("sandbox:", sb.sandboxId);

// Terminate the sandbox.
await sb.terminate();
