"use client";
import { useEffect } from "react";
import type { Dispatch, SetStateAction } from "react";

export type ReplayControlsProps = {
  // visibility
  isOpen?: boolean;
  onClose?: () => void;

  // playback
  playing: boolean;
  setPlaying: Dispatch<SetStateAction<boolean>>;

  // frames & index
  frames: any[];
  frameIndex: number;
  setFrameIndex: Dispatch<SetStateAction<number>>;

  // speed (fps)
  speed: number;
  setSpeed: Dispatch<SetStateAction<number>>;

  // fetch frames for a given window (ISO)
  loadWindow: (fromISO: string, toISO: string, sn?: string) => Promise<void>;

  // focused drone (optional)
  selectedSn?: string | null;

  // OPTIONAL: explicit time window controls (ISO strings)
  windowFrom?: string;
  windowTo?: string;
  setWindowFrom?: Dispatch<SetStateAction<string>>;
  setWindowTo?: Dispatch<SetStateAction<string>>;
};

export default function ReplayControls({
  isOpen = true,
  onClose,
  playing,
  setPlaying,
  frames,
  frameIndex,
  setFrameIndex,
  speed,
  setSpeed,
  loadWindow,
  selectedSn,
  windowFrom,
  windowTo,
  setWindowFrom,
  setWindowTo,
}: ReplayControlsProps) {
  if (!isOpen) return null;

  // auto-advance while playing
  useEffect(() => {
    if (!playing || frames.length === 0) return;
    const iv = setInterval(() => {
      setFrameIndex(i => (i + 1) % frames.length);
    }, 1000 / Math.max(1, speed));
    return () => clearInterval(iv);
  }, [playing, frames, speed, setFrameIndex]);

  const hasFrames = frames.length > 0;

  const applyWindow = async () => {
    if (!windowFrom || !windowTo) return;
    await loadWindow(windowFrom, windowTo, selectedSn ?? undefined);
    setFrameIndex(0);
  };

  return (
    <div className="absolute right-2 top-2 z-[65] flex flex-wrap items-center gap-2 rounded-md bg-white/90 px-2 py-1 text-xs shadow">
      {onClose && (
        <button
          className="mr-1 rounded bg-white px-2 py-1 shadow hover:bg-gray-50"
          onClick={onClose}
          title="Close replay"
        >
          ✕
        </button>
      )}

      <button
        className="rounded bg-black/80 px-2 py-1 text-white hover:bg-black disabled:opacity-40"
        onClick={() => setPlaying(p => !p)}
        disabled={!hasFrames}
        title={playing ? "Pause" : "Play"}
      >
        {playing ? "Pause" : "Play"}
      </button>

      <button
        className="rounded bg-white px-2 py-1 shadow hover:bg-gray-50 disabled:opacity-40"
        onClick={() => hasFrames && setFrameIndex(i => (i - 1 + frames.length) % frames.length)}
        disabled={!hasFrames}
        title="Prev frame"
      >
        ◀
      </button>

      <span className="tabular-nums">{hasFrames ? `${frameIndex + 1}/${frames.length}` : "0/0"}</span>

      <button
        className="rounded bg-white px-2 py-1 shadow hover:bg-gray-50 disabled:opacity-40"
        onClick={() => hasFrames && setFrameIndex(i => (i + 1) % frames.length)}
        disabled={!hasFrames}
        title="Next frame"
      >
        ▶
      </button>

      <label className="ml-2">
        Speed:
        <input
          type="number"
          className="ml-1 w-16 rounded border px-1 py-0.5"
          min={1}
          max={60}
          step={1}
          value={speed}
          onChange={(e) => {
            const v = Number(e.target.value);
            if (!Number.isNaN(v)) setSpeed(Math.max(1, Math.min(60, v)));
          }}
        />
        fps
      </label>

      {/* Optional window pickers if parent provides state */}
      {setWindowFrom && setWindowTo && (
        <>
          <input
            type="datetime-local"
            className="rounded border px-1 py-0.5"
            value={toLocalDatetime(windowFrom)}
            onChange={(e) => setWindowFrom(fromLocalDatetime(e.target.value))}
            title="From"
          />
          <span>→</span>
          <input
            type="datetime-local"
            className="rounded border px-1 py-0.5"
            value={toLocalDatetime(windowTo)}
            onChange={(e) => setWindowTo(fromLocalDatetime(e.target.value))}
            title="To"
          />
          <button
            className="rounded bg-white px-2 py-1 shadow hover:bg-gray-50"
            onClick={applyWindow}
            title="Load selected window"
          >
            Load
          </button>
        </>
      )}

      <button
        className="ml-2 rounded bg-white px-2 py-1 shadow hover:bg-gray-50"
        onClick={async () => {
          const end = new Date();
          const start = new Date(end.getTime() - 5 * 60 * 1000);
          await loadWindow(start.toISOString(), end.toISOString(), selectedSn ?? undefined);
          setFrameIndex(0);
        }}
        title="Load last 5 minutes"
      >
        Last 5m
      </button>
    </div>
  );
}

/** Helpers to convert ISO ⇄ <input type="datetime-local"> */
function toLocalDatetime(iso?: string) {
  if (!iso) return "";
  const d = new Date(iso);
  // YYYY-MM-DDTHH:mm
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}
function fromLocalDatetime(local: string) {
  // interpret as local time, return ISO
  const d = new Date(local);
  return isNaN(d.valueOf()) ? "" : d.toISOString();
}
