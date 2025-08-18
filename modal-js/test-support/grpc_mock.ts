import { vi } from "vitest";

export class MockGrpc {
  // Map of short RPC name -> FIFO queue of handlers
  private readonly methodHandlerQueues: Map<
    string,
    Array<(req: unknown) => unknown | Promise<unknown>>
  > = new Map();

  static async install(): Promise<MockGrpc> {
    const instance = new MockGrpc();
    vi.resetModules();

    const mockClient: Record<string, (req: unknown) => Promise<unknown>> =
      new Proxy(
        {},
        {
          get(_target, propKey) {
            if (typeof propKey !== "string") return undefined;
            return (req: unknown) => instance.dispatch(propKey, req);
          },
        },
      );

    vi.doMock("../src/client", async () => {
      const actual = (await vi.importActual<any>("../src/client")) as Record<
        string,
        unknown
      >;
      return {
        ...actual,
        client: mockClient,
      };
    });

    return instance;
  }

  async uninstall(): Promise<void> {
    this.assertExhausted();
    vi.unmock("../src/client");
    vi.resetModules();
    this.methodHandlerQueues.clear();
  }

  private readonly dispatch = async (
    methodKey: string,
    actualRequest: unknown,
  ): Promise<unknown> => {
    const queue = this.methodHandlerQueues.get(methodKey) ?? [];
    if (queue.length === 0) {
      throw new Error(
        `Unexpected gRPC call: ${methodKey} with request ${formatValue(actualRequest)}`,
      );
    }
    const handler = queue.shift()!;
    const response = await handler(actualRequest);
    return structuredClone(response);
  };

  handleUnary(
    rpcName: string,
    handler: (req: unknown) => unknown | Promise<unknown>,
  ) {
    const methodKey = rpcToClientMethodName(shortName(rpcName));
    const queue = this.methodHandlerQueues.get(methodKey) ?? [];
    queue.push(handler);
    this.methodHandlerQueues.set(methodKey, queue);
  }

  assertExhausted() {
    const outstanding = Array.from(this.methodHandlerQueues.entries()).filter(
      ([, q]) => q.length > 0,
    );
    if (outstanding.length > 0) {
      const details = outstanding
        .map(([k, q]) => `- ${k}: ${q.length} expectation(s) remaining`)
        .join("\n");
      throw new Error(`Not all expected gRPC calls were made:\n${details}`);
    }
  }
}

function rpcToClientMethodName(name: string): string {
  return name.length ? name[0].toLowerCase() + name.slice(1) : name;
}

function shortName(method: string): string {
  if (method.startsWith("/")) {
    const idx = method.lastIndexOf("/");
    if (idx >= 0 && idx + 1 < method.length) {
      return method.slice(idx + 1);
    }
  }
  return method;
}

function formatValue(v: unknown): string {
  try {
    return JSON.stringify(v, undefined, 2);
  } catch {
    return String(v);
  }
}
