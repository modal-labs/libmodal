import { client } from "./client";
import { environmentName as configEnvironmentName } from "./config";
import { ClientError, Status } from "nice-grpc";
import { InvalidError, NotFoundError } from "./errors";
import { ObjectCreationType } from "../proto/modal_proto/api";

/** Options for `Secret.fromName()`. */
export type SecretFromNameOptions = {
  environment?: string;
  requiredKeys?: string[];
};

/** Secrets provide a dictionary of environment variables for images. */
export class Secret {
  readonly secretId: string;

  /** @ignore */
  constructor(secretId: string) {
    this.secretId = secretId;
  }

  /** Reference a Secret by its name. */
  static async fromName(
    name: string,
    options?: SecretFromNameOptions,
  ): Promise<Secret> {
    try {
      const resp = await client.secretGetOrCreate({
        deploymentName: name,
        environmentName: configEnvironmentName(options?.environment),
        requiredKeys: options?.requiredKeys ?? [],
      });
      return new Secret(resp.secretId);
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.NOT_FOUND)
        throw new NotFoundError(err.details);
      if (
        err instanceof ClientError &&
        err.code === Status.FAILED_PRECONDITION &&
        err.details.includes("Secret is missing key")
      )
        throw new NotFoundError(err.details);
      throw err;
    }
  }

  /** Create a Secret from a plain object of key-value pairs. */
  static async fromObject(
    entries: Record<string, string>,
    options?: { environment?: string },
  ): Promise<Secret> {
    for (const [, value] of Object.entries(entries)) {
      if (value == null || typeof value !== "string") {
        throw new InvalidError(
          "entries must be an object mapping string keys to string values, but got:\n" +
            JSON.stringify(entries),
        );
      }
    }

    try {
      const resp = await client.secretGetOrCreate({
        objectCreationType: ObjectCreationType.OBJECT_CREATION_TYPE_EPHEMERAL,
        envDict: entries as Record<string, string>,
        environmentName: configEnvironmentName(options?.environment),
      });
      return new Secret(resp.secretId);
    } catch (err) {
      if (
        err instanceof ClientError &&
        (err.code === Status.INVALID_ARGUMENT ||
          err.code === Status.FAILED_PRECONDITION)
      )
        throw new InvalidError(err.details);
      throw err;
    }
  }
}
