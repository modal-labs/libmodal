import { Queue, QueueEmptyError } from "modal";
import { expect, test } from "vitest";

test("QueueInvalidName", async () => {
  for (const name of ["has space", "has/slash", "a".repeat(65)]) {
    await expect(Queue.lookup(name)).rejects.toThrow();
  }
});

test("QueueEphemeral", async () => {
  const queue = await Queue.ephemeral();
  await queue.put(123);
  expect(await queue.len()).toBe(1);
  expect(await queue.get()).toBe(123);
  queue.closeEphemeral();
});

test("QueueSuite1", async () => {
  const queue = await Queue.ephemeral();
  expect(await queue.len()).toBe(0);

  await queue.put(123);
  expect(await queue.len()).toBe(1);
  expect(await queue.get()).toBe(123);

  await queue.put(432);
  expect(await queue.get({ timeout: 0 })).toBe(432);

  await expect(queue.get({ timeout: 0 })).rejects.toThrow(QueueEmptyError);
  expect(await queue.len()).toBe(0);

  await queue.putMany([1, 2, 3]);
  const results: number[] = [];
  for await (const item of queue.iterate()) {
    results.push(item);
  }
  expect(results).toEqual([1, 2, 3]);
  queue.closeEphemeral();
});

test("QueueSuite2", async () => {
  const results: number[] = [];
  const producer = async (queue: Queue) => {
    for (let i = 0; i < 10; i++) {
      await queue.put(i);
    }
  };

  const consumer = async (queue: Queue) => {
    for await (const item of queue.iterate({ itemPollTimeout: 1000 })) {
      results.push(item);
    }
  };

  const queue = await Queue.ephemeral();
  await Promise.all([producer(queue), consumer(queue)]);
  expect(results).toEqual([0, 1, 2, 3, 4, 5, 6, 7, 8, 9]);
  queue.closeEphemeral();
});

test("QueuePutAndGetMany", async () => {
  const queue = await Queue.ephemeral();
  await queue.putMany([1, 2, 3]);
  expect(await queue.len()).toBe(3);
  expect(await queue.getMany(3)).toEqual([1, 2, 3]);
  queue.closeEphemeral();
});

test("QueueNonBlocking", async () => {
  // Assuming the queue is available, these operations
  // Should succeed immediately.
  const queue = await Queue.ephemeral();
  await queue.put(123, { timeout: 0 });
  expect(await queue.len()).toBe(1);
  expect(await queue.get({ timeout: 0 })).toBe(123);
  queue.closeEphemeral();
});
