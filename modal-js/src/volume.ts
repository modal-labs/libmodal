import { ObjectCreationType } from "../proto/modal_proto/api";
import { client } from "./client";
import { environmentName as configEnvironmentName } from "./config";
import { ClientError, Status } from "nice-grpc";
import { NotFoundError } from "./errors";

/** Options for `Volume.fromName()`. */
export type VolumeFromNameOptions = {
  environment?: string;
  createIfMissing?: boolean;
};

/** Volumes provide persistent storage that can be mounted in Modal functions. */
export class Volume {
  readonly volumeId: string;

  /** @ignore */
  constructor(volumeId: string) {
    this.volumeId = volumeId;
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
      return new Volume(resp.volumeId);
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.NOT_FOUND)
        throw new NotFoundError(err.details);
      throw err;
    }
  }
}
