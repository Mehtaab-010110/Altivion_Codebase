# Altivion Quick Start Guide

Get your drone detection system running in 5 minutes!

## Prerequisites Check

```bash
# Check PostgreSQL
pg_isready

# Check Python
python3 --version

# Check Node.js
node --version
npm --version
```

---

## Quick Setup (First Time Only)

### 1. Database Setup (30 seconds)

```bash
# Start PostgreSQL (if not running)
sudo service postgresql start

# Create database and schema
sudo -u postgres psql -c "CREATE DATABASE altivion;"
sudo -u postgres psql -d altivion -f scripts/setup_database.sql
```

### 2. Backend Configuration (1 minute)

```bash
# Copy environment file
cp api/.env.example api/.env

# Edit the file (use nano, vim, or any editor)
nano api/.env

# Required changes:
# - Update DATABASE_URL password if needed
# - Set API_KEY to any secure random string (e.g., "my-secret-key-12345")
```

**Example api/.env:**
```env
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/altivion
API_KEY=my-super-secret-api-key-2024
NODE_ONLINE_WINDOW_SEC=60
NODES_TOTAL=3
ENABLE_LISTEN=true
```

### 3. Frontend Configuration (1 minute)

```bash
# Copy environment file
cp altivion-web/.env.local.example altivion-web/.env.local

# Edit the file
nano altivion-web/.env.local

# Required: Set your Google Maps API key
# Get one free at: https://console.cloud.google.com/google/maps-apis
```

**Example altivion-web/.env.local:**
```env
NEXT_PUBLIC_API_BASE=http://127.0.0.1:8000
NEXT_PUBLIC_GOOGLE_MAPS_API_KEY=AIzaSyXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
NEXT_PUBLIC_DEBUG=false
```

> **Note:** Without a Google Maps API key, the map won't load. Get a free key at:
> https://console.cloud.google.com/google/maps-apis/credentials

### 4. Install Dependencies (2 minutes)

```bash
# Backend dependencies
cd api
pip install -r requirements.txt
cd ..

# Frontend dependencies
cd altivion-web
npm install
cd ..
```

---

## Running the System

### Option 1: Using Helper Scripts (Easiest)

**Terminal 1 - Start Backend:**
```bash
./start-backend.sh
```

**Terminal 2 - Start Frontend:**
```bash
./start-frontend.sh
```

### Option 2: Manual Start

**Terminal 1 - Backend:**
```bash
cd api
uvicorn main:app --host 127.0.0.1 --port 8000 --reload
```

**Terminal 2 - Frontend:**
```bash
cd altivion-web
npm run dev
```

---

## Verify It's Working

### 1. Backend Health Check
Open browser: http://localhost:8000/health

Should show:
```json
{"status": "ok"}
```

### 2. View Frontend
Open browser: http://localhost:3000

You should see:
- ✅ Altivion title
- ✅ Google Maps loaded
- ✅ Stats panel (top-left)
- ✅ Test drone markers (if test data exists)

### 3. Check WebSocket
Open browser DevTools (F12) → Console

Should see:
```
WebSocket connection established
```

---

## Send Test Data

### Quick Test via API:

```bash
curl -X POST http://localhost:8000/ingest \
  -H "Content-Type: application/json" \
  -H "X-API-Key: my-super-secret-api-key-2024" \
  -d '[{
    "SN": "TEST001",
    "UASID": "UAS001",
    "LAT": 37.7749,
    "LON": -122.4194,
    "HEIGHT_M": 150.0,
    "SPEED_H_MPS": 20.0,
    "DIRECTION_DEG": 90.0
  }]'
```

The drone should appear on the map within 5 seconds!

---

## Architecture Diagram

```
┌─────────────┐
│ Sensor Node │
└──────┬──────┘
       │ HTTP POST
       ↓
┌─────────────┐       ┌──────────────┐
│  Gateway    │──────→│ Kafka/Redpanda│
│   (Go)      │       └──────┬───────┘
└─────────────┘              │
                             ↓
                      ┌──────────────┐       ┌────────────┐
                      │ FastAPI      │──────→│ PostgreSQL │
                      │ Backend      │       │  Database  │
                      │ (port 8000)  │←──────┤            │
                      └──────┬───────┘       └────────────┘
                             │
                             │ WebSocket + REST
                             ↓
                      ┌──────────────┐
                      │  Next.js     │
                      │  Frontend    │
                      │ (port 3000)  │
                      └──────┬───────┘
                             │
                             ↓
                      ┌──────────────┐
                      │ Google Maps  │
                      │     UI       │
                      └──────────────┘
```

---

## Troubleshooting

### "Cannot connect to database"
```bash
# Check PostgreSQL is running
pg_isready

# Start if not running
sudo service postgresql start

# Verify database exists
sudo -u postgres psql -l | grep altivion
```

### "Google Maps not loading"
- Check NEXT_PUBLIC_GOOGLE_MAPS_API_KEY in altivion-web/.env.local
- Verify API key is enabled at: https://console.cloud.google.com/google/maps-apis
- Enable "Maps JavaScript API" for your project

### "No drones showing"
```bash
# Check if data exists
curl http://localhost:8000/latest

# Insert test data
sudo -u postgres psql -d altivion -c "
INSERT INTO drone_signals (sn, lat, lon, height_m)
VALUES ('TEST999', 37.7750, -122.4195, 100.0);
"
```

### "Port already in use"
```bash
# Find what's using the port
sudo lsof -i :8000  # For backend
sudo lsof -i :3000  # For frontend

# Kill the process or use different ports
```

---

## What's Next?

1. ✅ System is running
2. 📡 Set up Gateway service to receive real sensor data
3. 🔐 Configure production authentication
4. 🚀 Deploy to production
5. 📊 Add monitoring and alerting

---

## Useful Commands

```bash
# View API documentation
http://localhost:8000/docs

# Check backend logs
# (watch Terminal 1)

# Check frontend logs
# (watch Terminal 2)

# View database data
sudo -u postgres psql -d altivion -c "SELECT * FROM drone_signals ORDER BY ts DESC LIMIT 10;"

# Clear all data
sudo -u postgres psql -d altivion -c "TRUNCATE drone_signals;"
```

---

## Default Ports

- Backend API: `8000`
- Frontend: `3000`
- PostgreSQL: `5432`
- WebSocket: `8000/ws` (same as API)

---

## Need Help?

See the full documentation: [SETUP.md](SETUP.md)

Or check the codebase structure at the root README.
