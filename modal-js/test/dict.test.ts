import { expect, onTestFinished, test } from "vitest";
import { Dict, KeyError } from "../src";

test("DictInvalidName", async () => {
  for (const name of ["has space", "has/slash", "a".repeat(65)]) {
    await expect(Dict.lookup(name)).rejects.toThrow();
  }
});

test("DictEphemeralBasicOperations", async () => {
  const dict = await Dict.ephemeral();

  await dict.put("key1", "value1");
  const value1 = await dict.get("key1");
  expect(value1).toBe("value1");

  expect(await dict.contains("key1")).toBe(true);
  expect(await dict.contains("nonexistent")).toBe(false);

  const defaultValue = await dict.get("nonexistent", "default");
  expect(defaultValue).toBe("default");

  // Get non-existent key without default should throw KeyError
  await expect(dict.get("nonexistent")).rejects.toThrow(KeyError);

  // Get non-existent key with explicit null default should return null
  expect(await dict.get("nonexistent", null)).toBe(null);

  // Get non-existent key with explicit undefined default should return undefined
  expect(await dict.get("nonexistent", undefined)).toBeUndefined();

  await dict.update({ key2: "value2", key3: "value3" });
  expect(await dict.get("key2")).toBe("value2");
  expect(await dict.get("key3")).toBe("value3");

  expect(await dict.len()).toBe(3);

  await dict.update({ key1: "newValue1", key4: "value4" });
  expect(await dict.get("key1")).toBe("newValue1");
  expect(await dict.get("key2")).toBe("value2");
  expect(await dict.get("key4")).toBe("value4");

  expect(await dict.len()).toBe(4);

  const popped = await dict.pop("key1");
  expect(popped).toBe("newValue1");
  expect(await dict.contains("key1")).toBe(false);

  await expect(dict.pop("nonexistent")).rejects.toThrow(KeyError);

  await dict.clear();
  expect(await dict.len()).toBe(0);

  dict.closeEphemeral();
});

test("DictDifferentDataTypes", async () => {
  const dict = await Dict.ephemeral();

  await dict.put("string", "hello");
  await dict.put("number", 42);
  await dict.put("boolean", true);
  await dict.put("object", { nested: "value" });
  await dict.put("array", [1, 2, 3]);
  await dict.put("null", null);
  await dict.put(123, "numeric key");

  // Verify with get
  expect(await dict.get("string")).toBe("hello");
  expect(await dict.get("number")).toBe(42);
  expect(await dict.get("boolean")).toBe(true);
  expect(await dict.get("object")).toEqual({ nested: "value" });
  expect(await dict.get("array")).toEqual([1, 2, 3]);
  expect(await dict.get("null")).toBe(null);
  expect(await dict.get(123)).toBe("numeric key");

  // Verify with pop
  expect(await dict.pop("string")).toBe("hello");
  expect(await dict.pop("number")).toBe(42);
  expect(await dict.pop("boolean")).toBe(true);
  expect(await dict.pop("object")).toEqual({ nested: "value" });
  expect(await dict.pop("array")).toEqual([1, 2, 3]);
  expect(await dict.pop("null")).toBe(null);
  expect(await dict.pop(123)).toBe("numeric key");

  expect(await dict.len()).toBe(0);

  dict.closeEphemeral();
});

test("DictSkipIfExists", async () => {
  const dict = await Dict.ephemeral();

  const created1 = await dict.put("key", "value1");
  expect(created1).toBe(true);
  expect(await dict.get("key")).toBe("value1");

  // Second put without skipIfExists should overwrite
  const created2 = await dict.put("key", "value2");
  expect(created2).toBe(true);
  expect(await dict.get("key")).toBe("value2");

  // Put with skipIfExists should not overwrite
  const created3 = await dict.put("key", "value3", { skipIfExists: true });
  expect(created3).toBe(false);
  expect(await dict.get("key")).toBe("value2");

  // New key with skipIfExists should succeed
  const created4 = await dict.put("newkey", "newvalue", {
    skipIfExists: true,
  });
  expect(created4).toBe(true);
  expect(await dict.get("newkey")).toBe("newvalue");

  dict.closeEphemeral();
});

test("DictIteration", async () => {
  const dict = await Dict.ephemeral();

  const testData = {
    key1: "value1",
    key2: "value2",
    key3: "value3",
  };
  await dict.update(testData);

  const keys = [];
  for await (const key of dict.keys()) {
    keys.push(key);
  }
  expect(keys.sort()).toEqual(["key1", "key2", "key3"]);

  const values = [];
  for await (const value of dict.values()) {
    values.push(value);
  }
  expect(values.sort()).toEqual(["value1", "value2", "value3"]);

  const items: [string, string][] = [];
  for await (const [key, value] of dict.items()) {
    items.push([key, value]);
  }
  expect(items.length).toBe(3);
  items.forEach(([key, value]) => {
    expect(testData[key as keyof typeof testData]).toBe(value);
  });

  dict.closeEphemeral();
});

test("DictNonEphemeral", async () => {
  const dictName = `test-dict-${Date.now()}`;

  const dict1 = await Dict.lookup(dictName, { createIfMissing: true });
  onTestFinished(async () => {
    await Dict.delete(dictName);
    await expect(Dict.lookup(dictName)).rejects.toThrow(); // confirm deletion
  });
  await dict1.put("persistent", "data");

  const dict2 = await Dict.lookup(dictName);
  expect(await dict2.get("persistent")).toBe("data");
});

test("DictEmptyOperations", async () => {
  const dict = await Dict.ephemeral();

  expect(await dict.len()).toBe(0);
  expect(await dict.get("key", "default")).toBe("default");
  expect(await dict.contains("key")).toBe(false);
  await expect(dict.clear()).resolves.not.toThrow();
  await expect(dict.update({})).resolves.not.toThrow();
  expect(await dict.len()).toBe(0);

  const keys = [];
  for await (const key of dict.keys()) {
    keys.push(key);
  }
  expect(keys).toEqual([]);

  dict.closeEphemeral();
});
