import { describe, test, expect, vi, beforeEach, afterEach } from "vitest";
import jwt from "jsonwebtoken";
import { AuthTokenManager, DEFAULT_EXPIRY_OFFSET } from "../src/client";

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

    // Decoding valid JWT
    await manager.getToken();
    expect(manager.getExpiry()).toBe(expiry);

    // Decoding invalid JWT
    const tokenWithoutExp = jwt.sign({ sub: "test" }, "walter-test");
    mockClient.setAuthToken(tokenWithoutExp);

    const beforeCall = Math.floor(Date.now() / 1000);
    await manager.refreshToken();
    const afterCall = Math.floor(Date.now() / 1000);

    const newExpiry = manager.getExpiry();
    expect(newExpiry).toBeGreaterThanOrEqual(
      beforeCall + DEFAULT_EXPIRY_OFFSET,
    );
    expect(newExpiry).toBeLessThanOrEqual(afterCall + DEFAULT_EXPIRY_OFFSET);
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

  test("TestAuthToken_NeedsRefresh", async () => {
    const now = Math.floor(Date.now() / 1000);

    // Doesn't need refresh
    const freshToken = createTestJWT(now + 3600);
    manager.setToken(freshToken, now + 3600);
    expect(manager.needsRefresh()).toBe(false);

    // Needs refresh
    const nearExpiryToken = createTestJWT(now + 60);
    manager.setToken(nearExpiryToken, now + 60);
    expect(manager.needsRefresh()).toBe(true);
  });

  test("TestAuthToken_RefreshExpiredToken", async () => {
    const now = Math.floor(Date.now() / 1000);
    const expiringToken = createTestJWT(now - 60);
    const freshToken = createTestJWT(now + 3600);

    manager.setToken(expiringToken, now - 60);
    mockClient.setAuthToken(freshToken);

    // Start the background refresh
    await manager.start();

    // Brief wait for refresh to complete. TODO(walter): Can adjust if flaky.
    await new Promise((resolve) => setTimeout(resolve, 10));

    // Should have the new token cached
    expect(manager.getCurrentToken()).toBe(freshToken);
    expect(manager.needsRefresh()).toBe(false);
  });

  test("TestAuthToken_RefreshNearExpiryToken", async () => {
    const now = Math.floor(Date.now() / 1000);
    const expiringToken = createTestJWT(now + 60);
    const freshToken = createTestJWT(now + 3600);

    manager.setToken(expiringToken, now + 60);
    mockClient.setAuthToken(freshToken);

    expect(manager.needsRefresh()).toBe(true);

    // Start the refresh timer
    await manager.start();

    // Brief wait for refresh to complete. TODO(walter): Can adjust if flaky.
    await new Promise((resolve) => setTimeout(resolve, 10));

    // Should have the new token cached
    expect(manager.getCurrentToken()).toBe(freshToken);
    expect(manager.needsRefresh()).toBe(false);
  });
});
