import { ModalClient } from "modal";

const mc = new ModalClient();

const app = await mc.apps.lookup("libmodal-example", {
  createIfMissing: true,
});

const image = mc.images
  .fromRegistry("alpine:3.21")
  .dockerfileCommands(["RUN apk add --no-cache curl=$CURL_VERSION"], {
    secrets: [
      await mc.secrets.fromObject({
        CURL_VERSION: "8.12.1-r1",
      }),
    ],
  })
  .dockerfileCommands(["ENV SERVER=ipconfig.me"]);

const sb = await mc.sandboxes.create(app, image, {
  command: ["sh", "-c", "curl -Ls $SERVER"],
});
console.log("Created Sandbox with ID:", sb.sandboxId);

console.log("Sandbox output:", await sb.stdout.readText());
await sb.terminate();
