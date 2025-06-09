// Queue object, to be used with Modal Queues.

import {
  DeploymentNamespace,
  ObjectCreationType,
  QueueNextItemsRequest,
} from "../proto/modal_proto/api";
import { client } from "./client";
import { environmentName } from "./config";
import { InvalidError, QueueEmptyError, QueueFullError } from "./errors";
import { dumps, loads } from "./pickle";
import { ClientError, Status } from "nice-grpc";
import { DeleteOptions, EphemeralOptions, LookupOptions } from "./app";

// From: modal/_object.py
const ephemeralObjectHeartbeatSleep = 300_000; // 300 seconds

const queueInitialPutBackoff = 100; // 100 milliseconds
const queueDefaultPartitionTtl = 24 * 3600 * 1000; // 24 hours

/** Options to configure a `Queue.clear()` operation. */
export type QueueClearOptions = {
  /** Partition to clear, uses default partition if not set. */
  partition?: string;

  /** Set to clear all queue partitions. */
  all?: boolean;
};

/** Options to configure a `Queue.get()` or `Queue.getMany()` operation. */
export type QueueGetOptions = {
  /** How long to wait if the queue is empty (default: indefinite). */
  timeout?: number;

  /** Partition to fetch values from, uses default partition if not set. */
  partition?: string;
};

/** Options to configure a `Queue.put()` or `Queue.putMany()` operation. */
export type QueuePutOptions = {
  /** How long to wait if the queue is full (default: indefinite). */
  timeout?: number;

  /** Partition to add items to, uses default partition if not set. */
  partition?: string;

  /** TTL for the partition in seconds (default: 1 day). */
  partitionTtl?: number;
};

/** Options to configure a `Queue.len()` operation. */
export type QueueLenOptions = {
  /** Partition to compute length, uses default partition if not set. */
  partition?: string;

  /** Return the total length across all partitions. */
  total?: boolean;
};

/** Options to configure a `Queue.iterate()` operation. */
export type QueueIterateOptions = {
  /** How long to wait between successive items before exiting iteration (default: 0). */
  itemPollTimeout?: number;

  /** Partition to iterate, uses default partition if not set. */
  partition?: string;
};

/**
 * Distributed, FIFO queue for data flow in Modal apps.
 */
export class Queue {
  readonly queueId: string;
  readonly #ephemeral: boolean;
  readonly #abortController?: AbortController;

  /** @ignore */
  constructor(queueId: string, ephemeral: boolean = false) {
    this.queueId = queueId;
    this.#ephemeral = ephemeral;
    this.#abortController = ephemeral ? new AbortController() : undefined;
  }

