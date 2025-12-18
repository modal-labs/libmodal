// This example shows how to use Image mounts and directory snapshots in Sandboxes.
// First, we mount an empty image at a directory, clone a git repo into it, and
// take a snapshot. Then we create a new Sandbox and mount the snapshot, showing
// how you can persist and reuse directory state across Sandboxes.

import { ModalClient } from "modal";

const modal = new ModalClient();

const app = await modal.apps.fromName("libmodal-example", {
  createIfMissing: true,
});
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
