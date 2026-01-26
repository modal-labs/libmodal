import { expect, test, vi } from "vitest";
import {
  parseJwtExpiration,
  callWithRetriesOnTransientErrors,
  callWithAuthRetry,
} from "../src/task_command_router_client";
import { ClientError, Status } from "nice-grpc";

const mockLogger = {
  debug: vi.fn(),
  info: vi.fn(),
  warn: vi.fn(),
  error: vi.fn(),
};

function mockJwt(exp: number | string | null): string {
  const header = btoa(JSON.stringify({ alg: "HS256", typ: "JWT" }));
  const payload =
    exp !== null ? btoa(JSON.stringify({ exp })) : btoa(JSON.stringify({}));
  const signature = "fake-signature";
  return `${header}.${payload}.${signature}`;
}

test("parseJwtExpiration with valid JWT", () => {
  const exp = Math.floor(Date.now() / 1000) + 3600;
  const jwt = mockJwt(exp);
  const result = parseJwtExpiration(jwt, mockLogger);
  expect(result).toBe(exp);
});

test("parseJwtExpiration without exp claim", () => {
  const jwt = mockJwt(null);
  const result = parseJwtExpiration(jwt, mockLogger);
  expect(result).toBeNull();
});

test("parseJwtExpiration with malformed JWT (wrong number of parts)", () => {
  const jwt = "only.two";
  const result = parseJwtExpiration(jwt, mockLogger);
  expect(result).toBeNull();
});

test("parseJwtExpiration with invalid base64", () => {
  const jwt = "invalid.!!!invalid!!!.signature";
  const result = parseJwtExpiration(jwt, mockLogger);
  expect(result).toBeNull();
  expect(mockLogger.warn).toHaveBeenCalled();
});

test("parseJwtExpiration with non-numeric exp", () => {
  const jwt = mockJwt("not-a-number");
  const result = parseJwtExpiration(jwt, mockLogger);
  expect(result).toBeNull();
});

test("callWithRetriesOnTransientErrors success on first attempt", async () => {
  const func = vi.fn().mockResolvedValue("success");
  const result = await callWithRetriesOnTransientErrors(func);
  expect(result).toBe("success");
  expect(func).toHaveBeenCalledTimes(1);
});

test.each([
  [Status.DEADLINE_EXCEEDED, "timeout"],
  [Status.UNAVAILABLE, "unavailable"],
  [Status.CANCELLED, "cancelled"],
  [Status.INTERNAL, "internal error"],
  [Status.UNKNOWN, "unknown error"],
])(
  "callWithRetriesOnTransientErrors retries on %s",
  async (status, message) => {
    const func = vi
      .fn()
      .mockRejectedValueOnce(new ClientError("/test", status, message))
      .mockResolvedValue("success");
    const result = await callWithRetriesOnTransientErrors(func, 10);
    expect(result).toBe("success");
    expect(func).toHaveBeenCalledTimes(2);
  },
);

test("callWithRetriesOnTransientErrors non-retryable error", async () => {
  const error = new ClientError("/test", Status.INVALID_ARGUMENT, "invalid");
  const func = vi.fn().mockRejectedValue(error);
  await expect(callWithRetriesOnTransientErrors(func, 10)).rejects.toThrow(
    error,
  );
  expect(func).toHaveBeenCalledTimes(1);
});

test("callWithRetriesOnTransientErrors max retries exceeded", async () => {
  const error = new ClientError("/test", Status.UNAVAILABLE, "unavailable");
  const func = vi.fn().mockRejectedValue(error);
  const maxRetries = 3;
  await expect(
    callWithRetriesOnTransientErrors(func, 10, 2, maxRetries),
  ).rejects.toThrow(error);
  expect(func).toHaveBeenCalledTimes(maxRetries + 1);
});

test("callWithRetriesOnTransientErrors deadline exceeded", async () => {
  const error = new ClientError("/test", Status.UNAVAILABLE, "unavailable");
  const func = vi.fn().mockRejectedValue(error);
  const deadline = Date.now() + 50;
  await expect(
    callWithRetriesOnTransientErrors(func, 100, 2, null, deadline),
  ).rejects.toThrow("Deadline exceeded");
});

test("callWithAuthRetry success on first attempt", async () => {
  const func = vi.fn().mockResolvedValue("success");
  const onAuthError = vi.fn();
  const result = await callWithAuthRetry(func, onAuthError);
  expect(result).toBe("success");
  expect(func).toHaveBeenCalledTimes(1);
  expect(onAuthError).not.toHaveBeenCalled();
});

test("callWithAuthRetry retries on UNAUTHENTICATED error", async () => {
  const func = vi
    .fn()
    .mockRejectedValueOnce(
      new ClientError("/test", Status.UNAUTHENTICATED, "unauthenticated"),
    )
    .mockResolvedValue("success");
  const onAuthError = vi.fn().mockResolvedValue(undefined);
  const result = await callWithAuthRetry(func, onAuthError);
  expect(result).toBe("success");
  expect(func).toHaveBeenCalledTimes(2);
  expect(onAuthError).toHaveBeenCalledTimes(1);
});

test("callWithAuthRetry does not retry on non-UNAUTHENTICATED errors", async () => {
  const error = new ClientError("/test", Status.INVALID_ARGUMENT, "invalid");
  const func = vi.fn().mockRejectedValue(error);
  const onAuthError = vi.fn();
  await expect(callWithAuthRetry(func, onAuthError)).rejects.toThrow(
    error,
  );
  expect(func).toHaveBeenCalledTimes(1);
  expect(onAuthError).not.toHaveBeenCalled();
});

test("callWithAuthRetry throws if still UNAUTHENTICATED after retry", async () => {
  const error = new ClientError(
    "/test",
    Status.UNAUTHENTICATED,
    "still unauthenticated",
  );
  const func = vi.fn().mockRejectedValue(error);
  const onAuthError = vi.fn().mockResolvedValue(undefined);
  await expect(callWithAuthRetry(func, onAuthError)).rejects.toThrow(
    error,
  );
  expect(func).toHaveBeenCalledTimes(2);
  expect(onAuthError).toHaveBeenCalledTimes(1);
});
