import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    maxConcurrency: 10,
    slowTestThreshold: 5_000,
    testTimeout: 20_000,
  },
});
