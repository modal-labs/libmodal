import { Function_, TimeoutError } from "modal";
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

  // Try the same again; results should still be available.
  resultKwargs = await functionCall.get();
  expect(resultKwargs).toBe("output: hello");

  // Looking function that takes a long time to complete.
  const functionSleep_ = await Function_.lookup(
    "libmodal-test-support",
    "sleep",
  );

  // Spawn function with long running input.
  functionCall = await functionSleep_.spawn([], { t: 5 });
  expect(functionCall.functionCallId).toBeDefined();

  // Get is now expected to timeout.
  const promise = functionCall.get({ timeout: 1 / 100 });
  await expect(promise).rejects.toThrowError(TimeoutError);
});
