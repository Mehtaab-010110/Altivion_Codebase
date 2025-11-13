"use client";
import { Marker } from "@react-google-maps/api";
import type { DronePoint } from "@/types/drone";

export default function MarkersLayer({
  markers,
  onSelect,
}: {
  markers: DronePoint[];
  onSelect: (sn: string) => void;
}) {
  return (
    <>
      {markers.map((p) => (
        <Marker
          key={p.sn}
          position={{ lat: p.lat, lng: p.lon }}
          title={`${p.sn} | ${new Date(p.ts).toLocaleTimeString()}`}
          label={{ text: p.sn, color: "white", fontSize: "10px", fontWeight: "bold" }}
          onClick={() => onSelect(p.sn)}
          options={{ clickable: true, zIndex: 1000 }}
        />
      ))}
    </>
  );
}
