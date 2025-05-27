import { DeploymentNamespace } from "../proto/modal_proto/api";
import { client } from "./client";
import { environmentName as configEnvironmentName } from "./config";
import { ClientError, Status } from "nice-grpc";
import { NotFoundError } from "./errors";

export type SecretFromNameOptions = {
  namespace?: DeploymentNamespace;
  environment_name?: string;
  required_keys?: string[];
};

export class Secret {
  readonly secretId: string;

  constructor(secretId: string) {
    this.secretId = secretId;
  }

  static async from_name(
    name: string,
    options?: SecretFromNameOptions,
  ): Promise<Secret> {
    try {
      const resp = await client.secretGetOrCreate({
        deploymentName: name,
        namespace:
          options?.namespace ??
          DeploymentNamespace.DEPLOYMENT_NAMESPACE_WORKSPACE,
        environmentName: configEnvironmentName(options?.environment_name),
        requiredKeys: options?.required_keys ?? [],
      });
      return new Secret(resp.secretId);
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.NOT_FOUND)
        throw new NotFoundError(`Secret '${name}' not found`);
      throw err;
    }
  }
}
