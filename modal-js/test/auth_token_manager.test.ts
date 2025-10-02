import { describe, test, expect, vi, beforeEach, afterEach } from "vitest";
import jwt from "jsonwebtoken";
import { AuthTokenManager } from "../src/client";

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
    manager = new AuthTokenManager(mockClient as any);
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
    const result = await manager.refreshToken();
    expect(result).toBe(token);
    expect(manager.getCurrentToken()).toBe(token);
  });

  test("TestAuthToken_InitialFetch", async () => {
    const now = Math.floor(Date.now() / 1000);
    const token = createTestJWT(now + 3600);
    mockClient.setAuthToken(token);

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

  test("TestAuthToken_ConcurrentRefresh", async () => {
    const token = createTestJWT(Math.floor(Date.now() / 1000) + 3600);
    mockClient.setAuthToken(token);

    // Multiple refresh calls
    const promise1 = manager.refreshToken();
    const promise2 = manager.refreshToken();
    const promise3 = manager.refreshToken();

    // All should return the same promise
    expect(promise1).toBe(promise2);
    expect(promise2).toBe(promise3);

    // All should resolve to the same token
    const [result1, result2, result3] = await Promise.all([
      promise1,
      promise2,
      promise3,
    ]);
    expect(result1).toBe(token);
    expect(result2).toBe(token);
    expect(result3).toBe(token);

    // authTokenGet should have been called only once
    expect(mockClient.authTokenGet).toHaveBeenCalledTimes(1);
  });
});
