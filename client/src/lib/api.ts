import type {
  Bottle,
  CastBottleRequest,
  Journey,
  MapBrowseQuery,
  MapBrowseResult,
} from "./types";

const API_BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

export type { MapBrowseResult };

export async function castBottle(body: CastBottleRequest): Promise<Bottle> {
  const res = await fetch(`${API_BASE}/api/v1/bottles`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error ?? `cast failed (${res.status})`);
  }
  return res.json();
}

export async function browseMap(q: MapBrowseQuery): Promise<MapBrowseResult> {
  const params = new URLSearchParams({
    min_lat: String(q.min_lat),
    max_lat: String(q.max_lat),
    min_lng: String(q.min_lng),
    max_lng: String(q.max_lng),
    zoom: String(q.zoom),
  });
  const res = await fetch(`${API_BASE}/api/v1/discovery/map?${params}`);
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error ?? `map query failed (${res.status})`);
  }
  return res.json();
}

/** Open — read Message + Nickname; does not claim or remove. */
export async function openBottle(id: number): Promise<Bottle> {
  const res = await fetch(`${API_BASE}/api/v1/bottles/${id}`);
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error ?? `open failed (${res.status})`);
  }
  return res.json();
}

export async function getJourney(id: number): Promise<Journey> {
  const res = await fetch(`${API_BASE}/api/v1/bottles/${id}/journey`);
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error ?? `journey failed (${res.status})`);
  }
  return res.json();
}
