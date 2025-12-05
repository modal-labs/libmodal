import { expect, test, vi } from "vitest";
import { ModalClient } from "../src/client";
import { MockGrpcClient } from "../test-support/grpc_mock";

function createTestJWT(exp: number): string {
  const header = "x";
  const payload = Buffer.from(JSON.stringify({ exp })).toString("base64");
  const signature = "x";
  return `${header}.${payload}.${signature}`;
}

function createMockEnvironmentClient(): {
  mockClient: ModalClient;
  mock: MockGrpcClient;
  getCallCount: () => number;
} {
  let callCount = 0;
  const mock = new MockGrpcClient();

  mock.handleUnary("/AuthTokenGet", () => {
    return { token: createTestJWT(9999999999) };
  });

  const originalHandleUnary = mock.handleUnary.bind(mock);
  mock.handleUnary = (rpcName: string, handler: any) => {
    if (rpcName === "/EnvironmentGetOrCreate") {
      originalHandleUnary(rpcName, (req: any) => {
        callCount++;
        return handler(req);
      });
    } else {
      originalHandleUnary(rpcName, handler);
    }
  };

  const mockClient = new ModalClient({
    cpClient: mock as any,
    tokenId: "test-token-id",
    tokenSecret: "test-token-secret",
  });

  return {
    mockClient,
    mock,
    getCallCount: () => callCount,
  };
}

test("GetEnvironmentCached", async () => {
  // Temporarily unset the env var so we fetch from server
  vi.stubEnv("MODAL_IMAGE_BUILDER_VERSION", undefined);

  const { mockClient, mock, getCallCount } = createMockEnvironmentClient();

  mock.handleUnary("/EnvironmentGetOrCreate", () => {
    return {
      environmentId: "en-test123",
      metadata: {
        name: "main",
        settings: {
          imageBuilderVersion: "2024.10",
          webhookSuffix: "modal.run",
        },
      },
    };
  });

  const version1 = await mockClient.getImageBuilderVersion();
  expect(version1).toBe("2024.10");
  expect(getCallCount()).toBe(1);

  const version2 = await mockClient.getImageBuilderVersion();
  expect(version2).toBe("2024.10");
  expect(getCallCount()).toBe(1); // got "" from cache

  const { mockClient: mockClientDev, mock: mockDev } = createMockEnvironmentClient();

  mockDev.handleUnary("/EnvironmentGetOrCreate", () => {
    return {
      environmentId: "en-dev",
      metadata: {
        name: "dev",
        settings: {
          imageBuilderVersion: "2025.06",
          webhookSuffix: "",
        },
      },
    };
  });

  // Create a new client with "dev" environment
  const newMockClientDev = new ModalClient({
    cpClient: mockDev as any,
    tokenId: "test-token-id",
    tokenSecret: "test-token-secret",
    environment: "dev",
  });

  const versionDev = await newMockClientDev.getImageBuilderVersion();
  expect(versionDev).toBe("2025.06");

  mockClient.close();
  mockClientDev.close();
  newMockClientDev.close();

  vi.unstubAllEnvs();
});

test("ImageBuilderVersion_LocalConfigHasPrecedence", async () => {
  const { mockClient, mock, getCallCount } = createMockEnvironmentClient();

  mock.handleUnary("/EnvironmentGetOrCreate", () => {
    return {
      environmentId: "en-test",
      metadata: {
        name: "",
        settings: {
          imageBuilderVersion: "2024.10",
          webhookSuffix: "",
        },
      },
    };
  });

  const mockClientWithConfig = new ModalClient({
    cpClient: mock as any,
    tokenId: "test-token-id",
    tokenSecret: "test-token-secret",
  });

  mockClientWithConfig.profile.imageBuilderVersion = "2024.04";

  const version = await mockClientWithConfig.getImageBuilderVersion();
  expect(version).toBe("2024.04");
  expect(getCallCount()).toBe(0); // should not fetch from server

  mockClient.close();
  mockClientWithConfig.close();
});
