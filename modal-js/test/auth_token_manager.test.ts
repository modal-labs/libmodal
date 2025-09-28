import { expect, test } from "vitest";
import jwt from "jsonwebtoken";
import { AuthTokenManager, DEFAULT_EXPIRY_OFFSET } from "../src/client";

function createTestJWT(exp?: number): string {
  const payload: any = {};
  if (exp !== undefined && exp > 0) {
    payload.exp = exp;
  }
  return jwt.sign(payload, "walter-secret-key");
}

class authTokenMockClient {
  public callCount = 0;
  public authToken = "";
  public shouldError = false;

  async authTokenGet(): Promise<{ token: string }> {
    this.callCount++;
    if (this.shouldError) {
      throw new Error("authTokenGet mock error");
    }
    return { token: this.authToken };
  }

  setAuthToken(token: string) {
    this.authToken = token;
  }

  getCallCount() {
    return this.callCount;
  }
}

test("AuthTokenManager getToken initial fetch", async () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);

  const validToken = createTestJWT(Math.floor(Date.now() / 1000) + 3600);
  mockClient.setAuthToken(validToken);

  const token = await authTokenManager.getToken();
  expect(token).toBe(validToken);
  expect(mockClient.getCallCount()).toBe(1);
});

test("AuthTokenManager getToken cached", async () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);

  const firstToken = createTestJWT(Math.floor(Date.now() / 1000) + 3600);
  mockClient.setAuthToken(firstToken);

  // Set up initial token
  const token1 = await authTokenManager.getToken();
  expect(token1).toBe(firstToken);

  // Set a bogus token in the mock, and verify we get the cached valid token
  mockClient.setAuthToken("bogus");
  const token2 = await authTokenManager.getToken();
  expect(token2).toBe(firstToken);
  expect(mockClient.getCallCount()).toBe(1);
});

test("AuthTokenManager getToken expired", async () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);

  const expiredToken = createTestJWT(Math.floor(Date.now() / 1000) - 3600);
  const refreshedToken = createTestJWT(Math.floor(Date.now() / 1000) + 3600);

  // Set up expired token directly
  (authTokenManager as any).token = expiredToken;
  (authTokenManager as any).expiry = Math.floor(Date.now() / 1000) - 3600;

  // Set up new valid token
  mockClient.setAuthToken(refreshedToken);

  const token = await authTokenManager.getToken();
  expect(token).toBe(refreshedToken);
  expect(mockClient.getCallCount()).toBe(1);
});

test("AuthTokenManager getToken near expiry", async () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);

  // Set up expiring token that expires within the refresh window
  const nearExpiryTime = Math.floor(Date.now() / 1000) + 180;
  const nearExpiryToken = createTestJWT(nearExpiryTime);

  // Set up token directly that needs refresh
  (authTokenManager as any).token = nearExpiryToken;
  (authTokenManager as any).expiry = nearExpiryTime;

  // Set up new valid token for refresh
  const refreshedToken = createTestJWT(Math.floor(Date.now() / 1000) + 3600);
  mockClient.setAuthToken(refreshedToken);

  const token = await authTokenManager.getToken();
  expect(token).toBe(refreshedToken);
  expect(mockClient.getCallCount()).toBe(1);
});

test("AuthTokenManager getToken without exp claim", async () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);

  const tokenWithoutExp = createTestJWT(); // No exp claim
  mockClient.setAuthToken(tokenWithoutExp);

  const originalWarn = console.warn;
  let warningMessage = "";
  console.warn = (message: string) => {
    warningMessage = message;
  };

  try {
    const token = await authTokenManager.getToken();
    expect(token).toBe(tokenWithoutExp);

    expect(warningMessage).toContain(
      "Failed to decode x-modal-auth-token exp field",
    );

    // Verify that the manager set the default expiry
    const expiry = (authTokenManager as any).expiry;
    expect(expiry).toBeGreaterThan(Math.floor(Date.now() / 1000));
    expect(expiry).toBeLessThanOrEqual(
      Math.floor(Date.now() / 1000) + DEFAULT_EXPIRY_OFFSET,
    );
  } finally {
    console.warn = originalWarn;
  }
});

test("AuthTokenManager getToken empty response", async () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);

  mockClient.setAuthToken("");

  await expect(authTokenManager.getToken()).rejects.toThrow(
    "Empty auth token received",
  );
});

