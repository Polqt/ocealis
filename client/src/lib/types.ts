export type CastBottleRequest = {
  nickname: string;
  message_text: string;
  turnstile_token: string;
  start_lat?: number;
  start_lng?: number;
  bottle_style?: number;
};

export type Bottle = {
  id: number;
  nickname: string;
  message_text: string;
  status: string;
  visible_at: string;
  start_lat: number;
  start_lng: number;
};

export type BottleEvent = {
  id: number;
  bottle_id: number;
  event_type: string;
  lat: number;
  lng: number;
  created_at: string;
};

export type Journey = {
  bottle: Bottle;
  events: BottleEvent[];
};

export type MapBrowseQuery = {
  min_lat: number;
  max_lat: number;
  min_lng: number;
  max_lng: number;
  zoom: number;
};

export type MapBrowseResult = {
  mode: "heat" | "corks";
  heat?: { lat: number; lng: number; count: number }[];
  corks?: { id: number; lat: number; lng: number; is_seed?: boolean }[];
};
