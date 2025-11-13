from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.config import CORS_ALLOW_ORIGINS
from app.routes import router as api_router
from app.ws import register_ws

app = FastAPI(title="Altivion Ingest API")

# CORS for the Next.js frontend
app.add_middleware(
    CORSMiddleware,
    allow_origins=CORS_ALLOW_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# API routes
app.include_router(api_router)

# WebSocket route + optional PG LISTEN
register_ws(app)
