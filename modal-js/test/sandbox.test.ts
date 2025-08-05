import { App, Volume, sandboxFromId, Secret } from "modal";
import { expect, test, onTestFinished } from "vitest";

test("CreateOneSandbox", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  expect(app.appId).toBeTruthy();

  const image = await app.imageFromRegistry("alpine:3.21");
  expect(image.imageId).toBeTruthy();

  const sb = await app.createSandbox(image);
  expect(sb.sandboxId).toBeTruthy();
  await sb.terminate();
  expect(await sb.wait()).toBe(137);
});

test("PassCatToStdin", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");

  // Spawn a sandbox running the "cat" command.
  const sb = await app.createSandbox(image, { command: ["cat"] });

  // Write to the sandbox's stdin and read from its stdout.
  await sb.stdin.writeText("this is input that should be mirrored by cat");
  await sb.stdin.close();
  expect(await sb.stdout.readText()).toBe(
    "this is input that should be mirrored by cat",
  );

  // Terminate the sandbox.
  await sb.terminate();
});

test("IgnoreLargeStdout", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("python:3.13-alpine");

  const sb = await app.createSandbox(image);
  try {
    const p = await sb.exec(["python", "-c", `print("a" * 1_000_000)`], {
      stdout: "ignore",
    });
    expect(await p.stdout.readText()).toBe(""); // Stdout is ignored
    // Stdout should be consumed after cancel, without blocking the process.
    expect(await p.wait()).toBe(0);
  } finally {
    await sb.terminate();
  }
});

test("SandboxExecOptions", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");

  const sb = await app.createSandbox(image);
  try {
    // Test with a custom working directory and timeout.
    const p = await sb.exec(["pwd"], {
      workdir: "/tmp",
      timeout: 5000,
    });

    expect(await p.stdout.readText()).toBe("/tmp\n");
    expect(await p.wait()).toBe(0);
  } finally {
    await sb.terminate();
  }
});

test("SandboxWithVolume", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");

  const volume = await Volume.fromName("libmodal-test-sandbox-volume", {
    createIfMissing: true,
  });

  const sandbox = await app.createSandbox(image, {
    command: ["echo", "volume test"],
    volumes: { "/mnt/test": volume },
  });

  expect(sandbox).toBeDefined();
  expect(sandbox.sandboxId).toMatch(/^sb-/);

  const exitCode = await sandbox.wait();
  expect(exitCode).toBe(0);
});

test("SandboxWithTunnels", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");

  const sandbox = await app.createSandbox(image, {
    command: ["cat"],
    encryptedPorts: [8443],
    unencryptedPorts: [8080],
  });

  expect(sandbox).toBeDefined();
  expect(sandbox.sandboxId).toMatch(/^sb-/);

  const tunnels = await sandbox.tunnels();
  expect(Object.keys(tunnels)).toHaveLength(2);

  // Test encrypted tunnel (port 8443)
  const encryptedTunnel = tunnels[8443];
  expect(encryptedTunnel.host).toMatch(/\.modal\.host$/);
  expect(encryptedTunnel.port).toBe(443);
  expect(encryptedTunnel.url).toMatch(/^https:\/\//);
  expect(encryptedTunnel.tlsSocket).toEqual([
    encryptedTunnel.host,
    encryptedTunnel.port,
  ]);

  // Test unencrypted tunnel (port 8080)
  const unencryptedTunnel = tunnels[8080];
  expect(unencryptedTunnel.unencryptedHost).toMatch(/\.modal\.host$/);
  expect(typeof unencryptedTunnel.unencryptedPort).toBe("number");
  expect(unencryptedTunnel.tcpSocket).toEqual([
    unencryptedTunnel.unencryptedHost,
    unencryptedTunnel.unencryptedPort,
  ]);

  await sandbox.terminate();
});

test("CreateSandboxWithSecrets", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");

  const secret = await Secret.fromName("libmodal-test-secret", {
    requiredKeys: ["c"],
  });
  expect(secret).toBeDefined();

  const sandbox = await app.createSandbox(image, {
    command: ["printenv", "c"],
    secrets: [secret],
  });
  expect(sandbox).toBeDefined();

  const result = await sandbox.stdout.readText();
  expect(result).toBe("hello world\n");
});

test("SandboxPollAndReturnCode", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");

  const sandbox = await app.createSandbox(image, { command: ["cat"] });

  expect(await sandbox.poll()).toBeNull();

  // Send input to make the cat command complete
  await sandbox.stdin.writeText("hello, sandbox");
  await sandbox.stdin.close();

  expect(await sandbox.wait()).toBe(0);
  expect(await sandbox.poll()).toBe(0);
});

test("SandboxPollAfterFailure", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");

  const sandbox = await app.createSandbox(image, {
    command: ["sh", "-c", "exit 42"],
  });

  expect(await sandbox.wait()).toBe(42);
  expect(await sandbox.poll()).toBe(42);
});

test("SandboxExecSecret", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");

  const sb = await app.createSandbox(image);
  expect(sb.sandboxId).toBeTruthy();

  onTestFinished(async () => {
    await sb.terminate();
  });

  const secret = await Secret.fromName("libmodal-test-secret", {
    requiredKeys: ["c"],
  });
  const printSecret = await sb.exec(["printenv", "c"], {
    stdout: "pipe",
    secrets: [secret],
  });
  const secretText = await printSecret.stdout.readText();
  expect(secretText).toBe("hello world\n");
});

test("SandboxFromId", async () => {
  const app = await App.lookup("libmodal-test", { createIfMissing: true });
  const image = await app.imageFromRegistry("alpine:3.21");

  const sb = await app.createSandbox(image);
  onTestFinished(async () => {
    await sb.terminate();
  });
  const sbFromId = await sandboxFromId(sb.sandboxId);
  expect(sbFromId.sandboxId).toBe(sb.sandboxId);
});
