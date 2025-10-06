import { v4 as uuidv4 } from "uuid";
import {
  CallOptions,
  ClientError,
  ClientMiddleware,
  ClientMiddlewareCall,
  createChannel,
  createClientFactory,
  Metadata,
  Status,
} from "nice-grpc";

import { ClientType, ModalClientDefinition } from "../proto/modal_proto/api";
import { getProfile, type Profile } from "./config";

const defaultProfile = getProfile(process.env["MODAL_PROFILE"]);

// Start refreshing this many seconds before the token expires
export const REFRESH_WINDOW = 5 * 60;
// If the token doesn't have an expiry field, default to current time plus this value (not expected).
export const DEFAULT_EXPIRY_OFFSET = 20 * 60;

/**
 * Manages auth tokens, refreshing tokens REFRESH_WINDOW seconds before they expire.
 */
export class AuthTokenManager {
  private client: ReturnType<typeof createClient>;
  private currentToken: string = "";
  private tokenExpiry: number = 0;
  private refreshPromise: Promise<string> | null = null;
  private nextRefresh: NodeJS.Timeout | null = null;

  constructor(client: ReturnType<typeof createClient>) {
    this.client = client;
  }

  /**
   * Returns cached token.
   * Concurrent calls will share the same refresh promise to avoid duplicate requests.
   */
  async getToken(): Promise<string> {
    // Return existing, non-expired token
    if (this.currentToken && !this.isExpired()) {
      return this.currentToken;
    }

    // If already refreshing, wait for that promise
    if (this.refreshPromise) {
      return await this.refreshPromise;
    }

    throw new Error("No valid auth token available");
  }

  /**
   * Refreshes the auth token by calling the server and stores it.
   * Returns a promise that resolves to the new token. Concurrent calls will get the same promise to avoid duplicate refreshes.
   */
  refreshToken(): Promise<string> {
    if (this.refreshPromise) {
      return this.refreshPromise;
    }

    // Store the promise first, then add cleanup
    this.refreshPromise = (async () => {
      try {
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
          this.tokenExpiry =
            Math.floor(Date.now() / 1000) + DEFAULT_EXPIRY_OFFSET;
        }

        // Schedule next refresh
        this.scheduleRefresh();

        return token;
      } catch (error) {
        console.error("Failed to refresh auth token:", error);
        throw error;
      } finally {
        this.refreshPromise = null;
      }
    })();

