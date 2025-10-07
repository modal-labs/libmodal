import { tc } from "../test-support/test-client";
import { NotFoundError } from "modal";
import { expect, test } from "vitest";
import { createMockModalClients } from "../test-support/grpc_mock";
import { Function_ } from "../src/function";

test("FunctionCall", async () => {
  const function_ = await tc.functions.fromName(
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

test("FunctionCallJsMap", async () => {
  const function_ = await Function_.lookup(
    "libmodal-test-support",
    "identity_with_repr",
  );

  // Represent Python kwargs.
  const resultKwargs = await function_.remote([new Map([["a", "b"]])]);
  expect(resultKwargs).toStrictEqual([{ a: "b" }, "{'a': 'b'}"]);
});

test("FunctionCallDateTimeRoundtrip", async () => {
  const function_ = await Function_.lookup(
    "libmodal-test-support",
    "identity_with_repr",
  );

  // Test: Send a JS Date to Python and see how it's represented
  const testDate = new Date("2024-01-15T10:30:45.123Z");
  const result = await function_.remote([testDate]);

  // Parse the result - identity_with_repr returns [input, repr(input)]
  expect(Array.isArray(result)).toBe(true);
  expect(result).toHaveLength(2);

  const [identityResult, reprResult] = result as [unknown, string];

  console.log("JS sent:", testDate.toISOString());
  console.log("JS received back:", identityResult);
  console.log("Python repr:", reprResult);

  // Check the Python representation
  expect(reprResult).toContain("datetime.datetime");
  expect(reprResult).toContain("2024");
  console.log("✅ SUCCESS: JS Date was received as Python datetime.datetime");

  // Verify the roundtrip - we should get back a Date
  expect(identityResult).toBeInstanceOf(Date);
  const receivedDate = identityResult as Date;

  // Check precision - JavaScript Date has millisecond precision
  // Python datetime has microsecond precision
  // We should get back millisecond precision (lose sub-millisecond)
  const timeDiff = Math.abs(testDate.getTime() - receivedDate.getTime());
  console.log(
    `Time difference after roundtrip: ${timeDiff}ms (${timeDiff * 1000000}ns)`,
  );

  // JavaScript Date only has millisecond precision, so we should have no loss
  expect(timeDiff).toBeLessThan(1); // Less than 1 millisecond
  expect(receivedDate.getTime()).toBe(testDate.getTime());
});

test("FunctionCallLargeInput", async () => {
  const function_ = await tc.functions.fromName(
    "libmodal-test-support",
    "bytelength",
  );
  const len = 3 * 1000 * 1000; // More than 2 MiB, offload to blob storage
  const input = new Uint8Array(len);
  const result = await function_.remote([input]);
  expect(result).toBe(len);
});

test("FunctionNotFound", async () => {
  const promise = tc.functions.fromName(
    "libmodal-test-support",
    "not_a_real_function",
  );
  await expect(promise).rejects.toThrowError(NotFoundError);
});

test("FunctionCallInputPlane", async () => {
  const function_ = await tc.functions.fromName(
    "libmodal-test-support",
    "input_plane",
  );
  const result = await function_.remote(["hello"]);
  expect(result).toBe("output: hello");
});

test("FunctionGetCurrentStats", async () => {
  const { mockClient: mc, mockCpClient: mock } = createMockModalClients();

  mock.handleUnary("/FunctionGetCurrentStats", (req) => {
    expect(req).toMatchObject({ functionId: "fid-stats" });
    return { backlog: 3, numTotalTasks: 7 };
  });

  const function_ = new Function_(mc, "fid-stats");
  const stats = await function_.getCurrentStats();
  expect(stats).toEqual({ backlog: 3, numTotalRunners: 7 });

  mock.assertExhausted();
});

test("FunctionUpdateAutoscaler", async () => {
  const { mockClient: mc, mockCpClient: mock } = createMockModalClients();

  mock.handleUnary("/FunctionUpdateSchedulingParams", (req) => {
    expect(req).toMatchObject({
      functionId: "fid-auto",
      settings: {
        minContainers: 1,
        maxContainers: 10,
        bufferContainers: 2,
        scaledownWindow: 300,
      },
    });
    return {};
  });

  const function_ = new Function_(mc, "fid-auto");
  await function_.updateAutoscaler({
    minContainers: 1,
    maxContainers: 10,
    bufferContainers: 2,
    scaledownWindow: 300,
  });

  mock.handleUnary("/FunctionUpdateSchedulingParams", (req) => {
    expect(req).toMatchObject({
      functionId: "fid-auto",
      settings: { minContainers: 2 },
    });
    return {};
  });

  await function_.updateAutoscaler({ minContainers: 2 });

  mock.assertExhausted();
});

test("FunctionGetWebUrl", async () => {
  const { mockClient: mc, mockCpClient: mock } = createMockModalClients();

  mock.handleUnary("FunctionGet", (req) => {
    expect(req).toMatchObject({
      appName: "libmodal-test-support",
      objectTag: "web_endpoint",
    });
    return {
      functionId: "fid-web",
      handleMetadata: { webUrl: "https://endpoint.internal" },
    };
  });

  const web_endpoint = await mc.functions.fromName(
    "libmodal-test-support",
    "web_endpoint",
  );
  expect(await web_endpoint.getWebUrl()).toBe("https://endpoint.internal");

  mock.assertExhausted();
});

test("FunctionGetWebUrlOnNonWebFunction", async () => {
  const function_ = await tc.functions.fromName(
    "libmodal-test-support",
    "echo_string",
  );
  expect(await function_.getWebUrl()).toBeUndefined();
});

test("FunctionCallPreCborVersionError", async () => {
  // test that calling a pre 1.2 function raises an error
  const function_ = await tc.functions.fromName(
    "test-support-1-1",
    "identity_with_repr",
  );

  // Represent Python kwargs.
  const promise = function_.remote([], { s: "hello" });
  await expect(promise).rejects.toThrowError(
    /please redeploy it using Modal Python SDK version >= 1.2/,
  );
});
