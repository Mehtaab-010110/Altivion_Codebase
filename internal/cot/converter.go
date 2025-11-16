package cot

import (
	"encoding/xml"
	"fmt"
	"time"

	"silentraven/internal/models"
)

// Event represents a CoT XML event
type Event struct {
	XMLName xml.Name `xml:"event"`
	Version string   `xml:"version,attr"`
	UID     string   `xml:"uid,attr"`
	Type    string   `xml:"type,attr"`
	Time    string   `xml:"time,attr"`
	Start   string   `xml:"start,attr"`
	Stale   string   `xml:"stale,attr"`
	How     string   `xml:"how,attr"`
	Point   Point    `xml:"point"`
	Detail  Detail   `xml:"detail"`
}

type Point struct {
	Lat float64 `xml:"lat,attr"`
	Lon float64 `xml:"lon,attr"`
	Hae float64 `xml:"hae,attr"`
	CE  float64 `xml:"ce,attr"`
	LE  float64 `xml:"le,attr"`
}

type Detail struct {
	Contact Contact `xml:"contact"`
	Remarks string  `xml:"remarks"`
	Link    *Link   `xml:"link,omitempty"`
	Track   *Track  `xml:"track,omitempty"`
}

type Contact struct {
	Callsign string `xml:"callsign,attr"`
}

type Link struct {
	UID      string `xml:"uid,attr"`
	Type     string `xml:"type,attr"`
	Relation string `xml:"relation,attr"`
	Point    Point  `xml:"point"`
}

type Track struct {
	Course float64 `xml:"course,attr"`
	Speed  float64 `xml:"speed,attr"`
}

// ConvertToCoT converts DroneDetection to CoT XML
func ConvertToCoT(detection models.DroneDetection) ([]byte, error) {
	now := time.Now().UTC()
	stale := now.Add(2 * time.Minute)

	// Build UID
	uid := fmt.Sprintf("SilentRaven.UAS.%s", detection.UASID)

	// Determine threat level
	cotType := determineCoTType(detection)

	// Build callsign
	callsign := fmt.Sprintf("UAS-%s", detection.UASID[len(detection.UASID)-4:])

	event := Event{
		Version: "2.0",
		UID:     uid,
		Type:    cotType,
		Time:    now.Format(time.RFC3339),
		Start:   now.Format(time.RFC3339),
		Stale:   stale.Format(time.RFC3339),
		How:     "m-g", // machine-GPS
		Point: Point{
			Lat: detection.Latitude,
			Lon: detection.Longitude,
			Hae: detection.Height,
			CE:  10.0,
			LE:  15.0,
		},
		Detail: Detail{
			Contact: Contact{
				Callsign: callsign,
			},
			Remarks: buildRemarks(detection),
		},
	}

	// Add velocity if moving
	if detection.SpeedHorizontal > 0 {
		event.Detail.Track = &Track{
			Course: float64(detection.Direction), // ✅ FIXED - convert int to float64
			Speed:  detection.SpeedHorizontal,
		}
	}

	// Add operator link if available
	if detection.OperatorLatitude != 0 && detection.OperatorLongitude != 0 {
		event.Detail.Link = &Link{
			UID:      fmt.Sprintf("SilentRaven.Operator.%s", detection.UASID),
			Type:     "a-h-G-U-C", // Hostile ground unit
			Relation: "p-p",
			Point: Point{
				Lat: detection.OperatorLatitude,
				Lon: detection.OperatorLongitude,
				Hae: 0,
				CE:  20.0,
				LE:  20.0,
			},
		}
	}

	// Marshal to XML
	output, err := xml.MarshalIndent(event, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal XML: %w", err)
	}

	// Add XML declaration
	xmlHeader := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	return append(xmlHeader, output...), nil
}

func determineCoTType(detection models.DroneDetection) string {
	// For Sandbox demo, mark all as hostile for visibility
	// TODO: Add logic for authorized drone database lookup
	if detection.OperatorLatitude == 0 && detection.OperatorLongitude == 0 {
		return "a-u-A" // Unknown - no operator info
	}
	return "a-h-A-M-F-Q" // Hostile military fixed-wing quadcopter
}

func buildRemarks(detection models.DroneDetection) string {
	return fmt.Sprintf(`Remote-ID Detection
Node: %s
UASID: %s
Type: %s
Speed: %.1f m/s @ %.0f°
Altitude: %.0fm`,
		detection.SN,
		detection.UASID,
		detection.DroneType,
		detection.SpeedHorizontal,
		float64(detection.Direction),
		detection.Height,
	)
}
