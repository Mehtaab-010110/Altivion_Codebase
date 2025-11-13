from typing import Any, Dict, List, Tuple
from psycopg import connect
from app.config import DATABASE_URL
from app.models import Signal
import json

def _insert_many(rows: List[Signal]) -> None:
    """Insert many rows; also emit a pg_notify for the last row (best-effort)."""
    with connect(DATABASE_URL) as conn:
        with conn.cursor() as cur:
            cur.executemany(
                """
                INSERT INTO drone_signals
                (ts, sn, uasid, drone_type, direction_deg, speed_h_mps, speed_v_mps,
                 lat, lon, height_m, operator_lat, operator_lon)
                VALUES
                (%(ts)s, %(sn)s, %(uasid)s, %(drone_type)s, %(direction_deg)s,
                 %(speed_h_mps)s, %(speed_v_mps)s, %(lat)s, %(lon)s, %(height_m)s,
                 %(operator_lat)s, %(operator_lon)s)
                """,
                [r.model_dump(by_alias=False) for r in rows],
            )

            # Optional: fire a NOTIFY so LISTEN loop can also push
            try:
                last = rows[-1]
                payload = json.dumps(
                    {
                        "sn": last.sn,
                        "ts": last.ts.isoformat(),
                        "lat": last.lat,
                        "lon": last.lon,
                        "height_m": last.height_m,
                        "direction_deg": last.direction_deg,
                        "speed_h_mps": last.speed_h_mps,
                    }
                )
                cur.execute("SELECT pg_notify('signals', %s);", (payload,))
            except Exception:
                pass

        conn.commit()

def _fetch_all(q: str, params: Tuple = ()) -> List[Dict[str, Any]]:
    """Query helper for returning dict results."""
    with connect(DATABASE_URL) as conn:
        with conn.cursor() as cur:
            cur.execute(q, params)
            cols = [c[0] for c in cur.description]
            return [dict(zip(cols, row)) for row in cur.fetchall()]
