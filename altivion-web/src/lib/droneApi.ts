export const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://127.0.0.1:8000";
export const WS_URL   = API_BASE.replace("https", "wss").replace("http", "ws") + "/ws";
export const fetcher  = (url: string) => fetch(url).then((r) => r.json());

// Common SWR keys
export const LATEST_URL = `${API_BASE}/latest?minutes=120`;
export const TRACKS_URL = `${API_BASE}/tracks?minutes=120&max_points=4000`;
