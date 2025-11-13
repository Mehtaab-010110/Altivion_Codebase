import os

API_KEY = os.environ.get("API_KEY")
DATABASE_URL = os.environ.get("DATABASE_URL")

# Next.js on port 3000
CORS_ALLOW_ORIGINS = [
    "http://localhost:3000",
    "http://127.0.0.1:3000",
    "http://26.144.3.234:3000",
]

# Node heartbeat window (sec) and total nodes for the dashboard card
NODE_ONLINE_WINDOW_SEC = int(os.environ.get("NODE_ONLINE_WINDOW_SEC", "60"))
NODES_TOTAL = int(os.environ.get("NODES_TOTAL", "3"))

# Toggle the optional PG LISTEN loop (realtime from DB triggers)
ENABLE_LISTEN = os.environ.get("ENABLE_LISTEN", "true").lower() == "true"
