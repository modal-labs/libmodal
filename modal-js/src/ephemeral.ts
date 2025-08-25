export const ephemeralObjectHeartbeatSleep = 300_000; // 300 seconds

export type HeartbeatFunction = (objectId: string) => Promise<any>;

export class EphemeralHeartbeatManager {
  private readonly objectId: string;
  private readonly heartbeatFn: HeartbeatFunction;
  private readonly abortController: AbortController;
  private timerId?: NodeJS.Timeout;

  constructor(objectId: string, heartbeatFn: HeartbeatFunction) {
    this.objectId = objectId;
    this.heartbeatFn = heartbeatFn;
    this.abortController = new AbortController();

    this.startHeartbeat();
  }

  private startHeartbeat(): void {
    const signal = this.abortController.signal;
    (async () => {
      while (!signal.aborted) {
        await this.heartbeatFn(this.objectId);
        await Promise.race([
          new Promise((resolve) => {
            this.timerId = setTimeout(resolve, ephemeralObjectHeartbeatSleep);
            this.timerId.unref(); // Don't let the heartbeat timer prevent the process from exiting
          }),
          new Promise((resolve) => {
            signal.addEventListener("abort", resolve, { once: true });
          }),
        ]);
      }
    })();
  }

  close(): void {
    this.abortController.abort();
  }
}
