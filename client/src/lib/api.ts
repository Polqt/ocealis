import type { Bottle, CastBottleRequest } from "./types";

const API_BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

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
