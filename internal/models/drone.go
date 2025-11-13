package models

import (
	"time"
)

// DroneDetection represents a single drone detection event
type DroneDetection struct {
	ID                int64     `json:"id" db:"id"`
	DetectionTime     time.Time `json:"detection_time" db:"detection_time"`
	SN                string    `json:"sn" db:"sn"`
	UASID             string    `json:"uas_id" db:"uas_id"`
	DroneType         string    `json:"drone_type" db:"drone_type"`
	Latitude          float64   `json:"latitude" db:"latitude"`
	Longitude         float64   `json:"longitude" db:"longitude"`
	Height            float64   `json:"height" db:"height"`
	Direction         int       `json:"direction" db:"direction"`
	SpeedHorizontal   float64   `json:"speed_horizontal" db:"speed_horizontal"`
	SpeedVertical     float64   `json:"speed_vertical" db:"speed_vertical"`
	OperatorLatitude  float64   `json:"operator_latitude" db:"operator_latitude"`
	OperatorLongitude float64   `json:"operator_longitude" db:"operator_longitude"`
	NodeID            string    `json:"node_id" db:"node_id"`
	Signature         string    `json:"signature" db:"signature"`
	RawData           string    `json:"raw_data" db:"raw_data"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// IncomingPacket represents the raw data from sensor nodes
type IncomingPacket struct {
	SN                string  `json:"SN"`
	UASID             string  `json:"UASID"`
	DroneType         string  `json:"DroneType"`
	Direction         int     `json:"Direction"`
	SpeedHorizontal   float64 `json:"SpeedHorizontal"`
	SpeedVertical     float64 `json:"SpeedVertical"`
	Latitude          float64 `json:"Latitude"`
	Longitude         float64 `json:"Longitude"`
	Height            float64 `json:"Height"`
	OperatorLatitude  float64 `json:"OperatorLatitude"`
	OperatorLongitude float64 `json:"OperatorLongitude"`
	Signature         string  `json:"signature,omitempty"`
	NodeID            string  `json:"node_id,omitempty"`
	Timestamp         string  `json:"timestamp,omitempty"`
}

// APIResponse is a standard API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}
