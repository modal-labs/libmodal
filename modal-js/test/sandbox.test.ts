import { tc } from "../test-support/test-client";
import { parseGpuConfig } from "../src/app";
import { buildSandboxCreateRequestProto } from "../src/sandbox";
import { expect, test, onTestFinished } from "vitest";
import { buildContainerExecRequestProto } from "../src/sandbox";
import {
  GPUConfig,
  PTYInfo_PTYType,
  NetworkAccess_NetworkAccessType,
} from "../proto/modal_proto/api";

test("CreateOneSandbox", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sb = await tc.sandboxes.create(app, image);
  onTestFinished(async () => {
    await sb.terminate();
    expect(await sb.wait()).toBe(137);
  });
  expect(sb.sandboxId).toBeTruthy();
});

test("PassCatToStdin", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sb = await tc.sandboxes.create(app, image, { command: ["cat"] });
  onTestFinished(async () => await sb.terminate());

  await sb.stdin.writeText("this is input that should be mirrored by cat");
  await sb.stdin.close();
  expect(await sb.stdout.readText()).toBe(
    "this is input that should be mirrored by cat",
  );
});

test("IgnoreLargeStdout", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("python:3.13-alpine");

  const sb = await tc.sandboxes.create(app, image);
  onTestFinished(async () => await sb.terminate());

  const p = await sb.exec(["python", "-c", `print("a" * 1_000_000)`], {
    stdout: "ignore",
  });
  expect(await p.stdout.readText()).toBe(""); // Stdout is ignored
  // Stdout should be consumed after cancel, without blocking the process.
  expect(await p.wait()).toBe(0);
});

test("SandboxCreateOptions", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sb = await tc.sandboxes.create(app, image, {
    command: ["echo", "hello, params"],
    cloud: "aws",
    regions: ["us-east-1", "us-west-2"],
    verbose: true,
  });
  onTestFinished(async () => await sb.terminate());

  expect(sb.sandboxId).toMatch(/^sb-/);

  const exitCode = await sb.wait();
  expect(exitCode).toBe(0);

  await expect(
    tc.sandboxes.create(app, image, {
      cloud: "invalid-cloud",
    }),
  ).rejects.toThrow("INVALID_ARGUMENT");

  await expect(
    tc.sandboxes.create(app, image, {
      regions: ["invalid-region"],
    }),
  ).rejects.toThrow("INVALID_ARGUMENT");
});

test("SandboxExecOptions", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sb = await tc.sandboxes.create(app, image);
  onTestFinished(async () => await sb.terminate());
  const p = await sb.exec(["pwd"], {
    workdir: "/tmp",
    timeoutMs: 5000,
  });

  expect(await p.stdout.readText()).toBe("/tmp\n");
  expect(await p.wait()).toBe(0);
});

test("parseGpuConfig", () => {
  expect(parseGpuConfig(undefined)).toEqual(GPUConfig.create({}));
  expect(parseGpuConfig("T4")).toEqual({
    type: 0,
    count: 1,
    gpuType: "T4",
  });
  expect(parseGpuConfig("A10G")).toEqual({
    type: 0,
    count: 1,
    gpuType: "A10G",
  });
  expect(parseGpuConfig("A100-80GB")).toEqual({
    type: 0,
    count: 1,
    gpuType: "A100-80GB",
  });
  expect(parseGpuConfig("A100-80GB:3")).toEqual({
    type: 0,
    count: 3,
    gpuType: "A100-80GB",
  });
  expect(parseGpuConfig("T4:2")).toEqual({
    type: 0,
    count: 2,
    gpuType: "T4",
  });
  expect(parseGpuConfig("a100:4")).toEqual({
    type: 0,
    count: 4,
    gpuType: "A100",
  });

  expect(() => parseGpuConfig("T4:invalid")).toThrow(
    "Invalid GPU count: invalid. Value must be a positive integer.",
  );
  expect(() => parseGpuConfig("T4:")).toThrow(
    "Invalid GPU count: . Value must be a positive integer.",
  );
  expect(() => parseGpuConfig("T4:0")).toThrow(
    "Invalid GPU count: 0. Value must be a positive integer.",
  );
  expect(() => parseGpuConfig("T4:-1")).toThrow(
    "Invalid GPU count: -1. Value must be a positive integer.",
  );
});

