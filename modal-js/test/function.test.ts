import { Function_, NotFoundError } from "modal";
import { expect, test } from "vitest";

test("FunctionCall", async () => {
  const function_ = await Function_.lookup(
    "libmodal-test-support",
    "echo_string",
  );

  // Represent Python kwargs.
  const resultKwargs = await function_.remote([], { s: "hello" });
  expect(resultKwargs).toBe("output: hello");

  // Try the same, but with args.
  const resultArgs = await function_.remote(["hello"]);
  expect(resultArgs).toBe("output: hello");
});

test("FunctionCallLargeInput", async () => {
  const function_ = await Function_.lookup(
    "libmodal-test-support",
    "bytelength",
  );
  const len = 3 * 1000 * 1000; // More than 2 MiB, offload to blob storage
  const input = new Uint8Array(len);
  const result = await function_.remote([input]);
  expect(result).toBe(len);
});

test("FunctionNotFound", async () => {
  const promise = Function_.lookup(
    "libmodal-test-support",
    "not_a_real_function",
  );
  await expect(promise).rejects.toThrowError(NotFoundError);
});

test("FunctionCallInputPlane", async () => {
  const function_ = await Function_.lookup(
    "libmodal-test-support",
    "input_plane",
  );
  const result = await function_.remote(["hello"]);
  expect(result).toBe("output: hello");
});

test("FunctionGetCurrentStats", async () => {
  const function_ = await Function_.lookup(
    "libmodal-test-support",
    "echo_string",
  );
  const stats = await function_.getCurrentStats();
  expect(typeof stats.backlog).toBe("number");
  expect(typeof stats.numTotalRunners).toBe("number");
  expect(stats.backlog).toBeGreaterThanOrEqual(0);
  expect(stats.numTotalRunners).toBeGreaterThanOrEqual(0);
});

test("FunctionUpdateAutoscaler", async () => {
  const function_ = await Function_.lookup(
    "libmodal-test-support",
    "echo_string",
  );
  // Test updating various autoscaler settings - should not throw
  await function_.updateAutoscaler({
    minContainers: 1,
    maxContainers: 10,
    bufferContainers: 2,
    scaledownWindow: 300,
  });

  // Test partial updates
  await function_.updateAutoscaler({
    minContainers: 2,
  });

  await function_.updateAutoscaler({
    scaledownWindow: 600,
  });
});
