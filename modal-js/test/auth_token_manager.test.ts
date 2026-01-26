import { describe, test, expect, vi, beforeEach, afterEach } from "vitest";
import jwt from "jsonwebtoken";
import { ModalClient } from "../src/client";
import { AuthTokenManager, REFRESH_WINDOW } from "../src/auth_token_manager";
import { newLogger } from "../src/logger";

async function eventually(
  condition: () => boolean,
  timeoutMs: number = 1000,
  intervalMs: number = 10,
): Promise<void> {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    if (condition()) {
      return;
    }
    await new Promise((resolve) => setTimeout(resolve, intervalMs));
  }
  throw new Error(`Condition not met within ${timeoutMs}ms`);
}

class mockAuthClient {
  private authToken: string = "";

  setAuthToken(token: string) {
    this.authToken = token;
  }

  authTokenGet = vi.fn(async () => {
    return { token: this.authToken };
  });
}

function newMockAuthClient() {
  return new mockAuthClient();
}

// Creates a JWT token for testing
function createTestJWT(expiry: number): string {
  return jwt.sign({ exp: expiry }, "walter-test");
}

describe("AuthTokenManager", () => {
  let mockClient: mockAuthClient;
  let manager: AuthTokenManager;

  beforeEach(() => {
    mockClient = newMockAuthClient();
    manager = new AuthTokenManager(mockClient as any, newLogger());
  });

  afterEach(() => {
    manager.stop();
  });

  test("TestAuthToken_DecodeJWT", async () => {
    const now = Math.floor(Date.now() / 1000);
    const expiry = now + 1800;
    const token = createTestJWT(expiry);
    mockClient.setAuthToken(token);

    // Test by fetching a valid JWT and checking if it gets stored properly
    await manager.start();
    expect(manager.getCurrentToken()).toBe(token);
  });

  test("TestAuthToken_InitialFetch", async () => {
    const now = Math.floor(Date.now() / 1000);
    const token = createTestJWT(now + 3600);
    mockClient.setAuthToken(token);

    await manager.start();

    const firstToken = await manager.getToken();
    expect(firstToken).toBe(token);

    const secondToken = await manager.getToken();
    expect(secondToken).toBe(token);
  });

  test("TestAuthToken_IsExpired", async () => {
    const now = Math.floor(Date.now() / 1000);

    // Test not expired
    const validToken = createTestJWT(now + 3600);
    manager.setToken(validToken, now + 3600);
    expect(manager.isExpired()).toBe(false);

    // Test expired
    const expiredToken = createTestJWT(now - 60);
    manager.setToken(expiredToken, now - 60);
    expect(manager.isExpired()).toBe(true);
  });

  test("TestAuthToken_RefreshExpiredToken", async () => {
    const now = Math.floor(Date.now() / 1000);
    const expiringToken = createTestJWT(now - 60);
    const freshToken = createTestJWT(now + 3600);

    manager.setToken(expiringToken, now - 60);
    mockClient.setAuthToken(freshToken);

    // Start the background refresh
    await manager.start();

    // Wait for background refresh to update the token
    await eventually(() => manager.getCurrentToken() === freshToken);

    // Should have the new token cached
    expect(manager.getCurrentToken()).toBe(freshToken);
  });

  test("TestAuthToken_RefreshNearExpiryToken", async () => {
    const now = Math.floor(Date.now() / 1000);
    const expiringToken = createTestJWT(now + 60);
    const freshToken = createTestJWT(now + 3600);

    manager.setToken(expiringToken, now + 60);
    mockClient.setAuthToken(freshToken);

    // Start the refresh timer
    await manager.start();

    // Wait for background refresh to update the token
    await eventually(() => manager.getCurrentToken() === freshToken);

    // Should have the new token cached
    expect(manager.getCurrentToken()).toBe(freshToken);
  });

  test("TestAuthToken_ConcurrentGetToken", async () => {
    const token = createTestJWT(Math.floor(Date.now() / 1000) + 3600);
    mockClient.setAuthToken(token);

    // Start the manager to fetch initial token
    await manager.start();

    // Multiple concurrent getToken calls should all return the same token
    const [result1, result2, result3] = await Promise.all([
      manager.getToken(),
      manager.getToken(),
      manager.getToken(),
    ]);
    expect(result1).toBe(token);
    expect(result2).toBe(token);
    expect(result3).toBe(token);

    // authTokenGet should have been called only once (during start)
    expect(mockClient.authTokenGet).toHaveBeenCalledTimes(1);
  });

  test("TestAuthToken_ConcurrentGetTokenWithExpiredToken", async () => {
    vi.useFakeTimers();
    try {
      const now = Math.floor(Date.now() / 1000);
      mockClient.setAuthToken(createTestJWT(now + REFRESH_WINDOW + 60));

      await manager.start();
      mockClient.authTokenGet.mockClear();

      const expiredToken = createTestJWT(now - 10);
      manager.setToken(expiredToken, now - 10);

      const freshToken = createTestJWT(now + 3600);
      mockClient.setAuthToken(freshToken);

      const [result1, result2, result3] = await Promise.all([
        manager.getToken(),
        manager.getToken(),
        manager.getToken(),
      ]);

      expect(result1).toBe(freshToken);
      expect(result2).toBe(freshToken);
      expect(result3).toBe(freshToken);
      expect(mockClient.authTokenGet).toHaveBeenCalledTimes(1);
    } finally {
      vi.useRealTimers();
    }
  });

  test("TestAuthToken_GetToken_NoToken", async () => {
    await expect(manager.getToken()).rejects.toThrow(
      "No valid auth token available",
    );
  });

  test("TestAuthToken_StopCleansUpResources", async () => {
    const now = Math.floor(Date.now() / 1000);
    const token = createTestJWT(now + 3600);
    mockClient.setAuthToken(token);

    await manager.start();
    expect(manager.getCurrentToken()).toBe(token);

    manager.stop();

    await new Promise((resolve) => setTimeout(resolve, 100));
  });

});

describe("ModalClient with AuthTokenManager", () => {
  test("TestModalClient_CloseCleansUpAuthTokenManager", () => {
    const mockCpClient = newMockAuthClient();
    const client = new ModalClient({
      cpClient: mockCpClient as any,
    });

    client.close();
  });

  test("TestModalClient_MultipleInstancesHaveSeparateManagers", () => {
    const mockCpClient1 = newMockAuthClient();
    const mockCpClient2 = newMockAuthClient();

    const client1 = new ModalClient({
      cpClient: mockCpClient1 as any,
    });

    const client2 = new ModalClient({
      cpClient: mockCpClient2 as any,
    });

    client1.close();
    client2.close();
  });
});
