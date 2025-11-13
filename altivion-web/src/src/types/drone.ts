export type DronePoint = {
    sn: string;
    ts: string;
    lat: number;
    lon: number;
    height_m?: number;
    speed_h_mps?: number;
    direction_deg?: number;
  };
  
  export type Track = { sn: string; points: { ts: string; lat: number; lon: number }[] };
  export type TracksBySn = Record<string, { ts: string; lat: number; lon: number }[]>;
  