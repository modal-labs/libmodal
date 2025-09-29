// This example calls a Function defined in `libmodal_test_support.py`.

import { ModalClient } from "modal";

const mc = new ModalClient();

const echo = await mc.functions.fromName(
  "libmodal-test-support",
  "echo_string",
);

// Spawn the Function with kwargs.
const functionCall = await echo.spawn([], { s: "Hello world!" });
const ret = await functionCall.get();
console.log(ret);