test("SandboxWithVolume", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const volume = await tc.volumes.fromName("libmodal-test-sandbox-volume", {
    createIfMissing: true,
  });

  const sb = await tc.sandboxes.create(app, image, {
    command: ["echo", "volume test"],
    volumes: { "/mnt/test": volume },
  });
  onTestFinished(async () => await sb.terminate());

  expect(sb.sandboxId).toMatch(/^sb-/);

  const exitCode = await sb.wait();
  expect(exitCode).toBe(0);
});

test("SandboxWithReadOnlyVolume", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const volume = await tc.volumes.fromName("libmodal-test-sandbox-volume", {
    createIfMissing: true,
  });

  const readOnlyVolume = volume.readOnly();
  expect(readOnlyVolume.isReadOnly).toBe(true);

  const sb = await tc.sandboxes.create(app, image, {
    command: ["sh", "-c", "echo 'test' > /mnt/test/test.txt"],
    volumes: { "/mnt/test": readOnlyVolume },
  });
  onTestFinished(async () => await sb.terminate());

  expect(await sb.wait()).toBe(1);
  expect(await sb.stderr.readText()).toContain("Read-only file system");
});

test("SandboxWithTunnels", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sb = await tc.sandboxes.create(app, image, {
    command: ["cat"],
    encryptedPorts: [8443],
    unencryptedPorts: [8080],
  });
  onTestFinished(async () => await sb.terminate());

  expect(sb.sandboxId).toMatch(/^sb-/);

  const tunnels = await sb.tunnels();
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
});

test("CreateSandboxWithSecrets", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const secret = await tc.secrets.fromName("libmodal-test-secret", {
    requiredKeys: ["c"],
  });

  const sb = await tc.sandboxes.create(app, image, {
    command: ["printenv", "c"],
    secrets: [secret],
  });
  onTestFinished(async () => await sb.terminate());

  const result = await sb.stdout.readText();
  expect(result).toBe("hello world\n");
});

test("CreateSandboxWithNetworkAccessParams", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sb = await tc.sandboxes.create(app, image, {
    command: ["echo", "hello, network access"],
    blockNetwork: false,
    cidrAllowlist: ["10.0.0.0/8", "192.168.0.0/16"],
  });
  onTestFinished(async () => await sb.terminate());

  expect(sb.sandboxId).toMatch(/^sb-/);

  const exitCode = await sb.wait();
  expect(exitCode).toBe(0);

  await expect(
    tc.sandboxes.create(app, image, {
      blockNetwork: false,
      cidrAllowlist: ["not-an-ip/8"],
    }),
  ).rejects.toThrow("Invalid CIDR: not-an-ip/8");

  await expect(
    tc.sandboxes.create(app, image, {
      blockNetwork: true,
      cidrAllowlist: ["10.0.0.0/8"],
    }),
  ).rejects.toThrow(
    "cidrAllowlist cannot be used when blockNetwork is enabled",
  );
});

test("SandboxPollAndReturnCode", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sb = await tc.sandboxes.create(app, image, { command: ["cat"] });
  onTestFinished(async () => await sb.terminate());

  expect(await sb.poll()).toBeNull();

  // Send input to make the cat command complete
  await sb.stdin.writeText("hello, Sandbox");
  await sb.stdin.close();

  expect(await sb.wait()).toBe(0);
  expect(await sb.poll()).toBe(0);
});

test("SandboxPollAfterFailure", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sb = await tc.sandboxes.create(app, image, {
    command: ["sh", "-c", "exit 42"],
  });
  onTestFinished(async () => await sb.terminate());

  expect(await sb.wait()).toBe(42);
  expect(await sb.poll()).toBe(42);
});

test("SandboxExecSecret", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sb = await tc.sandboxes.create(app, image);
  onTestFinished(async () => await sb.terminate());

  const secret = await tc.secrets.fromName("libmodal-test-secret", {
    requiredKeys: ["c"],
  });
  const secret2 = await tc.secrets.fromObject({ d: "3" });
  const printSecret = await sb.exec(["printenv", "c", "d"], {
    stdout: "pipe",
    secrets: [secret, secret2],
  });
  const secretText = await printSecret.stdout.readText();
  expect(secretText).toBe("hello world\n3\n");
});

test("SandboxFromId", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sb = await tc.sandboxes.create(app, image);
  onTestFinished(async () => await sb.terminate());

  const sbFromId = await tc.sandboxes.fromId(sb.sandboxId);
  expect(sbFromId.sandboxId).toBe(sb.sandboxId);
});

