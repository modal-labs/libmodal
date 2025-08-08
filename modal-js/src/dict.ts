import {
  ObjectCreationType,
  DictEntry,
  DictContentsRequest,
} from "../proto/modal_proto/api";
import type { DeleteOptions, EphemeralOptions, LookupOptions } from "./app";
import { client } from "./client";
import { environmentName } from "./config";
import { InvalidError, KeyError, RequestSizeError } from "./errors";
import { dumps, loads } from "./pickle";
import { ClientError } from "nice-grpc";

// From: modal/_object.py
const ephemeralObjectHeartbeatSleep = 300_000; // 300 seconds

/** Options to configure a `Dict.put()` operation. */
export type DictPutOptions = {
  /** Skip adding the key if it already exists. */
  skipIfExists?: boolean;
};

// Dict is a distributed dictionary for key-value storage in Modal apps.
export class Dict {
  readonly dictId: string;
  readonly #ephemeral: boolean;
  readonly #abortController?: AbortController;

  /** @ignore */
  constructor(dictId: string, ephemeral: boolean = false) {
    this.dictId = dictId;
    this.#ephemeral = ephemeral;
    this.#abortController = ephemeral ? new AbortController() : undefined;
  }

  /**
   * Create a nameless, temporary Dict.
   * You will need to call `closeEphemeral()` to delete the Dict.
   */
  static async ephemeral(options: EphemeralOptions = {}): Promise<Dict> {
    const resp = await client.dictGetOrCreate({
      objectCreationType: ObjectCreationType.OBJECT_CREATION_TYPE_EPHEMERAL,
      environmentName: environmentName(options.environment),
    });

    const dict = new Dict(resp.dictId, true);
    const signal = dict.#abortController!.signal;
    (async () => {
      // Launch a background task to heartbeat the ephemeral Dict.
      while (true) {
        await client.dictHeartbeat({ dictId: resp.dictId });
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

    return dict;
  }

  /** Delete the ephemeral Dict. Only usable with `Dict.ephemeral()`. */
  closeEphemeral(): void {
    if (this.#ephemeral) {
      this.#abortController!.abort();
    } else {
      throw new InvalidError("Dict is not ephemeral.");
    }
  }

  /**
   * Lookup a Dict by name.
   */
  static async lookup(
    name: string,
    options: LookupOptions = {},
  ): Promise<Dict> {
    const resp = await client.dictGetOrCreate({
      deploymentName: name,
      objectCreationType: options.createIfMissing
        ? ObjectCreationType.OBJECT_CREATION_TYPE_CREATE_IF_MISSING
        : undefined,
      environmentName: environmentName(options.environment),
    });
    return new Dict(resp.dictId);
  }

  /** Delete a Dict by name. */
  static async delete(
    name: string,
    options: DeleteOptions = {},
  ): Promise<void> {
    const dict = await Dict.lookup(name, options);
    await client.dictDelete({ dictId: dict.dictId });
  }

  /**
   * Remove all items from the Dict.
   */
  async clear(): Promise<void> {
    await client.dictClear({ dictId: this.dictId });
  }

  /**
   * Get the value associated with a key.
   *
   * Returns `defaultValue` if key does not exist.
   * Throws `KeyError` if key does not exist and no default value is provided.
   */
  async get(key: any, defaultValue?: any): Promise<any> {
    const hasDefault = arguments.length > 1;
    const resp = await client.dictGet({
      dictId: this.dictId,
      key: dumps(key),
    });
    if (!resp.found) {
      if (!hasDefault) {
        throw new KeyError(`Key not found in Dict ${this.dictId}`);
      }
      return defaultValue;
    }
    return loads(resp.value!);
  }

  /**
   * Return if a key is present.
   */
  async contains(key: any): Promise<boolean> {
    const resp = await client.dictContains({
      dictId: this.dictId,
      key: dumps(key),
    });
    return resp.found;
  }

  /**
   * Return the length of the Dict.
   *
   * Note: This is an expensive operation and will return at most 100,000.
   */
  async len(): Promise<number> {
    const resp = await client.dictLen({ dictId: this.dictId });
    return resp.len;
  }

  /**
   * Update the Dict with additional items.
   */
  async update(items: Record<any, any>): Promise<void> {
    const updates: DictEntry[] = Object.entries(items).map(([k, v]) => ({
      key: dumps(k),
      value: dumps(v),
    }));

    try {
      await client.dictUpdate({
        dictId: this.dictId,
        updates,
      });
    } catch (e) {
      if (e instanceof ClientError && e.details?.includes("status = '413'")) {
        throw new RequestSizeError("Dict.update request is too large");
      }
      throw e;
    }
  }

  /**
   * Add a specific key-value pair to the Dict.
   *
   * Returns true if the key-value pair was added and false if it wasn't
   * because the key already existed and `skipIfExists` was set.
   */
  async put(
    key: any,
    value: any,
    options: DictPutOptions = {},
  ): Promise<boolean> {
    const updates: DictEntry[] = [
      {
        key: dumps(key),
        value: dumps(value),
      },
    ];

    try {
      const resp = await client.dictUpdate({
        dictId: this.dictId,
        updates,
        ifNotExists: options.skipIfExists,
      });
      return resp.created;
    } catch (e) {
      if (e instanceof ClientError && e.details?.includes("status = '413'")) {
        throw new RequestSizeError("Dict.put request is too large");
      }
      throw e;
    }
  }

  /**
   * Remove a key from the Dict, returning the value if it exists.
   */
  async pop(key: any): Promise<any> {
    const resp = await client.dictPop({
      dictId: this.dictId,
      key: dumps(key),
    });
    if (!resp.found) {
      throw new KeyError(`Key not found in Dict ${this.dictId}`);
    }
    return loads(resp.value!);
  }

  /**
   * Return an async iterator over the keys in the Dict.
   */
  async *keys(): AsyncGenerator<any, void, unknown> {
    const request: DictContentsRequest = {
      dictId: this.dictId,
      keys: true,
      values: false,
    };

    for await (const entry of client.dictContents(request)) {
      yield loads(entry.key);
    }
  }

  /**
   * Return an async iterator over the values in the Dict.
   */
  async *values(): AsyncGenerator<any, void, unknown> {
    const request: DictContentsRequest = {
      dictId: this.dictId,
      keys: false,
      values: true,
    };

    for await (const entry of client.dictContents(request)) {
      yield loads(entry.value);
    }
  }

  /**
   * Return an async iterator over the [key, value] pairs in this Dict.
   */
  async *items(): AsyncGenerator<[any, any], void, unknown> {
    const request: DictContentsRequest = {
      dictId: this.dictId,
      keys: true,
      values: true,
    };

    for await (const entry of client.dictContents(request)) {
      yield [loads(entry.key), loads(entry.value)];
    }
  }
}
