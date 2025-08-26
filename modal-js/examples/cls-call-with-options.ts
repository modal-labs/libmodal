// This example calls a Modal Cls defined in `libmodal_test_support.py`,
// and overrides the default options.

import { Cls, Secret } from "modal";

// Lookup a deployed Cls.
const cls = await Cls.lookup("libmodal-test-support", "EchoClsParametrized");
const instance = await cls.instance();
const method = instance.method("echo_env_var");

const instanceWithOptions = await cls
  .withOptions({
    secrets: [await Secret.fromObject({ SECRET_MESSAGE: "hello, secret" })],
  })
  .withConcurrency({ maxInputs: 1 })
  .instance();
const methodWithOptions = instanceWithOptions.method("echo_env_var");

// Call the Cls function, without the secret being set.
console.log(await method.remote(["SECRET_MESSAGE"]));

// Call the Cls function with overrides, and confirm that the secret is set.
console.log(await methodWithOptions.remote(["SECRET_MESSAGE"]));
