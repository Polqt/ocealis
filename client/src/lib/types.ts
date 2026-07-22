export type BottleStatus = "drifting" | "discovered" | "released" | "scheduled";

export type EventType = "released" | "drift" | "discovered" | "re_released";

export interface Bottle {
  id: number;
  sender_id: number;
  message_text: string;
  bottle_style: number;
  start_lat: number;
  start_lng: number;
  current_lat: number;
  current_lng: number;
  hops: number;
  scheduled_release: string;
  is_released: boolean;
  status: BottleStatus;
  created_at: string;
}

export interface BottleEvent {
  id: number;
  bottle_id: number;
  event_type: EventType;
  lat: number;
  lng: number;
  created_at: string;
}

export interface Journey {
  bottle: Bottle;
  events: BottleEvent[];
}

export interface User {
  id: number;
  nickname: string;
  avatar_url?: string;
  created_at: string;
}

export interface DriftPayload {
  bottle_id: number;
  lat: number;
  lng: number;
  hops: number;
  bottle_style: number;
  timestamp: string;
}

export interface WsEnvelope {
  type: "bottle_drift" | "bottle_discovered" | "bottle_released";
  payload: DriftPayload | { bottle_id: number };
}
