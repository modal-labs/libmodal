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
import { DeleteOptions, LookupOptions } from "./app";

// From: modal/_object.py
const empheralQueueHeartbeatSleep = 300;

export type QueueClearOptions = {
  partition?: string;
  all?: boolean;
};

export type QueueGetOptions = {
  block?: boolean;
  timeout?: number;
  partition?: string;
};

export type QueuePutOptions = {
  block?: boolean;
  timeout?: number;
  partition?: string;
  partitionTtl?: number;
};

export type QueueLenOptions = {
  partition?: string;
  total?: boolean;
};

export type QueueIterateOptions = {
  partition?: string;
  itemPollTimeout?: number; // in milliseconds
};

/**
 * Distributed, FIFO queue for data flow in Modal apps.
 */
export class Queue {
  readonly queueId: string;
  readonly ephemeral: boolean;
  readonly abortController?: AbortController;

  constructor(queueId: string, ephemeral: boolean = false) {
    this.queueId = queueId;
    this.ephemeral = ephemeral;
    this.abortController = ephemeral ? new AbortController() : undefined;
  }

  static #validatePartitionKey(partition: string | undefined): Uint8Array {
    if (partition) {
      const partitionKey = new TextEncoder().encode(partition);
      if (partitionKey.length === 0 || partitionKey.length > 64) {
        throw new InvalidError(
          "Queue partition key must be between 1 and 64 characters.",
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
  static async ephemeral({
    environment,
  }: { environment?: string } = {}): Promise<Queue> {
    const resp = await client.queueGetOrCreate({
      objectCreationType: ObjectCreationType.OBJECT_CREATION_TYPE_EPHEMERAL,
      environmentName: environmentName(environment),
    });

    const queue = new Queue(resp.queueId, true);
    (async () => {
      while (true) {
        await client.queueHeartbeat({ queueId: resp.queueId });
        await Promise.race([
          new Promise((resolve) =>
            setTimeout(resolve, empheralQueueHeartbeatSleep),
          ),
          new Promise((_, reject) => {
            queue.abortController!.signal.addEventListener(
              "abort",
              () => {
                reject(new Error("Aborted"));
              },
              { once: true },
            );
          }).catch(() => {}),
        ]);
      }
    })();

    return queue;
  }

  /**
   * Delete the ephemeral queue.
   */
  async closeEphemeral(): Promise<void> {
    if (this.ephemeral) {
      this.abortController!.abort();
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

  /**
   * Delete a queue by name.
   */
  static async delete(
    name: string,
    options: DeleteOptions = {},
  ): Promise<void> {
    const queue = await Queue.lookup(name, options);
    await client.queueDelete({ queueId: queue.queueId });
  }

  async #getNonblocking({
    partition,
    n_values,
  }: {
    partition?: string;
    n_values: number;
  }): Promise<any[]> {
    const request = {
      queueId: this.queueId,
      partitionKey: Queue.#validatePartitionKey(partition),
      timeout: 0,
      nValues: n_values,
    };

    const response = await client.queueGet(request);
    if (response.values) {
      return response.values.map((value) => loads(value));
    } else {
      return [];
    }
  }

  async #getBlocking({
    partition,
    timeout,
    n_values,
  }: {
    partition?: string;
    timeout?: number; // in milliseconds
    n_values: number;
  }): Promise<any[]> {
    let deadline: number | undefined = undefined;
    if (timeout !== undefined) {
      deadline = Date.now() + timeout;
    }

    while (true) {
      let requestTimeout = 50.0;
      if (deadline) {
        requestTimeout = Math.min(requestTimeout, deadline - Date.now());
      }

      const request = {
        queueId: this.queueId,
        partitionKey: Queue.#validatePartitionKey(partition),
        timeout: requestTimeout,
        nValues: n_values,
      };

      const response = await client.queueGet(request);

      if (response.values && response.values.length > 0) {
        return response.values.map((value) => loads(value));
      }

      if (deadline && Date.now() > deadline) {
        break;
      }
    }

    throw new QueueEmptyError("Queue is empty");
  }

  /**
   * Remove all objects from a queue partition.
   */
  async clear(
    { partition, all }: QueueClearOptions = { all: false },
  ): Promise<void> {
    if (partition && all) {
      throw new InvalidError(
        "Partition must be null when requesting to clear all.",
      );
    }

    const request = {
      queueId: this.queueId,
      partitionKey: Queue.#validatePartitionKey(partition),
      allPartitions: all,
    };

    await client.queueClear(request);
  }

  /**
   * Remove and return an object from the queue.
   */
  async get({
    block = true,
    timeout,
    partition,
  }: QueueGetOptions = {}): Promise<any> {
    const values = await this.getMany(1, { block, timeout, partition });
    if (values.length !== 0) {
      return values[0];
    } else {
      return null;
    }
  }

  /**
   * Remove and return up to `n_values` objects from the queue.
   */
  async getMany(
    n_values: number,
    { block = true, timeout, partition }: QueueGetOptions = {},
  ): Promise<any[]> {
    if (block) {
      return await this.#getBlocking({
        partition,
        timeout,
        n_values,
      });
    } else {
      if (timeout) {
        throw new InvalidError("Cannot pass timeout and block = false.");
      }
      return await this.#getNonblocking({
        partition,
        n_values,
      });
    }
  }

  /**
   * Add an object to the end of the queue.
   */
  async put(
    v: any,
    {
      block = true,
      timeout,
      partition,
      partitionTtl = 24 * 3600,
    }: QueuePutOptions = {},
  ): Promise<void> {
    await this.putMany([v], { block, timeout, partition, partitionTtl });
  }

  /**
   * Add several objects to the end of the queue.
   */
  async putMany(
    vs: any[],
    {
      block = true,
      timeout,
      partition,
      partitionTtl = 24 * 3600,
    }: QueuePutOptions = {},
  ): Promise<void> {
    if (block) {
      await this.#putManyBlocking({
        partition,
        partitionTtl,
        vs,
        timeout,
      });
    } else {
      if (timeout) {
        console.warn("Timeout is ignored for non-blocking put.");
      }
      await this.#putManyNonblocking({
        partition,
        partitionTtl,
        vs,
      });
    }
  }

  async #putManyBlocking({
    partition,
    partitionTtl,
    vs,
    timeout,
  }: {
    partition?: string;
    partitionTtl: number;
    vs: any[];
    timeout?: number;
  }): Promise<void> {
    const vs_encoded = vs.map((v) => dumps(v));
    const request = {
      queueId: this.queueId,
      partitionKey: Queue.#validatePartitionKey(partition),
      values: vs_encoded,
      partitionTtlSeconds: partitionTtl,
    };

    try {
      const retryOptions = {
        additionalStatusCodes: [Status.RESOURCE_EXHAUSTED],
        maxDelay: 30.0,
        maxRetries: undefined,
        totalTimeout: timeout,
      };
      await client.queuePut(request, retryOptions);
    } catch (e) {
      if (e instanceof ClientError && e.code === Status.RESOURCE_EXHAUSTED) {
        throw new QueueFullError("Queue.put() tried to put to a full queue.");
      } else {
        throw e;
      }
    }
  }

  async #putManyNonblocking({
    partition,
    partitionTtl,
    vs,
  }: {
    partition?: string;
    partitionTtl: number;
    vs: any[];
  }): Promise<void> {
    const vs_encoded = vs.map((v) => dumps(v));
    const request = {
      queueId: this.queueId,
      partitionKey: Queue.#validatePartitionKey(partition),
      values: vs_encoded,
      partitionTtlSeconds: partitionTtl,
    };

    try {
      await client.queuePut(request);
    } catch (e) {
      if (e instanceof ClientError && e.code === Status.RESOURCE_EXHAUSTED) {
        throw new QueueFullError("Queue is full");
      } else {
        throw e;
      }
    }
  }

  /**
   * Return the number of objects in the queue partition.
   */
  async len({
    partition,
    total = false,
  }: QueueLenOptions = {}): Promise<number> {
    if (partition && total) {
      throw new InvalidError(
        "Partition must be null when requesting total length.",
      );
    }

    const request = {
      queueId: this.queueId,
      partitionKey: Queue.#validatePartitionKey(partition),
      total,
    };

    const response = await client.queueLen(request);
    return response.len;
  }

  /**
   * Iterate through items in a queue without mutation.
   */
  async *iterate({
    partition,
    itemPollTimeout = 0.0,
  }: QueueIterateOptions = {}): AsyncGenerator<any, void, unknown> {
    let lastEntryId = undefined;
    const validatedPartitionKey = Queue.#validatePartitionKey(partition);
    let fetchDeadline = Date.now() + itemPollTimeout;

    const MAX_POLL_DURATION = 30_000;
    while (true) {
      const pollDuration = Math.max(
        0.0,
        Math.min(MAX_POLL_DURATION, fetchDeadline - Date.now()),
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
