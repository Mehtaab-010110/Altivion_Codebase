package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Build connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_SSLMODE"),
	)

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create database
	dbName := os.Getenv("DB_NAME")
	fmt.Printf("Creating database: %s\n", dbName)

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		log.Printf("Warning: Could not drop existing database: %v\n", err)
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		log.Fatal("Failed to create database:", err)
	}

	fmt.Println("✓ Database created successfully")

	// Connect to the new database
	connStr = fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		dbName,
		os.Getenv("DB_SSLMODE"),
	)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to new database:", err)
	}
	defer db.Close()

	// Enable TimescaleDB extension
	fmt.Println("Enabling TimescaleDB extension...")
	_, err = db.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb")
	if err != nil {
		log.Fatal("Failed to enable TimescaleDB:", err)
	}
	fmt.Println("✓ TimescaleDB extension enabled")

	// Create tables
	schema := `
	-- Drone detections table (will be converted to hypertable)
	CREATE TABLE drone_detections (
		id BIGSERIAL,
		detection_time TIMESTAMPTZ NOT NULL,
		sn VARCHAR(50) NOT NULL,
		uas_id VARCHAR(50) NOT NULL,
		drone_type VARCHAR(100),
		latitude DOUBLE PRECISION,
		longitude DOUBLE PRECISION,
		height DOUBLE PRECISION,
		direction INTEGER,
		speed_horizontal DOUBLE PRECISION,
		speed_vertical DOUBLE PRECISION,
		operator_latitude DOUBLE PRECISION,
		operator_longitude DOUBLE PRECISION,
		node_id VARCHAR(50) NOT NULL,
		signature TEXT,
		raw_data JSONB,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		PRIMARY KEY (detection_time, id)
	);

	-- Convert to hypertable (TimescaleDB magic)
	SELECT create_hypertable('drone_detections', 'detection_time', 
		chunk_time_interval => INTERVAL '1 day');

	-- Sensor nodes table
	CREATE TABLE sensor_nodes (
		node_id VARCHAR(50) PRIMARY KEY,
		node_name VARCHAR(100),
		latitude DOUBLE PRECISION,
		longitude DOUBLE PRECISION,
		status VARCHAR(20) DEFAULT 'active',
		last_heartbeat TIMESTAMPTZ,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	-- Alerts table
	CREATE TABLE alerts (
		id BIGSERIAL PRIMARY KEY,
		detection_id BIGINT,
		alert_type VARCHAR(50) NOT NULL,
		severity VARCHAR(20) NOT NULL,
		message TEXT,
		resolved BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		resolved_at TIMESTAMPTZ
	);

	-- Audit log for security compliance
	CREATE TABLE audit_log (
		id BIGSERIAL PRIMARY KEY,
		event_type VARCHAR(50) NOT NULL,
		user_id VARCHAR(50),
		action VARCHAR(100),
		resource VARCHAR(100),
		details JSONB,
		ip_address INET,
		timestamp TIMESTAMPTZ DEFAULT NOW()
	);

	-- Create indexes for performance
	CREATE INDEX idx_drone_uas_id ON drone_detections(uas_id, detection_time DESC);
	CREATE INDEX idx_drone_sn ON drone_detections(sn, detection_time DESC);
	CREATE INDEX idx_drone_node ON drone_detections(node_id, detection_time DESC);
	CREATE INDEX idx_drone_location ON drone_detections(latitude, longitude);
	CREATE INDEX idx_alerts_unresolved ON alerts(created_at) WHERE resolved = FALSE;
	CREATE INDEX idx_audit_timestamp ON audit_log(timestamp DESC);

	-- Create continuous aggregate for 5-minute summary statistics
	CREATE MATERIALIZED VIEW drone_stats_5min
	WITH (timescaledb.continuous) AS
	SELECT 
		time_bucket('5 minutes', detection_time) AS bucket,
		node_id,
		COUNT(*) as detection_count,
		COUNT(DISTINCT uas_id) as unique_drones,
		AVG(height) as avg_height,
		MAX(height) as max_height,
		AVG(speed_horizontal) as avg_speed
	FROM drone_detections
	GROUP BY bucket, node_id
	WITH NO DATA;

	-- Refresh policy for continuous aggregate
	SELECT add_continuous_aggregate_policy('drone_stats_5min',
		start_offset => INTERVAL '1 hour',
		end_offset => INTERVAL '5 minutes',
		schedule_interval => INTERVAL '5 minutes');

	-- Data retention policy (keep raw data for 90 days)
	SELECT add_retention_policy('drone_detections', INTERVAL '90 days');
	`

	fmt.Println("Creating database schema...")
	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal("Failed to create schema:", err)
	}

	fmt.Println("✓ Database schema created successfully")
	fmt.Println("\n🎉 Database setup complete!")
	fmt.Println("\nYou can now start the services:")
	fmt.Println("  go run cmd/gateway/main.go")
	fmt.Println("  go run cmd/ingestion/main.go")
	fmt.Println("  go run cmd/api/main.go")
}
