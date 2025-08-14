// Demonstrates how to get current statistics for a Modal Function.

import { Function_ } from "modal";

const func = await Function_.lookup("libmodal-test-support", "echo_string");

const stats = await func.getCurrentStats();

console.log("Function Statistics:");
console.log(`  Backlog: ${stats.backlog} inputs`);
console.log(`  Total Runners: ${stats.numTotalRunners} containers`);
