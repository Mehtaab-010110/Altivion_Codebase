# Altivion Setup Guide

This guide will help you set up and run the Altivion drone detection system.

## Prerequisites

- PostgreSQL 12+ installed and running
- Python 3.10+ installed
- Node.js 18+ and npm installed
- Google Maps API key

---

## Setup Steps

### 1. Start PostgreSQL

```bash
# Start PostgreSQL service (method varies by system)
sudo service postgresql start
# OR
sudo systemctl start postgresql
```

### 2. Create Database and Schema

```bash
# Switch to postgres user and create database
sudo -u postgres psql -c "CREATE DATABASE altivion;"

# Run the setup script
sudo -u postgres psql -d altivion -f scripts/setup_database.sql
```

### 3. Configure Backend Environment

```bash
# Copy example env file
cp api/.env.example api/.env

# Edit api/.env and set:
# - DATABASE_URL (update password if different)
# - API_KEY (generate a secure random key)
```

### 4. Configure Frontend Environment

```bash
# Copy example env file
cp altivion-web/.env.local.example altivion-web/.env.local

# Edit altivion-web/.env.local and set:
# - NEXT_PUBLIC_GOOGLE_MAPS_API_KEY (get from Google Cloud Console)
```

### 5. Install Backend Dependencies

```bash
cd api
pip install -r requirements.txt
# OR use virtual environment (recommended)
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
cd ..
```

### 6. Install Frontend Dependencies

```bash
cd altivion-web
npm install
cd ..
```

---

## Running the Application

### Start Backend API (Terminal 1)

```bash
cd api
# If using virtual environment
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Start FastAPI server
uvicorn main:app --host 127.0.0.1 --port 8000 --reload
```

You should see:
```
INFO:     Uvicorn running on http://127.0.0.1:8000 (Press CTRL+C to quit)
```

### Start Frontend (Terminal 2)

```bash
cd altivion-web
npm run dev
```

You should see:
```
> altivion-web@0.1.0 dev
> next dev

  ▲ Next.js 16.0.1
  - Local:        http://localhost:3000
```

---

## Testing Connectivity

### 1. Test Backend API

Open browser to: http://localhost:8000/health

Expected response:
```json
{"status": "ok"}
```

### 2. Test Data Endpoint

Open browser to: http://localhost:8000/data

Expected response (with test data):
```json
{
  "uptime_seconds": 10,
  "drones_today": 2,
  "drones_online": 0,
  "nodes_online": 0,
  "nodes_total": 3
}
```

### 3. Test Frontend

Open browser to: http://localhost:3000

You should see:
- Altivion title
- Google Maps loaded
- Stats panel showing system data
- Test drone markers on the map (if test data was inserted)

### 4. Test WebSocket Connection

With the frontend open:
- Open browser DevTools (F12)
- Go to Console tab
- You should see WebSocket connection logs
- Stats panel should update in real-time

---

## Sending Test Data

### Method 1: Via API Directly

```bash
curl -X POST http://localhost:8000/ingest \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-api-key-here" \
  -d '[{
    "SN": "DRONE123",
    "UASID": "UAS123",
    "LAT": 37.7749,
    "LON": -122.4194,
    "HEIGHT_M": 100.5,
    "SPEED_H_MPS": 15.0,
    "DIRECTION_DEG": 45.0
  }]'
```

### Method 2: Insert Directly to Database

```bash
sudo -u postgres psql -d altivion -c "
INSERT INTO drone_signals (sn, uasid, lat, lon, height_m, speed_h_mps, direction_deg)
VALUES ('DRONE999', 'UAS999', 37.7752, -122.4197, 200.0, 25.0, 180.0);
"
```

The new drone should appear on the map within 5 seconds (SWR polling interval).

---

## Troubleshooting

### Backend won't start

**Error: "could not connect to server"**
- Check PostgreSQL is running: `pg_isready`
- Verify DATABASE_URL in api/.env
- Check database exists: `sudo -u postgres psql -l | grep altivion`

**Error: "ModuleNotFoundError"**
- Install dependencies: `pip install -r requirements.txt`
- Activate virtual environment if using one

### Frontend won't start

**Error: "Cannot find module"**
- Install dependencies: `npm install`

**Error: "Google Maps JavaScript API error"**
- Check NEXT_PUBLIC_GOOGLE_MAPS_API_KEY in .env.local
- Verify API key is enabled for Maps JavaScript API in Google Cloud Console

### No drones showing on map

- Check browser Console for errors
- Verify backend is running: http://localhost:8000/health
- Check data exists: http://localhost:8000/latest
- Verify CORS is configured (should allow localhost:3000)
- Check WebSocket connection in DevTools Network tab

### WebSocket not connecting

- Verify backend allows WebSocket connections
- Check browser Console for connection errors
- Ensure NEXT_PUBLIC_API_BASE is set correctly
- Check firewall/proxy settings

---

## Architecture

```
Sensor Node → Gateway (Go) → Kafka/Redpanda → FastAPI Backend → PostgreSQL
                                                      ↓
                                                  WebSocket
                                                      ↓
                                              Frontend (Next.js)
                                                      ↓
                                                Google Maps UI
```

---

## API Endpoints

- `GET /health` - Health check
- `GET /data` - System statistics
- `GET /latest?minutes=120` - Latest drone positions
- `GET /tracks?minutes=120&max_points=4000` - Drone flight tracks
- `POST /ingest` - Ingest drone signals (requires API key)
- `WS /ws` - WebSocket for real-time updates

---

## Configuration Options

### Backend (api/.env)

- `DATABASE_URL` - PostgreSQL connection string
- `API_KEY` - Authentication key for /ingest endpoint
- `NODE_ONLINE_WINDOW_SEC` - Seconds for node to be considered online (default: 60)
- `NODES_TOTAL` - Total number of sensor nodes (default: 3)
- `ENABLE_LISTEN` - Enable PostgreSQL LISTEN for real-time updates (default: true)

### Frontend (altivion-web/.env.local)

- `NEXT_PUBLIC_API_BASE` - Backend API URL (default: http://127.0.0.1:8000)
- `NEXT_PUBLIC_GOOGLE_MAPS_API_KEY` - Google Maps API key (required)
- `NEXT_PUBLIC_DEBUG` - Enable debug HUD (default: false)

---

## Next Steps

1. Set up the Gateway service (Go) to receive sensor data
2. Configure Kafka/Redpanda for message queuing
3. Deploy to production environment
4. Set up monitoring and logging
5. Configure SSL/TLS for production

For questions or issues, refer to the documentation or contact support.