    return this.refreshPromise;
  }

  /**
   * Schedules the next token refresh based on the current token's expiry.
   */
  private scheduleRefresh() {
    if (this.nextRefresh) {
      clearTimeout(this.nextRefresh);
    }

    const now = Math.floor(Date.now() / 1000);
    const refreshTime = this.tokenExpiry - REFRESH_WINDOW;
    const delay = Math.max(0, refreshTime - now) * 1000;

    this.nextRefresh = setTimeout(async () => {
      try {
        await this.refreshToken();
      } catch (error) {
        console.error("Token refresh failed:", error);
      }
    }, delay);
  }

  /**
   * Fetches the initial token on start.
   * Throws an error if the initial token fetch fails.
   */
  async start(): Promise<void> {
    if (this.currentToken === "") {
      await this.refreshToken();
    } else {
      this.scheduleRefresh();
    }
  }

  /**
   * Stops the background refresh.
   */
  stop(): void {
    if (this.nextRefresh) {
      clearTimeout(this.nextRefresh);
      this.nextRefresh = null;
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

  /**
   * Checks if the current token is expired.
   */
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

let authTokenManager: AuthTokenManager | null = null;

/** gRPC client middleware to add auth token to request. */
function authMiddleware(profile: Profile): ClientMiddleware {
  return async function* authMiddleware<Request, Response>(
    call: ClientMiddlewareCall<Request, Response>,
    options: CallOptions,
  ) {
    if (!profile.tokenId || !profile.tokenSecret) {
      throw new Error(
        `Profile is missing token_id or token_secret. Please set them in .modal.toml, or as environment variables, or initializeClient().`,
      );
    }
    const { tokenId, tokenSecret } = profile;

    options.metadata ??= new Metadata();
    options.metadata.set(
      "x-modal-client-type",
      String(ClientType.CLIENT_TYPE_LIBMODAL_JS),
    );
    options.metadata.set("x-modal-client-version", "1.0.0"); // CLIENT VERSION: Behaves like this Python SDK version
    options.metadata.set("x-modal-token-id", tokenId);
    options.metadata.set("x-modal-token-secret", tokenSecret);

    // Skip auth token for AuthTokenGet requests to prevent it from getting stuck
    if (call.method.path !== "/modal.client.ModalClient/AuthTokenGet") {
      if (!authTokenManager) {
        throw new Error("Auth token manager not initialized");
      }

      const token = await authTokenManager.getToken();
      if (token) {
        options.metadata.set("x-modal-auth-token", token);
      }
    }

    return yield* call.next(call.request, options);
  };
}

type TimeoutOptions = {
  /** Timeout for this call, interpreted as a duration in milliseconds */
  timeout?: number;
};

/** gRPC client middleware to set timeout and retries on a call. */
const timeoutMiddleware: ClientMiddleware<TimeoutOptions> =
  async function* timeoutMiddleware(call, options) {
    if (!options.timeout || options.signal?.aborted) {
      return yield* call.next(call.request, options);
    }

    const { timeout, signal: origSignal, ...restOptions } = options;
    const abortController = new AbortController();
    const abortListener = () => abortController.abort();
    origSignal?.addEventListener("abort", abortListener);

    let timedOut = false;

    const timer = setTimeout(() => {
      timedOut = true;
      abortController.abort();
    }, timeout);

    try {
      return yield* call.next(call.request, {
        ...restOptions,
        signal: abortController.signal,
      });
    } finally {
      origSignal?.removeEventListener("abort", abortListener);
      clearTimeout(timer);

      if (timedOut) {
        // eslint-disable-next-line no-unsafe-finally
        throw new ClientError(
          call.method.path,
          Status.DEADLINE_EXCEEDED,
          `Timed out after ${timeout}ms`,
        );
      }
    }
  };

const retryableGrpcStatusCodes = new Set([
  Status.DEADLINE_EXCEEDED,
  Status.UNAVAILABLE,
  Status.CANCELLED,
  Status.INTERNAL,
  Status.UNKNOWN,
]);

export function isRetryableGrpc(err: unknown) {
  if (err instanceof ClientError) {
    return retryableGrpcStatusCodes.has(err.code);
  }
  return false;
}

/** Sleep helper that can be cancelled via an AbortSignal. */
const sleep = (ms: number, signal?: AbortSignal) =>
  new Promise<void>((resolve, reject) => {
    if (signal?.aborted) return reject(signal.reason);
    const t = setTimeout(resolve, ms);
    signal?.addEventListener(
      "abort",
      () => {
        clearTimeout(t);
        reject(signal.reason);
      },
      { once: true },
    );
  });

type RetryOptions = {
  /** Number of retries to take. */
  retries?: number;

  /** Base delay in milliseconds. */
  baseDelay?: number;

  /** Maximum delay in milliseconds. */
  maxDelay?: number;

  /** Exponential factor to multiply successive delays. */
  delayFactor?: number;

  /** Additional status codes to retry. */
  additionalStatusCodes?: Status[];
};

/** Middleware to retry transient errors and timeouts for unary requests. */
const retryMiddleware: ClientMiddleware<RetryOptions> =
  async function* retryMiddleware(call, options) {
    const {
      retries = 3,
      baseDelay = 100,
      maxDelay = 1000,
      delayFactor = 2,
      additionalStatusCodes = [],
      signal,
      ...restOptions
    } = options;

    if (call.requestStream || call.responseStream || !retries) {
      // Don't retry streaming calls, or if retries are disabled.
      return yield* call.next(call.request, restOptions);
    }

    const retryableCodes = new Set([
      ...retryableGrpcStatusCodes,
      ...additionalStatusCodes,
    ]);

    // One idempotency key for the whole call (all attempts).
    const idempotencyKey = uuidv4();

    const startTime = Date.now();
    let attempt = 0;
    let delayMs = baseDelay;

    while (true) {
      // Clone/augment metadata for this attempt.
      const metadata = new Metadata(restOptions.metadata ?? {});

      metadata.set("x-idempotency-key", idempotencyKey);
      metadata.set("x-retry-attempt", String(attempt));
      if (attempt > 0) {
        metadata.set(
          "x-retry-delay",
          ((Date.now() - startTime) / 1000).toFixed(3),
        );
      }

      try {
        // Forward the call.
        return yield* call.next(call.request, {
          ...restOptions,
          metadata,
          signal,
        });
      } catch (err) {
        // Immediately propagate non-retryable situations.
        if (
          !(err instanceof ClientError) ||
          !retryableCodes.has(err.code) ||
          attempt >= retries
        ) {
          throw err;
        }

        // Exponential back-off with a hard cap.
        await sleep(delayMs, signal);
        delayMs = Math.min(delayMs * delayFactor, maxDelay);
        attempt += 1;
      }
    }
  };

/** Map of server URL to input-plane client. */
const inputPlaneClients: Record<string, ReturnType<typeof createClient>> = {};

/** Returns a client for the given server URL, creating it if it doesn't exist. */
export const getOrCreateInputPlaneClient = (
  serverUrl: string,
): ReturnType<typeof createClient> => {
  const client = inputPlaneClients[serverUrl];
  if (client) {
    return client;
  }
  const profile = { ...clientProfile, serverUrl };
  const newClient = createClient(profile);
  inputPlaneClients[serverUrl] = newClient;
  return newClient;
};

function createClient(profile: Profile) {
  // Channels don't do anything until you send a request on them.
  // Ref: https://github.com/modal-labs/modal-client/blob/main/modal/_utils/grpc_utils.py
  const channel = createChannel(profile.serverUrl, undefined, {
    "grpc.max_receive_message_length": 100 * 1024 * 1024,
    "grpc.max_send_message_length": 100 * 1024 * 1024,
    "grpc-node.flow_control_window": 64 * 1024 * 1024,
  });
  return createClientFactory()
    .use(authMiddleware(profile))
    .use(retryMiddleware)
    .use(timeoutMiddleware)
    .create(ModalClientDefinition, channel);
}

export let clientProfile = defaultProfile;

export let client = createClient(clientProfile);

/** Options for initializing a client at runtime. */
export type ClientOptions = {
  tokenId: string;
  tokenSecret: string;
  environment?: string;
};

/**
 * Initialize the Modal client, passing in token authentication credentials.
 *
 * You should call this function at the start of your application if not
 * configuring Modal with a `.modal.toml` file or environment variables.
 * Auth token will be fetched on startup.
 */
export async function initializeClient(options: ClientOptions) {
  const mergedProfile = {
    ...defaultProfile,
    tokenId: options.tokenId,
    tokenSecret: options.tokenSecret,
    environment: options.environment || defaultProfile.environment,
  };
  clientProfile = mergedProfile;
  client = createClient(mergedProfile);

  if (authTokenManager) {
    authTokenManager.stop();
  }
  authTokenManager = new AuthTokenManager(client);
  await authTokenManager.start();
}
