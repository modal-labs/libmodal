import { App, Image } from "modal";
import { execFile } from "child_process";
import { promisify } from "util";
import { mkdirSync, writeFileSync, rmSync } from "fs";
import { join } from "path";
import { tmpdir } from "os";

const execFileAsync = promisify(execFile);

const app = await App.lookup("libmodal-example", { createIfMissing: true });

console.log("Starting rsync demo...\n");

const image = Image.fromRegistry("alpine:3.21").dockerfileCommands([
  "RUN apk add --no-cache rsync",
]);

const sandbox = await app.createSandbox(image, { encryptedPorts: [873] });
console.log("Created Sandbox with ID:", sandbox.sandboxId);

let tempDir: string | undefined;

try {
  await sandbox.exec(["mkdir", "-p", "/data", "/etc/rsync"]);

  // Write the rsync daemon config file to the Sandbox.
  // Could potentially use the Sandbox filesystem API for this since it's small.
  const rsyncConfig = `
[files]
path = /data
read only = false
uid = root
gid = root
`;
  await sandbox.exec([
    "sh",
    "-c",
    `echo '${rsyncConfig}' > /etc/rsync/rsyncd.conf`,
  ]);

  await sandbox.exec([
    "sh",
    "-c",
    "rsync --daemon --config=/etc/rsync/rsyncd.conf --port=873 && echo 'Daemon started' && sleep 300",
  ]);

  await new Promise((resolve) => setTimeout(resolve, 3000));
  console.log("rsync daemon started successfully");

  tempDir = join(tmpdir(), `modal-rsync-test-${Date.now()}`);
  mkdirSync(tempDir, { recursive: true });

  writeFileSync(join(tempDir, "test-file-1.txt"), "hello, file 1!");
  writeFileSync(join(tempDir, "test-file-2.txt"), "hello, file 2!");

  const subDir = join(tempDir, "subdir");
  mkdirSync(subDir);
  writeFileSync(join(subDir, "nested-file.txt"), "hello, nested file!");

  // Generating a large random file.
  // Note that upload time will be pessimistic since it won't benefit from compression.
  const { stderr: ddStderr } = await execFileAsync("dd", [
    "if=/dev/urandom",
    "of=" + join(tempDir!, "large-test-file.bin"),
    "bs=1M",
    "count=1229", // ~1.2 GiB
  ]);
  console.log("Large test file created successfully");
  if (ddStderr?.trim()) {
    console.log(`dd output: ${ddStderr.trim()}`);
  }

  console.log(`Created local test files`);

  const tunnels = await sandbox.tunnels();
  const tunnel = tunnels[873];

  console.log("\nRunning rsync to sync files to sandbox...");

  const rsyncCommand = [
    "rsync-ssl",
    "--verbose",
    "--compress",
    "--recursive",
    "--delete",
    `${tempDir}/`,
    `rsync://${tunnel.host}:${tunnel.port}/files/`,
  ];

  console.log(`Running: ${rsyncCommand.join(" ")}`);

  const rsyncStartTime = Date.now();

  const { stdout: rsyncStdout, stderr: rsyncStderr } = await execFileAsync(
    rsyncCommand[0],
    rsyncCommand.slice(1),
    {
      env: {
        ...process.env,
        RSYNC_SSL_TYPE: "openssl",
        RSYNC_SSL_CAPATH: "",
      },
    },
  );

  const rsyncDuration = ((Date.now() - rsyncStartTime) / 1000).toFixed(1);

  console.log("rsync completed successfully");
  if (rsyncStdout?.trim()) {
    console.log("rsync stdout:");
    console.log(rsyncStdout);
  }
  if (rsyncStderr?.trim()) {
    console.log("rsync stderr:");
    console.log(rsyncStderr);
  }

  console.log(`✅ Files synced successfully in ${rsyncDuration}s!`);

  console.log("\nVerifying synced files in sandbox...\n");
  const listFiles = await sandbox.exec(["find", "/data", "-type", "f"]);
  const fileList = await listFiles.stdout.readText();
  console.log("Files in sandbox /data directory:");
  console.log(fileList);

  console.log("File sizes in sandbox:");
  const listSizes = await sandbox.exec(["ls", "-lh", "/data/"]);
  const sizeList = await listSizes.stdout.readText();
  console.log(sizeList);

  const catFile1 = await sandbox.exec(["cat", "/data/test-file-1.txt"]);
  console.log("Content of test-file-1.txt:", await catFile1.stdout.readText());

  const catNestedFile = await sandbox.exec([
    "cat",
    "/data/subdir/nested-file.txt",
  ]);
  console.log("Content of nested file:", await catNestedFile.stdout.readText());

  console.log("✅ File verification successful!\n");
} finally {
  if (tempDir) {
    console.log("Cleaning up local test files...");
    try {
      rmSync(tempDir, { recursive: true, force: true });
    } catch (cleanupError) {
      console.error(
        "Failed to clean up temporary files from ${tempDir}:",
        cleanupError,
      );
    }
  }

  console.log("Terminating sandbox...");
  await sandbox.terminate();
}
