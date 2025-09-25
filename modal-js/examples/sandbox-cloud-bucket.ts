import { ModalClient, CloudBucketMount } from "modal";

const mc = new ModalClient();

const app = await mc.apps.fromName("libmodal-example", {
  createIfMissing: true,
});
const image = mc.images.fromRegistry("alpine:3.21");

const secret = await mc.secrets.fromName("libmodal-aws-bucket-secret");

const sb = await mc.sandboxes.create(app, image, {
  command: ["sh", "-c", "ls -la /mnt/s3-bucket"],
  cloudBucketMounts: {
    "/mnt/s3-bucket": new CloudBucketMount("my-s3-bucket", {
      secret,
      keyPrefix: "data/",
      readOnly: true,
    }),
  },
});

console.log("S3 Sandbox:", sb.sandboxId);
console.log(
  "Sandbox directory listing of /mnt/s3-bucket:",
  await sb.stdout.readText(),
);

await sb.terminate();
