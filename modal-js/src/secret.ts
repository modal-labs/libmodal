import { getDefaultClient } from "./client";
import { environmentName as configEnvironmentName } from "./config";
import { ClientError, Status } from "nice-grpc";
import { InvalidError, NotFoundError } from "./errors";
import { ObjectCreationType } from "../proto/modal_proto/api";
import { APIService } from "./api-service";

/** Options for `Secret.fromName()`. */
export type SecretFromNameOptions = {
  environment?: string;
  requiredKeys?: string[];
};

/**
 * Service for managing Secrets.
 */
export class SecretService extends APIService {
  /** Reference a Secret by its name. */
  async fromName(
    name: string,
    options?: SecretFromNameOptions,
  ): Promise<Secret> {
    try {
      const resp = await this.client.cpClient.secretGetOrCreate({
        deploymentName: name,
        environmentName: configEnvironmentName(options?.environment),
        requiredKeys: options?.requiredKeys ?? [],
      });
      return new Secret(resp.secretId, name);
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
  async fromObject(
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
      const resp = await this.client.cpClient.secretGetOrCreate({
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

/** Secrets provide a dictionary of environment variables for Images. */
export class Secret {
  readonly secretId: string;
  readonly name?: string;

  /** @ignore */
  constructor(secretId: string, name?: string) {
    this.secretId = secretId;
    this.name = name;
  }

  /**
   * @deprecated Use `client.secrets.fromName()` instead.
   */
  static async fromName(
    name: string,
    options?: SecretFromNameOptions,
  ): Promise<Secret> {
    return getDefaultClient().secrets.fromName(name, options);
  }

  /**
   * @deprecated Use `client.secrets.fromMap()` instead.
   */
  static async fromObject(
    entries: Record<string, string>,
    options?: { environment?: string },
  ): Promise<Secret> {
    return getDefaultClient().secrets.fromObject(entries, options);
  }
}

export async function mergeEnvAndSecrets(
  env?: Record<string, string>,
  secrets?: Secret[],
): Promise<Secret[]> {
  const result = [...(secrets || [])];
  if (env && Object.keys(env).length > 0) {
    result.push(await getDefaultClient().secrets.fromObject(env));
  }
  return result;
}
