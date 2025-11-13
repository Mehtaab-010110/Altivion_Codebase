type Props = {
  sn: string;
  ts: string;
  lat: number;
  lon: number;
  height_m?: number;
  speed_h_mps?: number;
  direction_deg?: number;
};

export default function DroneInfo({
  sn, ts, lat, lon, height_m, speed_h_mps, direction_deg,
}: Props) {
  const fmt = (n?: number, d = 2) =>
    typeof n === "number" ? n.toFixed(d) : "—";

  const speedKmh =
    typeof speed_h_mps === "number" ? speed_h_mps * 3.6 : undefined;

  return (
    <div className="text-[13px] text-slate-800 leading-[1.15]">
      <div className="mb-2 font-semibold text-slate-900">{sn}</div>

      <div className="grid grid-cols-[110px_1fr] gap-y-1 gap-x-2">
        <div className="text-slate-500">Last seen</div>
        <div className="font-medium">
          {new Date(ts).toLocaleString()}
        </div>

        <div className="text-slate-500">Lat / Lon</div>
        <div className="font-medium tabular-nums">
          {fmt(lat, 5)}, {fmt(lon, 5)}
        </div>

        <div className="text-slate-500">Height</div>
        <div className="font-medium">{fmt(height_m, 1)} m</div>

        <div className="text-slate-500">Speed (H)</div>
        <div className="font-medium">
          {typeof speed_h_mps === "number"
            ? `${fmt(speed_h_mps, 2)} m/s  (${fmt(speedKmh, 1)} km/h)`
            : "—"}
        </div>

        <div className="text-slate-500">Direction</div>
        <div className="font-medium">
          {typeof direction_deg === "number" ? `${direction_deg}°` : "—"}
        </div>
      </div>
    </div>
  );
}
