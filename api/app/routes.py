from typing import Any, Dict, List, Union, Optional
from datetime import datetime, timezone, timedelta
from fastapi import APIRouter, Depends, Header, HTTPException, Query
from starlette.concurrency import run_in_threadpool

from app.config import API_KEY, NODE_ONLINE_WINDOW_SEC, NODES_TOTAL
from app.models import Signal
from app.db import _insert_many, _fetch_all
from app.ws import manager
import json

router = APIRouter()

# Security
def require_api_key(x_api_key: str = Header(default="")):
    if not API_KEY or x_api_key != API_KEY:
        raise HTTPException(status_code=401, detail="Invalid or missing API key.")

# Runtime node heartbeat state (kept in-memory)
STARTED_AT = datetime.now(timezone.utc)
NODE_HEARTBEATS: Dict[str, datetime] = {}

@router.get("/health")
def health():
    return {"status": "ok"}

@router.post("/ingest", dependencies=[Depends(require_api_key)])
async def ingest(payload: Union[Signal, List[Signal]]):
    rows = payload if isinstance(payload, list) else [payload]
    if not rows:
        raise HTTPException(status_code=400, detail="Empty payload.")

    # Write to DB off the event loop
    await run_in_threadpool(_insert_many, rows)

    # Realtime push to browsers
    for r in rows:
        msg = {
            "sn": r.sn,
            "ts": (r.ts.isoformat() if hasattr(r.ts, "isoformat") else str(r.ts)),
            "lat": r.lat,
            "lon": r.lon,
            "height_m": r.height_m,
            "speed_h_mps": r.speed_h_mps,
            "direction_deg": r.direction_deg,
        }
        await manager.broadcast(json.dumps(msg))

    return {"inserted": len(rows)}

@router.get("/latest")
def latest(minutes: int = 10):
    q = """
    SELECT DISTINCT ON (sn)
      sn, ts, lat, lon, height_m, speed_h_mps, direction_deg
    FROM public.drone_signals
    WHERE ts > now() - (%s)::interval
      AND lat IS NOT NULL AND lon IS NOT NULL
    ORDER BY sn, ts DESC;
    """
    return _fetch_all(q, (f"{minutes} minutes",))

@router.get("/latest_in_view")
def latest_in_view(swLat: float, swLng: float, neLat: float, neLng: float, minutes: int = 10):
    loLat, hiLat = min(swLat, neLat), max(swLat, neLat)
    loLng, hiLng = min(swLng, neLng), max(swLng, neLng)
    q = """
    WITH latest AS (
      SELECT DISTINCT ON (sn)
        sn, ts, lat, lon, height_m, speed_h_mps, direction_deg
      FROM public.drone_signals
      WHERE ts > now() - (%s)::interval
        AND lat IS NOT NULL AND lon IS NOT NULL
      ORDER BY sn, ts DESC
    )
    SELECT * FROM latest
    WHERE lat BETWEEN %s AND %s
      AND lon BETWEEN %s AND %s;
    """
    return _fetch_all(q, (f"{minutes} minutes", loLat, hiLat, loLng, hiLng))

@router.get("/tracks")
def tracks(minutes: int = 60, max_points: int = 1000):
    q = """
    SELECT sn, ts, lat, lon
    FROM public.drone_signals
    WHERE ts > now() - (%s)::interval
      AND lat IS NOT NULL AND lon IS NOT NULL
    ORDER BY sn, ts ASC
    LIMIT %s;
    """
    rows = _fetch_all(q, (f"{minutes} minutes", max_points))
    grouped: Dict[str, List[Dict[str, Any]]] = {}
    for r in rows:
        grouped.setdefault(r["sn"], []).append(
            {
                "ts": r["ts"].isoformat() if hasattr(r["ts"], "isoformat") else r["ts"],
                "lat": r["lat"],
                "lon": r["lon"],
            }
        )
    return [{"sn": sn, "points": pts} for sn, pts in grouped.items()]


# 🆕 NEW: Time-window query for replay
@router.get("/tracks_window")
def tracks_window(
    from_ts: str = Query(..., alias="from"),
    to_ts: str = Query(..., alias="to"),
    sn: Optional[str] = None,
    max_points: int = 20000,
):
    """
    Return tracks grouped by SN between [from, to] (ISO 8601 timestamps).
    Optional ?sn= filters to a single drone.
    """
    q = f"""
    SELECT sn, ts, lat, lon
    FROM public.drone_signals
    WHERE ts BETWEEN %s AND %s
      AND lat IS NOT NULL AND lon IS NOT NULL
      {"AND sn = %s" if sn else ""}
    ORDER BY sn, ts ASC
    LIMIT %s;
    """
    params = (from_ts, to_ts, sn, max_points) if sn else (from_ts, to_ts, max_points)
    rows = _fetch_all(q, params)

    grouped: Dict[str, List[Dict[str, Any]]] = {}
    for r in rows:
        grouped.setdefault(r["sn"], []).append({
            "ts": r["ts"].isoformat() if hasattr(r["ts"], "isoformat") else r["ts"],
            "lat": r["lat"], "lon": r["lon"],
        })
    return [{"sn": sn_, "points": pts} for sn_, pts in grouped.items()]


@router.post("/node_heartbeat")
def node_heartbeat(node_id: str):
    if not node_id:
        raise HTTPException(status_code=400, detail="Missing node_id")
    NODE_HEARTBEATS[node_id] = datetime.now(timezone.utc)
    return {"ok": True, "node_id": node_id, "last_seen": NODE_HEARTBEATS[node_id].isoformat()}

@router.get("/data")
def data(minutes_online_window: int = 2):
    now = datetime.now(timezone.utc)
    uptime_seconds = int((now - STARTED_AT).total_seconds())

    q_today = """
        SELECT COUNT(DISTINCT uasid) AS count
        FROM public.drone_signals
        WHERE uasid IS NOT NULL
          AND ts >= date_trunc('day', now());
    """
    today_unique = _fetch_all(q_today)[0]["count"]

    q_online = """
        SELECT COUNT(DISTINCT sn) AS c
        FROM public.drone_signals
        WHERE ts > now() - (%s)::interval;
    """
    online_drone_count = _fetch_all(q_online, (f"{minutes_online_window} minutes",))[0]["c"]

    active_nodes = 0
    cutoff = now - timedelta(seconds=NODE_ONLINE_WINDOW_SEC)
    for last_seen in NODE_HEARTBEATS.values():
        if last_seen >= cutoff:
            active_nodes += 1

    return {
        "uptime_seconds": uptime_seconds,
        "today_unique_uasids": today_unique,
        "online_drones": online_drone_count,
        "nodes_active": active_nodes,
        "nodes_total": NODES_TOTAL,
        "as_of": now.isoformat(),
    }
