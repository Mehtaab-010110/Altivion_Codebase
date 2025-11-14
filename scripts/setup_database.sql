-- Altivion Database Setup Script
-- Creates the drone_signals table for storing drone detection data

-- Create database (run separately if needed)
-- CREATE DATABASE altivion;

-- Connect to the database
-- \c altivion;

-- Optional: Enable TimescaleDB extension for time-series optimization
-- CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Create the drone_signals table
CREATE TABLE IF NOT EXISTS drone_signals (
    id SERIAL PRIMARY KEY,
    ts TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sn VARCHAR(100),
    uasid VARCHAR(100),
    drone_type VARCHAR(50),
    direction_deg DOUBLE PRECISION,
    speed_h_mps DOUBLE PRECISION,
    speed_v_mps DOUBLE PRECISION,
    lat DOUBLE PRECISION NOT NULL,
    lon DOUBLE PRECISION NOT NULL,
    height_m DOUBLE PRECISION,
    operator_lat DOUBLE PRECISION,
    operator_lon DOUBLE PRECISION
);

-- Create index on timestamp for faster queries
CREATE INDEX IF NOT EXISTS idx_drone_signals_ts ON drone_signals(ts DESC);

-- Create index on serial number for track queries
CREATE INDEX IF NOT EXISTS idx_drone_signals_sn ON drone_signals(sn);

-- Optional: Convert to TimescaleDB hypertable for better time-series performance
-- SELECT create_hypertable('drone_signals', 'ts', if_not_exists => TRUE);

-- Create a table for node heartbeats
CREATE TABLE IF NOT EXISTS node_heartbeats (
    node_id VARCHAR(100) PRIMARY KEY,
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert some test data (optional)
INSERT INTO drone_signals (sn, uasid, lat, lon, height_m, speed_h_mps, direction_deg)
VALUES
    ('DRONE001', 'UAS001', 37.7749, -122.4194, 100.5, 15.2, 45.0),
    ('DRONE002', 'UAS002', 37.7750, -122.4195, 150.0, 20.5, 90.0),
    ('DRONE001', 'UAS001', 37.7751, -122.4196, 105.0, 16.0, 50.0)
ON CONFLICT DO NOTHING;

COMMIT;