test("AuthTokenManager concurrent getToken", async () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);

  const validToken = createTestJWT(Math.floor(Date.now() / 1000) + 3600);
  mockClient.setAuthToken(validToken);

  // Make concurrent calls
  const numCalls = 10;
  const promises = Array.from({ length: numCalls }, () =>
    authTokenManager.getToken(),
  );
  const results = await Promise.all(promises);

  // All should return the same token
  results.forEach((token) => expect(token).toBe(validToken));

  // Should have made only one call to the server
  expect(mockClient.getCallCount()).toBe(1);
});

test("AuthTokenManager concurrent refresh", async () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);

  // Set up expiring token that needs refresh
  const nearExpiryTime = Math.floor(Date.now() / 1000) + 180;
  (authTokenManager as any).token = "old.but.valid.token";
  (authTokenManager as any).expiry = nearExpiryTime;

  const newToken = createTestJWT(Math.floor(Date.now() / 1000) + 3600);
  mockClient.setAuthToken(newToken);

  // Make concurrent calls
  const numCalls = 10;
  const promises = Array.from({ length: numCalls }, () =>
    authTokenManager.getToken(),
  );
  const results = await Promise.all(promises);

  // All should succeed
  results.forEach((token) => expect(typeof token).toBe("string"));

  // At least one call should have returned the new token
  // Typically, all calls should have returned the new token, but due to timing, some may have returned the old token
  expect(results).toContain(newToken);

  // Should have made only one call to authTokenGet
  expect(mockClient.getCallCount()).toBe(1);

  // The new token should be cached now
  const finalToken = await authTokenManager.getToken();
  expect(finalToken).toBe(newToken);
});

test("AuthTokenManager decodeJWT valid", () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);
  const validToken = createTestJWT(Math.floor(Date.now() / 1000) + 3600);

  const exp = authTokenManager.decodeJWT(validToken);
  expect(exp).toBeGreaterThan(Math.floor(Date.now() / 1000));
});

test("AuthTokenManager decodeJWT without exp claim", () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);
  const tokenWithoutExp = createTestJWT(); // No exp claim

  const exp = authTokenManager.decodeJWT(tokenWithoutExp);
  expect(exp).toBe(0);
});

test("AuthTokenManager decodeJWT invalid format", () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);

  const exp = authTokenManager.decodeJWT("invalid.token");
  expect(exp).toBe(0); // Should return 0 for invalid format
});

test("AuthTokenManager needsRefresh true", () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);
  (authTokenManager as any).expiry = Math.floor(Date.now() / 1000) + 180;

  expect(authTokenManager.needsRefresh()).toBe(true);
});

test("AuthTokenManager needsRefresh false", () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);
  (authTokenManager as any).expiry = Math.floor(Date.now() / 1000) + 600;

  expect(authTokenManager.needsRefresh()).toBe(false);
});

test("AuthTokenManager isExpired true", () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);
  (authTokenManager as any).expiry = Math.floor(Date.now() / 1000) - 60;

  expect(authTokenManager.isExpired()).toBe(true);
});

test("AuthTokenManager isExpired false", () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);
  (authTokenManager as any).expiry = Math.floor(Date.now() / 1000) + 60;

  expect(authTokenManager.isExpired()).toBe(false);
});

test("AuthTokenManager multiple refresh cycles", async () => {
  const mockClient = new authTokenMockClient();
  const authTokenManager = new AuthTokenManager(mockClient as any);

  const exp = Math.floor(Date.now() / 1000) + 3600;
  const tokens = [createTestJWT(exp), createTestJWT(exp), createTestJWT(exp)];

  // First call
  mockClient.setAuthToken(tokens[0]);
  const token0 = await authTokenManager.getToken();
  expect(token0).toBe(tokens[0]);

  // Expire the token
  (authTokenManager as any).expiry = Math.floor(Date.now() / 1000) - 100;

  // Second call
  mockClient.setAuthToken(tokens[1]);
  const token1 = await authTokenManager.getToken();
  expect(token1).toBe(tokens[1]);

  // Expire again
  (authTokenManager as any).expiry = Math.floor(Date.now() / 1000) - 100;

  // Third call
  mockClient.setAuthToken(tokens[2]);
  const token2 = await authTokenManager.getToken();
  expect(token2).toBe(tokens[2]);
});
