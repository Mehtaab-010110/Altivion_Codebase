import asyncio
from typing import Set
from fastapi import WebSocket, WebSocketDisconnect, FastAPI
from psycopg import AsyncConnection
from app.config import DATABASE_URL, ENABLE_LISTEN

class WSManager:
    def __init__(self):
        self.active: Set[WebSocket] = set()
        self._lock = asyncio.Lock()

    async def connect(self, ws: WebSocket):
        await ws.accept()
        async with self._lock:
            self.active.add(ws)

    async def disconnect(self, ws: WebSocket):
        async with self._lock:
            self.active.discard(ws)

    async def broadcast(self, message: str):
        dead = []
        async with self._lock:
            for ws in list(self.active):
                try:
                    await ws.send_text(message)
                except Exception:
                    dead.append(ws)
            for ws in dead:
                self.active.discard(ws)

manager = WSManager()
_listen_task = None

async def listen_pg_notifications():
    """Optional async LISTEN loop using psycopg3 iterator API."""
    while True:
        try:
            conn: AsyncConnection = await AsyncConnection.connect(DATABASE_URL)
            await conn.set_autocommit(True)
            async with conn.cursor() as cur:
                await cur.execute("LISTEN signals;")
            print("LISTEN(iterator): started")
            async for notify in conn.notifies():
                await manager.broadcast(notify.payload)
        except Exception as e:
            print("LISTEN(iterator) error:", repr(e))
            await asyncio.sleep(2)

def register_ws(app: FastAPI):
    """Attach WS route + startup listener (if enabled)."""
    @app.websocket("/ws")
    async def websocket_endpoint(ws: WebSocket):
        await manager.connect(ws)
        try:
            while True:
                await ws.receive_text()  # keepalive only
        except WebSocketDisconnect:
            await manager.disconnect(ws)

    @app.on_event("startup")
    async def _startup():
        global _listen_task
        if ENABLE_LISTEN and (_listen_task is None or _listen_task.done()):
            _listen_task = asyncio.create_task(listen_pg_notifications())
