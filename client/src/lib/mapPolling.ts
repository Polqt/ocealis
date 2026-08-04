const OCEAN_POLL_MS = 60_000;

export type IntervalScheduler = {
  setInterval(callback: () => void, delay: number): unknown;
  clearInterval(id: unknown): void;
};

const browserScheduler: IntervalScheduler = {
  setInterval: (callback, delay) => window.setInterval(callback, delay),
  clearInterval: id => window.clearInterval(id as number),
};

export function startMapPolling(
  refresh: () => void,
  scheduler: IntervalScheduler = browserScheduler,
): () => void {
  const intervalID = scheduler.setInterval(refresh, OCEAN_POLL_MS);
  return () => scheduler.clearInterval(intervalID);
}
