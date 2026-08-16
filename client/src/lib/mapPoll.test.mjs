import assert from "node:assert/strict";
import test from "node:test";

import { OCEAN_POLL_INTERVAL_MS, startMapPolling } from "./mapPoll.ts";

test("Ocean map polling refreshes Cork positions and stops on cleanup", () => {
  let scheduled;
  let cancelled;
  let refreshes = 0;

  const stop = startMapPolling(
    () => {
      refreshes += 1;
    },
    (callback, intervalMs) => {
      scheduled = { callback, intervalMs };
      return 17;
    },
    handle => {
      cancelled = handle;
    },
  );

  assert.equal(scheduled.intervalMs, OCEAN_POLL_INTERVAL_MS);
  scheduled.callback();
  assert.equal(refreshes, 1);

  stop();
  assert.equal(cancelled, 17);
});
