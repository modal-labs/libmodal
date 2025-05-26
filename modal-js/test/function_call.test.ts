import { Function_, FunctionTimeoutError } from "modal";
import { expect, test } from "vitest";

test("FunctionSpawn", async () => {
  const function_ = await Function_.lookup(
    "libmodal-test-support",
    "echo_string",
  );

  // Spawn function with kwargs.
  var functionCall = await function_.spawn([], { s: "hello" });
  expect(functionCall.functionCallId).toBeDefined();

  // Get results after spawn.
  var resultKwargs = await functionCall.get();
  expect(resultKwargs).toBe("output: hello");

  // Try the same again; same results should still be available.
  resultKwargs = await functionCall.get();
  expect(resultKwargs).toBe("output: hello");

  // Lookup function that takes a long time to complete.
  const functionSleep_ = await Function_.lookup(
    "libmodal-test-support",
    "sleep",
  );

  // Spawn with long running input.
  functionCall = await functionSleep_.spawn([], { t: 5 });
  expect(functionCall.functionCallId).toBeDefined();

  // Getting outputs with timeout raises error.
  const promise = functionCall.get({ timeout: 1000 }); // 1000ms
  await expect(promise).rejects.toThrowError(FunctionTimeoutError);
});
