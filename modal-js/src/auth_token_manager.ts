// Start refreshing this many seconds before the token expires
import type { Logger } from "./logger";

export const REFRESH_WINDOW = 5 * 60;
// If the token doesn't have an expiry field, default to current time plus this value (not expected).
export const DEFAULT_EXPIRY_OFFSET = 20 * 60;

export class AuthTokenManager {
  private client: any;
  private logger: Logger;
  private currentToken: string = "";
  private tokenExpiry: number = 0;
  private running: boolean = false;
  private fetchPromise: Promise<void> | null = null;

  constructor(client: any, logger: Logger) {
    this.client = client;
    this.logger = logger;
  }

  /**
   * Returns a valid auth token.
   * If the current token is expired and the manager is running, triggers an on-demand refresh.
   */
  async getToken(): Promise<string> {
    if (this.currentToken && !this.isExpired()) {
      return this.currentToken;
    }

    if (this.running) {
      await this.runFetch();
      if (this.currentToken && !this.isExpired()) {
        return this.currentToken;
      }
    }

    throw new Error("No valid auth token available");
  }

  /**
   * Fetches a new auth token from the server and stores it.
   */
  private async fetchToken(): Promise<void> {
    const response = await this.client.authTokenGet({});
    const token = response.token;

    if (!token) {
      throw new Error(
        "Internal error: did not receive auth token from server, please contact Modal support",
      );
    }

    this.currentToken = token;

    // Parse JWT expiry
    const exp = this.decodeJWT(token);
    if (exp > 0) {
      this.tokenExpiry = exp;
    } else {
      this.logger.warn("x-modal-auth-token does not contain exp field");
      // We'll use the token, and set the expiry to DEFAULT_EXPIRY_OFFSET from now.
      this.tokenExpiry = Math.floor(Date.now() / 1000) + DEFAULT_EXPIRY_OFFSET;
    }

    const now = Math.floor(Date.now() / 1000);
    const expiresIn = this.tokenExpiry - now;
    const refreshIn = this.tokenExpiry - now - REFRESH_WINDOW;
    this.logger.debug(
      "Fetched auth token",
      "expires_in",
      `${expiresIn}s`,
      "refresh_in",
      `${refreshIn}s`,
    );
  }

  /**
   * Fetches the initial token and starts the refresh loop.
   * Throws an error if the initial token fetch fails.
   */
  async start(): Promise<void> {
    if (this.running) {
      return;
    }

    this.running = true;
    try {
      await this.runFetch();
    } catch (error) {
      this.running = false;
      throw error;
    }
  }

  /**
   * Stops the background refresh.
   */
  stop(): void {
    this.running = false;
  }

  /**
   * Extracts the exp claim from a JWT token.
   */
  private decodeJWT(token: string): number {
    try {
      const parts = token.split(".");
      if (parts.length !== 3) {
        return 0;
      }

      let payload = parts[1];
      while (payload.length % 4 !== 0) {
        payload += "=";
      }

      const decoded = atob(payload);
      const claims = JSON.parse(decoded);
      return claims.exp || 0;
    } catch {
      return 0;
    }
  }

  isExpired(): boolean {
    const now = Math.floor(Date.now() / 1000);
    return now >= this.tokenExpiry;
  }

  private runFetch(): Promise<void> {
    if (!this.fetchPromise) {
      this.fetchPromise = (async () => {
        try {
          await this.fetchToken();
        } finally {
          this.fetchPromise = null;
        }
      })();
    }
    return this.fetchPromise;
  }

  getCurrentToken(): string {
    return this.currentToken;
  }

  setToken(token: string, expiry: number): void {
    this.currentToken = token;
    this.tokenExpiry = expiry;
  }
}
