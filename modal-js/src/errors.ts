/** Function execution exceeds the allowed time limit. */
export class TimeoutError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "TimeoutError";
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
