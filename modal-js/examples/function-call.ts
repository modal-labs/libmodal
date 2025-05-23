// This example calls a function defined in `libmodal_test_support.py`.

import { Function_ } from "modal";

const echo = await Function_.lookup("libmodal-test-support", "echo_string");

// Call the function with args.
let ret = await echo.remote(["Hello world!"]);
console.log(ret);

// Call the function with kwargs.
ret = await echo.remote([], { s: "Hello world!" });
console.log(ret);