  static #validatePartitionKey(partition: string | undefined): Uint8Array {
    if (partition) {
      const partitionKey = new TextEncoder().encode(partition);
      if (partitionKey.length === 0 || partitionKey.length > 64) {
        throw new InvalidError(
          "Queue partition key must be between 1 and 64 bytes.",
        );
      }
      return partitionKey;
    }
    return new Uint8Array();
  }

  /**
   * Create a nameless, temporary queue.
   * You will need to call `closeEphemeral()` to delete the queue.
   */
  static async ephemeral(options: EphemeralOptions = {}): Promise<Queue> {
    const resp = await client.queueGetOrCreate({
      objectCreationType: ObjectCreationType.OBJECT_CREATION_TYPE_EPHEMERAL,
      environmentName: environmentName(options.environment),
    });

    const queue = new Queue(resp.queueId, true);
    const signal = queue.#abortController!.signal;
    (async () => {
      // Launch a background task to heartbeat the ephemeral queue.
      while (true) {
        await client.queueHeartbeat({ queueId: resp.queueId });
        await Promise.race([
          new Promise((resolve) =>
            setTimeout(resolve, ephemeralObjectHeartbeatSleep),
          ),
          new Promise((resolve) => {
            signal.addEventListener("abort", resolve, { once: true });
          }),
        ]);
      }
    })();

    return queue;
  }

  /** Delete the ephemeral queue. Only usable with `Queue.ephemeral()`. */
  closeEphemeral(): void {
    if (this.#ephemeral) {
      this.#abortController!.abort();
    } else {
      throw new InvalidError("Queue is not ephemeral.");
    }
  }

  /**
   * Lookup a queue by name.
   */
  static async lookup(
    name: string,
    options: LookupOptions = {},
  ): Promise<Queue> {
    const resp = await client.queueGetOrCreate({
      deploymentName: name,
      objectCreationType: options.createIfMissing
        ? ObjectCreationType.OBJECT_CREATION_TYPE_CREATE_IF_MISSING
        : undefined,
      namespace: DeploymentNamespace.DEPLOYMENT_NAMESPACE_WORKSPACE,
      environmentName: environmentName(options.environment),
    });
    return new Queue(resp.queueId);
  }

  /** Delete a queue by name. */
  static async delete(
    name: string,
    options: DeleteOptions = {},
  ): Promise<void> {
    const queue = await Queue.lookup(name, options);
    await client.queueDelete({ queueId: queue.queueId });
  }

  /**
   * Remove all objects from a queue partition.
   */
  async clear(options: QueueClearOptions = {}): Promise<void> {
    if (options.partition && options.all) {
      throw new InvalidError(
        "Partition must be null when requesting to clear all.",
      );
    }
    await client.queueClear({
      queueId: this.queueId,
      partitionKey: Queue.#validatePartitionKey(options.partition),
      allPartitions: options.all,
    });
  }

  async #get(n: number, partition?: string, timeout?: number): Promise<any[]> {
    const partitionKey = Queue.#validatePartitionKey(partition);

    const startTime = Date.now();
    let pollTimeout = 50_000;
    if (timeout !== undefined) {
      pollTimeout = Math.min(pollTimeout, timeout);
    }

    while (true) {
      const response = await client.queueGet({
        queueId: this.queueId,
        partitionKey,
        timeout: pollTimeout / 1000,
        nValues: n,
      });
      if (response.values && response.values.length > 0) {
        return response.values.map((value) => loads(value));
      }
      if (timeout !== undefined) {
        const remaining = timeout - (Date.now() - startTime);
        if (remaining <= 0) {
          const message = `Queue ${this.queueId} did not return values within ${timeout}ms.`;
          throw new QueueEmptyError(message);
        }
        pollTimeout = Math.min(pollTimeout, remaining);
      }
    }
  }

  /**
   * Remove and return the next object from the queue.
   *
   * By default, this will wait until at least one item is present in the queue.
   * If `timeout` is set, raises `QueueEmptyError` if no items are available
   * within that timeout in milliseconds.
   */
  async get(options: QueueGetOptions = {}): Promise<any | null> {
    const values = await this.#get(1, options.partition, options.timeout);
    return values[0]; // Must have length >= 1 if returned.
  }

  /**
   * Remove and return up to `n` objects from the queue.
   *
   * By default, this will wait until at least one item is present in the queue.
   * If `timeout` is set, raises `QueueEmptyError` if no items are available
   * within that timeout in milliseconds.
   */
  async getMany(n: number, options: QueueGetOptions = {}): Promise<any[]> {
    return await this.#get(n, options.partition, options.timeout);
  }

  async #put(
    values: any[],
    timeout?: number,
    partition?: string,
    partitionTtl?: number,
  ): Promise<void> {
    const valuesEncoded = values.map((v) => dumps(v));
    const partitionKey = Queue.#validatePartitionKey(partition);

    let delay = queueInitialPutBackoff;
    const deadline = timeout ? Date.now() + timeout : undefined;
    while (true) {
      try {
        await client.queuePut({
          queueId: this.queueId,
          values: valuesEncoded,
          partitionKey,
          partitionTtlSeconds:
            (partitionTtl || queueDefaultPartitionTtl) / 1000,
        });
        break;
      } catch (e) {
        if (e instanceof ClientError && e.code === Status.RESOURCE_EXHAUSTED) {
          // Queue is full, retry with exponential backoff up to the deadline.
          delay = Math.min(delay * 2, 30_000);
          if (deadline !== undefined) {
            const remaining = deadline - Date.now();
            if (remaining <= 0)
              throw new QueueFullError(`Put failed on ${this.queueId}.`);
            delay = Math.min(delay, remaining);
          }
          await new Promise((resolve) => setTimeout(resolve, delay));
        } else {
          throw e;
        }
      }
    }
  }

  /**
   * Add an item to the end of the queue.
   *
   * If the queue is full, this will retry with exponential backoff until the
   * provided `timeout` is reached, or indefinitely if `timeout` is not set.
   * Raises `QueueFullError` if the queue is still full after the timeout.
   */
  async put(v: any, options: QueuePutOptions = {}): Promise<void> {
    await this.#put(
      [v],
      options.timeout,
      options.partition,
      options.partitionTtl,
    );
  }

  /**
   * Add several items to the end of the queue.
   *
   * If the queue is full, this will retry with exponential backoff until the
   * provided `timeout` is reached, or indefinitely if `timeout` is not set.
   * Raises `QueueFullError` if the queue is still full after the timeout.
   */
  async putMany(values: any[], options: QueuePutOptions = {}): Promise<void> {
    await this.#put(
      values,
      options.timeout,
      options.partition,
      options.partitionTtl,
    );
  }

  /** Return the number of objects in the queue. */
  async len(options: QueueLenOptions = {}): Promise<number> {
    if (options.partition && options.total) {
      throw new InvalidError(
        "Partition must be null when requesting total length.",
      );
    }
    const resp = await client.queueLen({
      queueId: this.queueId,
      partitionKey: Queue.#validatePartitionKey(options.partition),
      total: options.total,
    });
    return resp.len;
  }

  /** Iterate through items in a queue without mutation. */
  async *iterate(
    options: QueueIterateOptions = {},
  ): AsyncGenerator<any, void, unknown> {
    const { partition, itemPollTimeout = 0 } = options;

    let lastEntryId = undefined;
    const validatedPartitionKey = Queue.#validatePartitionKey(partition);
    let fetchDeadline = Date.now() + itemPollTimeout;

    const maxPollDuration = 30_000;
    while (true) {
      const pollDuration = Math.max(
        0.0,
        Math.min(maxPollDuration, fetchDeadline - Date.now()),
      );
      const request: QueueNextItemsRequest = {
        queueId: this.queueId,
        partitionKey: validatedPartitionKey,
        itemPollTimeout: pollDuration / 1000,
        lastEntryId: lastEntryId || "",
      };

      const response = await client.queueNextItems(request);
      if (response.items && response.items.length > 0) {
        for (const item of response.items) {
          yield loads(item.value);
          lastEntryId = item.entryId;
        }
        fetchDeadline = Date.now() + itemPollTimeout;
      } else if (Date.now() > fetchDeadline) {
        break;
      }
    }
  }
}
