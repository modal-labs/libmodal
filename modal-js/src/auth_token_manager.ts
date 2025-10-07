// Start refreshing this many seconds before the token expires
export const REFRESH_WINDOW = 5 * 60;
// If the token doesn't have an expiry field, default to current time plus this value (not expected).
export const DEFAULT_EXPIRY_OFFSET = 20 * 60;

export class AuthTokenManager {
  private client: any;
  private currentToken: string = "";
  private tokenExpiry: number = 0;
  private stopped: boolean = false;
  private timeoutId: NodeJS.Timeout | null = null;
  private initialTokenPromise: Promise<void> | null = null;

  constructor(client: any) {
    this.client = client;
  }

  /**
   * Returns the current cached token.
   * If the initial token fetch is still in progress, waits for it to complete.
   */
  async getToken(): Promise<string> {
    // If initial fetch is in progress, wait for it
    if (this.initialTokenPromise) {
      await this.initialTokenPromise;
    }

    if (this.currentToken && !this.isExpired()) {
      return this.currentToken;
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
      console.warn("Failed to decode x-modal-auth-token exp field");
      // We'll use the token, and set the expiry to DEFAULT_EXPIRY_OFFSET from now.
      this.tokenExpiry = Math.floor(Date.now() / 1000) + DEFAULT_EXPIRY_OFFSET;
    }
  }

  /**
   * Background loop that refreshes tokens REFRESH_WINDOW seconds before they expire.
   */
  private async backgroundRefresh(): Promise<void> {
    while (!this.stopped) {
      const now = Math.floor(Date.now() / 1000);
      const refreshTime = this.tokenExpiry - REFRESH_WINDOW;
      const delay = Math.max(0, refreshTime - now) * 1000;

      console.log(`Refreshing token in ${Math.floor(delay / 1000)} seconds`);

      // Sleep until it's time to refresh
      await new Promise<void>((resolve) => {
        this.timeoutId = setTimeout(resolve, delay);
        this.timeoutId.unref();
      });

      if (this.stopped) {
        return;
      }

      // Fetch new token
      try {
        await this.fetchToken();
      } catch (error) {
        console.error("Failed to refresh auth token:", error);
        // Sleep for 5 seconds before trying again on failure
        await new Promise((resolve) => setTimeout(resolve, 5000));
      }
    }
  }

  /**
   * Fetches the initial token and starts the refresh loop.
   * Throws an error if the initial token fetch fails.
   */
  async start(): Promise<void> {
    // Fetch initial token and store the promise so getToken() can wait for it
    this.initialTokenPromise = this.fetchToken();

    try {
      await this.initialTokenPromise;
    } finally {
      this.initialTokenPromise = null;
    }

    // Start background refresh loop, do not await
    this.stopped = false;
    this.backgroundRefresh();
  }

  /**
   * Stops the background refresh.
   */
  stop(): void {
    this.stopped = true;
    if (this.timeoutId) {
      clearTimeout(this.timeoutId);
      this.timeoutId = null;
    }
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

  getCurrentToken(): string {
    return this.currentToken;
  }

  setToken(token: string, expiry: number): void {
    this.currentToken = token;
    this.tokenExpiry = expiry;
  }
}