test("SandboxWithWorkdir", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sb = await tc.sandboxes.create(app, image, {
    command: ["pwd"],
    workdir: "/tmp",
  });
  onTestFinished(async () => await sb.terminate());

  expect(await sb.stdout.readText()).toBe("/tmp\n");
});

test("SandboxWithWorkdirValidation", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  await expect(
    tc.sandboxes.create(app, image, {
      workdir: "relative/path",
    }),
  ).rejects.toThrow("workdir must be an absolute path, got: relative/path");
});

test("SandboxSetTagsAndList", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sb = await tc.sandboxes.create(app, image);
  onTestFinished(async () => await sb.terminate());

  const unique = `${Math.random()}`;

  const foundBefore: string[] = [];
  for await (const s of tc.sandboxes.list({ tags: { "test-key": unique } })) {
    foundBefore.push(s.sandboxId);
  }
  expect(foundBefore.length).toBe(0);

  await sb.setTags({ "test-key": unique });

  const foundAfter: string[] = [];
  for await (const s of tc.sandboxes.list({ tags: { "test-key": unique } })) {
    foundAfter.push(s.sandboxId);
  }
  expect(foundAfter).toEqual([sb.sandboxId]);
});

test("SandboxSetMultipleTagsAndList", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sb = await tc.sandboxes.create(app, image);
  onTestFinished(async () => await sb.terminate());

  const tagA = `A-${Math.random()}`;
  const tagB = `B-${Math.random()}`;
  const tagC = `C-${Math.random()}`;

  expect(await sb.getTags()).toEqual({});

  await sb.setTags({ "key-a": tagA, "key-b": tagB, "key-c": tagC });

  expect(await sb.getTags()).toEqual({
    "key-a": tagA,
    "key-b": tagB,
    "key-c": tagC,
  });

  let ids: string[] = [];
  for await (const s of tc.sandboxes.list({ tags: { "key-a": tagA } })) {
    ids.push(s.sandboxId);
  }
  expect(ids).toEqual([sb.sandboxId]);

  ids = [];
  for await (const s of tc.sandboxes.list({
    tags: { "key-a": tagA, "key-b": tagB },
  })) {
    ids.push(s.sandboxId);
  }
  expect(ids).toEqual([sb.sandboxId]);

  ids = [];
  for await (const s of tc.sandboxes.list({
    tags: { "key-a": tagA, "key-b": tagB, "key-d": "not-set" },
  })) {
    ids.push(s.sandboxId);
  }
  expect(ids.length).toBe(0);
});

test("SandboxListByAppId", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sb = await tc.sandboxes.create(app, image);
  onTestFinished(async () => await sb.terminate());

  let count = 0;
  for await (const s of tc.sandboxes.list({ appId: app.appId })) {
    expect(s.sandboxId).toMatch(/^sb-/);
    count++;
    if (count > 0) break;
  }
  expect(count).toBeGreaterThan(0);
});

test("NamedSandbox", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  const sandboxName = `test-sandbox-${Math.random().toString().substring(2, 10)}`;

  const sb = await tc.sandboxes.create(app, image, {
    name: sandboxName,
    command: ["sleep", "60"],
  });
  onTestFinished(async () => await sb.terminate());

  const sb1FromName = await tc.sandboxes.fromName("libmodal-test", sandboxName);
  expect(sb1FromName.sandboxId).toBe(sb.sandboxId);
  const sb2FromName = await tc.sandboxes.fromName("libmodal-test", sandboxName);
  expect(sb2FromName.sandboxId).toBe(sb1FromName.sandboxId);

  await expect(
    tc.sandboxes.create(app, image, {
      name: sandboxName,
      command: ["sleep", "60"],
    }),
  ).rejects.toThrow("already exists");
});

test("NamedSandboxNotFound", async () => {
  await expect(
    tc.sandboxes.fromName("libmodal-test", "non-existent-sandbox"),
  ).rejects.toThrow("not found");
});

test("buildContainerExecRequestProto without PTY", async () => {
  const req = await buildContainerExecRequestProto("task-123", ["bash"]);

  expect(req.ptyInfo).toBeUndefined();
});

