import { App, Secret, CloudBucketMount } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await app.imageFromRegistry("alpine:3.21");

const secret = await Secret.fromName("libmodal-aws-bucket-secret");

const sb = await app.createSandbox(image, {
  command: ["sh", "-c", "ls -la /mnt/s3-bucket"],
  cloudBucketMounts: {
    "/mnt/s3-bucket": new CloudBucketMount("my-s3-bucket", {
      secret,
      keyPrefix: "data/",
      readOnly: true,
    }),
  },
});

console.log("S3 sandbox:", sb.sandboxId);
console.log(
  "Sandbox directory listing of /mnt/s3-bucket:",
  await sb.stdout.readText(),
);

await sb.terminate();
