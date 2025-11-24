package cot

import (
	"encoding/xml"
	"fmt"
	"time"

	"silentraven/internal/models"
)

// CoT XML structures
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
	Ce  float64 `xml:"ce,attr"`
	Le  float64 `xml:"le,attr"`
}

type Detail struct {
	Contact Contact `xml:"contact"`
	Remarks string  `xml:"remarks"`
	Track   Track   `xml:"track"`
}

type Contact struct {
	Callsign string `xml:"callsign,attr"`
}

type Track struct {
	Course float64 `xml:"course,attr"`
	Speed  float64 `xml:"speed,attr"`
}

// ConvertToCoT converts IncomingPacket to CoT XML
func ConvertToCoT(detection models.IncomingPacket) ([]byte, error) {
	now := time.Now().UTC()
	stale := now.Add(2 * time.Minute)

	// Safe UASID extraction
	uidSuffix := detection.UASID
	if len(detection.UASID) > 4 {
		uidSuffix = detection.UASID[len(detection.UASID)-4:]
	}
	uid := fmt.Sprintf("SilentRaven.UAS.%s", uidSuffix)

	// CoT type: a-h-A-M-F-Q
	// a = atom (entity)
	// h = hostile
	// A = Air
	// M = Military
	// F = Fixed wing (or rotary if needed)
	// Q = Unmanned Aerial System
	cotType := "a-h-A-M-F-Q"

	// Build remarks with detection details
	remarks := fmt.Sprintf(`Remote-ID Detection
Node: %s
Type: %s
Speed: %.1f m/s @ %d°`,
		detection.SN,
		detection.DroneType,
		detection.SpeedHorizontal,
		detection.Direction)

	event := Event{
		Version: "2.0",
		UID:     uid,
		Type:    cotType,
		Time:    now.Format(time.RFC3339),
		Start:   now.Format(time.RFC3339),
		Stale:   stale.Format(time.RFC3339),
		How:     "m-g", // m-g = machine generated
		Point: Point{
			Lat: detection.Latitude,
			Lon: detection.Longitude,
			Hae: detection.Height,
			Ce:  10.0, // Circular error (meters)
			Le:  15.0, // Linear error (meters)
		},
		Detail: Detail{
			Contact: Contact{
				Callsign: fmt.Sprintf("UAS-%s", uidSuffix),
			},
			Remarks: remarks,
			Track: Track{
				Course: float64(detection.Direction),
				Speed:  detection.SpeedHorizontal,
			},
		},
	}

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(event, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal CoT XML: %w", err)
	}

	// Add XML declaration
	fullXML := []byte(xml.Header + string(xmlData))

	return fullXML, nil
}
