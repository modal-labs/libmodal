// This example calls a Modal Cls defined in `libmodal_test_support.py`.

import { Cls } from "modal";

// Lookup a deployed Cls.
const cls = await Cls.lookup("libmodal-test-support", "EchoCls");
const instance = await cls.instance();
const method = instance.method("echo_string");

// Call the Cls function with args.
let ret = await method.remote(["Hello world!"]);
console.log(ret);

// Call the Cls function with kwargs.
ret = await method.remote([], { s: "Hello world!" });
console.log(ret);
