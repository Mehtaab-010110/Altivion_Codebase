"use client";

import { useEffect, useMemo, useState } from "react";
import useSWR, { mutate } from "swr";
import type { DronePoint, Track, TracksBySn } from "@/types/drone";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://127.0.0.1:8000";
const WS_URL   = API_BASE.replace("https", "wss").replace("http", "ws") + "/ws";
const fetcher  = (u: string) => fetch(u).then(r => r.json());

export default function useRealtimeDrones() {
  // baseline
  const { data: latestBase, error: latestErr } = useSWR<DronePoint[]>(
    `${API_BASE}/latest?minutes=120`, fetcher,
    { refreshInterval: 5000, revalidateOnFocus: false, revalidateOnReconnect: false }
  );

  const { data: tracksBase, error: tracksErr } = useSWR<Track[]>(
    `${API_BASE}/tracks?minutes=120&max_points=4000`, fetcher,
    { refreshInterval: 5000, revalidateOnFocus: false, revalidateOnReconnect: false }
  );

  // live
  const [wsOk, setWsOk] = useState(false);
  const [wsMsgs, setWsMsgs] = useState(0);
  const [liveLatest, setLiveLatest] = useState<Record<string, DronePoint>>({});
  const [liveTracks, setLiveTracks] = useState<TracksBySn>({});

  useEffect(() => {
    const ws = new WebSocket(WS_URL);
    ws.onopen = () => setWsOk(true);
    ws.onclose = () => setWsOk(false);
    ws.onerror = () => setWsOk(false);
    ws.onmessage = (ev) => {
      try {
        const p = JSON.parse(ev.data);
        if (!p?.sn || typeof p.lat !== "number" || typeof p.lon !== "number") return;
        setWsMsgs(c => c + 1);

        const dp: DronePoint = {
          sn: p.sn, ts: p.ts || new Date().toISOString(),
          lat: p.lat, lon: p.lon,
          height_m: p.height_m, direction_deg: p.direction_deg, speed_h_mps: p.speed_h_mps,
        };

        setLiveLatest(prev => ({ ...prev, [dp.sn]: dp }));
        setLiveTracks(prev => {
          const arr = prev[dp.sn] ? [...prev[dp.sn]] : [];
          arr.push({ ts: dp.ts, lat: dp.lat, lon: dp.lon });
          if (arr.length > 4000) arr.splice(0, arr.length - 4000);
          return { ...prev, [dp.sn]: arr };
        });

        // nudge SWR-driven bits
        mutate(`${API_BASE}/latest?minutes=120`);
        mutate(`${API_BASE}/tracks?minutes=120&max_points=4000`);
      } catch {}
    };
    return () => ws.close();
  }, []);

  // fallback polling only if ws down
  useEffect(() => {
    if (wsOk) return;
    const id = setInterval(async () => {
      try {
        const res = await fetch(`${API_BASE}/latest?minutes=2`);
        const ar: DronePoint[] = await res.json();
        const map: Record<string, DronePoint> = {};
        ar.forEach(d => (map[d.sn] = d));
        setLiveLatest(map);
      } catch {}
    }, 1500);
    return () => clearInterval(id);
  }, [wsOk]);

  // merge for markers
  const markers: DronePoint[] = useMemo(() => {
    const base: Record<string, DronePoint> = {};
    (latestBase || []).forEach(d => {
      if (typeof d.lat === "number" && typeof d.lon === "number") base[d.sn] = d;
    });
    for (const [sn, dp] of Object.entries(liveLatest)) base[sn] = dp;
    return Object.values(base);
  }, [latestBase, liveLatest]);

  // merge for tracks
  const tracks: TracksBySn = useMemo(() => {
    const out: TracksBySn = {};
    (tracksBase || []).forEach(t => (out[t.sn] = t.points.slice()));
    for (const [sn, pts] of Object.entries(liveTracks)) {
      const a = out[sn] ? out[sn] : [];
      const lastTs = a.length ? a[a.length - 1].ts : "";
      const extras = lastTs ? pts.filter(p => p.ts > lastTs) : pts;
      out[sn] = [...a, ...extras];
      if (out[sn].length > 4000) out[sn].splice(0, out[sn].length - 4000);
    }
    return out;
  }, [tracksBase, liveTracks]);

  return {
    markers, tracks,
    wsOk, wsMsgs,
    latestErr, tracksErr,
  };
}
