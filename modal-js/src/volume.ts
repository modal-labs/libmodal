import { ObjectCreationType } from "../proto/modal_proto/api";
import { getDefaultClient, type ModalClient } from "./client";
import { ClientError, Status } from "nice-grpc";
import { NotFoundError, InvalidError } from "./errors";
import { EphemeralHeartbeatManager } from "./ephemeral";
import type { EphemeralOptions } from "./app";

/** Options for `Volume.fromName()`. */
export type VolumeFromNameOptions = {
  environment?: string;
  createIfMissing?: boolean;
};

/**
 * Service for managing Volumes.
 */
export class VolumeService {
  readonly #client: ModalClient;
  constructor(client: ModalClient) {
    this.#client = client;
  }

  /**
   * Reference a Volume by its name.
   */
  async fromName(
    name: string,
    options?: VolumeFromNameOptions,
  ): Promise<Volume> {
    try {
      const resp = await this.#client.cpClient.volumeGetOrCreate({
        deploymentName: name,
        environmentName: this.#client.environmentName(options?.environment),
        objectCreationType: options?.createIfMissing
          ? ObjectCreationType.OBJECT_CREATION_TYPE_CREATE_IF_MISSING
          : ObjectCreationType.OBJECT_CREATION_TYPE_UNSPECIFIED,
      });
      return new Volume(resp.volumeId, name);
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.NOT_FOUND)
        throw new NotFoundError(err.details);
      throw err;
    }
  }

  /**
   * Create a nameless, temporary Volume.
   * It persists until closeEphemeral() is called, or the process exits.
   */
  async ephemeral(options: EphemeralOptions = {}): Promise<Volume> {
    const resp = await this.#client.cpClient.volumeGetOrCreate({
      objectCreationType: ObjectCreationType.OBJECT_CREATION_TYPE_EPHEMERAL,
      environmentName: this.#client.environmentName(options.environment),
    });

    const ephemeralHbManager = new EphemeralHeartbeatManager(() =>
      this.#client.cpClient.volumeHeartbeat({ volumeId: resp.volumeId }),
    );

    return new Volume(resp.volumeId, undefined, false, ephemeralHbManager);
  }
}

/** Volumes provide persistent storage that can be mounted in Modal Functions. */
export class Volume {
  readonly volumeId: string;
  readonly name?: string;
  private _readOnly: boolean = false;
  readonly #ephemeralHbManager?: EphemeralHeartbeatManager;

  /** @ignore */
  constructor(
    volumeId: string,
    name?: string,
    readOnly: boolean = false,
    ephemeralHbManager?: EphemeralHeartbeatManager,
  ) {
    this.volumeId = volumeId;
    this.name = name;
    this._readOnly = readOnly;
    this.#ephemeralHbManager = ephemeralHbManager;
  }

  /**
   * @deprecated Use `client.volumes.fromName()` instead.
   */
  static async fromName(
    name: string,
    options?: VolumeFromNameOptions,
  ): Promise<Volume> {
    return getDefaultClient().volumes.fromName(name, options);
  }

  /** Configure Volume to mount as read-only. */
  readOnly(): Volume {
    return new Volume(this.volumeId, this.name, true, this.#ephemeralHbManager);
  }

  get isReadOnly(): boolean {
    return this._readOnly;
  }

  /**
   * @deprecated Use `client.volumes.ephemeral()` instead.
   */
  static async ephemeral(options: EphemeralOptions = {}): Promise<Volume> {
    return getDefaultClient().volumes.ephemeral(options);
  }

  /** Delete the ephemeral Volume. Only usable with emphemeral Volumes. */
  closeEphemeral(): void {
    if (this.#ephemeralHbManager) {
      this.#ephemeralHbManager.stop();
    } else {
      throw new InvalidError("Volume is not ephemeral.");
    }
  }
}
