// This example calls a Function defined in `libmodal_test_support.py`.

import { Function_ } from "modal";

const echo = await Function_.lookup("cbor-app", "f");

// Call the Function with args.
let ret = await echo.remote(["Hello world!"]);
console.log(ret);

// Call the Function with kwargs.
ret = await echo.remote([], { arg: "Hello world!" });
console.log(ret);
