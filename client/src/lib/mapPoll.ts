export const OCEAN_POLL_INTERVAL_MS = 60_000;

type IntervalHandle = ReturnType<typeof setInterval>;
type ScheduleInterval = (callback: () => void, intervalMs: number) => IntervalHandle;
type CancelInterval = (handle: IntervalHandle) => void;

export function startMapPolling(
  refresh: () => void,
  schedule: ScheduleInterval = globalThis.setInterval,
  cancel: CancelInterval = globalThis.clearInterval,
): () => void {
  const handle = schedule(refresh, OCEAN_POLL_INTERVAL_MS);
  return () => cancel(handle);
}
