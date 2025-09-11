import { App, Image, Secret } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });

const image = Image.fromRegistry("alpine:3.21")
  .dockerfileCommands(["RUN apk add --no-cache curl=$CURL_VERSION"], {
    secrets: [
      await Secret.fromObject({
        CURL_VERSION: "8.12.1-r1",
      }),
    ],
  })
  .dockerfileCommands(["ENV SERVER=ipconfig.me"]);

const sb = await app.createSandbox(image, {
  command: ["sh", "-c", "curl -Ls $SERVER"],
});
console.log("Created Sandbox with ID:", sb.sandboxId);

console.log("Sandbox output:", await sb.stdout.readText());
await sb.terminate();
