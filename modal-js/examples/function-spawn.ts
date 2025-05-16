// This example calls a function defined in `libmodal_test_support.py`.

import { Function_ } from "modal";

const echo = await Function_.lookup("libmodal-test-support", "echo_string");

// Spawn the function with kwargs.
const functionCall = await echo.spawn([], { s: "Hello world!" });
const ret = await functionCall.get();
console.log(ret);
