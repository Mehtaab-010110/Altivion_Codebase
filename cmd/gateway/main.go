package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"github.com/segmentio/kafka-go"

	"silentraven/internal/models"
	"silentraven/pkg/config"
)

type Gateway struct {
	config    *config.Config
	writer    *kafka.Writer
	router    *mux.Router
	db        *sql.DB
	wsClients map[*websocket.Conn]bool
	wsMutex   sync.Mutex
	upgrader  websocket.Upgrader
}

// DronePoint represents a single drone position
type DronePoint struct {
	SN           string   `json:"sn"`
	TS           string   `json:"ts"`
	Lat          *float64 `json:"lat"`
	Lon          *float64 `json:"lon"`
	HeightM      *float64 `json:"height_m"`
	SpeedHMps    *float64 `json:"speed_h_mps"`
	DirectionDeg *int     `json:"direction_deg"`
}

// TrackPoint represents a point in a track
type TrackPoint struct {
	TS  string  `json:"ts"`
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Track represents a drone's track
type Track struct {
	SN     string       `json:"sn"`
	Points []TrackPoint `json:"points"`
}

func main() {
	log.Println("🚀 Starting SilentRaven Gateway Service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Create gateway instance
	gateway := NewGateway(cfg)
	defer gateway.Close()

	// Setup routes
	gateway.setupRoutes()

	// Setup CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.GetAPIAddress(),
		Handler:      corsHandler.Handler(gateway.router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("✅ Gateway listening on %s", cfg.GetAPIAddress())
		log.Println("📡 Ready to receive drone detections")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down gateway...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Gateway stopped gracefully")
}

// NewGateway creates a new gateway instance
func NewGateway(cfg *config.Config) *Gateway {
	// Create Kafka writer for Redpanda
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.KafkaBrokers),
		Topic:        cfg.KafkaTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    10,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}

	log.Printf("✅ Connected to Redpanda at %s", cfg.KafkaBrokers)

	// Connect to database
	db, err := sql.Open("postgres", cfg.GetDatabaseURL())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("✅ Connected to TimescaleDB")

	return &Gateway{
		config:    cfg,
		writer:    writer,
		router:    mux.NewRouter(),
		db:        db,
		wsClients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}
}

// Close closes all connections
func (g *Gateway) Close() {
	if g.writer != nil {
		g.writer.Close()
	}
	if g.db != nil {
		g.db.Close()
	}
}

// setupRoutes configures HTTP routes
func (g *Gateway) setupRoutes() {
	// Health check
	g.router.HandleFunc("/health", g.handleHealth).Methods("GET")

	// Receive drone detection
	g.router.HandleFunc("/api/v1/detection", g.handleDetection).Methods("POST")

	// Test endpoint
	g.router.HandleFunc("/api/v1/test", g.handleTest).Methods("GET")

	// Frontend API routes
	g.router.HandleFunc("/latest", g.handleLatest).Methods("GET")
	g.router.HandleFunc("/tracks", g.handleTracks).Methods("GET")
	g.router.HandleFunc("/ws", g.handleWebSocket).Methods("GET")
}

// handleHealth returns service health status
func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := models.APIResponse{
		Success: true,
		Message: "Gateway service is healthy",
		Data: map[string]string{
			"service": "gateway",
			"status":  "running",
			"version": "1.0.0",
		},
	}
	sendJSON(w, http.StatusOK, response)
}

