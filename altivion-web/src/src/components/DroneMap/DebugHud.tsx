"use client";

export default function DebugHud({
  show, wsOk, wsMsgs, markerCount, sns,
}: {
  show: boolean;
  wsOk: boolean;
  wsMsgs: number;
  markerCount: number;
  sns: string[];
}) {
  if (!show) return null;
  return (
    <div className="absolute right-2 top-10 z-[60] rounded-md bg-black/70 px-3 py-2 text-xs text-white shadow pointer-events-none">
      <div>WS: {wsOk ? "connected" : "offline"} | msgs: {wsMsgs}</div>
      <div>markers: {markerCount}</div>
      <div>sn: {sns.join(", ") || "-"}</div>
    </div>
  );
}
