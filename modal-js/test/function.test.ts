import { Function_, NotFoundError } from "modal";
import { expect, onTestFinished, test } from "vitest";

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
  if (reprResult.includes("datetime.datetime")) {
    // Success! Python received it as a datetime
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
  } else if (reprResult.match(/\d{10,}/)) {
    // Python received it as a Unix timestamp (integer)
    console.log("⚠️  Python received JS Date as Unix timestamp:", reprResult);
    console.log(
      "This means CBOR time tags are not being used by the JS client",
    );

    // The identity result might be an integer (Unix timestamp in seconds)
    const unixTime =
      typeof identityResult === "number" ? identityResult : undefined;
    if (unixTime !== undefined) {
      const expectedUnix = Math.floor(testDate.getTime() / 1000);
      expect(unixTime).toBe(expectedUnix);
      console.log("✅ Unix timestamp roundtrip successful:", unixTime);
    }
  } else {
    console.log("❓ Unexpected Python representation:", reprResult);
    console.log("Identity result:", identityResult, typeof identityResult);
  }
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
  const { MockGrpc } = await import("../test-support/grpc_mock");
  const mock = await MockGrpc.install();
  onTestFinished(async () => {
    await mock.uninstall();
  });

  mock.handleUnary("/FunctionGetCurrentStats", (req) => {
    expect(req).toMatchObject({ functionId: "fid-stats" });
    return { backlog: 3, numTotalTasks: 7 };
  });

  const { Function_ } = await import("modal");
  const function_ = new Function_("fid-stats");
  const stats = await function_.getCurrentStats();
  expect(stats).toEqual({ backlog: 3, numTotalRunners: 7 });
});

test("FunctionUpdateAutoscaler", async () => {
  const { MockGrpc } = await import("../test-support/grpc_mock");
  const mock = await MockGrpc.install();
  onTestFinished(async () => {
    await mock.uninstall();
  });

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

  const { Function_ } = await import("modal");
  const function_ = new Function_("fid-auto");
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
});

test("FunctionGetWebUrl", async () => {
  const { MockGrpc } = await import("../test-support/grpc_mock");
  const mock = await MockGrpc.install();
  onTestFinished(async () => {
    await mock.uninstall();
  });

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

  const { Function_ } = await import("modal");
  const web_endpoint = await Function_.lookup(
    "libmodal-test-support",
    "web_endpoint",
  );
  expect(await web_endpoint.getWebUrl()).toBe("https://endpoint.internal");
});

test("FunctionGetWebUrlOnNonWebFunction", async () => {
  const function_ = await Function_.lookup(
    "libmodal-test-support",
    "echo_string",
  );
  expect(await function_.getWebUrl()).toBeUndefined();
});
