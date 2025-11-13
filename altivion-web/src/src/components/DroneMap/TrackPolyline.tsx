"use client";

import { Polyline } from "@react-google-maps/api";

export default function TrackPolyline(props: {
  sn: string;
  points: { lat: number; lon: number }[];
}) {
  const { sn, points } = props;
  if (!points || points.length < 2) return null;
  return (
    <Polyline
      key={`track-${sn}`}
      path={points.map((p) => ({ lat: p.lat, lng: p.lon }))}
      options={{
        geodesic: true,
        strokeColor: "#8B0000",
        strokeOpacity: 0.95,
        strokeWeight: 3,
      }}
    />
  );
}
