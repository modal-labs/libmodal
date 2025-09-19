import { ObjectCreationType } from "../proto/modal_proto/api";
import { client } from "./client";
import { environmentName as configEnvironmentName } from "./config";
import { ClientError, Status } from "nice-grpc";
import { NotFoundError, InvalidError } from "./errors";
import { EphemeralHeartbeatManager } from "./ephemeral";
import type { EphemeralOptions } from "./app";

/** Options for `Volume.fromName()`. */
export type VolumeFromNameOptions = {
  environment?: string;
  createIfMissing?: boolean;
};

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

  static async fromName(
    name: string,
    options?: VolumeFromNameOptions,
  ): Promise<Volume> {
    try {
      const resp = await client.volumeGetOrCreate({
        deploymentName: name,
        environmentName: configEnvironmentName(options?.environment),
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

  /** Configure Volume to mount as read-only. */
  readOnly(): Volume {
    return new Volume(this.volumeId, this.name, true, this.#ephemeralHbManager);
  }

  get isReadOnly(): boolean {
    return this._readOnly;
  }

  /**
   * Create a nameless, temporary Volume.
   * You will need to call `closeEphemeral()` to delete the Volume.
   */
  static async ephemeral(options: EphemeralOptions = {}): Promise<Volume> {
    const resp = await client.volumeGetOrCreate({
      objectCreationType: ObjectCreationType.OBJECT_CREATION_TYPE_EPHEMERAL,
      environmentName: configEnvironmentName(options.environment),
    });

    const ephemeralHbManager = new EphemeralHeartbeatManager(() =>
      client.volumeHeartbeat({ volumeId: resp.volumeId }),
    );

    return new Volume(resp.volumeId, undefined, false, ephemeralHbManager);
  }

  /** Delete the ephemeral Volume. Only usable with `Volume.ephemeral()`. */
  closeEphemeral(): void {
    if (this.#ephemeralHbManager) {
      this.#ephemeralHbManager.stop();
    } else {
      throw new InvalidError("Volume is not ephemeral.");
    }
  }
}
