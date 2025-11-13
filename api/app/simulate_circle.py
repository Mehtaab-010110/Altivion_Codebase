import requests
import math
import time
from datetime import datetime, timezone

API_KEY = "LsSBMGpGcpSoeNQtOfWyfHuvLX9JKLqHDuLEb9xTih0"  # <-- your API key
API_URL = "http://localhost:8000/ingest"               # or your container IP if remote

# Center point (e.g. Lethbridge)
CENTER_LAT = 49.701
CENTER_LON = -112.818

RADIUS_M = 100  # radius of circle (meters)
NUM_POINTS = 50
DRONE_SN = "TEST-CIRCLE-001"

def offset_latlon(lat, lon, dx, dy):
    """Offset lat/lon by meters east (dx) and north (dy)."""
    dlat = dy / 111320  # ~ meters per degree latitude
    dlon = dx / (40075000 * math.cos(math.radians(lat)) / 360)
    return lat + dlat, lon + dlon

def simulate_circle():
    print(f"Simulating drone {DRONE_SN} around ({CENTER_LAT}, {CENTER_LON})...")
    for i in range(NUM_POINTS):
        angle = (i / NUM_POINTS) * 2 * math.pi
        x = RADIUS_M * math.cos(angle)   # east-west
        y = RADIUS_M * math.sin(angle)   # north-south
        lat, lon = offset_latlon(CENTER_LAT, CENTER_LON, x, y)
        ts = datetime.now(timezone.utc).isoformat()

        data = {
            "SN": DRONE_SN,
            "UASID": "SIM01",
            "DroneType": "TestDrone",
            "Direction": int(math.degrees(angle) % 360),
            "SpeedHorizontal": 10.0,
            "Latitude": lat,
            "Longitude": lon,
            "Height": 120.0,
        }

        try:
            resp = requests.post(API_URL, headers={
                "Content-Type": "application/json",
                "X-API-Key": API_KEY
            }, json=data, timeout=5)
            print(f"{i+1:02d}/{NUM_POINTS}: HTTP {resp.status_code}")
        except Exception as e:
            print("Error:", e)

        time.sleep(1)  # seconds between points

if __name__ == "__main__":
    simulate_circle()
