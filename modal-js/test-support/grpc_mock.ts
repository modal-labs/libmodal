import { vi, expect } from "vitest";

// Usage per test:
//   const mock = await MockGrpc.install();
//   mock.on('FunctionMap').expect({ ... }).reply({ ... });
//   mock.on('FunctionGetOutputs').expect({ ... }).reply({ ... });
//   // run code under test (import after install or within vi.isolateModules)
//   mock.assertExhausted();
//   await mock.uninstall();

class ExpectationBuilder {
  private readonly method: string;
  private readonly mock: MockGrpc;
  private expectedRequest: unknown | undefined;

  constructor(method: string, mock: MockGrpc) {
    this.method = method;
    this.mock = mock;
  }

  expect(expectedRequest: unknown) {
    this.expectedRequest = expectedRequest;
    return this;
  }

  reply(response: unknown) {
    if (this.expectedRequest === undefined) {
      throw new Error("Call .expect(...) before .reply(...)");
    }
    this.mock.enqueueExpectation(this.method, {
      expected: this.expectedRequest,
      response,
    });
  }
}

export class MockGrpc {
  // Map of client method name -> FIFO queue of expectations
  private readonly queues: Map<
    string,
    Array<{ expected: unknown; response: unknown }>
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
    vi.unmock("../src/client");
    vi.resetModules();
    this.queues.clear();
  }

  private readonly dispatch = async (
    methodKey: string,
    actualRequest: unknown,
  ): Promise<unknown> => {
    const queue = this.queues.get(methodKey) ?? [];
    if (queue.length === 0) {
      throw new Error(
        `Unexpected gRPC call: ${methodKey} with request ${formatValue(actualRequest)}`,
      );
    }
    const { expected, response } = queue.shift()!;

    try {
      if (
        typeof actualRequest === "object" &&
        actualRequest !== null &&
        typeof expected === "object" &&
        expected !== null
      ) {
        expect(actualRequest).toMatchObject(expected);
      } else {
        expect(actualRequest).toEqual(expected);
      }
    } catch (error) {
      throw new Error(
        `gRPC request mismatch for ${methodKey}:\n${(error as Error).message}`,
      );
    }

    return structuredClone(response);
  };

  on(rpcName: "FunctionGetCurrentStats" | "FunctionUpdateSchedulingParams") {
    const method = rpcToClientMethodName(rpcName);
    return new ExpectationBuilder(method, this);
  }

  enqueueExpectation(
    methodKey: string,
    entry: { expected: unknown; response: unknown },
  ) {
    const queue = this.queues.get(methodKey) ?? [];
    queue.push(entry);
    this.queues.set(methodKey, queue);
  }

  assertExhausted() {
    const outstanding = Array.from(this.queues.entries()).filter(
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

function formatValue(v: unknown): string {
  try {
    return JSON.stringify(v, undefined, 2);
  } catch {
    return String(v);
  }
}
