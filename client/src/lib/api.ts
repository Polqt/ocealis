import type { Bottle, Journey, User } from "./types";

const API_URL = (import.meta.env.VITE_API_URL as string | undefined) ?? "http://localhost:8080";

const TOKEN_KEY = "ocealis_token";
const USER_KEY = "ocealis_user";

export function getToken(): string | null {
  if (typeof localStorage === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function getStoredUser(): User | null {
  if (typeof localStorage === "undefined") return null;
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as User;
  } catch {
    return null;
  }
}

function storeSession(token: string, user: User) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Content-Type", "application/json");
  const token = getToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);

  const res = await fetch(`${API_URL}${path}`, { ...init, headers });
  if (!res.ok) {
    let message = res.statusText;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      /* ignore */
    }
    throw new Error(message);
  }
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

function randomNickname(): string {
  const n = Math.floor(1000 + Math.random() * 9000);
  return `drifter${n}`;
}

export async function ensureAnonSession(): Promise<User> {
  const existing = getStoredUser();
  const token = getToken();
  if (existing && token) {
    try {
      return await request<User>("/api/v1/users/profile");
    } catch {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
    }
  }

  const created = await request<{ user: User; token: string }>("/api/v1/users", {
    method: "POST",
    body: JSON.stringify({ nickname: randomNickname() })
  });
  storeSession(created.token, created.user);
  return created.user;
}

export function listOceanBottles(limit = 100): Promise<{ data: Bottle[] }> {
  return request(`/api/v1/ocean/bottles?limit=${limit}`);
}

export function createBottle(input: {
  message_text: string;
  bottle_style: number;
  start_lat: number;
  start_lng: number;
}): Promise<Bottle> {
  return request("/api/v1/bottles", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function getJourney(id: number): Promise<Journey> {
  return request(`/api/v1/bottles/${id}/journey`);
}

export function discoverBottle(id: number, user_lat: number, user_lng: number): Promise<Journey> {
  return request(`/api/v1/bottles/${id}/discover`, {
    method: "POST",
    body: JSON.stringify({ user_lat, user_lng })
  });
}

export function releaseBottle(id: number, lat: number, lng: number): Promise<Bottle> {
  return request(`/api/v1/bottles/${id}/release`, {
    method: "POST",
    body: JSON.stringify({ lat, lng })
  });
}

export function apiBase(): string {
  return API_URL;
}

export function wsUrl(): string {
  const base = API_URL.replace(/^http/, "ws");
  return `${base}/ws`;
}
