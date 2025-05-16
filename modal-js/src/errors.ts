/** Function execution exceeds the allowed time limit. */
export class FunctionTimeoutError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "FunctionTimeoutError";
  }
}

/** An error on the Modal server, or a Python exception. */
export class RemoteError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "RemoteError";
  }
}

/** A retryable internal error from Modal. */
export class InternalFailure extends Error {
  constructor(message: string) {
    super(message);
    this.name = "InternalFailure";
  }
}

/** Some resource was not found. */
export class NotFoundError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "NotFoundError";
  }
}
