import type { DriftPayload, WsEnvelope } from "./types";
import { wsUrl } from "./api";

export type OceanWsHandlers = {
  onDrift: (payload: DriftPayload) => void;
  onReleased: (bottleId: number) => void;
  onDiscovered: (bottleId: number) => void;
  onStatus?: (connected: boolean) => void;
};

export function connectOceanWs(handlers: OceanWsHandlers): () => void {
  let socket: WebSocket | null = null;
  let closed = false;
  let retry = 0;
  let timer: ReturnType<typeof setTimeout> | undefined;

  const connect = () => {
    if (closed) return;
    socket = new WebSocket(wsUrl());

    socket.onopen = () => {
      retry = 0;
      handlers.onStatus?.(true);
      socket?.send(JSON.stringify({ action: "subscribe", topic: "ocean:all" }));
    };

    socket.onmessage = event => {
      try {
        const msg = JSON.parse(String(event.data)) as WsEnvelope;
        if (msg.type === "bottle_drift") {
          handlers.onDrift(msg.payload as DriftPayload);
        } else if (msg.type === "bottle_released") {
          handlers.onReleased((msg.payload as { bottle_id: number }).bottle_id);
        } else if (msg.type === "bottle_discovered") {
          handlers.onDiscovered((msg.payload as { bottle_id: number }).bottle_id);
        }
      } catch {
        /* ignore malformed */
      }
    };

    socket.onclose = () => {
      handlers.onStatus?.(false);
      if (closed) return;
      const delay = Math.min(8000, 500 * 2 ** retry);
      retry += 1;
      timer = setTimeout(connect, delay);
    };

    socket.onerror = () => {
      socket?.close();
    };
  };

  connect();

  return () => {
    closed = true;
    if (timer) clearTimeout(timer);
    socket?.close();
  };
}
