"use client";
import { Polyline } from "@react-google-maps/api";
import type { TracksBySn } from "@/types/drone";

export default function TracksLayer({ tracks }: { tracks: TracksBySn }) {
  return (
    <>
      {Object.entries(tracks).map(([sn, pts]) => {
        if (!pts || pts.length < 2) return null;
        return (
          <Polyline
            key={`track-${sn}`}
            path={pts.map((p) => ({ lat: p.lat, lng: p.lon }))}
            options={{ geodesic: true, strokeColor: "#8B0000", strokeOpacity: 0.95, strokeWeight: 3 }}
          />
        );
      })}
    </>
  );
}
