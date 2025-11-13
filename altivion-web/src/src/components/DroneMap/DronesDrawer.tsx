"use client";

import { useMemo } from "react";
import type { DronePoint } from "@/types/drone";

const safeTime = (ts?: string) => {
  if (!ts) return "-";
  const n = Date.parse(ts);
  if (Number.isNaN(n)) return "-";
  try {
    return new Date(n).toLocaleTimeString();
  } catch {
    return "-";
  }
};

type Props = {
  isOpen: boolean;
  onClose: () => void;

  drones: DronePoint[];
  tracks: Record<string, { ts: string; lat: number; lon: number }[]>;

  selectedSn: string | null;
  setSelectedSn: (sn: string | null) => void;

  focusMode: boolean;
  setFocusMode: (v: boolean) => void;

  // Replay wiring
  replayOpen: boolean;
  setReplayOpen: (v: boolean) => void;
  replayPlaying: boolean;
  setReplayPlaying: (v: boolean) => void;
  replaySpeed: number;
  setReplaySpeed: (n: number) => void;
  replayFrom: string;
  setReplayFrom: (iso: string) => void;
  replayTo: string;
  setReplayTo: (iso: string) => void;
  frames: string[];
  frameIndex: number;
  setFrameIndex: (i: number) => void;
  loadWindow: (fromISO: string, toISO: string, sn?: string | null) => Promise<void>;
};

export default function DronesDrawer(props: Props) {
  const {
    isOpen,
    onClose,
    drones,
    tracks,
    selectedSn,
    setSelectedSn,
    focusMode,
    setFocusMode,
    replayOpen,
    setReplayOpen,
    replayPlaying,
    setReplayPlaying,
    replaySpeed,
    setReplaySpeed,
    replayFrom,
    replayTo,
    frames,
    frameIndex,
    setFrameIndex,
    loadWindow,
  } = props;

  const sorted = useMemo(() => [...drones].sort((a, b) => a.sn.localeCompare(b.sn)), [drones]);
  const hasFrames = frames.length > 0;

  return (
    <div
      className={`fixed right-0 top-0 z-[80] h-full w-[360px] transform bg-white shadow-xl transition-transform ${
        isOpen ? "translate-x-0" : "translate-x-full"
      }`}
    >
      <div className="flex items-center justify-between border-b px-4 py-3">
        <h3 className="text-sm font-semibold text-slate-900">Detected Drones</h3>
        <button onClick={onClose} className="text-xs text-slate-600 hover:text-slate-900">
          Close
        </button>
      </div>

      <div className="space-y-4 overflow-y-auto px-4 py-3 text-slate-800">
        {/* Focus toggle */}
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={focusMode && !!selectedSn} onChange={(e) => setFocusMode(e.target.checked)} />
          Show only selected drone on map
          {focusMode && selectedSn && (
            <button className="ml-2 text-xs text-blue-600 underline" onClick={() => setSelectedSn(null)}>
              Clear selection
            </button>
          )}
        </label>

        {/* Drone list */}
        <div className="space-y-2">
          {sorted.map((d) => {
            const active = d.sn === selectedSn;
            const pts = tracks[d.sn]?.length || 0;
            return (
              <button
                key={d.sn}
                onClick={() => setSelectedSn(active ? null : d.sn)}
                className={`w-full rounded-lg border px-3 py-2 text-left hover:bg-gray-50 ${
                  active ? "border-blue-600 bg-slate-900 text-white" : "border-slate-200 bg-white text-slate-900"
                }`}
              >
                <div className="flex items-center justify-between">
                  <div className="font-semibold text-xs">{d.sn}</div>
                  <div className={`text-[10px] ${active ? "opacity-80" : "text-slate-500"}`}>{safeTime(d.ts)}</div>
                </div>
                <div className={`mt-1 grid grid-cols-2 gap-x-2 text-[11px] ${active ? "opacity-90" : "text-slate-600"}`}>
                  <div>Lat/Lon</div>
                  <div className="text-right">
                    {d.lat.toFixed(5)}, {d.lon.toFixed(5)}
                  </div>
                  <div>Height</div>
                  <div className="text-right">{d.height_m ?? 0} m</div>
                  <div>Speed</div>
                  <div className="text-right">{(d.speed_h_mps ?? 0).toFixed(2)} m/s</div>
                  <div>Direction</div>
                  <div className="text-right">{d.direction_deg ?? 0}°</div>
                  <div>Track pts</div>
                  <div className="text-right">{pts}</div>
                </div>
              </button>
            );
          })}
        </div>

        {/* ---------------- REPLAY SECTION ---------------- */}
        <div className="mt-4 rounded-lg border border-slate-200 p-3">
          <div className="mb-2 flex items-center justify-between">
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={replayOpen}
                onChange={(e) => {
                  setReplayOpen(e.target.checked);
                }}
              />
              Replay mode
            </label>
            {replayOpen && (
              <button className="text-xs text-blue-600 underline" onClick={() => loadWindow(replayFrom, replayTo, selectedSn)}>
                Reload window
              </button>
            )}
          </div>

          {replayOpen && (
            <div className="space-y-3">
              {/* read-only window */}
              <div className="rounded bg-slate-50 p-2 text-[11px] text-slate-700">
                <div className="flex justify-between">
                  <span className="font-medium">From</span>
                  <span>{new Date(replayFrom).toLocaleString()}</span>
                </div>
                <div className="flex justify-between">
                  <span className="font-medium">To</span>
                  <span>{new Date(replayTo).toLocaleString()}</span>
                </div>
              </div>

              {/* transport */}
              <div className="flex items-center gap-2">
                <button
                  className={`rounded px-2 py-1 text-xs ${
                    hasFrames ? "bg-slate-900 text-white hover:bg-slate-800" : "bg-slate-200 text-slate-500 cursor-not-allowed"
                  }`}
                  disabled={!hasFrames}
                  onClick={() => setReplayPlaying(!replayPlaying)}
                >
                  {replayPlaying ? "Pause" : "Play"}
                </button>

                <label className="ml-1 flex items-center gap-2 text-xs">
                  Speed
                  <input
                    type="number"
                    min={1}
                    max={30}
                    step={1}
                    value={replaySpeed}
                    onChange={(e) => setReplaySpeed(Math.max(1, Number(e.target.value) || 1))}
                    className="w-16 rounded border border-slate-300 px-1 py-0.5 text-xs"
                  />
                  fps
                </label>

                <span className="ml-auto text-[11px] text-slate-600">frames: {frames.length}</span>
              </div>

              {/* scrubber */}
              <div>
                <input
                  type="range"
                  min={0}
                  max={Math.max(0, frames.length - 1)}
                  value={Math.min(frameIndex, Math.max(0, frames.length - 1))}
                  onChange={(e) => setFrameIndex(Number(e.target.value))}
                  className="w-full"
                  disabled={!hasFrames}
                />
                <div className="mt-1 flex justify-between text-[10px] text-slate-600">
                  <span>{frames[0] ? new Date(frames[0]).toLocaleTimeString() : "-"}</span>
                  <span>
                    {frames.length ? new Date(frames[Math.min(frameIndex, frames.length - 1)]).toLocaleTimeString() : "-"}
                  </span>
                  <span>{frames.length ? new Date(frames[frames.length - 1]).toLocaleTimeString() : "-"}</span>
                </div>
              </div>

              {/* scope */}
              <div className="text-[11px] text-slate-700">Scope: {selectedSn ? `Drone ${selectedSn}` : "All drones"}</div>
            </div>
          )}
        </div>
        {/* -------------- END REPLAY SECTION -------------- */}
      </div>
    </div>
  );
}
