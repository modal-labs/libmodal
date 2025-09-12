import { expect, test } from "vitest";
import { sandboxCreateRequestProto } from "../src/app";
import { PTYInfo_PTYType } from "../proto/modal_proto/api";

test("sandboxCreateRequestProto without PTY", () => {
  const req = sandboxCreateRequestProto("app-123", "img-456");

  const definition = req.definition!;
  expect(definition.ptyInfo).toBeUndefined();
});

test("sandboxCreateRequestProto with PTY", () => {
  const req = sandboxCreateRequestProto("app-123", "img-456", {
    pty: true,
  });

  const definition = req.definition!;
  const ptyInfo = definition.ptyInfo!;
  expect(ptyInfo.enabled).toBe(true);
  expect(ptyInfo.winszRows).toBe(24);
  expect(ptyInfo.winszCols).toBe(80);
  expect(ptyInfo.envTerm).toBe("xterm-256color");
  expect(ptyInfo.envColorterm).toBe("truecolor");
  expect(ptyInfo.ptyType).toBe(PTYInfo_PTYType.PTY_TYPE_SHELL);
});
