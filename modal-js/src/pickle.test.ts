import { expect, test } from "vitest";
import { dumps, loads, type Protocol } from "./pickle";

test("PickleUnpickle", () => {
  const sample = {
    a: 1,
    b: [2, 3, true, null],
    c: new Uint8Array([4, 5, 6]),
    d: "hello 🎉",
  };
  for (const proto of [3, 4, 5] as Protocol[]) {
    const pkl = dumps(sample, proto);
    const back = loads(pkl);
    expect(back).toEqual(sample);
  }
});