// handleTest returns test response
func (g *Gateway) handleTest(w http.ResponseWriter, r *http.Request) {
	response := models.APIResponse{
		Success: true,
		Message: "Test endpoint working",
		Data: map[string]string{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}
	sendJSON(w, http.StatusOK, response)
}

// handleLatest returns latest position for each drone
func (g *Gateway) handleLatest(w http.ResponseWriter, r *http.Request) {
	minutes := 10
	if m := r.URL.Query().Get("minutes"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil {
			minutes = parsed
		}
	}

	query := `
		SELECT DISTINCT ON (sn)
			sn, ts, lat, lon, height_m, speed_h_mps, direction_deg
		FROM public.drone_signals
		WHERE ts > NOW() - $1::interval
			AND lat IS NOT NULL AND lon IS NOT NULL
		ORDER BY sn, ts DESC
	`

	rows, err := g.db.Query(query, strconv.Itoa(minutes)+" minutes")
	if err != nil {
		log.Printf("❌ Query error: %v", err)
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []DronePoint
	for rows.Next() {
		var p DronePoint
		var ts time.Time
		err := rows.Scan(&p.SN, &ts, &p.Lat, &p.Lon, &p.HeightM, &p.SpeedHMps, &p.DirectionDeg)
		if err != nil {
			log.Printf("❌ Scan error: %v", err)
			continue
		}
		p.TS = ts.Format(time.RFC3339)
		results = append(results, p)
	}

	if results == nil {
		results = []DronePoint{}
	}

	sendJSON(w, http.StatusOK, results)
}

// handleTracks returns historical tracks grouped by drone
func (g *Gateway) handleTracks(w http.ResponseWriter, r *http.Request) {
	minutes := 60
	maxPoints := 1000

	if m := r.URL.Query().Get("minutes"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil {
			minutes = parsed
		}
	}
	if mp := r.URL.Query().Get("max_points"); mp != "" {
		if parsed, err := strconv.Atoi(mp); err == nil {
			maxPoints = parsed
		}
	}

	query := `
		SELECT sn, ts, lat, lon
		FROM public.drone_signals
		WHERE ts > NOW() - $1::interval
			AND lat IS NOT NULL AND lon IS NOT NULL
		ORDER BY sn, ts ASC
		LIMIT $2
	`

	rows, err := g.db.Query(query, strconv.Itoa(minutes)+" minutes", maxPoints)
	if err != nil {
		log.Printf("❌ Query error: %v", err)
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Group by SN
	grouped := make(map[string][]TrackPoint)
	for rows.Next() {
		var sn string
		var ts time.Time
		var lat, lon float64

		err := rows.Scan(&sn, &ts, &lat, &lon)
		if err != nil {
			log.Printf("❌ Scan error: %v", err)
			continue
		}

		point := TrackPoint{
			TS:  ts.Format(time.RFC3339),
			Lat: lat,
			Lon: lon,
		}
		grouped[sn] = append(grouped[sn], point)
	}

	// Convert to array format
	var results []Track
	for sn, points := range grouped {
		results = append(results, Track{
			SN:     sn,
			Points: points,
		})
	}

	if results == nil {
		results = []Track{}
	}

	sendJSON(w, http.StatusOK, results)
}

// handleWebSocket upgrades connection and streams real-time updates
func (g *Gateway) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := g.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade failed: %v", err)
		return
	}

	g.wsMutex.Lock()
	g.wsClients[conn] = true
	g.wsMutex.Unlock()

	log.Printf("📡 WebSocket client connected (total: %d)", len(g.wsClients))

	// Keep connection alive
	defer func() {
		g.wsMutex.Lock()
		delete(g.wsClients, conn)
		g.wsMutex.Unlock()
		conn.Close()
		log.Printf("📡 WebSocket client disconnected (total: %d)", len(g.wsClients))
	}()

	// Read messages (to detect disconnect)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// BroadcastToWebSockets sends message to all connected WebSocket clients
func (g *Gateway) BroadcastToWebSockets(message []byte) {
	g.wsMutex.Lock()
	defer g.wsMutex.Unlock()

	for conn := range g.wsClients {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("❌ WebSocket send error: %v", err)
			conn.Close()
			delete(g.wsClients, conn)
		}
	}
}

// handleDetection processes incoming drone detection
func (g *Gateway) handleDetection(w http.ResponseWriter, r *http.Request) {
	// Parse incoming packet
	var packet models.IncomingPacket
	if err := json.NewDecoder(r.Body).Decode(&packet); err != nil {
		log.Printf("❌ Invalid JSON: %v", err)
		response := models.APIResponse{
			Success: false,
			Error:   "Invalid JSON format",
		}
		sendJSON(w, http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if packet.SN == "" || packet.UASID == "" {
		log.Println("❌ Missing required fields")
		response := models.APIResponse{
			Success: false,
			Error:   "Missing required fields: SN or UASID",
		}
		sendJSON(w, http.StatusBadRequest, response)
		return
	}

	// Add timestamp if not present
	if packet.Timestamp == "" {
		packet.Timestamp = time.Now().Format(time.RFC3339)
	}

	// Log received detection
	log.Printf("📡 Received detection: UASID=%s, SN=%s, Type=%s",
		packet.UASID, packet.SN, packet.DroneType)

	// Save to database
	query := `
		INSERT INTO drone_signals (sn, ts, uasid, drone_type, lat, lon, height_m, speed_h_mps, speed_v_mps, direction_deg, operator_lat, operator_lon)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (sn, ts) DO UPDATE SET
			uasid = EXCLUDED.uasid,
			lat = EXCLUDED.lat,
			lon = EXCLUDED.lon,
			height_m = EXCLUDED.height_m
	`

	ts, _ := time.Parse(time.RFC3339, packet.Timestamp)
	if ts.IsZero() {
		ts = time.Now()
	}

	_, err := g.db.Exec(query,
		packet.SN,
		ts,
		packet.UASID,
		packet.DroneType,
		packet.Latitude,
		packet.Longitude,
		packet.Height,
		packet.SpeedHorizontal,
		packet.SpeedVertical,
		packet.Direction,
		packet.OperatorLatitude,
		packet.OperatorLongitude,
	)

	if err != nil {
		log.Printf("❌ Failed to save to database: %v", err)
		response := models.APIResponse{
			Success: false,
			Error:   "Failed to save detection",
		}
		sendJSON(w, http.StatusInternalServerError, response)
		return
	}

	log.Printf("✅ Saved to database: %s", packet.UASID)

	// Try to publish to Redpanda (non-blocking, don't fail if it errors)
	go func() {
		packetJSON, err := json.Marshal(packet)
		if err != nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = g.writer.WriteMessages(ctx,
			kafka.Message{
				Key:   []byte(packet.UASID),
				Value: packetJSON,
				Time:  time.Now(),
			},
		)
		if err != nil {
			log.Printf("⚠️ Kafka publish failed (non-critical): %v", err)
		}
	}()

	// Broadcast to WebSocket clients (only if we have valid coordinates)
	if packet.Latitude != 0 && packet.Longitude != 0 {
		wsMsg := map[string]interface{}{
			"sn":            packet.SN,
			"ts":            packet.Timestamp,
			"lat":           packet.Latitude,
			"lon":           packet.Longitude,
			"height_m":      packet.Height,
			"direction_deg": packet.Direction,
			"speed_h_mps":   packet.SpeedHorizontal,
		}
		wsMsgJSON, _ := json.Marshal(wsMsg)
		g.BroadcastToWebSockets(wsMsgJSON)
	}

	// Send success response
	response := models.APIResponse{
		Success: true,
		Message: "Detection received and queued",
		Data: map[string]string{
			"uas_id":    packet.UASID,
			"sn":        packet.SN,
			"timestamp": packet.Timestamp,
		},
	}
	sendJSON(w, http.StatusOK, response)
}

// sendJSON sends JSON response
func sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
