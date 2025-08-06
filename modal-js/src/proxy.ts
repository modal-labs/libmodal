import { client } from "./client";
import { environmentName as configEnvironmentName } from "./config";
import { ClientError, Status } from "nice-grpc";
import { NotFoundError } from "./errors";

/** Options for `Proxy.fromName()`. */
export type ProxyFromNameOptions = {
  environment?: string;
};

/** Proxy objects give your Modal containers a static outbound IP address. */
export class Proxy {
  readonly proxyId: string;

  /** @ignore */
  constructor(proxyId: string) {
    this.proxyId = proxyId;
  }

  /** Reference a Proxy by its name. */
  static async fromName(
    name: string,
    options?: ProxyFromNameOptions,
  ): Promise<Proxy> {
    try {
      const resp = await client.proxyGet({
        name,
        environmentName: configEnvironmentName(options?.environment),
      });
      if (!resp.proxy?.proxyId) {
        throw new NotFoundError(`Proxy '${name}' not found`);
      }
      return new Proxy(resp.proxy.proxyId);
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.NOT_FOUND)
        throw new NotFoundError(`Proxy '${name}' not found`);
      throw err;
    }
  }
}
