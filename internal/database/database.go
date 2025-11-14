package database

import (
	"database/sql"
	"fmt"
	"log"
	"silentraven/internal/models"
	"silentraven/pkg/config"
	"time"

	_ "github.com/lib/pq"
)

// DB wraps database connection and operations
type DB struct {
	conn *sql.DB
}

// New creates a new database connection
func New(cfg *config.Config) (*DB, error) {
	connStr := cfg.GetDBConnectionString()

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	log.Println("✅ Database connected successfully")

	return &DB{conn: conn}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// InsertDroneDetection inserts a new drone detection
func (db *DB) InsertDroneDetection(detection *models.DroneDetection) error {
	query := `
		INSERT INTO drone_detections (
			detection_time, sn, uas_id, drone_type, latitude, longitude, height,
			direction, speed_horizontal, speed_vertical, operator_latitude, 
			operator_longitude, node_id, signature, raw_data
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING id, created_at
	`

	err := db.conn.QueryRow(
		query,
		detection.DetectionTime,
		detection.SN,
		detection.UASID,
		detection.DroneType,
		detection.Latitude,
		detection.Longitude,
		detection.Height,
		detection.Direction,
		detection.SpeedHorizontal,
		detection.SpeedVertical,
		detection.OperatorLatitude,
		detection.OperatorLongitude,
		detection.NodeID,
		detection.Signature,
		detection.RawData,
	).Scan(&detection.ID, &detection.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert drone detection: %w", err)
	}

	return nil
}

// GetRecentDetections retrieves recent drone detections
func (db *DB) GetRecentDetections(limit int) ([]models.DroneDetection, error) {
	query := `
		SELECT 
			id, detection_time, sn, uas_id, drone_type, latitude, longitude, height,
			direction, speed_horizontal, speed_vertical, operator_latitude, 
			operator_longitude, node_id, signature, raw_data, created_at
		FROM drone_detections
		ORDER BY detection_time DESC
		LIMIT $1
	`

	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query detections: %w", err)
	}
	defer rows.Close()

	var detections []models.DroneDetection
	for rows.Next() {
		var d models.DroneDetection
		err := rows.Scan(
			&d.ID, &d.DetectionTime, &d.SN, &d.UASID, &d.DroneType,
			&d.Latitude, &d.Longitude, &d.Height, &d.Direction,
			&d.SpeedHorizontal, &d.SpeedVertical, &d.OperatorLatitude,
			&d.OperatorLongitude, &d.NodeID, &d.Signature, &d.RawData, &d.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan detection: %w", err)
		}
		detections = append(detections, d)
	}

	return detections, nil
}

// Health checks database connectivity
func (db *DB) Health() error {
	return db.conn.Ping()
}
