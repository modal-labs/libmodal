// We use `Image.build` to create an Image object on Modal
// that eagerly pulls from the registry. The first Sandbox created with this Image
// will ues this "pre-warmed" Image and will start faster.
import { ModalClient } from "modal";

const mc = new ModalClient();

const app = await mc.apps.lookup("libmodal-example", {
  createIfMissing: true,
});

// With `.build(app)`, we create an Image object on Modal that eagerly pulls
// from the registry.
const image = await mc.images.fromRegistry("alpine:3.21").build(app);
console.log("image id:", image.imageId);

const imageId = image.imageId;
// You can save the Image ID and create a new Image object that referes to it.
const image2 = await mc.images.fromId(imageId);

const sb = await mc.sandboxes.create(app, image2, { command: ["cat"] });
console.log("Sandbox:", sb.sandboxId);

await sb.terminate();
