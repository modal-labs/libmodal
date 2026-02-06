import { describe, expect, test } from "vitest";
import { ModalClient } from "../src/client";
import { Sandbox } from "../src/sandbox";
import { ClientError, Status } from "nice-grpc";

function makeClient(cpClient: any): ModalClient {
  return new ModalClient({
    cpClient: cpClient as any,
    tokenId: "test-id",
    tokenSecret: "test-secret",
  });
}

function textItem(data: string) {
  return { data };
}

function batch(entryId: string, items: Array<{ data: string }>, eof = false) {
  return { entryId, items, eof };
}

describe("SandboxGetLogs lazy and retry behavior", () => {
  test("testSandboxGetLogsNotCalledUntilStdoutIsAccessed", async () => {
    let calls = 0;
    const cpClient = {
      // Return an empty, immediate EOF stream
      async *sandboxGetLogs(_req: any) {
        calls++;
        yield batch("1-0", [], true);
      },
    };
    const client = makeClient(cpClient);
    const sb = new Sandbox(client, "sb-123");

    // Constructor should not trigger logs
    expect(calls).toBe(0);

    // Accessing stdout doesn't start pulling until read
    const reader = sb.stdout.getReader();
    await reader.read();
    reader.releaseLock();

    expect(calls).toBe(1);
  });

  test("testSandboxGetLogsNotCalledUntilStderrIsAccessed", async () => {
    let calls = 0;
    const cpClient = {
      async *sandboxGetLogs(_req: any) {
        calls++;
        yield batch("1-0", [], true);
      },
    };
    const client = makeClient(cpClient);
    const sb = new Sandbox(client, "sb-456");

    expect(calls).toBe(0);

    const reader = sb.stderr.getReader();
    await reader.read();
    reader.releaseLock();

    expect(calls).toBe(1);
  });

  test("testSandboxGetLogsRetriesAfterDelayOnRetriableError", async () => {
    const callTimes: number[] = [];
    let attempt = 0;
    const cpClient = {
      sandboxGetLogs(_req: any) {
        callTimes.push(Date.now());
        attempt++;
        if (attempt === 1) {
          // First attempt: retryable error
          throw new ClientError(
            "/modal.client.ModalClient/SandboxGetLogs",
            Status.UNAVAILABLE,
            "transient",
          );
        }
        // Second attempt: immediate EOF
        return (async function* () {
          yield batch("1-0", [], true);
        })();
      },
    };
    const client = makeClient(cpClient);
    const sb = new Sandbox(client, "sb-789");

    const reader = sb.stdout.getReader();
    await reader.read();
    reader.releaseLock();

    expect(callTimes.length).toBeGreaterThanOrEqual(2);
    const delta = callTimes[1] - callTimes[0];
    // Expect at least ~10ms backoff; keep upper bound generous for CI variance
    expect(delta).toBeGreaterThanOrEqual(8);
    expect(delta).toBeLessThan(500);
  });

  test("testSandboxGetLogsRetryDelayResetsAfterSuccessfulRead", async () => {
    const callTimes: number[] = [];
    let attempt = 0;
    const cpClient = {
      sandboxGetLogs(_req: any) {
        callTimes.push(Date.now());
        attempt++;
        if (attempt === 1) {
          // First: retryable error (will set next delay to 20ms internally)
          throw new ClientError(
            "/modal.client.ModalClient/SandboxGetLogs",
            Status.UNAVAILABLE,
            "transient-1",
          );
        } else if (attempt === 2) {
          // Second: successful read (resets delay back to initial)
          return (async function* () {
            yield batch("1-0", [textItem("hi")], false);
            // end of stream without eof -> outer loop will re-enter
          })();
        } else if (attempt === 3) {
          // Third: retryable error; delay should be reset to initial (~10ms)
          throw new ClientError(
            "/modal.client.ModalClient/SandboxGetLogs",
            Status.UNAVAILABLE,
            "transient-2",
          );
        }
        // Fourth: complete
        return (async function* () {
          yield batch("1-1", [], true);
        })();
      },
    };
    const client = makeClient(cpClient);
    const sb = new Sandbox(client, "sb-000");

    const reader = sb.stdout.getReader();
    await reader.read();
    reader.releaseLock();

    // We expect at least 4 invocations
    expect(callTimes.length).toBeGreaterThanOrEqual(4);
    const deltaAfterReset = callTimes[3] - callTimes[2];
    expect(deltaAfterReset).toBeGreaterThanOrEqual(8);
    expect(deltaAfterReset).toBeLessThan(500);
  });

  test("testCancellingStdoutIteratorClosesIterator", async () => {
    let cancelled = false;
    let seenSignalAbort = false;

    const cpClient = {
      sandboxGetLogs(_req: any, opts?: { signal?: AbortSignal }) {
        const signal = opts?.signal;
        return (async function* () {
          try {
            // Emit one item so the reader starts
            yield batch("1-0", [textItem("hello")], false);
            while (true) {
              if (signal?.aborted) {
                seenSignalAbort = true;
                break;
              }
              await setTimeout(10);
            }
          } finally {
            cancelled = true;
          }
        })();
      },
    };
    const client = makeClient(cpClient);
    const sb = new Sandbox(client, "sb-cancel");

    const reader = sb.stdout.getReader();
    const first = await reader.read(); // pull first chunk
    expect(first.done).toBe(false);
    // Cancel consumption
    await reader.cancel();

    // Give the generator a moment to run its finally block
    await new Promise((r) => setTimeout(r, 20));
    expect(seenSignalAbort).toBe(true);
    expect(cancelled).toBe(true);
  });
});

