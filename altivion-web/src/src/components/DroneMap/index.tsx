"use client";

import { GoogleMap, LoadScriptNext, InfoWindow } from "@react-google-maps/api";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import StatsPanel from "@/components/StatsPanel";
import DroneInfo from "./DroneInfo";
import DronesDrawer from "./DronesDrawer";
import MarkersLayer from "./MarkersLayer";
import TracksLayer from "./TracksLayer";
import DebugHud from "./DebugHud";
import useRealtimeDrones from "@/hooks/useRealtimeDrones";
import type { DronePoint } from "@/types/drone";

type Track = { sn: string; points: { ts: string; lat: number; lon: number }[] };

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://127.0.0.1:8000";
const containerStyle = { width: "100%", height: "80vh" };

export default function DroneMap() {
  const [center] = useState({ lat: 49.7, lng: -112.8 });
  const [zoom] = useState(12);

  const { markers, tracks, wsOk, wsMsgs, latestErr, tracksErr } = useRealtimeDrones();

  // selection + drawer + focus
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [selectedSn, setSelectedSn] = useState<string | null>(null);
  const [focusMode, setFocusMode] = useState(false);

  // debug toggle
  const envDefaultDebug = (process.env.NEXT_PUBLIC_DEBUG || "0") === "1";
  const [showDebug, setShowDebug] = useState(envDefaultDebug);

  // map ref
  const mapRef = useRef<google.maps.Map | null>(null);
  const onMapLoad = useCallback((map: google.maps.Map) => {
    mapRef.current = map;
    map.setOptions({
      streetViewControl: false,
      fullscreenControl: true,
      mapTypeControl: true,
      mapTypeControlOptions: { position: google.maps.ControlPosition.TOP_RIGHT },
      zoomControlOptions: { position: google.maps.ControlPosition.RIGHT_CENTER },
    });
    const pts = markers.map((m) => ({ lat: m.lat, lng: m.lon }));
    if (pts.length > 1) {
      const b = new google.maps.LatLngBounds();
      pts.forEach((p) => b.extend(p));
      map.fitBounds(b);
    }
  }, [markers]);

  // ------------ REPLAY MODE ------------
  const [replayOpen, setReplayOpen] = useState(false);
  const [replayPlaying, setReplayPlaying] = useState(false);
  const [replaySpeed, setReplaySpeed] = useState(4); // fps

  // default 15-minute window
  const makeWindow = () => {
    const to = new Date();
    const from = new Date(to.getTime() - 15 * 60 * 1000);
    return { fromISO: from.toISOString(), toISO: to.toISOString() };
  };
  const initial = makeWindow();
  const [replayFrom, setReplayFrom] = useState(initial.fromISO);
  const [replayTo, setReplayTo] = useState(initial.toISO);

  const [replayTracks, setReplayTracks] = useState<Track[]>([]);

  // Build frames from valid timestamps
  const replayFrames = useMemo(() => {
    const set = new Set<string>();
    for (const t of replayTracks) {
      for (const p of t.points || []) {
        if (p.ts && !Number.isNaN(Date.parse(p.ts))) set.add(p.ts);
      }
    }
    return Array.from(set).sort();
  }, [replayTracks]);

  const [frameIndex, setFrameIndex] = useState(0);

  // Fetch a time window
  const loadWindow = useCallback(
    async (fromISO: string, toISO: string, sn?: string | null) => {
      try {
        const params = new URLSearchParams({ from: fromISO, to: toISO });
        if (sn) params.set("sn", sn);
        const res = await fetch(`${API_BASE}/tracks_window?${params.toString()}`);
        const data: Track[] = await res.json();

        const clean: Track[] = (Array.isArray(data) ? data : []).map((t) => ({
          sn: t.sn,
          points: (t.points || []).filter((p) => p.ts && !Number.isNaN(Date.parse(p.ts))),
        }));

        setReplayTracks(clean);
        setFrameIndex(0);
        setReplayPlaying(false);
      } catch {
        setReplayTracks([]);
        setFrameIndex(0);
        setReplayPlaying(false);
      }
    },
    []
  );

  // Auto-load when you open Replay
  useEffect(() => {
    if (!replayOpen) return;
    const { fromISO, toISO } = makeWindow();
    setReplayFrom(fromISO);
    setReplayTo(toISO);
    loadWindow(fromISO, toISO, selectedSn).catch(() => {});
  }, [replayOpen, selectedSn, loadWindow]);

  // Also refresh window every time you change selected drone while Replay is open
  useEffect(() => {
    if (!replayOpen) return;
    loadWindow(replayFrom, replayTo, selectedSn).catch(() => {});
  }, [selectedSn]); // intentionally minimal deps

  // Advance frames when playing
  useEffect(() => {
    if (!replayOpen || !replayPlaying || replayFrames.length === 0) return;
    const iv = setInterval(() => {
      setFrameIndex((i) => (i + 1) % replayFrames.length);
    }, 1000 / Math.max(1, replaySpeed));
    return () => clearInterval(iv);
  }, [replayOpen, replayPlaying, replayFrames, replaySpeed]);

  // LIVE vs REPLAY layers
  const liveMarkersToShow = useMemo(
    () => (focusMode && selectedSn ? markers.filter((m) => m.sn === selectedSn) : markers),
    [markers, focusMode, selectedSn]
  );
  const liveTracksToShow = useMemo(() => {
    if (focusMode && selectedSn) return { [selectedSn]: tracks[selectedSn] || [] };
    return tracks;
  }, [tracks, focusMode, selectedSn]);

  const replayCutoffTs = replayFrames[frameIndex] || "";
  const replayTracksSliced = useMemo(() => {
    if (!replayOpen) return {};
    const out: Record<string, { ts: string; lat: number; lon: number }[]> = {};
    replayTracks.forEach((t) => {
      out[t.sn] = (t.points || []).filter((p) => p.ts <= replayCutoffTs);
    });
    return out;
  }, [replayOpen, replayTracks, replayCutoffTs]);

  const replayMarkersToShow = useMemo(() => {
    if (!replayOpen) return [];
    const arr: DronePoint[] = [];
    replayTracks.forEach((t) => {
      const pts = (t.points || []).filter((p) => p.ts <= replayCutoffTs);
      if (pts.length) {
        const last = pts[pts.length - 1];
        arr.push({ sn: t.sn, ts: last.ts, lat: last.lat, lon: last.lon });
      }
    });
    return arr;
  }, [replayOpen, replayTracks, replayCutoffTs]);

  // Selected info bubble (guard invalid dates)
  const selectedPoint: DronePoint | undefined = useMemo(() => {
    const pool = replayOpen ? replayMarkersToShow : liveMarkersToShow;
    return selectedSn ? pool.find((m) => m.sn === selectedSn) : undefined;
  }, [selectedSn, liveMarkersToShow, replayMarkersToShow, replayOpen]);
  const hasValidSelectedPoint =
    !!selectedPoint && !Number.isNaN(Date.parse((selectedPoint!.ts as string) || ""));

  return (
    <div className="flex w-full flex-col items-center">
      <LoadScriptNext googleMapsApiKey={process.env.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY!}>
        <div className="relative w-full">
          {/* Buttons */}
          <div className="absolute right-2 top-16 z-[70] flex gap-2">
            <button
              onClick={() => setShowDebug((s) => !s)}
              className="rounded-md bg-white px-3 py-1 text-xs font-medium text-slate-900 shadow hover:bg-slate-50 border border-slate-200"
            >
              {showDebug ? "Debug: off" : "Debug: on"}
            </button>
            <button
              onClick={() => {
                setReplayOpen((o) => !o);
                setReplayPlaying(false);
              }}
              className="rounded-md bg-white px-3 py-1 text-xs font-medium text-slate-900 shadow hover:bg-slate-50 border border-slate-200"
            >
              {replayOpen ? "Exit replay" : "Replay"}
            </button>
            <button
              onClick={() => setDrawerOpen(true)}
              className="rounded-md bg-white px-3 py-1 text-xs font-medium text-slate-900 shadow hover:bg-slate-50 border border-slate-200"
            >
              Drones
            </button>
          </div>

          <DebugHud
            show={showDebug}
            wsOk={wsOk}
            wsMsgs={wsMsgs}
            markerCount={markers.length}
            sns={markers.map((m) => m.sn)}
          />

          <div className="pointer-events-none">{typeof StatsPanel !== "undefined" && <StatsPanel />}</div>

          <GoogleMap mapContainerStyle={containerStyle} center={center} zoom={zoom} onLoad={onMapLoad} mapTypeId="satellite">
            {replayOpen ? (
              <>
                <TracksLayer tracks={replayTracksSliced} />
                <MarkersLayer
                  markers={replayMarkersToShow}
                  onSelect={(sn) => {
                    setSelectedSn(sn);
                    const p = replayMarkersToShow.find((m) => m.sn === sn);
                    if (p && mapRef.current) mapRef.current.panTo({ lat: p.lat, lng: p.lon });
                  }}
                />
              </>
            ) : (
              <>
                <TracksLayer tracks={liveTracksToShow} />
                <MarkersLayer
                  markers={liveMarkersToShow}
                  onSelect={(sn) => {
                    setSelectedSn(sn);
                    const p = liveMarkersToShow.find((m) => m.sn === sn);
                    if (p && mapRef.current) mapRef.current.panTo({ lat: p.lat, lng: p.lon });
                  }}
                />
              </>
            )}

            {hasValidSelectedPoint && selectedPoint && (
              <InfoWindow position={{ lat: selectedPoint.lat, lng: selectedPoint.lon }} onCloseClick={() => setSelectedSn(null)}>
                <div className="min-w-[220px]">
                  <DroneInfo {...selectedPoint} />
                </div>
              </InfoWindow>
            )}
          </GoogleMap>

          <DronesDrawer
            isOpen={drawerOpen}
            onClose={() => setDrawerOpen(false)}
            drones={markers}
            tracks={tracks}
            selectedSn={selectedSn}
            setSelectedSn={(sn) => {
              setSelectedSn(sn);
              if (sn) {
                const pool = replayOpen ? replayMarkersToShow : liveMarkersToShow;
                const p = pool.find((m) => m.sn === sn);
                if (p && mapRef.current) mapRef.current.panTo({ lat: p.lat, lng: p.lon });
              }
            }}
            focusMode={focusMode}
            setFocusMode={setFocusMode}
            // replay wiring
            replayOpen={replayOpen}
            setReplayOpen={(v) => {
              setReplayOpen(v);
              if (v) {
                const { fromISO, toISO } = makeWindow();
                setReplayFrom(fromISO);
                setReplayTo(toISO);
                loadWindow(fromISO, toISO, selectedSn).catch(() => {});
              } else {
                setReplayPlaying(false);
              }
            }}
            replayPlaying={replayPlaying}
            setReplayPlaying={setReplayPlaying}
            replaySpeed={replaySpeed}
            setReplaySpeed={setReplaySpeed}
            replayFrom={replayFrom}
            setReplayFrom={setReplayFrom}
            replayTo={replayTo}
            setReplayTo={setReplayTo}
            frames={replayFrames}
            frameIndex={frameIndex}
            setFrameIndex={setFrameIndex}
            loadWindow={loadWindow}
          />
        </div>
      </LoadScriptNext>

      <div className="mt-2 text-center text-xs text-gray-500">
        {(latestErr || tracksErr) && "API error (check console)."}
      </div>
    </div>
  );
}
