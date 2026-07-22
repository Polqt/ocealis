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
