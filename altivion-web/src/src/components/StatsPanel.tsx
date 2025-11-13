"use client";
import { useEffect, useMemo, useRef, useState } from "react";
import useSWR from "swr";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://127.0.0.1:8000";
const WS_URL   = "ws://127.0.0.1:8000/ws";
const fetcher  = (url: string) => fetch(url).then(r => r.json());

export default function StatsPanel() {
  const [uptime, setUptime] = useState(0);
  const [todayUnique, setTodayUnique] = useState(0);

  // live counters from WS
  const lastSeen = useRef<Map<string, number>>(new Map()); // sn -> epoch ms
  const [onlineCount, setOnlineCount] = useState(0);
  const [nodesActive, setNodesActive] = useState(0);
  const [nodesTotal, setNodesTotal]   = useState(3);
  const [wsOk, setWsOk] = useState(false);

  // One baseline load (no fast polling)
  const { data: base } = useSWR(`${API_BASE}/data`, fetcher, {
    refreshInterval: 30000, // occasional correction
    revalidateOnFocus: false,
  });

  // apply baseline when it changes
  useEffect(() => {
    if (!base) return;
    setUptime(base.uptime_seconds ?? 0);
    setTodayUnique(base.today_unique_uasids ?? 0);
    setNodesActive(base.nodes_active ?? 0);
    setNodesTotal(base.nodes_total ?? 3);
  }, [base]);

  // tick uptime every second for a “live” feel
  useEffect(() => {
    const id = setInterval(() => setUptime(u => u + 1), 1000);
    return () => clearInterval(id);
  }, []);

  // subscribe to WS; update online drones instantly
  useEffect(() => {
    const ws = new WebSocket(WS_URL);
    ws.onopen = () => setWsOk(true);
    ws.onclose = () => setWsOk(false);
    ws.onmessage = (ev) => {
      try {
        const p = JSON.parse(ev.data) as { sn?: string; ts?: string; lat?: number; lon?: number };
        if (!p?.sn) return;
        lastSeen.current.set(p.sn, Date.now());
      } catch {}
    };
    return () => ws.close();
  }, []);

  // compute online drones from “seen in last N sec”
  useEffect(() => {
    const WINDOW_MS = 15_000; // consider online if seen in last 15s
    const id = setInterval(() => {
      const now = Date.now();
      let c = 0;
      for (const t of lastSeen.current.values()) {
        if (now - t <= WINDOW_MS) c++;
      }
      setOnlineCount(c);
    }, 1000);
    return () => clearInterval(id);
  }, []);

  const uptimeText = useMemo(() => {
    const h = Math.floor(uptime / 3600);
    const m = Math.floor((uptime % 3600) / 60);
    const s = uptime % 60;
    return `${h}h ${m}m ${s}s`;
  }, [uptime]);

  return (
    <div className="absolute top-4 left-4 z-50 rounded-2xl bg-black/70 text-white px-4 py-3 shadow-xl backdrop-blur">
      <div className="text-sm font-semibold mb-1">
        System Stats {wsOk && <span className="ml-2 text-green-400">LIVE</span>}
      </div>
      <div className="text-xs text-gray-300">Uptime: {uptimeText}</div>
      <div className="text-xs text-gray-300">Today’s detections: {todayUnique}</div>
      <div className="text-xs text-gray-300">Online drones: {onlineCount}</div>
      <div className="text-xs text-gray-300">Nodes: {nodesActive}/{nodesTotal} Online</div>
    </div>
  );
}
