import { client } from "./client";
import { environmentName as configEnvironmentName } from "./config";
import { ClientError, Status } from "nice-grpc";
import { NotFoundError } from "./errors";

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
}
