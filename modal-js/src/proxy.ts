import { getDefaultClient } from "./client";
import { ClientError, Status } from "nice-grpc";
import { NotFoundError } from "./errors";
import { APIService } from "./api-service";

/**
 * Service for managing Proxies.
 */
export class ProxyService extends APIService {
  /**
   * Reference a Proxy by its name.
   */
  async fromName(name: string, options?: ProxyFromNameOptions): Promise<Proxy> {
    try {
      const resp = await this.client.cpClient.proxyGet({
        name,
        environmentName: this.client.environmentName(options?.environment),
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

  /**
   * @deprecated Use `client.proxies.fromName()` instead.
   */
  static async fromName(
    name: string,
    options?: ProxyFromNameOptions,
  ): Promise<Proxy> {
    return getDefaultClient().proxies.fromName(name, options);
  }
}
