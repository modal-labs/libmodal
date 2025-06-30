// This example configures a client using a `CUSTOM_MODAL_ID` and `CUSTOM_MODAL_SECRET` environment variable.

import { Function_, initializeClient } from "modal";

const modalId = process.env.CUSTOM_MODAL_ID;
if (!modalId) {
  throw new Error("CUSTOM_MODAL_ID environment variable not set");
}
const modalSecret = process.env.CUSTOM_MODAL_SECRET;
if (!modalSecret) {
  throw new Error("CUSTOM_MODAL_SECRET environment variable not set");
}

initializeClient({ tokenId: modalId, tokenSecret: modalSecret });

const echo = await Function_.lookup("libmodal-test-support", "echo_string");
console.log(echo);
