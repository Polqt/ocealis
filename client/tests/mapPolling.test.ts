import assert from "node:assert/strict";
import test from "node:test";

import { startMapPolling, type IntervalScheduler } from "../src/lib/mapPolling.ts";

test("polls the Ocean and stops polling on cleanup", () => {
  let poll: (() => void) | undefined;
  let cleared: unknown;
  const intervalID = { id: 7 };
  const scheduler: IntervalScheduler = {
    setInterval(callback, delay) {
      assert.equal(delay, 60_000);
      poll = callback;
      return intervalID;
    },
    clearInterval(id) {
      cleared = id;
    },
  };
  let refreshes = 0;

  const stop = startMapPolling(() => {
    refreshes++;
  }, scheduler);

  assert.ok(poll, "expected an Ocean poll to be scheduled");
  poll();
  assert.equal(refreshes, 1);

  stop();
  assert.equal(cleared, intervalID);
});