test("buildContainerExecRequestProto with PTY", async () => {
  const req = await buildContainerExecRequestProto("task-123", ["bash"], {
    pty: true,
  });

  const ptyInfo = req.ptyInfo!;
  expect(ptyInfo.enabled).toBe(true);
  expect(ptyInfo.winszRows).toBe(24);
  expect(ptyInfo.winszCols).toBe(80);
  expect(ptyInfo.envTerm).toBe("xterm-256color");
  expect(ptyInfo.envColorterm).toBe("truecolor");
  expect(ptyInfo.ptyType).toBe(PTYInfo_PTYType.PTY_TYPE_SHELL);
  expect(ptyInfo.noTerminateOnIdleStdin).toBe(true);
});

test("buildContainerExecRequestProto_defaults", async () => {
  const req = await buildContainerExecRequestProto("task-123", ["bash"]);

  expect(req.workdir).toBeUndefined();
  expect(req.timeoutSecs).toBe(0);
  expect(req.secretIds).toEqual([]);
  expect(req.ptyInfo).toBeUndefined();
});

test("buildSandboxCreateRequestProto without PTY", async () => {
  const req = await buildSandboxCreateRequestProto("app-123", "img-456");

  const definition = req.definition!;
  expect(definition.ptyInfo).toBeUndefined();
});

test("buildSandboxCreateRequestProto with PTY", async () => {
  const req = await buildSandboxCreateRequestProto("app-123", "img-456", {
    pty: true,
  });

  const definition = req.definition!;
  const ptyInfo = definition.ptyInfo!;
  expect(ptyInfo.enabled).toBe(true);
  expect(ptyInfo.winszRows).toBe(24);
  expect(ptyInfo.winszCols).toBe(80);
  expect(ptyInfo.envTerm).toBe("xterm-256color");
  expect(ptyInfo.envColorterm).toBe("truecolor");
  expect(ptyInfo.ptyType).toBe(PTYInfo_PTYType.PTY_TYPE_SHELL);
});

test("buildSandboxCreateRequestProto with CPU and CPULimit", async () => {
  const req = await buildSandboxCreateRequestProto("app-123", "img-456", {
    cpu: 2.0,
    cpuLimit: 4.5,
  });

  const resources = req.definition!.resources!;
  expect(resources.milliCpu).toBe(2000);
  expect(resources.milliCpuMax).toBe(4500);
});

test("buildSandboxCreateRequestProto CPULimit lower than CPU", async () => {
  await expect(
    buildSandboxCreateRequestProto("app-123", "img-456", {
      cpu: 4.0,
      cpuLimit: 2.0,
    }),
  ).rejects.toThrow("cpu (4) cannot be higher than cpuLimit (2)");
});

test("buildSandboxCreateRequestProto CPULimit without CPU", async () => {
  await expect(
    buildSandboxCreateRequestProto("app-123", "img-456", {
      cpuLimit: 4.0,
    }),
  ).rejects.toThrow("must also specify cpu when cpuLimit is specified");
});

test("buildSandboxCreateRequestProto with Memory and MemoryLimit", async () => {
  const req = await buildSandboxCreateRequestProto("app-123", "img-456", {
    memoryMiB: 1024,
    memoryLimitMiB: 2048,
  });

  const resources = req.definition!.resources!;
  expect(resources.memoryMb).toBe(1024);
  expect(resources.memoryMbMax).toBe(2048);
});

test("buildSandboxCreateRequestProto MemoryLimit lower than Memory", async () => {
  await expect(
    buildSandboxCreateRequestProto("app-123", "img-456", {
      memoryMiB: 2048,
      memoryLimitMiB: 1024,
    }),
  ).rejects.toThrow(
    "the memoryMiB request (2048) cannot be higher than memoryLimitMiB (1024)",
  );
});

test("buildSandboxCreateRequestProto MemoryLimit without Memory", async () => {
  await expect(
    buildSandboxCreateRequestProto("app-123", "img-456", {
      memoryLimitMiB: 2048,
    }),
  ).rejects.toThrow(
    "must also specify memoryMiB when memoryLimitMiB is specified",
  );
});

test("buildSandboxCreateRequestProto negative CPU", async () => {
  await expect(
    buildSandboxCreateRequestProto("app-123", "img-456", {
      cpu: -1.0,
    }),
  ).rejects.toThrow("must be a positive number");
});

test("buildSandboxCreateRequestProto negative Memory", async () => {
  await expect(
    buildSandboxCreateRequestProto("app-123", "img-456", {
      memoryMiB: -100,
    }),
  ).rejects.toThrow("must be a positive number");
});

