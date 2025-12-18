// This example shows how to mount Images in the Sandbox filesystem and take snapshots
// of them.
//
// The feature is still experimental in the sense that the API is subject to change.
//
// High level, it allows you to:
// - Mount any Modal Image at a specific directory within the Sandbox filesystem.
// - Take a snapshot of that directory, which will create a new Modal Image with
//   the updated contents of the directory.
//
// You can only snapshot directories that have previously been mounted using
// `Sandbox.experimentalMountImage`. If you want to mount an empty directory,
// you can pass undefined as the image parameter.
//
// For exmaple, you can use this to mount user specific dependencies into a running
// Sandbox, that is started with a base Image with shared system dependencies. This
// way, you can update system dependencies and user projects independently.

import { ModalClient } from "modal";

const modal = new ModalClient();

const app = await modal.apps.fromName("libmodal-example", {
  createIfMissing: true,
});
// The base Image you use for the Sandbox must have a /usr/bin/mount binary.
const baseImage = modal.images.fromRegistry("debian:12-slim");

const sb = await modal.sandboxes.create(app, baseImage);

// You must mount an Image at a directory in the Sandbox filesystem before you
// can snapshot it. You can pass undefined as the image parameter to mount an
// empty directory.
//
// The target directory must exist before you can mount it:
await (await sb.exec(["mkdir", "-p", "/repo"])).wait();
await sb.experimentalMountImage("/repo");

const gitClone = await sb.exec([
  "git",
  "clone",
  "https://github.com/modal-labs/libmodal.git",
  "/repo",
]);
await gitClone.wait();

const repoSnapshot = await sb.experimentalSnapshotDirectory("/repo");
console.log(
  "Took a snapshot of the /repo directory, Image ID:",
  repoSnapshot.imageId,
);

await sb.terminate();

// Start a new Sandbox, and mount the repo directory:
const sb2 = await modal.sandboxes.create(app, baseImage);

await (await sb2.exec(["mkdir", "-p", "/repo"])).wait();
await sb2.experimentalMountImage("/repo", repoSnapshot);

const repoLs = await sb2.exec(["ls", "/repo"]);
console.log(
  "Contents of /repo directory in new Sandbox sb2:\n",
  await repoLs.stdout.readText(),
);

await sb2.terminate();
await modal.images.delete(repoSnapshot.imageId);