test("ConnectToken", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("python:3.12-alpine");

  const sb = await tc.sandboxes.create(app, image);
  onTestFinished(async () => {
    await sb.terminate();
  });

  const creds = await sb.createConnectToken({ userMetadata: "abc" });
  expect(creds.token).toBeTruthy();
  expect(creds.url).toBeTruthy();
});

test("buildSandboxCreateRequestProto_defaults", async () => {
  const req = await buildSandboxCreateRequestProto("app-123", "img-456");
  const def = req.definition!;

  expect(def.timeoutSecs).toBe(300);
  expect(def.entrypointArgs).toEqual([]);
  expect(def.networkAccess?.networkAccessType).toBe(
    NetworkAccess_NetworkAccessType.OPEN,
  );
  expect(def.networkAccess?.allowedCidrs).toEqual([]);
  expect(def.verbose).toBe(false);
  expect(def.cloudProviderStr).toBe("");
  expect(def.resources?.milliCpu).toBe(0);
  expect(def.resources?.memoryMb).toBe(0);
  expect(def.ptyInfo).toBeUndefined();
  expect(def.idleTimeoutSecs).toBeUndefined();
  expect(def.workdir).toBeUndefined();
  expect(def.schedulerPlacement).toBeUndefined();
  expect(def.proxyId).toBeUndefined();
  expect(def.volumeMounts).toEqual([]);
  expect(def.cloudBucketMounts).toEqual([]);
  expect(def.secretIds).toEqual([]);
  expect(def.openPorts?.ports).toEqual([]);
  expect(def.name).toBeUndefined();
});

test("sandboxInvalidTimeouts", async () => {
  const app = await tc.apps.fromName("libmodal-test", {
    createIfMissing: true,
  });
  const image = tc.images.fromRegistry("alpine:3.21");

  await expect(
    tc.sandboxes.create(app, image, { timeoutMs: 0 }),
  ).rejects.toThrow(/timeoutMs must be positive/);

  await expect(
    tc.sandboxes.create(app, image, { timeoutMs: -1000 }),
  ).rejects.toThrow(/timeoutMs must be positive/);

  await expect(
    tc.sandboxes.create(app, image, { timeoutMs: 1500 }),
  ).rejects.toThrow(/timeoutMs must be a multiple of 1000ms/);

  await expect(
    tc.sandboxes.create(app, image, { idleTimeoutMs: 0 }),
  ).rejects.toThrow(/idleTimeoutMs must be positive/);

  await expect(
    tc.sandboxes.create(app, image, { idleTimeoutMs: -2000 }),
  ).rejects.toThrow(/idleTimeoutMs must be positive/);

  await expect(
    tc.sandboxes.create(app, image, { idleTimeoutMs: 2500 }),
  ).rejects.toThrow(/idleTimeoutMs must be a multiple of 1000ms/);

  const sandbox = await tc.sandboxes.create(app, image);
  onTestFinished(async () => await sandbox.terminate());

  await expect(
    sandbox.exec(["echo", "test"], { timeoutMs: 0 }),
  ).rejects.toThrow(/timeoutMs must be positive/);

  await expect(
    sandbox.exec(["echo", "test"], { timeoutMs: -5000 }),
  ).rejects.toThrow(/timeoutMs must be positive/);

  await expect(
    sandbox.exec(["echo", "test"], { timeoutMs: 1500 }),
  ).rejects.toThrow(/timeoutMs must be a multiple of 1000ms/);
});

test("buildSandboxCreateRequestProto with experimentalOptions", async () => {
  const req = await buildSandboxCreateRequestProto("app-123", "img-456", {
    experimentalOptions: { enable_docker: true },
  });

  const def = req.definition!;
  expect(def.experimentalOptions).toEqual({ enable_docker: true });
});

test("buildSandboxCreateRequestProto with _experimentalEnableSnapshot", async () => {
  const req = await buildSandboxCreateRequestProto("app-123", "img-456", {
    _experimentalEnableSnapshot: true,
  });

  const def = req.definition!;
  expect(def.enableSnapshot).toBe(true);
});

test("buildSandboxCreateRequestProto defaults experimental fields", async () => {
  const req = await buildSandboxCreateRequestProto("app-123", "img-456");
  const def = req.definition!;

  expect(def.experimentalOptions).toEqual({});
  expect(def.enableSnapshot).toBe(false);
});
